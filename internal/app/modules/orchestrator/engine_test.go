package orchestrator

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/irving-frias/drupal-watcher/internal/app/eventbus"
	"github.com/irving-frias/drupal-watcher/pkg/core"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestDefaultDebounce(t *testing.T) {
	if d := DefaultDebounce(); d != 800 {
		t.Errorf("expected 800, got %d", d)
	}
}

func TestResolveCommand_MatchLongest(t *testing.T) {
	cmds := map[string]string{
		".html.twig": "cc render",
		".twig":      "cc render",
		".php":       "cc plugin",
	}
	result := resolveCommand("node.html.twig", cmds)
	got := strings.Join(result, " ")
	if got != "cc render" {
		t.Errorf("expected 'cc render', got %q", got)
	}
}

func TestResolveCommand_Fallback(t *testing.T) {
	result := resolveCommand("unknown.xyz", map[string]string{".php": "cc plugin"})
	got := strings.Join(result, " ")
	if got != "cr" {
		t.Errorf("expected 'cr' fallback, got %q", got)
	}
}

func TestResolveCommand_NilMap(t *testing.T) {
	result := resolveCommand("file.php", nil)
	got := strings.Join(result, " ")
	if got != "cr" {
		t.Errorf("expected 'cr', got %q", got)
	}
}

func TestResolveCommand_Exact(t *testing.T) {
	cmds := map[string]string{
		".php": "cc plugin",
		".inc": "cc theme-registry",
	}
	result := resolveCommand("custom.inc", cmds)
	got := strings.Join(result, " ")
	if got != "cc theme-registry" {
		t.Errorf("expected 'cc theme-registry', got %q", got)
	}
}

func TestNewEngineCommandBuilder(t *testing.T) {
	builder := NewEngineCommandBuilder(map[string]string{
		".php": "cc plugin",
	})
	cmd := builder("test.php")
	if cmd != "cc plugin" {
		t.Errorf("expected 'cc plugin', got %q", cmd)
	}
}

func TestShouldProcess(t *testing.T) {
	e := &Engine{
		Filters: []core.EventFilter{
			&mockFilter{allow: true},
		},
	}
	if !e.shouldProcess(core.FileEvent{Path: "test.php"}) {
		t.Error("expected event to be processed")
	}
}

func TestShouldProcess_Rejected(t *testing.T) {
	e := &Engine{
		Filters: []core.EventFilter{
			&mockFilter{allow: false},
		},
	}
	if e.shouldProcess(core.FileEvent{Path: "test.php"}) {
		t.Error("expected event to be rejected")
	}
}

func TestShouldProcess_RejectedByAny(t *testing.T) {
	e := &Engine{
		Filters: []core.EventFilter{
			&mockFilter{allow: true},
			&mockFilter{allow: false},
		},
	}
	if e.shouldProcess(core.FileEvent{Path: "test.php"}) {
		t.Error("expected event to be rejected by second filter")
	}
}

