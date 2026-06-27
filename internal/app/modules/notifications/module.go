package notifications

import (
	"context"

	"github.com/irving-frias/drupal-watcher/internal/app"
	"github.com/irving-frias/drupal-watcher/internal/app/common"
	"github.com/irving-frias/drupal-watcher/internal/app/eventbus"
)

type Module struct {
	svc NotificationService
}

var _ app.Module = (*Module)(nil)

func (m *Module) Name() string { return "notifications" }

func (m *Module) DependsOn() []app.Module { return nil }

func (m *Module) Init(container *app.Container) error {
	bus := container.MustGet(common.SvcEventBus).(*eventbus.EventBus)
	_ = bus
	return nil
}

func (m *Module) Start(ctx context.Context) error { return nil }

func (m *Module) Stop(ctx context.Context) error { return nil }
