package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/irving-frias/drupal-watcher/internal/app/common"
	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/internal/drush"
	"github.com/irving-frias/drupal-watcher/pkg/core"
	"github.com/samber/do/v2"
)

// Register provides Config, WorkDir, and DrupalRoot to the injector.
func Register(root string) func(i do.Injector) error {
	return func(i do.Injector) error {
		mgr := config.NewManager()
		c, err := mgr.LoadConfig(root)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		cfg := &c

		workDir := resolveAbs(root)
		do.ProvideValue(i, common.WorkDir(workDir))

		dr := workDir
		if cfg.DrupalRoot != nil {
			dr = filepath.Join(workDir, *cfg.DrupalRoot)
		}
		do.ProvideValue(i, common.DrupalRoot(dr))

		normalizeRoutes(cfg, workDir)
		resolveSites(cfg, dr, root)

		do.ProvideValue(i, cfg)
		return nil
	}
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
