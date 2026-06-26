package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/internal/orchestrator"
	"github.com/irving-frias/drupal-watcher/internal/ui"
	"github.com/irving-frias/drupal-watcher/pkg/adapters"
	"github.com/irving-frias/drupal-watcher/pkg/core"
)

func main() {
	root := flag.String("root", "", "Drupal root directory")
	debounce := flag.Int("debounce", 800, "Debounce interval in ms")
	noTUI := flag.Bool("no-tui", false, "Disable TUI")
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	mgr := config.NewManager()
	if *configPath != "" {
		mgr.SetCustomConfigPath(*configPath)
	}

	r := *root
	if r == "" && len(flag.Args()) > 0 {
		r = flag.Args()[0]
	}

	cfg, err := mgr.LoadConfig(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if *debounce > 0 {
		cfg.Debounce = *debounce
	}

	logger := adapters.NewSlogLogger("")

	skipDirs := append(adapters.DefaultSkipDirs(), cfg.ExcludePatterns...)

	fsnWatcher, err := adapters.NewFSNotifyWatcher(cfg.Routes, skipDirs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create watcher: %v\n", err)
		os.Exit(1)
	}
	defer fsnWatcher.Close()

	drushExec := adapters.NewDrushExecutor(cfg)

	patternFilter := adapters.NewPatternFilter(cfg.Patterns)
	excludeFilter := adapters.NewExcludeFilter(cfg.ExcludePatterns)

	filters := []core.EventFilter{patternFilter, excludeFilter}

	var resolvedSites []core.SiteInfo
	for _, s := range cfg.GetResolvedSites() {
		resolvedSites = append(resolvedSites, core.SiteInfo{Name: s.Name, URI: s.URI})
	}

	drupalRoot := ""
	if cfg.DrupalRoot != nil {
		drupalRoot = filepath.Join(r, *cfg.DrupalRoot)
	}

	var eventChan chan core.EngineEvent
	if !*noTUI {
		eventChan = make(chan core.EngineEvent, 100)
	}

	engineCfg := core.EngineConfig{
		Watcher:            fsnWatcher,
		Executor:           drushExec,
		SiteExecutorFactory: func(site core.SiteInfo) core.CommandExecutor {
			return adapters.NewSiteAwareDrushExecutor(cfg, site.Name, site.URI)
		},
		Filters:            filters,
		PostProcessors:     []core.PostProcessor{},
		EventChan:          eventChan,
		Logger:             logger,
		Debounce:           cfg.Debounce,
		Patterns:           cfg.Patterns,
		ExcludePatterns:    cfg.ExcludePatterns,
		CommandsPerPattern: cfg.CommandsPerPattern,
		ResolvedSites:      resolvedSites,
		DrupalRoot:         drupalRoot,
	}

	engine := orchestrator.New(engineCfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	if !*noTUI {
		go func() {
			if err := engine.Run(ctx); err != nil && err != context.Canceled {
				logger.Error("engine error", "error", err)
			}
		}()

		if err := ui.Run(eventChan, engine); err != nil {
			logger.Error("TUI error", "error", err)
		}
	} else {
		if err := engine.Run(ctx); err != nil && err != context.Canceled {
			fmt.Fprintf(os.Stderr, "Engine error: %v\n", err)
			os.Exit(1)
		}
	}
}
