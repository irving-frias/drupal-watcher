package common

type ServiceName string

const (
	SvcEventBus     ServiceName = "eventbus"
	SvcConfig       ServiceName = "config"
	SvcWorkDir      ServiceName = "workdir"
	SvcDrupalRoot   ServiceName = "drupal.root"
	SvcWatcher      ServiceName = "watcher"
	SvcExecutor     ServiceName = "executor"
	SvcOrchestrator ServiceName = "orchestrator"
)
