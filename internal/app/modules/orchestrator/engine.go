package orchestrator

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/irving-frias/drupal-watcher/internal/app/eventbus"
	"github.com/irving-frias/drupal-watcher/internal/metrics"
	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type Engine struct {
	Watcher            core.Watcher
	Executor           core.CommandExecutor
	SiteExecutorFactory func(site core.SiteInfo) core.CommandExecutor
	Filters            []core.EventFilter
	LintCheckers        map[string]core.LintChecker
	SkipLint            bool
	EventBus            *eventbus.EventBus
	Logger             *slog.Logger
	Debounce           int
	LazyRebuildMs      int
	Patterns           []string
	ExcludePatterns    []string
	CommandsPerPattern map[string]string
	ResolvedSites      []core.SiteInfo
	DrupalRoot         string
	Routes             []string

	pending atomic.Bool

	mu          sync.Mutex
	changedFiles map[string]struct{}
	lastFile    string
	timer       *time.Timer

	lazyMu       sync.Mutex
	lazyPending  bool
	lazyTimer    *time.Timer
	lazyFiles    map[string]struct{}

	lastPubFile string
	lastPubTime time.Time
	dedupMu     sync.Mutex

	startTime time.Time
	cancel    context.CancelFunc

	stats struct {
		changes atomic.Int64
		clears  atomic.Int64
	}
}

func NewEngine(cfg EngineConfig) *Engine {
	lr := cfg.LazyRebuildMs
	if lr <= 0 {
		lr = 2000
	}
	return &Engine{
		Watcher:            cfg.Watcher,
		Executor:           cfg.Executor,
		SiteExecutorFactory: cfg.SiteExecutorFactory,
		Filters:            cfg.Filters,
		LintCheckers:       cfg.LintCheckers,
		SkipLint:           cfg.SkipLint,
		EventBus:           cfg.EventBus,
		Logger:             cfg.Logger,
		Debounce:           cfg.Debounce,
		LazyRebuildMs:      lr,
		Patterns:           cfg.Patterns,
		ExcludePatterns:    cfg.ExcludePatterns,
		CommandsPerPattern: cfg.CommandsPerPattern,
		ResolvedSites:      cfg.ResolvedSites,
		DrupalRoot:         cfg.DrupalRoot,
		Routes:             cfg.Routes,
		changedFiles:       make(map[string]struct{}),
		lazyFiles:          make(map[string]struct{}),
		startTime:          time.Now(),
	}
}

func (e *Engine) Run(ctx context.Context) error {
	ctx, e.cancel = context.WithCancel(ctx)
	events, errs := e.Watcher.Start(ctx)

	e.Logger.Info("orchestrator started",
		"debounce_ms", e.Debounce,
		"filters", len(e.Filters),
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case err, ok := <-errs:
			if !ok {
				return nil
			}
			e.Logger.Error("watcher error", "error", err)
			metrics.RecordError()
			if e.EventBus != nil {
				e.EventBus.Publish(eventbus.TopicError, core.EngineEvent{
					Type:      core.EventError,
					Error:     err,
					Timestamp: time.Now(),
				})
			}

		case event, ok := <-events:
			if !ok {
				return nil
			}

			if !e.shouldProcess(event) {
				continue
			}

			e.mu.Lock()
			e.lastFile = event.Path
			e.changedFiles[event.Path] = struct{}{}
			dispFile := event.Path
			e.mu.Unlock()
			e.pending.Store(true)

			// Publish file.change immediately (dedup within 500ms for same file)
			e.dedupMu.Lock()
			dup := dispFile == e.lastPubFile && time.Since(e.lastPubTime) < 500*time.Millisecond
			if !dup {
				e.lastPubFile = dispFile
				e.lastPubTime = time.Now()
			}
			e.dedupMu.Unlock()

			if !dup && e.EventBus != nil {
				e.stats.changes.Add(1)
				metrics.RecordChange()
				e.EventBus.Publish(eventbus.TopicFileChange, core.EngineEvent{
					Type:      core.EventChange,
					File:      dispFile,
					Changes:   1,
					Timestamp: time.Now(),
				})
			}

			e.mu.Lock()
			if e.timer != nil {
				e.timer.Stop()
			}
			e.timer = time.AfterFunc(time.Duration(e.Debounce)*time.Millisecond, func() {
				if !e.pending.Load() {
					return
				}
				e.pending.Store(false)
				e.processBatch()
			})
			e.mu.Unlock()
		}
	}
}

func (e *Engine) shouldProcess(event core.FileEvent) bool {
	for _, f := range e.Filters {
		if !f.ShouldProcess(event) {
			return false
		}
	}
	return true
}

func (e *Engine) hasCRCommand(cmds []string) bool {
	for _, c := range cmds {
		if c == "cr" || c == "cache:rebuild" {
			return true
		}
	}
	return false
}

