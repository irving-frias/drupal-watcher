# Drupal Watcher — Project Map

> Auto-generated dependency and structure map.
> Regenerate with: `go run ./scripts/project_map.go` (if available) or manually review.

## Entry Point

```
cmd/drupal-watcher/main.go
  └─ main() → crea App con 5 módulos:
       config → watcher → executor → orchestrator → ui
     Escribe PID, corre health check, arranca App
```

## Package Dependency Graph

```
cmd/drupal-watcher (main)
  ├── internal/app            — Module system, Container, App lifecycle
  ├── internal/app/common     — ServiceName, PkgVersion
  ├── internal/app/modules/config     → app, config, drush, core
  ├── internal/app/modules/watcher    → app, config, adapters, core
  ├── internal/app/modules/executor   → app, config, adapters, core
  ├── internal/app/modules/orchestrator → app, eventbus, config, hooks, metrics, adapters, core
  ├── internal/app/modules/ui         → app, eventbus, tui bridge, ui
  ├── internal/config         → utils, core
  ├── internal/health         → (stdlib)
  └── internal/validate       → config

internal/
  ├── app/                    — Module system
  │   ├── app.go              → App.Start/Stop/Done, ordena módulos por dependencia
  │   ├── container.go        → Container.Set/Get/MustGet (DI)
  │   ├── module.go           → Module interface
  │   ├── common/
  │   │   ├── types.go        → ServiceName constants (SvcEventBus, SvcConfig, etc.)
  │   │   └── version.go      → PkgVersion()
  │   ├── eventbus/
  │   │   └── bus.go          → EventBus (pub/sub): New, Subscribe, Publish
  │   │                        Topics: file.change, cache.clear, error, config.update, engine.start|stop
  │   └── modules/
  │       ├── config/         → Carga config.yaml desde container, registra Manager
  │       ├── watcher/        → Crea FSNotifyWatcher desde config, registra en container
  │       ├── executor/       → Crea DrushExecutor desde config, registra en container
  │       ├── orchestrator/   → Engine (event loop), module lifecycle
  │       └── ui/             → Bridge (EventBus → TUI channel)
  │           └── providers/tui/bridge.go
  │
  ├── config/                 — Config Manager, PID/starttime files, Drupal root detection
  ├── drush/                  — Drush command resolution, execution, site alias loading
  ├── health/                 — Health check file writer (cada 30s)
  ├── hooks/
  │   ├── builtin/drush_clear.go
  │   └── examples/slack.go
  ├── metrics/                — Runtime stats: changes, clears, errors per minute
  ├── training/               — Training suggestions (training.json)
  ├── ui/                     — Bubble Tea TUI (model, view, update, styles, messages, powermode)
  ├── utils/                  — Colors (pterm), formatting, section printing
  ├── validate/               — Config validation
  └── xdebug/                 — Xdebug detection + disable

pkg/
  ├── core/                   — Domain interfaces & models (cero dependencias internas)
  │   ├── interfaces.go       → Watcher, CommandExecutor, EventFilter, PostProcessor, EngineConfig, SiteInfo
  │   ├── models.go           → FileEvent, ExecutionResult, EngineEvent (EventChange|CacheClear|Error)
  │   └── lint.go             → LintChecker interface, LintResult
  └── adapters/               — Implementaciones concretas
      ├── fsnotify_watcher.go → FSNotifyWatcher (core.Watcher)
      ├── polling_watcher.go  → PollingWatcher (core.Watcher)
      ├── hybrid_watcher.go   → HybridWatcher (fsnotify + polling dedup)
      ├── drush_executor.go   → DrushExecutor, SiteAwareDrushExecutor
      ├── regex_filter.go     → PatternFilter, ExcludeFilter, DotfileFilter
      ├── php_lint.go         → PhpLintChecker (php -l)
      ├── phpcs_lint.go       → PhpCsLintChecker (phpcs Drupal standards)
      ├── yaml_lint.go        → YamlLintChecker
      ├── lint_cache.go       → CachingLintChecker (SHA-1, 5min TTL)
      └── slog_logger.go      → SlogLogger factory
```

