package ui

import (
	"context"

	"github.com/irving-frias/drupal-watcher/internal/app/common"
	"github.com/irving-frias/drupal-watcher/internal/app/eventbus"
	"github.com/irving-frias/drupal-watcher/internal/app/modules/ui/providers/tui"
	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/samber/do/v2"
)

// Register is a no-op; the UI has no services to provide.
func Register(i do.Injector) error {
	return nil
}

// Run starts the TUI. Call this after all services are resolved.
// It blocks until the user quits.
func Run(ctx context.Context, i do.Injector) error {
	bus := do.MustInvoke[*eventbus.EventBus](i)
	workDir := do.MustInvoke[common.WorkDir](i)
	cfg := do.MustInvoke[*config.Config](i)
	return tui.RunWithBus(ctx, bus, string(workDir), cfg.GetShowLogo())
}
