package common

import "context"

type ServiceName string

const (
	SvcEventBus      ServiceName = "eventbus"
	SvcConfigService ServiceName = "config.service"
	SvcWatcherService ServiceName = "watcher.service"
	SvcExecutorService ServiceName = "executor.service"
	SvcOrchestrator  ServiceName = "orchestrator"
	SvcUIService     ServiceName = "ui.service"
	SvcNotification  ServiceName = "notification.service"
)

type Lifecycle interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Initializer interface {
	Init(ctx context.Context, container ServiceRegistry) error
}

type ServiceRegistry interface {
	Set(name ServiceName, svc any)
	Get(name ServiceName) any
	MustGet(name ServiceName) any
}
