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
	dedup      *dedup
	bufSize    int
}

func NewHybridWatcher(fsnotify *FSNotifyWatcher, poll *PollingWatcher, dedupWindow time.Duration, opts WatcherOptions) *HybridWatcher {
	if dedupWindow <= 0 {
		dedupWindow = time.Second
	}
	bufSize := opts.BufferSize
	if bufSize <= 0 {
		bufSize = 500
	}
	return &HybridWatcher{
		fsnotify: fsnotify,
		poll:     poll,
		dedupWin: dedupWindow,
		dedup:    newDedup(dedupWindow),
		bufSize:  bufSize,
	}
}

func (w *HybridWatcher) Start(ctx context.Context) (<-chan core.FileEvent, <-chan error) {
	outEvents := make(chan core.FileEvent, w.bufSize)
	outErrs := make(chan error, 2)

	fsEvents, fsErrs := w.fsnotify.Start(ctx)
	pollEvents, pollErrs := w.poll.Start(ctx)

	var wg sync.WaitGroup
	wg.Add(2)

	dedup := w.dedup

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
	w.dedup.stop()
	if err := w.fsnotify.Close(); err != nil {
		return err
	}
	return w.poll.Close()
}

type dedup struct {
	mu       sync.Mutex
	window   time.Duration
	seen     map[string]time.Time
	stopCh   chan struct{}
}

func newDedup(window time.Duration) *dedup {
	d := &dedup{
		window: window,
		seen:   make(map[string]time.Time),
		stopCh: make(chan struct{}),
	}
	go d.cleanupLoop()
	return d
}

func (d *dedup) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			d.mu.Lock()
			cutoff := time.Now().Add(-d.window)
			for path, t := range d.seen {
				if t.Before(cutoff) {
					delete(d.seen, path)
				}
			}
			d.mu.Unlock()
		case <-d.stopCh:
			return
		}
	}
}

func (d *dedup) stop() {
	close(d.stopCh)
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
