package watcher

import (
	"context"
	"fmt"

	"github.com/irving-frias/drupal-watcher/internal/app"
	"github.com/irving-frias/drupal-watcher/internal/app/common"
	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/pkg/adapters"
	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type Module struct {
	watcher core.Watcher
}

var _ app.Module = (*Module)(nil)

func (m *Module) Name() string { return "watcher" }

func (m *Module) DependsOn() []app.Module { return nil }

func (m *Module) Init(container *app.Container) error {
	cfg := container.MustGet(common.SvcConfig).(*config.Config)
	skipDirs := append(adapters.DefaultSkipDirs(), cfg.ExcludePatterns...)
	w, err := adapters.NewFSNotifyWatcher(cfg.Routes, skipDirs)
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	m.watcher = w
	container.Set(common.SvcWatcher, m.watcher)
	return nil
}

func (m *Module) Start(ctx context.Context) error { return nil }

func (m *Module) Stop(ctx context.Context) error {
	if m.watcher != nil {
		return m.watcher.Close()
	}
	return nil
}
