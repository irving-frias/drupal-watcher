package watcher_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/irving-frias/drupal-watcher/internal/watcher"
)

type mockConfig struct {
	routes             []string
	patterns           []string
	excludePatterns    []string
	debounce           int
	drushCmd           *string
	drushCommand       string
	drushArgs          []string
	postClearCommands  []string
	commandsPerPattern map[string]string
	drupalRoot         *string
}

func (m mockConfig) GetRoutes() []string                  { return m.routes }
func (m mockConfig) GetPatterns() []string                { return m.patterns }
func (m mockConfig) GetExcludePatterns() []string         { return m.excludePatterns }
func (m mockConfig) GetDebounce() int                     { return m.debounce }
func (m mockConfig) GetDrushCmd() *string                 { return m.drushCmd }
func (m mockConfig) GetDrushCommand() string               { return m.drushCommand }
func (m mockConfig) GetDrushArgs() []string                { return m.drushArgs }
func (m mockConfig) GetPostClearCommands() []string        { return m.postClearCommands }
func (m mockConfig) GetCommandsPerPattern() map[string]string { return m.commandsPerPattern }
func (m mockConfig) GetDrupalRoot() *string                 { return m.drupalRoot }

func TestWatcherStartStop(t *testing.T) {
	tmp := t.TempDir()
	customDir := filepath.Join(tmp, "custom")
	if err := os.MkdirAll(customDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := mockConfig{
		routes:           []string{customDir},
		patterns:         []string{".php"},
		debounce:         100,
		drushCommand:     "cr",
	}

	h, err := watcher.Start(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}

	if h == nil {
		t.Fatal("expected non-nil handle")
	}

	if h.WatchCount <= 0 {
		t.Errorf("expected positive WatchCount, got %d", h.WatchCount)
	}

	watcher.Stop(h)
}

func TestWatcherDrushNotFound(t *testing.T) {
	tmp := t.TempDir()
	customDir := filepath.Join(tmp, "custom")
	if err := os.MkdirAll(customDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := mockConfig{
		routes:           []string{customDir},
		patterns:         []string{".php"},
		debounce:         100,
		drushCommand:     "nonexistent",
	}

	h, err := watcher.Start(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	watcher.Stop(h)
}

func TestGatherDirs(t *testing.T) {
	// This is a basic test to ensure gatherDirs doesn't panic
	// The actual function is unexported, so we test it via Start/Stop
	tmp := t.TempDir()
	dir1 := filepath.Join(tmp, "modules", "custom")
	dir2 := filepath.Join(tmp, "themes", "custom")
	if err := os.MkdirAll(dir1, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir2, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := mockConfig{
		routes:   []string{tmp},
		patterns: []string{".php"},
		debounce: 100,
	}

	h, err := watcher.Start(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	watcher.Stop(h)
}
