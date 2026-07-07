package executor

import (
	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/pkg/adapters"
	"github.com/irving-frias/drupal-watcher/pkg/core"
	"github.com/samber/do/v2"
)

// Register provides core.CommandExecutor to the injector.
func Register(i do.Injector) error {
	cfg := do.MustInvoke[*config.Config](i)
	exec := adapters.NewDrushExecutor(cfg)
	do.ProvideValue(i, core.CommandExecutor(exec))
	return nil
}
