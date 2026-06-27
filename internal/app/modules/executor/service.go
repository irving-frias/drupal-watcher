package executor

import (
	"context"

	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type ExecutorService interface {
	Execute(ctx context.Context, commands []string, dir string) core.ExecutionResult
}

type executorService struct {
	core.CommandExecutor
}

func NewExecutorService(exec core.CommandExecutor) ExecutorService {
	return &executorService{exec}
}