func TestAffectedSites_NoSites(t *testing.T) {
	e := &Engine{ResolvedSites: nil}
	result := e.affectedSites(map[string]struct{}{"test.php": {}})
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestAffectedSites_SpecificSite(t *testing.T) {
	e := &Engine{
		ResolvedSites: []core.SiteInfo{
			{Name: "default", URI: "http://default"},
			{Name: "admin", URI: "http://admin"},
		},
	}
	files := map[string]struct{}{
		"/drupal/sites/admin/modules/custom/foo.module": {},
	}
	result := e.affectedSites(files)
	if len(result) != 1 || result[0].Name != "admin" {
		t.Errorf("expected [admin], got %v", result)
	}
}

func TestAffectedSites_SharedFile(t *testing.T) {
	e := &Engine{
		ResolvedSites: []core.SiteInfo{
			{Name: "default", URI: "http://default"},
			{Name: "admin", URI: "http://admin"},
		},
	}
	files := map[string]struct{}{
		"/drupal/modules/custom/shared.module": {},
	}
	result := e.affectedSites(files)
	if len(result) != 2 {
		t.Errorf("expected 2 sites for shared file, got %d", len(result))
	}
}

func TestAffectedSites_MultiSiteTriggerAll(t *testing.T) {
	e := &Engine{
		ResolvedSites: []core.SiteInfo{
			{Name: "default", URI: "http://default"},
			{Name: "admin", URI: "http://admin"},
		},
	}
	files := map[string]struct{}{
		"/drupal/sites/admin/modules/custom/a.module": {},
		"/drupal/sites/default/modules/custom/b.module": {},
	}
	result := e.affectedSites(files)
	if len(result) != 2 {
		t.Errorf("expected 2 sites for multi-site change, got %d", len(result))
	}
}

func TestIsWatchedFile(t *testing.T) {
	e := &Engine{
		Routes: []string{"docroot/modules/custom", "docroot/themes/custom"},
	}
	if !e.isWatchedFile("docroot/modules/custom/my_module/module.php") {
		t.Error("expected watched file to match")
	}
	if e.isWatchedFile("docroot/vendor/some_pkg/file.php") {
		t.Error("expected unwatched file to be rejected")
	}
}

func TestProcessBatch(t *testing.T) {
	mu := sync.Mutex{}
	calls := []string{}

	exec := &mockExecutor{
		fn: func(cmds []string) core.ExecutionResult {
			mu.Lock()
			calls = append(calls, strings.Join(cmds, " "))
			mu.Unlock()
			return core.ExecutionResult{ExitCode: 0}
		},
	}

	e := &Engine{
		Executor: exec,
		Debounce: 50,
		CommandsPerPattern: map[string]string{
			".php": "cc plugin",
		},
		LintCheckers: nil,
		EventBus:     eventbus.New(),
		Logger:       testLogger(),
		changedFiles: map[string]struct{}{"test.php": {}},
		lastFile:     "test.php",
	}

	e.processBatch()

	if len(calls) == 0 {
		t.Fatal("expected executor to be called")
	}
	if calls[0] != "cc plugin" {
		t.Errorf("expected 'cc plugin', got %q", calls[0])
	}
}

func TestProcessBatch_DeduplicatesCommands(t *testing.T) {
	mu := sync.Mutex{}
	calls := []string{}

	exec := &mockExecutor{
		fn: func(cmds []string) core.ExecutionResult {
			mu.Lock()
			calls = append(calls, strings.Join(cmds, " "))
			mu.Unlock()
			return core.ExecutionResult{ExitCode: 0}
		},
	}

	e := &Engine{
		Executor: exec,
		Debounce: 50,
		CommandsPerPattern: map[string]string{
			".php": "cc plugin",
			".inc": "cc plugin",
		},
		LintCheckers: nil,
		EventBus:     eventbus.New(),
		Logger:       testLogger(),
		changedFiles: map[string]struct{}{
			"test.php": {},
			"test.inc": {},
		},
		lastFile: "test.php",
	}

	e.processBatch()

	mu.Lock()
	defer mu.Unlock()

	if len(calls) != 1 {
		t.Errorf("expected 1 unique command, got %d: %v", len(calls), calls)
	}
}

func TestEngineRun_ContextCancellation(t *testing.T) {
	exec := &mockExecutor{
		fn: func(cmds []string) core.ExecutionResult {
			return core.ExecutionResult{ExitCode: 0}
		},
	}

	watcher := &mockWatcher{
		eventsCh: make(chan core.FileEvent),
		errsCh:   make(chan error),
	}

	e := &Engine{
		Watcher:  watcher,
		Executor: exec,
		Logger:   testLogger(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := e.Run(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

type mockFilter struct {
	allow bool
}

func (m *mockFilter) ShouldProcess(core.FileEvent) bool {
	return m.allow
}

type mockExecutor struct {
	fn func([]string) core.ExecutionResult
}

func (m *mockExecutor) Execute(_ context.Context, cmds []string, _ string) core.ExecutionResult {
	return m.fn(cmds)
}

type mockWatcher struct {
	eventsCh chan core.FileEvent
	errsCh   chan error
	started  bool
}

func (m *mockWatcher) Start(_ context.Context) (<-chan core.FileEvent, <-chan error) {
	m.started = true
	return m.eventsCh, m.errsCh
}

func (m *mockWatcher) Add(path string) error  { return nil }
func (m *mockWatcher) Remove(path string) error { return nil }
func (m *mockWatcher) Close() error {
	close(m.eventsCh)
	close(m.errsCh)
	return nil
}

func TestStats(t *testing.T) {
	e := &Engine{}
	e.stats.changes.Store(5)
	e.stats.clears.Store(3)

	changes, clears := e.Stats()
	if changes != 5 {
		t.Errorf("expected 5 changes, got %d", changes)
	}
	if clears != 3 {
		t.Errorf("expected 3 clears, got %d", clears)
	}
}

func TestString(t *testing.T) {
	e := &Engine{Debounce: 500}
	s := e.String()
	if !strings.Contains(s, "500") {
		t.Errorf("expected engine string to contain debounce, got %q", s)
	}
}

func TestStartTime(t *testing.T) {
	before := time.Now()
	e := NewEngine(EngineConfig{
		Logger:   nil,
		Debounce: 800,
	})
	after := time.Now()
	st := e.StartTime()
	if st.Before(before) || st.After(after) {
		t.Error("start time should be between before and after")
	}
}

func TestValidateEngineConfig(t *testing.T) {
	cfg := ValidateEngineConfig(EngineConfig{})
	if cfg.Debounce != 800 {
		t.Errorf("expected debounce 800, got %d", cfg.Debounce)
	}
	if cfg.Logger == nil {
		t.Error("expected non-nil logger after validation")
	}
}
