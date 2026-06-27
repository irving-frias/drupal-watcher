package watcher

import (
	"context"

	"github.com/irving-frias/drupal-watcher/internal/app"
	"github.com/irving-frias/drupal-watcher/internal/app/common"
)

type Module struct {
	svc WatcherService
}

var _ app.Module = (*Module)(nil)

func (m *Module) Name() string { return "watcher" }

func (m *Module) DependsOn() []app.Module { return nil }

func (m *Module) Init(container *app.Container) error {
	cfgSvc := container.MustGet(common.SvcConfigService).(ConfigProvider)
	_ = cfgSvc
	return nil
}

func (m *Module) Start(ctx context.Context) error { return nil }

func (m *Module) Stop(ctx context.Context) error { return nil }
