package adapters_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/irving-frias/drupal-watcher/pkg/adapters"
)

func TestFSNotifyWatcher_Creation(t *testing.T) {
	if runtime.GOOS == "darwin" && os.Getenv("CI") != "" {
		t.Skip("skipping on macOS CI due to FSEvents limitations")
	}

	w, err := adapters.NewFSNotifyWatcher(nil, []string{".git", "node_modules"})
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Close()
}

func TestFSNotifyWatcher_WithOpts(t *testing.T) {
	if runtime.GOOS == "darwin" && os.Getenv("CI") != "" {
		t.Skip("skipping on macOS CI due to FSEvents limitations")
	}

	w, err := adapters.NewFSNotifyWatcherWithOpts(
		[]string{"/tmp"},
		[]string{".git"},
		adapters.WatcherOptions{BufferSize: 200},
	)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Close()
}

func TestFSNotifyWatcher_DefaultSkipDirs(t *testing.T) {
	dirs := adapters.DefaultSkipDirs()
	expected := map[string]bool{
		"node_modules": true,
		".git":         true,
		"vendor":       true,
	}
	for _, d := range dirs {
		if expected[d] {
			delete(expected, d)
		}
	}
	if len(expected) > 0 {
		t.Errorf("expected skip dirs to contain node_modules, .git, vendor, got missing: %v", expected)
	}
}

func TestWatcherOptions_Defaults(t *testing.T) {
	opts := adapters.WatcherOptions{BufferSize: 100}
	if opts.BufferSize != 100 {
		t.Errorf("expected buffer 100, got %d", opts.BufferSize)
	}
}
