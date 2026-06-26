package builtin

import (
	"context"
	"fmt"

	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type DrushClear struct{}

func NewDrushClear() *DrushClear {
	return &DrushClear{}
}

func (d *DrushClear) Name() string {
	return "DrushClear"
}

func (d *DrushClear) Process(ctx context.Context, event core.FileEvent, result core.ExecutionResult) error {
	if result.ExitCode == 0 {
		return nil
	}
	return fmt.Errorf("drush cache clear failed (exit %d): %s", result.ExitCode, result.Stderr)
}
