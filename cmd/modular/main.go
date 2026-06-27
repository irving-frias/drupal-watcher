package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/irving-frias/drupal-watcher/internal/app/eventbus"
	"github.com/irving-frias/drupal-watcher/internal/app/modules/orchestrator"
	"github.com/irving-frias/drupal-watcher/internal/app/modules/ui/providers/tui"
	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/internal/drush"
	"github.com/irving-frias/drupal-watcher/internal/hooks/builtin"
	"github.com/irving-frias/drupal-watcher/pkg/adapters"
	"github.com/irving-frias/drupal-watcher/pkg/core"
	"github.com/pterm/pterm"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	cfg, err := loadConfig(root)
	if err != nil {
		pterm.Error.Printfln("Config: %v", err)
		os.Exit(1)
	}

	dr := drupalRoot(root, cfg)
	checkDrush(cfg)

	if err := config.WritePid(root); err != nil {
		pterm.Error.Printfln("PID: %v", err)
		os.Exit(1)
	}
	defer config.RemovePid(root)

	resolveSites(cfg, root, dr)

	fsnWatcher, err := createWatcher(cfg)
	if err != nil {
		pterm.Error.Printfln("Watcher: %v", err)
		os.Exit(1)
	}
	defer fsnWatcher.Close()

	bus := eventbus.New()

	engine := orchestrator.NewEngine(orchestrator.EngineConfig{
		Watcher:            fsnWatcher,
		Executor:           adapters.NewDrushExecutor(cfg),
		Filters:            createFilters(cfg),
		PostProcessors:     []core.PostProcessor{&builtin.DrushClear{}},
		EventBus:           bus,
		Logger:             adapters.NewDiscardLogger(),
		Debounce:           cfg.Debounce,
		Patterns:           cfg.Patterns,
		ExcludePatterns:    cfg.ExcludePatterns,
		CommandsPerPattern: cfg.CommandsPerPattern,
		ResolvedSites:      cfg.GetResolvedSites(),
		DrupalRoot:         dr,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		cancel()
	}()

	pterm.Info.Printfln("Routes: %s", pterm.Cyan(stringsJoin(cfg.Routes, ", ")))
	pterm.Info.Printfln("Patterns: %s", pterm.Cyan(stringsJoin(cfg.Patterns, ", ")))
	pterm.Success.Printfln("Modular watcher PID %d.", os.Getpid())

	go func() {
		if err := engine.Run(ctx); err != nil && err != context.Canceled {
			pterm.Error.Printfln("Engine: %v", err)
		}
	}()

	if err := tui.RunWithBus(ctx, bus); err != nil {
		pterm.Warning.Printfln("TUI: %v", err)
	}
}

func loadConfig(root string) (*config.Config, error) {
	mgr := config.NewManager()
	c, err := mgr.LoadConfig(root)
	return &c, err
}

func drupalRoot(root string, cfg *config.Config) string {
	if cfg.DrupalRoot != nil {
		return filepath.Join(root, *cfg.DrupalRoot)
	}
	return root
}

func checkDrush(cfg *config.Config) {
	if !drush.HealthCheck(cfg) {
		pterm.Warning.Println("Drush not available. Cache clears will fail.")
	}
}

func createWatcher(cfg *config.Config) (*adapters.FSNotifyWatcher, error) {
	skipDirs := append(adapters.DefaultSkipDirs(), cfg.ExcludePatterns...)
	return adapters.NewFSNotifyWatcher(cfg.Routes, skipDirs)
}

func createFilters(cfg *config.Config) []core.EventFilter {
	return []core.EventFilter{
		adapters.NewPatternFilter(cfg.Patterns),
		adapters.NewExcludeFilter(cfg.ExcludePatterns),
		adapters.NewDotfileFilter(),
	}
}

func resolveSites(cfg *config.Config, root, drupalRoot string) {
	if len(cfg.Sites) > 0 {
		all, _ := drush.LoadSitesYml(drupalRoot, root)
		filtered := drush.FilterSites(all, cfg.Sites, nil)
		sites := make([]core.SiteInfo, 0, len(filtered))
		for name, s := range filtered {
			sites = append(sites, core.SiteInfo{Name: name, URI: s.URI})
		}
		cfg.SetResolvedSites(sites)
		return
	}

	if drush.HasMultiSite(drupalRoot) {
		all, _ := drush.LoadSitesYml(drupalRoot, root)
		sites := make([]core.SiteInfo, 0, len(all))
		for name, s := range all {
			sites = append(sites, core.SiteInfo{Name: name, URI: s.URI})
		}
		cfg.SetResolvedSites(sites)
	}
}

func stringsJoin(elems []string, sep string) string {
	if len(elems) == 0 {
		return ""
	}
	result := elems[0]
	for _, e := range elems[1:] {
		result += sep + e
	}
	return result
}
