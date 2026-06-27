package adapters

import (
	"context"
	"sync"
	"time"

	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type HybridWatcher struct {
	fsnotify   *FSNotifyWatcher
	poll       *PollingWatcher
	dedupWin   time.Duration
}

func NewHybridWatcher(fsnotify *FSNotifyWatcher, poll *PollingWatcher, dedupWindow time.Duration) *HybridWatcher {
	if dedupWindow <= 0 {
		dedupWindow = time.Second
	}
	return &HybridWatcher{
		fsnotify: fsnotify,
		poll:     poll,
		dedupWin: dedupWindow,
	}
}

func (w *HybridWatcher) Start(ctx context.Context) (<-chan core.FileEvent, <-chan error) {
	outEvents := make(chan core.FileEvent, 200)
	outErrs := make(chan error, 2)

	fsEvents, fsErrs := w.fsnotify.Start(ctx)
	pollEvents, pollErrs := w.poll.Start(ctx)

	var wg sync.WaitGroup
	wg.Add(2)

	dedup := newDedup(w.dedupWin)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-fsEvents:
				if !ok {
					return
				}
				if dedup.isDuplicate(evt.Path) {
					continue
				}
				dedup.record(evt.Path)
				select {
				case outEvents <- evt:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-pollEvents:
				if !ok {
					return
				}
				if dedup.isDuplicate(evt.Path) {
					continue
				}
				dedup.record(evt.Path)
				select {
				case outEvents <- evt:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	go func() {
		wg.Wait()
		close(outEvents)
		close(outErrs)
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-fsErrs:
				if !ok {
					fsErrs = nil
					continue
				}
				select {
				case outErrs <- err:
				case <-ctx.Done():
					return
				}
			case err, ok := <-pollErrs:
				if !ok {
					pollErrs = nil
					continue
				}
				select {
				case outErrs <- err:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outEvents, outErrs
}

func (w *HybridWatcher) Add(path string) error {
	if err := w.fsnotify.Add(path); err != nil {
		return err
	}
	return w.poll.Add(path)
}

func (w *HybridWatcher) Remove(path string) error {
	if err := w.fsnotify.Remove(path); err != nil {
		return err
	}
	return w.poll.Remove(path)
}

func (w *HybridWatcher) Close() error {
	if err := w.fsnotify.Close(); err != nil {
		return err
	}
	return w.poll.Close()
}

type dedup struct {
	mu     sync.Mutex
	window time.Duration
	seen   map[string]time.Time
}

func newDedup(window time.Duration) *dedup {
	return &dedup{
		window: window,
		seen:   make(map[string]time.Time),
	}
}

func (d *dedup) isDuplicate(path string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	t, ok := d.seen[path]
	if !ok || time.Since(t) > d.window {
		return false
	}
	return true
}

func (d *dedup) record(path string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seen[path] = time.Now()
}