func (e *Engine) scheduleLazyRebuild(files map[string]struct{}, dispFile string) {
	e.lazyMu.Lock()
	defer e.lazyMu.Unlock()

	for f := range files {
		e.lazyFiles[f] = struct{}{}
	}

	if e.lazyTimer != nil {
		e.lazyTimer.Stop()
	}
	e.lazyPending = true
	e.lazyTimer = time.AfterFunc(time.Duration(e.LazyRebuildMs)*time.Millisecond, func() {
		e.lazyMu.Lock()
		if !e.lazyPending {
			e.lazyMu.Unlock()
			return
		}
		e.lazyPending = false
		files := e.lazyFiles
		e.lazyFiles = make(map[string]struct{})
		lastFile := dispFile
		e.lazyMu.Unlock()

		e.executeCacheClear(files, lastFile)
	})
}

func (e *Engine) flushLazyRebuild() {
	e.lazyMu.Lock()
	if e.lazyTimer != nil {
		e.lazyTimer.Stop()
		e.lazyTimer = nil
	}
	if !e.lazyPending {
		e.lazyMu.Unlock()
		return
	}
	e.lazyPending = false
	files := e.lazyFiles
	e.lazyFiles = make(map[string]struct{})
	e.lazyMu.Unlock()

	if len(files) > 0 {
		e.executeCacheClear(files, "")
	}
}

func (e *Engine) processBatch() {
	e.mu.Lock()
	files := e.changedFiles
	e.changedFiles = make(map[string]struct{})
	dispFile := e.lastFile
	e.mu.Unlock()

	if len(files) == 0 {
		return
	}

	changes := int64(len(files))

	if !e.SkipLint && len(e.LintCheckers) > 0 {
		if fail := e.lintFiles(files); fail != nil {
			e.Logger.Warn("lint failed", "file", fail.File, "error", fail.Error)
			if e.EventBus != nil {
				e.EventBus.Publish(eventbus.TopicError, core.EngineEvent{
					Type:      core.EventError,
					File:      fail.File,
					Error:     fmt.Errorf("Lint: %s", fail.Error),
					Timestamp: time.Now(),
				})
			}
			return
		}
	}

	e.mu.Lock()
	cmds := e.resolveCommands(files)
	cmdStr := strings.Join(cmds, " + ")
	hasCR := e.hasCRCommand(cmds)
	e.mu.Unlock()

	e.Logger.Info("change detected",
		"file", dispFile,
		"changes", changes,
		"commands", cmdStr,
	)

	if hasCR && e.LazyRebuildMs > 0 {
		e.scheduleLazyRebuild(files, dispFile)
		return
	}

	e.executeCacheClear(files, dispFile)
}

func (e *Engine) executeCacheClear(files map[string]struct{}, dispFile string) {
	e.mu.Lock()
	cmds := e.resolveCommands(files)
	cmdStr := strings.Join(cmds, " + ")
	e.mu.Unlock()

	if len(cmds) == 0 {
		return
	}

	changes := int64(len(files))
	affected := e.affectedSites(files)

	if len(affected) == 0 {
		result := e.Executor.Execute(context.Background(), cmds, e.DrupalRoot)
		e.stats.clears.Add(1)
		metrics.RecordClear("")
		e.runPostProcessors(context.Background(), core.FileEvent{Path: dispFile}, result)
		e.publishDrushResult(result, cmdStr, int(changes), dispFile, "")
	} else {
		e.executeMultiSite(affected, cmds, cmdStr, int(changes), dispFile)
	}
}

func (e *Engine) resolveCommands(files map[string]struct{}) []string {
	seen := make(map[string]struct{})
	var cmds []string
	for f := range files {
		args := resolveCommand(f, e.CommandsPerPattern)
		cmdStr := strings.Join(args, " ")
		if _, ok := seen[cmdStr]; !ok {
			seen[cmdStr] = struct{}{}
			cmds = append(cmds, cmdStr)
		}
	}
	return cmds
}

func (e *Engine) executeMultiSite(sites []core.SiteInfo, cmds []string, cmdStr string, changes int, dispFile string) {
	concurrency := 3
	type siteJob struct {
		site core.SiteInfo
	}
	jobs := make(chan siteJob, len(sites))
	results := make(chan struct{}, len(sites))

	for w := 0; w < concurrency; w++ {
		go func() {
			for j := range jobs {
				s := j.site
				exec := e.Executor
				if e.SiteExecutorFactory != nil {
					exec = e.SiteExecutorFactory(s)
				}
				result := exec.Execute(context.Background(), cmds, e.DrupalRoot)
				e.stats.clears.Add(1)
				metrics.RecordClear(s.Name)
				e.runPostProcessors(context.Background(), core.FileEvent{Path: dispFile}, result)
				e.publishDrushResult(result, cmdStr, changes, dispFile, s.Name)
				results <- struct{}{}
			}
		}()
	}

	for _, site := range sites {
		jobs <- siteJob{site: site}
	}
	close(jobs)

	for i := 0; i < len(sites); i++ {
		<-results
	}
}

