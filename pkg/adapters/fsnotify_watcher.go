package adapters

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type WatcherOptions struct {
	BufferSize   int
	PollInterval time.Duration
	SkipDirs     []string
}

type FSNotifyWatcher struct {
	watcher   *fsnotify.Watcher
	routes    []string
	skipDirs  []string
	bufSize   int
}

func NewFSNotifyWatcher(routes, skipDirs []string) (*FSNotifyWatcher, error) {
	return NewFSNotifyWatcherWithOpts(routes, skipDirs, WatcherOptions{BufferSize: 100})
}

func NewFSNotifyWatcherWithOpts(routes, skipDirs []string, opts WatcherOptions) (*FSNotifyWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	bufSize := opts.BufferSize
	if bufSize <= 0 {
		bufSize = 100
	}
	fw := &FSNotifyWatcher{
		watcher:  w,
		routes:   routes,
		skipDirs: skipDirs,
		bufSize:  bufSize,
	}
	return fw, nil
}

func (f *FSNotifyWatcher) Start(ctx context.Context) (<-chan core.FileEvent, <-chan error) {
	eventCh := make(chan core.FileEvent, f.bufSize)
	errCh := make(chan error, 1)

	watchDirs := gatherDirs(f.routes, f.skipDirs)
	for _, dir := range watchDirs {
		if err := f.watcher.Add(dir); err != nil {
			errCh <- err
		}
	}

	go func() {
		defer close(eventCh)
		defer close(errCh)
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-f.watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Chmod) == 0 {
					continue
				}

				if event.Op&fsnotify.Create != 0 && runtime.GOOS == "linux" {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						base := filepath.Base(event.Name)
						if !isSkippedDir(base, f.skipDirs) {
							f.Add(event.Name)
						}
					}
				}

				fe := core.FileEvent{
					Path: event.Name,
					Op:   fsnotifyOpToCore(event.Op),
					IsDir: func() bool {
						info, err := os.Stat(event.Name)
						return err == nil && info.IsDir()
					}(),
				}
				select {
				case eventCh <- fe:
				case <-ctx.Done():
					return
				}
			case err, ok := <-f.watcher.Errors:
				if !ok {
					return
				}
				select {
				case errCh <- err:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return eventCh, errCh
}

func (f *FSNotifyWatcher) Add(path string) error {
	if runtime.GOOS != "linux" {
		return f.watcher.Add(path)
	}
	return filepath.WalkDir(path, func(walkPath string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if isSkippedDir(d.Name(), f.skipDirs) {
			return filepath.SkipDir
		}
		return f.watcher.Add(walkPath)
	})
}

func (f *FSNotifyWatcher) Remove(path string) error {
	return f.watcher.Remove(path)
}

func (f *FSNotifyWatcher) Close() error {
	return f.watcher.Close()
}

func fsnotifyOpToCore(op fsnotify.Op) core.Op {
	switch {
	case op&fsnotify.Create != 0:
		return core.Create
	case op&fsnotify.Write != 0:
		return core.Write
	case op&fsnotify.Remove != 0:
		return core.Remove
	case op&fsnotify.Rename != 0:
		return core.Rename
	case op&fsnotify.Chmod != 0:
		return core.Chmod
	default:
		return core.Write
	}
}

func isSkippedDir(name string, skipDirs []string) bool {
	for _, s := range skipDirs {
		if name == s {
			return true
		}
	}
	return false
}

var dirsSkippedByDefault = []string{
	"node_modules", ".git", ".svn", ".hg",
	"contrib", "vendor", "bower_components",
	"files", "images", "fonts",
}

func gatherDirs(routes []string, skipDirs []string) []string {
	recursive := runtime.GOOS == "darwin"
	dirSet := make(map[string]bool, len(routes))

	for _, route := range routes {
		abs, err := filepath.Abs(route)
		if err != nil {
			continue
		}
		if recursive {
			dirSet[abs] = true
			continue
		}
		filepath.WalkDir(abs, func(path string, d os.DirEntry, err error) error {
			if err != nil || !d.IsDir() {
				return nil
			}
			if isSkippedDir(d.Name(), skipDirs) {
				return filepath.SkipDir
			}
			dirSet[path] = true
			return nil
		})
	}

	dirs := make([]string, 0, len(dirSet))
	for d := range dirSet {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	return dirs
}

func DefaultSkipDirs() []string {
	return append([]string{}, dirsSkippedByDefault...)
}
