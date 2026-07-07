package app

import (
	"context"

	"github.com/irving-frias/drupal-watcher/internal/app/eventbus"
	"github.com/samber/do/v2"
)

// RegisterFn registers services in the injector.
type RegisterFn func(i do.Injector) error

// Setup creates and wires the injector, returning it ready for use.
func Setup(fns ...RegisterFn) (do.Injector, error) {
	i := do.New()

	// EventBus is always provided first — other modules depend on it.
	bus := eventbus.New()
	do.ProvideValue(i, bus)

	for _, fn := range fns {
		if err := fn(i); err != nil {
			return nil, err
		}
	}
	return i, nil
}

// Shutdown gracefully stops all services that implement Shutdowner.
func Shutdown(i do.Injector, ctx context.Context) error {
	return i.ShutdownWithContext(ctx)
}
