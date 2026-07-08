package ui

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/irving-frias/drupal-watcher/internal/app"
	"github.com/irving-frias/drupal-watcher/internal/app/eventbus"
	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/internal/ui"
	"github.com/irving-frias/drupal-watcher/pkg/core"
	"github.com/samber/do/v2"
)

type engineInfo struct {
	changes atomic.Int64
	clears  atomic.Int64
	start   time.Time
}

func (e *engineInfo) Stats() (int64, int64) {
	return e.changes.Load(), e.clears.Load()
}

func (e *engineInfo) StartTime() time.Time {
	return e.start
}

// Register is a no-op; the UI has no services to provide.
func Register(i do.Injector) error {
	return nil
}

// Run starts the TUI. Call this after all services are resolved.
// It blocks until the user quits.
func Run(ctx context.Context, i do.Injector) error {
	bus := do.MustInvoke[*eventbus.EventBus](i)
	workDir := do.MustInvoke[app.WorkDir](i)
	cfg := do.MustInvoke[*config.Config](i)

	info := &engineInfo{start: time.Now()}
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

	return ui.RunContext(ctx, eventChan, info, string(workDir), cfg.GetShowLogo())
}
