package executor

import (
	"context"

	"github.com/irving-frias/drupal-watcher/internal/app"
	"github.com/irving-frias/drupal-watcher/internal/app/common"
	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/pkg/adapters"
	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type Module struct {
	exec core.CommandExecutor
}

var _ app.Module = (*Module)(nil)

func (m *Module) Name() string { return "executor" }

func (m *Module) DependsOn() []app.Module { return nil }

func (m *Module) Init(container *app.Container) error {
	cfg := container.MustGet(common.SvcConfig).(*config.Config)
	m.exec = adapters.NewDrushExecutor(cfg)
	container.Set(common.SvcExecutor, m.exec)
	return nil
}

func (m *Module) Start(ctx context.Context) error { return nil }

func (m *Module) Stop(ctx context.Context) error { return nil }
