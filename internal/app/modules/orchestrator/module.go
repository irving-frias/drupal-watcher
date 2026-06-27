package orchestrator

import (
	"context"
	"os"

	"github.com/irving-frias/drupal-watcher/internal/app"
	"github.com/irving-frias/drupal-watcher/internal/app/common"
	"github.com/irving-frias/drupal-watcher/internal/app/eventbus"
	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/internal/hooks/builtin"
	"github.com/irving-frias/drupal-watcher/pkg/adapters"
	"github.com/irving-frias/drupal-watcher/pkg/core"
	"github.com/pterm/pterm"
)

type Module struct {
	engine *Engine
}

var _ app.Module = (*Module)(nil)

func (m *Module) Name() string { return "orchestrator" }

func (m *Module) DependsOn() []app.Module { return nil }

func (m *Module) Init(container *app.Container) error {
	cfg := container.MustGet(common.SvcConfig).(*config.Config)
	watcher := container.MustGet(common.SvcWatcher).(core.Watcher)
	exec := container.MustGet(common.SvcExecutor).(core.CommandExecutor)
	bus := container.MustGet(common.SvcEventBus).(*eventbus.EventBus)
	dr := container.MustGet(common.SvcDrupalRoot).(string)

	filters := []core.EventFilter{
		adapters.NewPatternFilter(cfg.Patterns),
		adapters.NewExcludeFilter(cfg.ExcludePatterns),
		adapters.NewDotfileFilter(),
	}

	lintCheckers := buildLintCheckers(cfg)

	m.engine = NewEngine(EngineConfig{
		Watcher:            watcher,
		Executor:           exec,
		Filters:            filters,
		PostProcessors:     []core.PostProcessor{&builtin.DrushClear{}},
		LintCheckers:       lintCheckers,
		SkipLint:           cfg.SkipLint,
		EventBus:           bus,
		Logger:             adapters.NewDiscardLogger(),
		Debounce:           cfg.Debounce,
		Patterns:           cfg.Patterns,
		ExcludePatterns:    cfg.ExcludePatterns,
		CommandsPerPattern: cfg.CommandsPerPattern,
		ResolvedSites:      cfg.GetResolvedSites(),
		DrupalRoot:         dr,
		Routes:             cfg.Routes,
	})

	container.Set(common.SvcOrchestrator, m.engine)

	pterm.Info.Printfln("Routes: %s", pterm.Cyan(stringJoin(cfg.Routes, ", ")))
	pterm.Info.Printfln("Patterns: %s", pterm.Cyan(stringJoin(cfg.Patterns, ", ")))
	pterm.Success.Printfln("Modular watcher PID %d.", os.Getpid())
	return nil
}

func (m *Module) Start(ctx context.Context) error {
	go func() {
		if err := m.engine.Run(ctx); err != nil && err != context.Canceled {
			pterm.Error.Printfln("Engine: %v", err)
		}
	}()
	return nil
}

func (m *Module) Stop(ctx context.Context) error { return nil }

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

func stringJoin(elems []string, sep string) string {
	if len(elems) == 0 {
		return ""
	}
	result := elems[0]
	for _, e := range elems[1:] {
		result += sep + e
	}
	return result
}
