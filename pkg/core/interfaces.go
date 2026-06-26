package core

import (
	"context"
	"log/slog"
)

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

type PostProcessor interface {
	Name() string
	Process(ctx context.Context, event FileEvent, result ExecutionResult) error
}

type EngineConfig struct {
	Watcher            Watcher
	Executor           CommandExecutor
	SiteExecutorFactory func(site SiteInfo) CommandExecutor
	Filters            []EventFilter
	PostProcessors     []PostProcessor
	EventChan          chan<- EngineEvent
	Logger             *slog.Logger
	Debounce           int
	Patterns           []string
	ExcludePatterns    []string
	CommandsPerPattern map[string]string
	ResolvedSites      []SiteInfo
	DrupalRoot         string
}

type SiteInfo struct {
	Name string
	URI  string
}
