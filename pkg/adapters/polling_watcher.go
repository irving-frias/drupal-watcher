package adapters

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type dirSnapshot map[string]time.Time

type PollingWatcher struct {
	routes    []string
	skipDirs  []string
	interval  time.Duration
	mu        sync.Mutex
	snapshot  dirSnapshot
}

func NewPollingWatcher(routes, skipDirs []string, interval time.Duration) *PollingWatcher {
	return &PollingWatcher{
		routes:   routes,
		skipDirs: skipDirs,
		interval: interval,
		snapshot: make(dirSnapshot),
	}
}

func (w *PollingWatcher) Start(ctx context.Context) (<-chan core.FileEvent, <-chan error) {
	events := make(chan core.FileEvent, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		w.mu.Lock()
		w.snapshot = w.scan()
		w.mu.Unlock()

		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w.mu.Lock()
				prev := w.snapshot
				curr := w.scan()
				w.snapshot = curr
				w.mu.Unlock()
				for _, evt := range diffSnapshots(prev, curr) {
					select {
					case events <- evt:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return events, errs
}

func (w *PollingWatcher) Add(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, r := range w.routes {
		if r == path {
			return nil
		}
	}
	w.routes = append(w.routes, path)
	return w.rescan()
}

func (w *PollingWatcher) Remove(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	filtered := make([]string, 0, len(w.routes))
	for _, r := range w.routes {
		if r != path {
			filtered = append(filtered, r)
		}
	}
	w.routes = filtered
	return w.rescan()
}

func (w *PollingWatcher) Close() error {
	return nil
}

func (w *PollingWatcher) rescan() error {
	w.snapshot = w.scan()
	return nil
}

func (w *PollingWatcher) scan() dirSnapshot {
	snap := make(dirSnapshot)
	for _, route := range w.routes {
		w.walkDir(route, snap)
	}
	return snap
}

func (w *PollingWatcher) walkDir(dir string, snap dirSnapshot) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		if w.shouldSkip(e.Name(), path) {
			continue
		}
		if e.IsDir() {
			w.walkDir(path, snap)
		} else {
			info, err := e.Info()
			if err != nil {
				continue
			}
			snap[path] = info.ModTime()
		}
	}
}

func (w *PollingWatcher) shouldSkip(name, path string) bool {
	for _, skip := range w.skipDirs {
		if strings.Contains(path, skip) || name == skip {
			return true
		}
		if strings.HasPrefix(skip, "/.") && strings.HasPrefix(name, ".") {
			return true
		}
	}
	return false
}

func diffSnapshots(prev, curr dirSnapshot) []core.FileEvent {
	seen := make(map[string]bool)
	var events []core.FileEvent

	for path, curMod := range curr {
		seen[path] = true
		prevMod, existed := prev[path]
		if !existed {
			events = append(events, core.FileEvent{Path: path, Op: core.Create, IsDir: false})
		} else if !curMod.Equal(prevMod) {
			events = append(events, core.FileEvent{Path: path, Op: core.Write, IsDir: false})
		}
	}

	for path := range prev {
		if !seen[path] {
			events = append(events, core.FileEvent{Path: path, Op: core.Remove, IsDir: false})
		}
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Path < events[j].Path
	})

	return events
}


