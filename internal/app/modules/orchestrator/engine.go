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
	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type Engine struct {
	Watcher            core.Watcher
	Executor           core.CommandExecutor
	SiteExecutorFactory func(site core.SiteInfo) core.CommandExecutor
	Filters            []core.EventFilter
	PostProcessors     []core.PostProcessor
	LintCheckers       map[string]core.LintChecker
	SkipLint           bool
	EventBus           *eventbus.EventBus
	Logger             *slog.Logger
	Debounce           int
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

	startTime time.Time

	stats struct {
		changes atomic.Int64
		clears  atomic.Int64
	}
}

func NewEngine(cfg EngineConfig) *Engine {
	return &Engine{
		Watcher:            cfg.Watcher,
		Executor:           cfg.Executor,
		SiteExecutorFactory: cfg.SiteExecutorFactory,
		Filters:            cfg.Filters,
		PostProcessors:     cfg.PostProcessors,
		LintCheckers:       cfg.LintCheckers,
		SkipLint:           cfg.SkipLint,
		EventBus:           cfg.EventBus,
		Logger:             cfg.Logger,
		Debounce:           cfg.Debounce,
		Patterns:           cfg.Patterns,
		ExcludePatterns:    cfg.ExcludePatterns,
		CommandsPerPattern: cfg.CommandsPerPattern,
		ResolvedSites:      cfg.ResolvedSites,
		DrupalRoot:         cfg.DrupalRoot,
		Routes:             cfg.Routes,
		changedFiles:       make(map[string]struct{}),
		startTime:          time.Now(),
	}
}

func (e *Engine) Run(ctx context.Context) error {
	events, errs := e.Watcher.Start(ctx)

	e.Logger.Info("orchestrator started",
		"debounce_ms", e.Debounce,
		"filters", len(e.Filters),
		"post_processors", len(e.PostProcessors),
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
			e.mu.Unlock()
			e.pending.Store(true)

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
	e.stats.changes.Add(changes)

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

	cmdStr := strings.Join(cmds, " + ")

	e.Logger.Info("change detected",
		"file", dispFile,
		"changes", changes,
		"commands", cmdStr,
	)

	if e.EventBus != nil {
		e.EventBus.Publish(eventbus.TopicFileChange, core.EngineEvent{
			Type:      core.EventChange,
			File:      dispFile,
			Changes:   int(changes),
			Timestamp: time.Now(),
		})
	}

	affected := e.affectedSites(files)

	if len(affected) == 0 {
		result := e.Executor.Execute(context.Background(), cmds, e.DrupalRoot)
		e.stats.clears.Add(1)
		e.runPostProcessors(context.Background(), core.FileEvent{Path: dispFile}, result)
		e.publishDrushResult(result, cmdStr, int(changes), dispFile, "")
	} else {
		var wg sync.WaitGroup
		for _, site := range affected {
			wg.Add(1)
			go func(s core.SiteInfo) {
				defer wg.Done()
				exec := e.Executor
				if e.SiteExecutorFactory != nil {
					exec = e.SiteExecutorFactory(s)
				}
				result := exec.Execute(context.Background(), cmds, e.DrupalRoot)
				e.stats.clears.Add(1)
				e.runPostProcessors(context.Background(), core.FileEvent{Path: dispFile}, result)
				e.publishDrushResult(result, cmdStr, int(changes), dispFile, s.Name)
			}(site)
		}
		wg.Wait()
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
	for _, pp := range e.PostProcessors {
		if err := pp.Process(ctx, event, result); err != nil {
			e.Logger.Error("post-processor failed",
				"name", pp.Name(),
				"error", err,
			)
		}
	}
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

func NewEngineCommandBuilder(commandsPerPattern map[string]string) func(file string) string {
	return func(file string) string {
		args := resolveCommand(file, commandsPerPattern)
		return strings.Join(args, " ")
	}
}

func (e *Engine) String() string {
	return fmt.Sprintf("Engine{debounce=%dms, filters=%d, postProcs=%d}",
		e.Debounce, len(e.Filters), len(e.PostProcessors))
}

func DefaultDebounce() int {
	return 800
}

type EngineConfig struct {
	Watcher             core.Watcher
	Executor            core.CommandExecutor
	SiteExecutorFactory func(site core.SiteInfo) core.CommandExecutor
	Filters             []core.EventFilter
	PostProcessors      []core.PostProcessor
	LintCheckers        map[string]core.LintChecker
	SkipLint            bool
	EventBus            *eventbus.EventBus
	Logger              *slog.Logger
	Debounce            int
	Patterns            []string
	ExcludePatterns     []string
	CommandsPerPattern  map[string]string
	ResolvedSites       []core.SiteInfo
	DrupalRoot          string
	Routes              []string
}

func ValidateEngineConfig(cfg EngineConfig) EngineConfig {
	if cfg.Debounce <= 0 {
		cfg.Debounce = DefaultDebounce()
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return cfg
}
