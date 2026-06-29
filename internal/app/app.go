package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/irving-frias/drupal-watcher/internal/app/common"
	"github.com/irving-frias/drupal-watcher/internal/app/eventbus"
)

const shutdownTimeout = 10 * time.Second

type App struct {
	container *Container
	modules   []Module
	cancel    context.CancelFunc
	done      chan struct{}
	stopOnce  sync.Once
}

func New(modules ...Module) *App {
	return &App{
		container: NewContainer(),
		modules:   modules,
		done:      make(chan struct{}),
	}
}

func (a *App) Container() *Container {
	return a.container
}

func (a *App) Start(ctx context.Context) error {
	bus := eventbus.New()
	a.container.Set(common.SvcEventBus, bus)

	ordered := sortModules(a.modules)
	for _, m := range ordered {
		if err := m.Init(a.container); err != nil {
			return fmt.Errorf("module %s init: %w", m.Name(), err)
		}
	}

	ctx, a.cancel = context.WithCancel(ctx)
	a.setupSignalHandler()

	for _, m := range ordered {
		if err := m.Start(ctx); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("module %s start: %w", m.Name(), err)
		}
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	var stopErr error
	a.stopOnce.Do(func() {
		if a.cancel != nil {
			a.cancel()
		}

		shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
		defer cancel()

		for i := len(a.modules) - 1; i >= 0; i-- {
			if err := a.modules[i].Stop(shutdownCtx); err != nil {
				if shutdownCtx.Err() != nil {
					stopErr = fmt.Errorf("shutdown timed out after %s", shutdownTimeout)
					break
				}
				stopErr = fmt.Errorf("module %s stop: %w", a.modules[i].Name(), err)
			}
		}

		close(a.done)
	})
	return stopErr
}

func (a *App) Done() <-chan struct{} {
	return a.done
}

func (a *App) setupSignalHandler() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigCh
		a.cancel()
	}()
}

func sortModules(modules []Module) []Module {
	sorted := make([]Module, len(modules))
	copy(sorted, modules)

	sort.SliceStable(sorted, func(i, j int) bool {
		return dependsOn(sorted[j], sorted[i])
	})

	return sorted
}

func dependsOn(a, b Module) bool {
	for _, dep := range a.DependsOn() {
		if dep == b {
			return true
		}
	}
	return false
}
