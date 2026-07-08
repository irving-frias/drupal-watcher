package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/irving-frias/drupal-watcher/internal/app"
	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/internal/drush"
	"github.com/irving-frias/drupal-watcher/pkg/adapters"
	"github.com/irving-frias/drupal-watcher/pkg/core"
	"github.com/samber/do/v2"
)

// RegisterConfig provides Config, WorkDir, and DrupalRoot to the injector.
func RegisterConfig(root string) func(i do.Injector) error {
	return func(i do.Injector) error {
		mgr := config.NewManager()
		c, err := mgr.LoadConfig(root)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		cfg := &c

		workDir := resolveAbs(root)
		do.ProvideValue(i, app.WorkDir(workDir))

		dr := workDir
		if cfg.DrupalRoot != nil {
			dr = filepath.Join(workDir, *cfg.DrupalRoot)
		}
		do.ProvideValue(i, app.DrupalRoot(dr))

		normalizeRoutes(cfg, workDir)
		resolveSites(cfg, dr, root)

		do.ProvideValue(i, cfg)
		return nil
	}
}

// RegisterExecutor provides core.CommandExecutor to the injector.
func RegisterExecutor(i do.Injector) error {
	cfg := do.MustInvoke[*config.Config](i)
	exec := adapters.NewDrushExecutor(cfg)
	do.ProvideValue(i, core.CommandExecutor(exec))
	return nil
}

// RegisterWatcher provides core.Watcher to the injector.
func RegisterWatcher(i do.Injector) error {
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

func normalizeRoutes(cfg *config.Config, workDir string) {
	for i, r := range cfg.Routes {
		if !filepath.IsAbs(r) {
			cfg.Routes[i] = filepath.Join(workDir, r)
		}
	}
}

func resolveSites(cfg *config.Config, drupalRoot string, workDir string) {
	if !drush.HasMultiSite(drupalRoot) {
		return
	}

	sites, err := drush.LoadSitesYml(drupalRoot, workDir)
	if err != nil {
		return
	}

	var filtered map[string]drush.Site
	if len(cfg.Sites) > 0 {
		filtered = drush.FilterSites(sites, cfg.Sites, nil)
	} else {
		filtered = sites
	}
	if len(filtered) == 0 {
		return
	}

	var siteList []core.SiteInfo
	for _, s := range filtered {
		siteList = append(siteList, core.SiteInfo{Name: s.Name, URI: s.URI})
	}
	cfg.SetResolvedSites(siteList)

	routeSet := make(map[string]bool, len(cfg.Routes))
	for _, r := range cfg.Routes {
		routeSet[r] = true
	}
	for _, site := range siteList {
		siteRoot := filepath.Join(drupalRoot, "sites", site.Name)
		for _, sub := range []string{"modules", "themes", "profiles", "custom"} {
			dir := filepath.Join(siteRoot, sub)
			if info, err := os.Stat(dir); err == nil && info.IsDir() && !routeSet[dir] {
				routeSet[dir] = true
				cfg.Routes = append(cfg.Routes, dir)
			}
		}
	}
}

func resolveAbs(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return abs
}
