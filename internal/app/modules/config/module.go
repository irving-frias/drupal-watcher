package config

import (
	"context"

	"github.com/irving-frias/drupal-watcher/internal/app"
	"github.com/irving-frias/drupal-watcher/internal/app/common"
)

type Module struct {
	svc ConfigService
}

var _ app.Module = (*Module)(nil)

func (m *Module) Name() string { return "config" }

func (m *Module) DependsOn() []app.Module { return nil }

func (m *Module) Init(container *app.Container) error {
	m.svc = NewConfigService()
	container.Set(common.SvcConfigService, m.svc)
	return nil
}

func (m *Module) Start(ctx context.Context) error { return nil }

func (m *Module) Stop(ctx context.Context) error { return nil }
