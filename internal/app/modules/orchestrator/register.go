package orchestrator

import (
	"context"
	"os"
	"strings"

	"github.com/irving-frias/drupal-watcher/internal/app"
	"github.com/irving-frias/drupal-watcher/internal/app/eventbus"
	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/internal/metrics"
	"github.com/irving-frias/drupal-watcher/pkg/adapters"
	"github.com/irving-frias/drupal-watcher/pkg/core"
	"github.com/pterm/pterm"
	"github.com/samber/do/v2"
)

// Register provides *Engine to the injector.
// Engine implements Shutdowner for graceful stop.
func Register(i do.Injector) error {
	cfg := do.MustInvoke[*config.Config](i)
	watcher := do.MustInvoke[core.Watcher](i)
	exec := do.MustInvoke[core.CommandExecutor](i)
	bus := do.MustInvoke[*eventbus.EventBus](i)
	dr := do.MustInvoke[app.DrupalRoot](i)

	filters := []core.EventFilter{
		adapters.NewPatternFilter(cfg.Patterns),
		adapters.NewExcludeFilter(cfg.ExcludePatterns),
		adapters.NewDotfileFilter(),
	}

	lintCheckers := buildLintCheckers(cfg)
	lintCheckers = wrapWithCache(lintCheckers)

	engine := NewEngine(EngineConfig{
		Watcher:            watcher,
		Executor:           exec,
		Filters:            filters,
		LintCheckers:       lintCheckers,
		SkipLint:           cfg.SkipLint,
		EventBus:           bus,
		Logger:             adapters.NewDiscardLogger(),
		Debounce:           cfg.Debounce,
		Patterns:           cfg.Patterns,
		ExcludePatterns:    cfg.ExcludePatterns,
		CommandsPerPattern: cfg.CommandsPerPattern,
		ResolvedSites:      cfg.GetResolvedSites(),
		DrupalRoot:         string(dr),
		Routes:             cfg.Routes,
	})

	pterm.Info.Printfln("Routes: %s", pterm.Cyan(strings.Join(cfg.Routes, ", ")))
	pterm.Info.Printfln("Patterns: %s", pterm.Cyan(strings.Join(cfg.Patterns, ", ")))
	pterm.Success.Printfln("Modular watcher PID %d.", os.Getpid())

	do.ProvideValue(i, engine)
	return nil
}

// Run starts the engine in a goroutine and blocks until ctx is done.
// Call this after all services are resolved.
func Run(i do.Injector, ctx context.Context) error {
	engine := do.MustInvoke[*Engine](i)
	metrics.Init()

	go func() {
		if err := engine.Run(ctx); err != nil && err != context.Canceled {
			pterm.Error.Printfln("Engine: %v", err)
		}
	}()
	<-ctx.Done()
	return ctx.Err()
}

func buildLintCheckers(cfg *config.Config) map[string]core.LintChecker {
	if cfg.SkipLint {
		return nil
	}
	m := make(map[string]core.LintChecker)
	exts := cfg.GetLintCommands()
	if exts == nil {
		return nil
	}

	phpcsStd := cfg.GetPhpCsStandard()

	for ext := range exts {
		switch ext {
		case ".php":
			if phpcsStd != "" {
				m[ext] = adapters.NewPhpCsLintChecker(phpcsStd)
			} else {
				m[ext] = adapters.NewPhpLintChecker()
			}
		case ".yml", ".yaml":
			m[ext] = adapters.NewYamlLintChecker()
		}
	}

	phpLike := map[string]bool{".php": true, ".module": true, ".inc": true, ".theme": true, ".install": true, ".profile": true, ".engine": true}

	if phpcsStd != "" {
		for _, ext := range []string{".css", ".js"} {
			if _, ok := m[ext]; !ok {
				m[ext] = adapters.NewPhpCsLintChecker(phpcsStd)
			}
		}
	}

	for _, p := range cfg.Patterns {
		if phpLike[p] {
			if _, ok := m[p]; !ok {
				if phpcsStd != "" {
					m[p] = adapters.NewPhpCsLintChecker(phpcsStd)
				} else {
					m[p] = adapters.NewPhpLintChecker()
				}
			}
		}
	}

	return m
}

func wrapWithCache(checkers map[string]core.LintChecker) map[string]core.LintChecker {
	wrapped := make(map[string]core.LintChecker, len(checkers))
	for ext, chk := range checkers {
		wrapped[ext] = adapters.NewCachingLintChecker(chk)
	}
	return wrapped
}
