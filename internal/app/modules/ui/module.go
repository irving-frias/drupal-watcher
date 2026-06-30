package ui

import (
	"context"

	"github.com/irving-frias/drupal-watcher/internal/app"
	"github.com/irving-frias/drupal-watcher/internal/app/common"
	"github.com/irving-frias/drupal-watcher/internal/app/eventbus"
	"github.com/irving-frias/drupal-watcher/internal/app/modules/ui/providers/tui"
	"github.com/irving-frias/drupal-watcher/internal/config"
)

type Module struct {
	bus     *eventbus.EventBus
	workDir string
	gifPath string
}

var _ app.Module = (*Module)(nil)

func (m *Module) Name() string { return "ui" }

func (m *Module) DependsOn() []app.Module { return nil }

func (m *Module) Init(container *app.Container) error {
	_ = container.MustGet(common.SvcOrchestrator)
	m.bus = container.MustGet(common.SvcEventBus).(*eventbus.EventBus)
	m.workDir = container.MustGet(common.SvcWorkDir).(string)
	cfg := container.MustGet(common.SvcConfig).(*config.Config)
	m.gifPath = cfg.GIFBackground
	return nil
}


func (m *Module) Start(ctx context.Context) error {
	return tui.RunWithBus(ctx, m.bus, m.workDir, m.gifPath)
}

func (m *Module) Stop(ctx context.Context) error { return nil }
