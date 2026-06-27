package watcher

import (
	"context"

	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type WatcherService interface {
	Start(ctx context.Context) (<-chan core.FileEvent, <-chan error)
	Add(path string) error
	Remove(path string) error
	Close() error
}

type watcherService struct {
	inner core.Watcher
}

func NewWatcherService(inner core.Watcher) WatcherService {
	return &watcherService{inner: inner}
}

func (s *watcherService) Start(ctx context.Context) (<-chan core.FileEvent, <-chan error) {
	return s.inner.Start(ctx)
}

func (s *watcherService) Add(path string) error {
	return s.inner.Add(path)
}

func (s *watcherService) Remove(path string) error {
	return s.inner.Remove(path)
}

func (s *watcherService) Close() error {
	return s.inner.Close()
}
