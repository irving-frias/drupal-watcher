package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/irving-frias/drupal-watcher/internal/app/common"
	"github.com/irving-frias/drupal-watcher/internal/app/eventbus"
)

type App struct {
	container *Container
	modules   []Module
	cancel    context.CancelFunc
	done      chan struct{}
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
	for _, m := range ordered {
		if err := m.Start(ctx); err != nil {
			return fmt.Errorf("module %s start: %w", m.Name(), err)
		}
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	if a.cancel != nil {
		a.cancel()
	}

	var lastErr error
	for i := len(a.modules) - 1; i >= 0; i-- {
		if err := a.modules[i].Stop(ctx); err != nil {
			lastErr = fmt.Errorf("module %s stop: %w", a.modules[i].Name(), err)
		}
	}

	close(a.done)
	return lastErr
}

func (a *App) Done() <-chan struct{} {
	return a.done
}

func (a *App) WaitForSignal(ctx context.Context) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		select {
		case <-sigCh:
			a.Stop(context.Background())
		case <-ctx.Done():
		case <-a.done:
		}
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