## Module System

```
app.Module (interface)
  ├── Name() string
  ├── DependsOn() []Module
  ├── Init(container *Container) error   → registra servicios en container
  ├── Start(ctx) error                    → arranca goroutines
  └── Stop(ctx) error                     → cleanup

Service Names (common.ServiceName)
  SvcEventBus     → *eventbus.EventBus
  SvcConfig       → *config.Manager
  SvcWorkDir      → string (project root)
  SvcDrupalRoot   → string (abs Drupal root)
  SvcWatcher      → core.Watcher (FSNotifyWatcher)
  SvcExecutor     → core.CommandExecutor (DrushExecutor)
  SvcOrchestrator → *orchestrator.Engine

App order: config → watcher → executor → orchestrator → ui (ui bloquea)
```

## Engine Event Loop (orchestrator/engine.go)

```
Watcher.Start(ctx)
  ├── FileEvent channel (fsnotify/poll/hybrid)
  └── Error channel
       │
       ▼
  EventFilter.ShouldProcess(event) → filtra por patrón/exclusión/dotfile
       │
       ▼
  Debounce timer (800ms) → acumula cambios
       │
       ▼
  Lint check → cada archivo .php → php -l, .yml → yaml parser
       │
       ▼
  CommandsPerPattern → mapa extensión → drush command
       │
       ▼
  CommandExecutor.Execute() → drush
       │
       ▼
  PostProcessor.Process() → hooks
       │
       ▼
  EventBus.Publish(TopicCacheClear, EngineEvent)
```

## Data Flow

```
[Watcher] ──FileEvent──▶ [Engine] ──EngineEvent──▶ [EventBus]
                                                       │
                                          ┌────────────┼────────────┐
                                          ▼            ▼            ▼
                                     [TUI]      [notifications]  [future modules]
```

## Key Interfaces (pkg/core/)

```
Watcher           → Start(ctx) (<-chan FileEvent, <-chan error), Add/Remove/Close
CommandExecutor   → Execute(ctx, []string, dir) ExecutionResult
EventFilter       → ShouldProcess(FileEvent) bool
PostProcessor     → Name() string, Process(ctx, FileEvent, ExecutionResult) error
LintChecker       → Lint(filePath) *LintResult
```

## Key Events (pkg/core/)

```
EngineEvent
  ├── Type:      EventChange | EventCacheClear | EventError
  ├── File:      string (path del archivo modificado)
  ├── Changes:   int (cantidad de cambios en el batch)
  ├── Commands:  string (drush commands ejecutados)
  ├── ExitCode:  int
  ├── Duration:  time.Duration
  ├── SiteName:  string (multi-site)
  └── Timestamp: time.Time
```

## EventBus Topics

```
file.change     → EngineEvent (cada cambio detectado)
cache.clear     → EngineEvent (después de drush)
error           → EngineEvent (error de lint/watcher)
config.update   → (config reload)
engine.start    → (orchestrator arrancó)
engine.stop     → (orchestrator terminó)
```

## TUI (internal/ui/) — Bubble Tea

```
Model (struct)
  ├── eventChan     ← core.EngineEvent (desde bridge)
  ├── engineInfo    → Stats(), StartTime()
  ├── status        → PID, Uptime, Changes, Clears, AllocMB
  ├── events []eventLine (ring buffer 100)
  ├── viewport      → scrollable event log
  ├── input         → command input
  ├── siteFilter / siteClears → multi-site filtering
  ├── powerMode     → *PowerMode (animaciones)
  └── ... (help, star, xdebug, training, etc.)

Message types:
  tickMsg           → cada 1s (refresh status, animaciones)
  engineEventMsg    → evento del watcher
  fsCompleteMsg     → filesystem scan completado
  powerPulseMsg     → pulso de PowerMode

Init():
  tea.Batch(tickCmd, listenForEvents, textinput.Blink)

Update():
  WindowSizeMsg → layout recalculation
  KeyMsg        → input, history, completions, f2, f4, ?, r, ctrl+x, etc.
  engineEventMsg → pushEvent + powerMode.Punch(changes)
  tickMsg       → updateStatus + powerMode.Tick() + render

View():
  JoinVertical(status, events, star?, xdebug?, input)
  PowerMode modify border colors dinámicamente
```

