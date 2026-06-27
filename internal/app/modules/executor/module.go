package executor

import (
	"context"

	"github.com/irving-frias/drupal-watcher/internal/app"
	"github.com/irving-frias/drupal-watcher/internal/app/common"
)

type Module struct {
	svc ExecutorService
}

var _ app.Module = (*Module)(nil)

func (m *Module) Name() string { return "executor" }

func (m *Module) DependsOn() []app.Module { return nil }

func (m *Module) Init(container *app.Container) error {
	_ = container.MustGet(common.SvcConfigService)
	return nil
}

func (m *Module) Start(ctx context.Context) error { return nil }

func (m *Module) Stop(ctx context.Context) error { return nil }
