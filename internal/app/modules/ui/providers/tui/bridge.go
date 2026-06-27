package tui

import (
	"context"
	"sync"
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
	mu      sync.RWMutex
}

func (e *EngineInfo) Stats() (int64, int64) {
	return e.changes.Load(), e.clears.Load()
}

func (e *EngineInfo) StartTime() time.Time {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.start
}

func (e *EngineInfo) SetChanges(c int64) { e.changes.Store(c) }
func (e *EngineInfo) SetClears(c int64)  { e.clears.Store(c) }
func (e *EngineInfo) SetStartTime(t time.Time) {
	e.mu.Lock()
	e.start = t
	e.mu.Unlock()
}

func RunWithBus(ctx context.Context, bus *eventbus.EventBus) error {
	info := &EngineInfo{start: time.Now()}
	eventChan := make(chan core.EngineEvent, 100)
	defer close(eventChan)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if bus == nil {
			return
		}
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
			info.SetClears(info.clears.Load() + 1)
			select {
			case eventChan <- evt:
			default:
			}
		})
		<-ctx.Done()
	}()
	defer wg.Wait()

	return ui.Run(eventChan, info)
}
