package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/irving-frias/drupal-watcher/internal/app"
	"github.com/irving-frias/drupal-watcher/internal/app/common"
	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/internal/drush"
	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type Module struct {
	WorkDir    string
	DrupalRoot string
	cfg        *config.Config
}

var _ app.Module = (*Module)(nil)

func (m *Module) Name() string { return "config" }

func (m *Module) DependsOn() []app.Module { return nil }

func (m *Module) Init(container *app.Container) error {
	mgr := config.NewManager()
	c, err := mgr.LoadConfig(m.WorkDir)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	m.cfg = &c
	container.Set(common.SvcConfig, m.cfg)
	container.Set(common.SvcWorkDir, m.WorkDir)

	dr := resolveAbs(m.WorkDir)
	if m.cfg.DrupalRoot != nil {
		dr = filepath.Join(dr, *m.cfg.DrupalRoot)
	}
	m.DrupalRoot = dr
	container.Set(common.SvcDrupalRoot, dr)

	m.normalizeRoutes()
	m.resolveSites(dr)

	return nil
}

func (m *Module) Start(ctx context.Context) error { return nil }

func (m *Module) Stop(ctx context.Context) error { return nil }

func (m *Module) resolveSites(drupalRoot string) {
	if !drush.HasMultiSite(drupalRoot) {
		return
	}

	sites, err := drush.LoadSitesYml(drupalRoot, m.WorkDir)
	if err != nil {
		return
	}

	var filtered map[string]drush.Site
	if len(m.cfg.Sites) > 0 {
		filtered = drush.FilterSites(sites, m.cfg.Sites, nil)
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
	m.cfg.SetResolvedSites(siteList)

	routeSet := make(map[string]bool, len(m.cfg.Routes))
	for _, r := range m.cfg.Routes {
		routeSet[r] = true
	}
	for _, site := range siteList {
		siteRoot := filepath.Join(drupalRoot, "sites", site.Name)
		for _, sub := range []string{"modules", "themes", "profiles", "custom"} {
			dir := filepath.Join(siteRoot, sub)
			if info, err := os.Stat(dir); err == nil && info.IsDir() && !routeSet[dir] {
				routeSet[dir] = true
				m.cfg.Routes = append(m.cfg.Routes, dir)
			}
		}
	}
}

func (m *Module) GetConfig() *config.Config { return m.cfg }

func resolveAbs(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return abs
}

func (m *Module) normalizeRoutes() {
	for i, r := range m.cfg.Routes {
		m.cfg.Routes[i] = resolveAbs(r)
	}
}
