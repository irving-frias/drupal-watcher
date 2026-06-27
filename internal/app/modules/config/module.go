package config

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/irving-frias/drupal-watcher/internal/app"
	"github.com/irving-frias/drupal-watcher/internal/app/common"
	"github.com/irving-frias/drupal-watcher/internal/config"
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

	dr := m.WorkDir
	if m.cfg.DrupalRoot != nil {
		dr = filepath.Join(m.WorkDir, *m.cfg.DrupalRoot)
	}
	m.DrupalRoot = dr
	container.Set(common.SvcDrupalRoot, dr)
	return nil
}

func (m *Module) Start(ctx context.Context) error { return nil }

func (m *Module) Stop(ctx context.Context) error { return nil }

func (m *Module) GetConfig() *config.Config { return m.cfg }