## PowerMode System (powermode.go)

```
PowerLevel: Normal(0-2) → Warm(3-5) → Hot(6-10) → Power(11+)

Particles:
  ├── ParticleSpark  → ✦ ✧ ⚡ ★ ♦  (rápidas, corta vida)
  ├── ParticleFire   → 🔥 💥 ⚡     (ondulantes, media vida)
  └── ParticleSmoke  → · ‧ ∘ ° ≈  (lentas, larga vida)

Cooldown ❄:
  ─ Sin eventos 3s → border azul, icono ❄
  ─ Humo sube desde abajo (más si nivel era alto)
  ─ 30 ticks de desvanecimiento

Skull 💀:
  ─ 50+ cambios en un batch → skull flota, border blanco/rojo
  ─ 100+ → 2 skulls, 200+ → 3 skulls
  ─ PowerMode maxea energía, explota partículas

Events:
  Punch(changes) → combo++, energía, partículas, skull check
  Tick()         → decaimiento combo/energía, cooldown, animación partículas
```

## Adapter Implementations

```
Watcher:
  FSNotifyWatcher   → fsnotify (natal)
  PollingWatcher    → filepath.Walk cada N ms
  HybridWatcher     → ambos + dedup 1s

Executor:
  DrushExecutor         → drush simple
  SiteAwareDrushExecutor → drush multi-site con --uri

LintChecker:
  PhpLintChecker     → php -l
  PhpCsLintChecker   → phpcs (Drupal standards)
  YamlLintChecker    → yaml.Parse
  CachingLintChecker → wrapper con SHA-1 + 5min TTL

Filter:
  PatternFilter   → match por extensión
  ExcludeFilter   → exclude por substring
  DotfileFilter   → exclude .*
```

## Exported Functions Index

| Package | Function | Purpose |
|---------|----------|---------|
| config | `NewManager()` | Config manager |
| config | `WritePid/RemovePid/CheckPid` | PID lifecycle |
| config | `WriteStarttime/RemoveStarttime/GetStarttime` | Start time tracking |
| drush | `GetCmd/ResetCmdCache` | Drush binary resolution |
| drush | `Run/RunCacheClears/RunPostClearCommands` | Drush execution |
| drush | `HealthCheck/PromptConfirm/NotifyOS` | Drush utilities |
| drush | `LoadSitesYml/HasMultiSite/FilterSites` | Site alias management |
| app | `New/Start/Stop/Done` | App lifecycle |
| app | `NewContainer/Set/Get/MustGet` | DI container |
| eventbus | `New/Subscribe/Publish` | Event bus |
| ui | `NewModel/Run/RunContext` | TUI lifecycle |
| ui | `NewPowerMode` | PowerMode |
| core | various interfaces | Domain contracts |
| adapters | various `New*` | Implementation factories |

## Test Packages

```
internal/app/modules/orchestrator/ — engine tests
internal/config/                  — config load/save, PID
internal/drush/                   — command execution
internal/health/                  — health check
internal/metrics/                 — stats tracking
internal/training/                — suggestion loading
internal/ui/                      — model, powermode
internal/utils/                   — formatting
internal/validate/                — config validation
pkg/adapters/                     — watchers, lint, filters
```

## Configuration (configs/config.yaml)

```
routes, patterns, debounce, commandsPerPattern, skipLint,
lintCommands, phpCsStandard, watchMode (auto|fsnotify|poll|hybrid),
pollInterval, eventBufferSize
```
