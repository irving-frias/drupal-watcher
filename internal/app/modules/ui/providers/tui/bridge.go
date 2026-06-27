package tui

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/irving-frias/drupal-watcher/internal/app/eventbus"
	"github.com/irving-frias/drupal-watcher/internal/ui"
	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type EngineInfo struct {
	changes atomic.Int64
	clears  atomic.Int64
	start   time.Time
}

func (e *EngineInfo) Stats() (int64, int64) {
	return e.changes.Load(), e.clears.Load()
}

func (e *EngineInfo) StartTime() time.Time {
	return e.start
}

func RunWithBus(ctx context.Context, bus *eventbus.EventBus) error {
	info := &EngineInfo{start: time.Now()}
	eventChan := make(chan core.EngineEvent, 100)

	if bus != nil {
		bus.Subscribe(eventbus.TopicFileChange, func(event any) {
			evt, ok := event.(core.EngineEvent)
			if !ok {
				return
			}
			select {
			case eventChan <- evt:
			default:
			}
		})
		bus.Subscribe(eventbus.TopicCacheClear, func(event any) {
			evt, ok := event.(core.EngineEvent)
			if !ok {
				return
			}
			info.clears.Add(1)
			select {
			case eventChan <- evt:
			default:
			}
		})
		bus.Subscribe(eventbus.TopicError, func(event any) {
			evt, ok := event.(core.EngineEvent)
			if !ok {
				return
			}
			select {
			case eventChan <- evt:
			default:
			}
		})
	}

	return ui.RunContext(ctx, eventChan, info)
}