func (e *Engine) isWatchedFile(path string) bool {
	for _, route := range e.Routes {
		if !strings.HasPrefix(path, route) {
			continue
		}
		if len(path) == len(route) || path[len(route)] == filepath.Separator {
			return true
		}
	}
	return false
}

func (e *Engine) lintFiles(files map[string]struct{}) *core.LintResult {
	var (
		mu     sync.Mutex
		wg     sync.WaitGroup
		fail   *core.LintResult
	)
	for f := range files {
		if !e.isWatchedFile(f) {
			continue
		}
		ext := filepath.Ext(f)
		if ext == ".info" {
			ext = ".yml"
		}
		checker, ok := e.LintCheckers[ext]
		if !ok {
			continue
		}
		wg.Add(1)
		go func(path string, chk core.LintChecker) {
			defer wg.Done()
			if result := chk.Lint(path); result != nil {
				mu.Lock()
				if fail == nil {
					fail = result
				}
				mu.Unlock()
			}
		}(f, checker)
	}
	wg.Wait()
	return fail
}

func (e *Engine) affectedSites(files map[string]struct{}) []core.SiteInfo {
	sites := e.ResolvedSites
	if len(sites) == 0 {
		return nil
	}

	type marker struct {
		name string
		path string
	}
	markers := make([]marker, len(sites))
	for i, s := range sites {
		markers[i] = marker{name: s.Name, path: string(filepath.Separator) + "sites" + string(filepath.Separator) + s.Name + string(filepath.Separator)}
	}

	sharedFile := false
	siteSet := make(map[string]bool)

	for f := range files {
		found := false
		for _, m := range markers {
			if strings.Contains(f, m.path) {
				siteSet[m.name] = true
				found = true
				break
			}
		}
		if !found {
			sharedFile = true
		}
	}

	if sharedFile || len(siteSet) > 1 {
		return sites
	}

	for name := range siteSet {
		for _, s := range sites {
			if s.Name == name {
				return []core.SiteInfo{s}
			}
		}
	}

	return sites
}

func (e *Engine) runPostProcessors(ctx context.Context, event core.FileEvent, result core.ExecutionResult) {
	if result.ExitCode == 0 {
		return
	}
	e.Logger.Error("cache clear failed",
		"exit_code", result.ExitCode,
		"stderr", result.Stderr,
	)
}

func (e *Engine) publishDrushResult(result core.ExecutionResult, cmdStr string, changes int, dispFile, siteName string) {
	if e.EventBus == nil {
		return
	}
	e.EventBus.Publish(eventbus.TopicCacheClear, core.EngineEvent{
		Type:      core.EventCacheClear,
		File:      dispFile,
		Changes:   changes,
		Commands:  cmdStr,
		ExitCode:  result.ExitCode,
		Duration:  result.Duration,
		Stderr:    strings.TrimSpace(result.Stderr),
		SiteName:  siteName,
		Timestamp: time.Now(),
	})
}

func (e *Engine) Stats() (changes, clears int64) {
	return e.stats.changes.Load(), e.stats.clears.Load()
}

func (e *Engine) StartTime() time.Time {
	return e.startTime
}

func resolveCommand(file string, commandsPerPattern map[string]string) []string {
	if commandsPerPattern == nil {
		return []string{"cr"}
	}

	type kv struct {
		pattern string
		command string
	}
	var sorted []kv
	for k, v := range commandsPerPattern {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return len(sorted[i].pattern) > len(sorted[j].pattern)
	})

	for _, kv := range sorted {
		if strings.HasSuffix(file, kv.pattern) {
			return strings.Fields(kv.command)
		}
	}
	return []string{"cr"}
}

// Shutdown cancels the engine's context, stopping all goroutines.
// Implements do.Shutdowner for graceful DI lifecycle management.
func (e *Engine) Shutdown() error {
	if e.cancel != nil {
		e.cancel()
	}
	return nil
}

func (e *Engine) String() string {
	return fmt.Sprintf("Engine{debounce=%dms, filters=%d}",
		e.Debounce, len(e.Filters))
}

func DefaultDebounce() int {
	return 800
}

type EngineConfig struct {
	Watcher             core.Watcher
	Executor            core.CommandExecutor
	SiteExecutorFactory func(site core.SiteInfo) core.CommandExecutor
	Filters             []core.EventFilter
	LintCheckers        map[string]core.LintChecker
	SkipLint            bool
	EventBus            *eventbus.EventBus
	Logger              *slog.Logger
	Debounce            int
	LazyRebuildMs       int
	Patterns            []string
	ExcludePatterns     []string
	CommandsPerPattern  map[string]string
	ResolvedSites       []core.SiteInfo
	DrupalRoot          string
	Routes              []string
}
