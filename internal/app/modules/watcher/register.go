package watcher

import (
	"fmt"
	"time"

	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/pkg/adapters"
	"github.com/irving-frias/drupal-watcher/pkg/core"
	"github.com/samber/do/v2"
)

// Register provides core.Watcher to the injector.
func Register(i do.Injector) error {
	cfg := do.MustInvoke[*config.Config](i)
	skipDirs := append(adapters.DefaultSkipDirs(), cfg.ExcludePatterns...)
	bufSize := cfg.EventBufferSize
	if bufSize <= 0 {
		bufSize = 500
	}

	opts := adapters.WatcherOptions{
		BufferSize:   bufSize,
		PollInterval: time.Duration(cfg.PollInterval) * time.Millisecond,
		SkipDirs:     skipDirs,
	}

	var w core.Watcher
	switch cfg.WatchMode {
	case "poll":
		w = adapters.NewPollingWatcherWithOpts(cfg.Routes, skipDirs, time.Duration(cfg.PollInterval)*time.Millisecond, opts)

	case "hybrid":
		fsn, err := adapters.NewFSNotifyWatcherWithOpts(cfg.Routes, skipDirs, opts)
		if err != nil {
			return fmt.Errorf("create fsnotify for hybrid: %w", err)
		}
		poll := adapters.NewPollingWatcherWithOpts(cfg.Routes, skipDirs, time.Duration(cfg.PollInterval)*time.Millisecond, opts)
		w = adapters.NewHybridWatcher(fsn, poll, time.Second, opts)

	default: // "auto" or "fsnotify"
		fsn, err := adapters.NewFSNotifyWatcherWithOpts(cfg.Routes, skipDirs, opts)
		if err != nil {
			w = adapters.NewPollingWatcherWithOpts(cfg.Routes, skipDirs, time.Duration(cfg.PollInterval)*time.Millisecond, opts)
		} else {
			w = fsn
		}
	}

	do.ProvideValue(i, w)
	return nil
}
