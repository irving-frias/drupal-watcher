package adapters

import (
	"context"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type fileEntry struct {
	modTime time.Time
}

type dirSnapshot map[string]fileEntry

type PollingWatcher struct {
	routes     []string
	skipDirs   []string
	skipSet    map[string]bool
	interval   time.Duration
	maxBackoff time.Duration
	bufSize    int
	mu         sync.Mutex
	snapshot   dirSnapshot
}

func NewPollingWatcher(routes, skipDirs []string, interval time.Duration) *PollingWatcher {
	return NewPollingWatcherWithOpts(routes, skipDirs, interval, WatcherOptions{BufferSize: 100})
}

func NewPollingWatcherWithOpts(routes, skipDirs []string, interval time.Duration, opts WatcherOptions) *PollingWatcher {
	skipSet := make(map[string]bool, len(skipDirs))
	for _, s := range skipDirs {
		skipSet[s] = true
	}
	bufSize := opts.BufferSize
	if bufSize <= 0 {
		bufSize = 100
	}
	return &PollingWatcher{
		routes:     routes,
		skipDirs:   skipDirs,
		skipSet:    skipSet,
		interval:   interval,
		maxBackoff: interval * 4,
		bufSize:    bufSize,
		snapshot:   make(dirSnapshot),
	}
}

func (w *PollingWatcher) Start(ctx context.Context) (<-chan core.FileEvent, <-chan error) {
	events := make(chan core.FileEvent, w.bufSize)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		w.mu.Lock()
		w.snapshot = w.scan()
		w.mu.Unlock()

		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		idleCount := 0

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

				diffs := diffSnapshots(prev, curr)
				for _, evt := range diffs {
					select {
					case events <- evt:
					case <-ctx.Done():
						return
					}
				}

				if len(diffs) == 0 {
					idleCount++
					if idleCount >= 3 {
						ticker.Stop()
						backoff := w.interval * time.Duration(1+idleCount/3)
						if backoff > w.maxBackoff {
							backoff = w.maxBackoff
						}
						ticker = time.NewTicker(backoff)
					}
				} else {
					idleCount = 0
					if ticker.C != time.NewTicker(w.interval).C {
						ticker.Stop()
						ticker = time.NewTicker(w.interval)
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
		filepath.WalkDir(route, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if w.skipSet[d.Name()] {
					return fs.SkipDir
				}
				return nil
			}
			if w.shouldSkip(d.Name(), path) {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			snap[path] = fileEntry{modTime: info.ModTime()}
			return nil
		})
	}
	return snap
}

func (w *PollingWatcher) shouldSkip(name, path string) bool {
	if w.skipSet[name] {
		return true
	}
	for _, skip := range w.skipDirs {
		if strings.Contains(path, skip) {
			return true
		}
		if strings.HasPrefix(skip, "/.") && strings.HasPrefix(name, ".") {
			return true
		}
	}
	return false
}

func diffSnapshots(prev, curr dirSnapshot) []core.FileEvent {
	if len(curr) == 0 && len(prev) == 0 {
		return nil
	}

	events := make([]core.FileEvent, 0, maxInt(len(curr), len(prev)))

	for path, cur := range curr {
		prevEntry, existed := prev[path]
		if !existed {
			events = append(events, core.FileEvent{Path: path, Op: core.Create, IsDir: false})
		} else if !cur.modTime.Equal(prevEntry.modTime) {
			events = append(events, core.FileEvent{Path: path, Op: core.Write, IsDir: false})
		}
	}

	for path := range prev {
		if _, exists := curr[path]; !exists {
			events = append(events, core.FileEvent{Path: path, Op: core.Remove, IsDir: false})
		}
	}

	if len(events) > 1 {
		sort.Slice(events, func(i, j int) bool {
			return events[i].Path < events[j].Path
		})
	}

	return events
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
