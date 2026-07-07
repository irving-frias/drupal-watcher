package core

import "context"

type Watcher interface {
	Start(ctx context.Context) (<-chan FileEvent, <-chan error)
	Add(path string) error
	Remove(path string) error
	Close() error
}

type CommandExecutor interface {
	Execute(ctx context.Context, commands []string, dir string) ExecutionResult
}

type EventFilter interface {
	ShouldProcess(event FileEvent) bool
}

type SiteInfo struct {
	Name string
	URI  string
}
