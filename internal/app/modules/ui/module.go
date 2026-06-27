package ui

import (
	"context"

	"github.com/irving-frias/drupal-watcher/internal/app"
	"github.com/irving-frias/drupal-watcher/internal/app/common"
	"github.com/irving-frias/drupal-watcher/internal/app/eventbus"
	"github.com/irving-frias/drupal-watcher/internal/app/modules/ui/providers/tui"
)

type Module struct {
	bus *eventbus.EventBus
}

var _ app.Module = (*Module)(nil)

func (m *Module) Name() string { return "ui" }

func (m *Module) DependsOn() []app.Module { return nil }

func (m *Module) Init(container *app.Container) error {
	_ = container.MustGet(common.SvcOrchestrator)
	m.bus = container.MustGet(common.SvcEventBus).(*eventbus.EventBus)
	return nil
}

func (m *Module) Start(ctx context.Context) error {
	return tui.RunWithBus(ctx, m.bus)
}

func (m *Module) Stop(ctx context.Context) error { return nil }
