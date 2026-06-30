# Drupal Watcher

File watcher for Drupal development. Monitors custom modules and themes and auto-runs `drush` cache clears whenever a file changes. Works with DDEV, Lando, and any local environment.

## Requirements

- **PHP 8.1+** (for Composer-based install)
- **Drush** installed in your Drupal project

> Go is **not required**. The binary is auto-downloaded during `composer install`.

> **Development tool only.** Install with `--dev` to exclude from production deployments via `composer install --no-dev`.

## Installation

```bash
composer require --dev irving-frias/drupal-watcher
vendor/bin/drupal-watcher help
```

On first run, the shell launcher downloads the correct binary for your OS/architecture from GitHub Releases. No compilation needed.

> If `vendor/bin/drupal-watcher` doesn't exist (e.g. on Windows), use the full path: `vendor/irving-frias/drupal-watcher/bin/drupal-watcher`

### Manual download

Download the binary for your platform from [GitHub Releases](https://github.com/irving-frias/drupal-watcher/releases), make it executable, and run it from your Drupal project root.

## Quick Start

```bash
cd /path/to/drupal/project
vendor/bin/drupal-watcher start
```

On first run, a `configs/config.yaml` is auto-created with defaults. Edit it to customize routes, patterns, and cache clear commands. You can also override any config value via environment variables (e.g. `DRUPAL_WATCHER_DEBOUNCE=150`).

You can validate your configuration and environment with:

```bash
vendor/bin/drupal-watcher validate
```

This checks YAML syntax, Drupal root, watched routes, drush, PHPCS, sites, and `commandsPerPattern`.

The TUI opens automatically. Events appear in real-time, and you can type commands at the prompt:

```
  ● drupal-watcher  PID: 12345  Uptime: 5m
  Memory: 2.1 MB ▂▃▄▅▆▇█  |  Changes: 14  |  Clears: 3  |  ⚡ x5  ▓▓▓▓░░░░  |  site1: 2  site2: 1

  ┌──────────────────────────────────────────────────────────────┐
  │ 10:00:01  ℹ  Waiting for file changes...                     │
  │ 10:02:15  ℹ  Change detected: docroot/modules/custom/foo.module │
  │ 10:02:16  ✔  drush cc plugin (312ms, exit 0)                │
  │ 10:03:22  ✖  Error in docroot/modules/custom/bad.php:       │
  │            PHP Parse error: syntax error, unexpected ...     │
  └──────────────────────────────────────────────────────────────┘

  ┌──────────────────────────────────────────────────────────────┐
  > help                                                         │
  └──────────────────────────────────────────────────────────────┘
```

Use `--no-tui` to run the classic interactive CLI instead.

## Commands

| Command                    | Description                            |
|----------------------------|----------------------------------------|
| `start`                    | Start watching (opens TUI by default)  |
| `validate`                 | Validate config, paths, drush, PHPCS   |
| `tui`                      | Terminal UI (default for `start`)      |
| `status`                   | Show running status and uptime         |
| `list` / `config`          | Display current configuration          |
| `add` <route> [pattern]    | Add route and/or pattern to watch      |
| `remove` / `rm` <route>    | Remove route and/or pattern            |
| `restart`                  | Restart the watcher                    |
| `stop` / `reset`           | Stop the watcher and clear PID         |
| `version`                  | Show version and Go runtime            |
| `help`                     | Show usage information                 |

## Options

| Flag                    | Description                                   |
|-------------------------|-----------------------------------------------|
| `--root <path>`         | Drupal root directory (default: cwd)          |
| `--debounce <ms>`       | Debounce interval (default: 800ms)            |
| `--no-dotfiles`         | Exclude dotfiles from watching                |
| `--no-tui`              | Disable TUI, use interactive CLI mode         |
| `--notify`              | Send desktop notification on cache clear      |
| `--log-file <path>`     | Write logs to file                            |
| `--config <path>`       | Custom config file path                       |
| `--commands-per-pattern <json>` | Override per-pattern commands        |
| `--site <names>`        | Only watch specified sites (comma-separated)  |
| `--exclude-site <names>` | Watch all sites except specified              |
| `--uri <uri>`           | Single-site mode with explicit URI            |
| `--help` / `-h`         | Show help                                     |
| `--version` / `-V`      | Show version and Go runtime                   |

### --notify

Sends a native OS desktop notification each time a cache clear completes:

```bash
vendor/bin/drupal-watcher start --notify
```

Uses the `beeep` Go library for cross-platform desktop notifications. No OS-specific configuration required — works on macOS (via `osascript`), Linux (via `notify-send` or D-Bus), and Windows (via Toast notifications).

### --root

Point the watcher at a different Drupal root:

```bash
vendor/bin/drupal-watcher start --root /var/www/html
vendor/bin/drupal-watcher status --root /var/www/html
```

### --commands-per-pattern

Override per-pattern drush commands without editing the config file:

```bash
vendor/bin/drupal-watcher start --commands-per-pattern '{"css":"cc css-js","js":"cc css-js"}'
```

Accepts a JSON object mapping file extensions to drush commands.

## TUI Commands

While the TUI is running, type commands at the prompt:

| Command                    | Description                            |
|----------------------------|----------------------------------------|
| Command                    | Description                            |
|----------------------------|----------------------------------------|
| `status`                   | Show stats, memory, and kernel watches |
| `stats`                    | Clear counts per site                  |
| `filter <site>`            | Filter events by site name             |
| `help`                     | Show available commands and keybinds   |
| `star`                     | Open GitHub repo in browser            |
| `dismiss`                  | Hide the star banner permanently       |
| `powermode`                | Toggle PowerMode visual effects        |
| `dashboard`                | Toggle live dashboard panel            |
| `stop` / `quit` / `exit`   | Stop the watcher                       |

### TUI keybinds

| Key       | Description                                |
|-----------|--------------------------------------------|
| `Ctrl+C` / `Ctrl+D` | Quit                              |
| `?`       | Toggle help / Esc to close                 |
| `↑` / `↓` | Navigate command history                   |
| `PgUp` / `PgDn` | Page up/down in event log              |
| `Home`    | Scroll to top                              |
| `End`     | Toggle auto-scroll                         |
| `Tab`     | Complete commands / site names             |
| `Insert`  | File system path scan for autocomplete     |
| `Delete`  | Cancel pending completions                 |
| `F2`      | Open interactive filter panel (extension)  |
| `F4`      | Toggle PowerMode visual effects            |
| `r`       | Get a context-aware training suggestion    |
| `Ctrl+X`  | Disable Xdebug if detected                 |

The status line shows memory usage with a live sparkline, change/clear counters, and per-site clear breakdowns when multiple sites are active.

## PowerMode

PowerMode adds energetic visual feedback when the watcher detects rapid file changes. Inspired by the VS Code Power Mode extension, it turns bursts of activity into a more engaging terminal experience.

When multiple file changes or cache clears arrive within a short window (2s), a combo counter builds up and visual effects intensify:

| Combo | Level | Effects |
|---|---|---|
| 0-2 | Normal | No effects |
| 3-5 | Warm | Status bar border turns orange, combo counter and energy bar appear |
| 6-10 | Hot | Border shifts to deep orange, particles animate, screen pulses subtly |
| 11+ | Power | Red animated border, intense particles (✦ ✧ ⚡ ★ ♦), full energy bar, screen pulses |

The **combo counter** (`⚡ x5`) and **energy bar** (`▓▓▓▓░░░░`) appear on the status line when activity ramps up. Energy decays during idle periods.

PowerMode also plays **synthesized tones** through your speakers when levels increase:
- **Warm**: single C5 tone (523 Hz)
- **Hot**: ascending C5→E5 two-tone arpeggio
- **Power**: ascending C5→E5→G5→C6 four-tone arpeggio
- Each combo increment plays a quick pitch blip that rises with combo count
- Sound is optional and silently degrades if audio initialization fails

Sound requires CGO and is only available in locally-compiled binaries (e.g. `go install`). Pre-built binaries from GitHub Releases use `CGO_ENABLED=0` for portability, so audio is disabled in those builds.

Toggle PowerMode on/off at any time with `F4` or the `powermode` command.

## Interactive CLI Commands

When running with `--no-tui`, type commands at the prompt:

| Command                | Description                            |
|------------------------|----------------------------------------|
| `status`               | Show stats, memory, and kernel watches |
| `list` / `config`      | Show current configuration             |
| `stats`                | Show runtime statistics and memory     |
| `add <route>`          | Add a route and restart watcher        |
| `remove <route>`       | Remove a route and restart watcher     |
| `reload`               | Reload config from file                |
| `help`                 | Show available commands                |
| `stop` / `quit` / `exit` | Stop the watcher                     |

## Configuration

Config is read from `configs/config.yaml` (auto-created on first run). Falls back to `watcher.config.json` for legacy setups.

```yaml
routes:
  - docroot/modules/custom
  - docroot/themes/custom
patterns:
  - .php; .module; .inc; .yml; .html.twig; .twig; .css; .js
debounce: 800
commandsPerPattern:
  .html.twig: cc render
  .twig: cc render
  .theme: cc theme-registry
  .module: cc plugin
  .inc: cc plugin
  .php: cc plugin
  .yml: cc plugin
  .info.yml: cr
  .services.yml: cr
  .routing.yml: cr
  .permissions.yml: cr
  .links.menu.yml: cr
  .css: cc css-js
  .js: cc css-js
skipLint: false
lintCommands:
  .php: php -l
  .yml: yaml
  .yaml: yaml
phpCsStandard: ""
watchMode: auto
pollInterval: 2000
eventBufferSize: 500
```

Any value can be overridden via environment variables with the `DRUPAL_WATCHER_` prefix:

| Env var | Config key | Example |
|---|---|---|
| `DRUPAL_WATCHER_DEBOUNCE` | `debounce` | `DRUPAL_WATCHER_DEBOUNCE=500` |
| `DRUPAL_WATCHER_DRUSH_COMMAND` | `drushCommand` | `DRUPAL_WATCHER_DRUSH_COMMAND=cc all` |
| `DRUPAL_WATCHER_SKIP_LINT` | `skipLint` | `DRUPAL_WATCHER_SKIP_LINT=true` |
| `DRUPAL_WATCHER_WATCH_MODE` | `watchMode` | `DRUPAL_WATCHER_WATCH_MODE=poll` |
| `DRUPAL_WATCHER_POLL_INTERVAL` | `pollInterval` | `DRUPAL_WATCHER_POLL_INTERVAL=1000` |

| Field                 | Description                                                  |
|-----------------------|--------------------------------------------------------------|
| `routes`              | Directories to watch (relative to Drupal root)               |
| `patterns`            | File extensions to trigger cache clears on                   |
| `debounce`            | Milliseconds to wait before running drush after a change     |
| `drushCmd`            | Custom path to the drush binary (auto-detected if omitted)   |
| `drushCommand`        | Default drush command (default: `cr`)                        |
| `drushArgs`           | Extra arguments to pass to drush                             |
| `commandsPerPattern`  | Maps file extensions to specific drush commands              |
| `postClearCommands`   | Shell commands to run after each cache clear                 |
| `excludePatterns`     | Path substrings to exclude from watching                     |
| `Sites`               | Site names to watch in multi-site setups (resolved via `drush/sites.yml`) |
| `skipLint`            | Disable lint checking before cache clear                     |
| `lintCommands`        | Per-extension lint commands (default: `php -l` for PHP, Go yaml parser for YAML). Only checks files inside `routes`. |
| `phpCsStandard`       | PHPCS standard for PHP linting, e.g. `"auto"`, `"Drupal"`, `"DrupalStrict"`. When set, replaces `php -l` with `phpcs` using Drupal coding standards. `"auto"` detects Drupal 11 → `DrupalStrict`, else `Drupal`. Empty string (default) keeps `php -l`. |
| `watchMode`           | File watching mode: `auto`, `fsnotify`, `poll`, `hybrid`     |
| `pollInterval`        | Polling interval in ms (default 2000, only used in poll/hybrid modes) |

**commandsPerPattern** maps file extensions to drush commands. The most specific match wins (e.g., `.info.yml` matches before `.yml`). Falls back to `cr` if no pattern matches.

### Watch modes

| Mode | Description |
|---|---|
| `auto` (default) | Tries fsnotify first; falls back to polling if OS limits are hit |
| `fsnotify` | Native file system events only (lowest latency) |
| `poll` | Periodic file tree scan at `pollInterval` (works around OS limits) |
| `hybrid` | Runs both fsnotify and polling simultaneously; events are deduplicated within a 1s window. Provides the reliability of polling with the low latency of fsnotify. |

Config via `watchMode` in `configs/config.yaml` (or legacy `watcher.config.json`) or override per session with `DRUPAL_WATCHER_WATCH_MODE=poll`. Polling mode is useful in large projects that exceed `fs.inotify.max_user_watches` on Linux, or when running in shared filesystems like NFS.

## How it works

1. `drupal-watcher start` loads config, detects the Drupal docroot, and writes a PID file
2. Uses `fsnotify` to watch all subdirectories under the configured routes (falls back to polling if fsnotify fails, or use hybrid mode for both)
3. When files change, debounces (default 800ms) collecting all changes into a batch. File changes appear instantly in the TUI event log — only the cache clear waits for the debounce window.
4. **PHP and YAML files are linted** before running drush (`php -l` or `phpcs` for PHP, Go yaml parser for YAML). Lint results are cached with a SHA-1 content hash and 5-minute TTL — unchanged files are not re-checked. Only files inside watched `routes` are checked. If linting fails, the cache clear is skipped and the error (with file path) is displayed in the TUI.

   When `phpCsStandard` is set in the config, PHP files are checked with `phpcs` using Drupal coding standards (auto-detects `DrupalStrict` for Drupal 11, `Drupal` for Drupal 10). Requires `drupal/coder` and `squizlabs/php_codesniffer` installed via Composer.
5. Compatible cache clear commands are merged into a single `drush` call (e.g. `drush cc render,plugin,css-js`)
6. If any change requires a full rebuild (`cr`), it applies a **lazy rebuild** — a separate 2-second debounce timer accumulates all changes and executes a single `drush cr` at the end of the burst, avoiding redundant full rebuilds
7. Drush output and post-clear commands are displayed in the TUI or printed to the terminal
8. A health file is written to `~/.cache/drupal-watcher/health` every 30s (cleaned up on shutdown)
9. Metrics (changes, clears, errors per minute) are tracked in-memory for the training mode and `stats` command
10. `Ctrl+C` (or `SIGTERM`) cancels the context, stops all modules with a 10s timeout, removes PID and health files

### Drush optimizations

The watcher applies several optimizations to minimize overhead:

| Optimization | Description |
|---|---|
| **Binary caching** | Resolved `drush` path cached after first lookup, avoids repeated `$PATH` scans |
| **Batch cache clears** | Multiple `cc <type>` commands merged into a single `drush cc type1,type2,...` call |
| **`cr` overrides** | If any change requires `drush cr`, a 2-second lazy rebuild timer accumulates all changes first |
| **Lint cache** | SHA-1 content hash avoids re-linting unchanged files (5-minute TTL, max 1000 entries) |
| **Parallel multi-site** | Per-site `drush` runs in a worker pool (up to 3 concurrent) for multi-site projects |
| **Quiet mode** | Drush runs with `--quiet --no-ansi` by default for minimal output overhead |

On large debounce windows, rapid changes to different file types share a single PHP bootstrap instead of spawning separate processes.

### Drupal root detection

The watcher scans for `docroot/`, `web/`, `public/`, or `html/` directories containing `core/`, `modules/`, `themes/`, or `index.php`. The detected root is stored in the config file.

### Cache clear per pattern

| Extension          | Drush command       |
|--------------------|---------------------|
| `.html.twig`       | `cc render`         |
| `.twig`            | `cc render`         |
| `.theme`           | `cc theme-registry` |
| `.module`          | `cc plugin`         |
| `.inc`             | `cc plugin`         |
| `.php`             | `cc plugin`         |
| `.yml`             | `cc plugin`         |
| `.info.yml`        | `cr`                |
| `.services.yml`    | `cr`                |
| `.routing.yml`     | `cr`                |
| `.permissions.yml` | `cr`                |
| `.links.menu.yml`  | `cr`                |
| `.css`             | `cc css-js`         |
| `.js`              | `cc css-js`         |

### Twig debug mode (development only)

Enable Twig development mode without touching settings.php:

```bash
drush twig:debug on   # enables Twig debug + auto-disable cache
drush twig:debug off  # restores production settings
```

Available since Drush 12.1+. Handles `twig.config` settings automatically — no manual cache clears needed.

### Pre-warming caches (Drush 13+)

`drush cache:warm` pre-builds caches so the first request isn't slow after a rebuild.
This is optional — add it to `postClearCommands` if you want automatic warming:

```json
"postClearCommands": ["drush cache:warm"]
```

> **Note:** Warming can be slow on large sites. Not recommended during active development.

## Multi-site

Drupal Watcher supports [multi-site](https://www.drupal.org/docs/developing/multisite-drupal) setups with a single watcher process. When multiple sites are detected under `sites/`, drush runs in parallel goroutines for each site on every file change.

### Auto-detection

The watcher checks for directories under `sites/` beyond `default/` that contain `settings.php`. If only `sites/default/` exists, single-site mode is used (no changes to existing workflows).

### Site aliases (required for multi-site)

When multiple sites are detected, you **must** configure site aliases so the watcher knows each site's URI. The watcher supports both formats Drush accepts:

#### Combined file: `drush/sites.yml`

```yaml
# docroot/drush/sites.yml
site1:
  uri: 'https://site1.local'
site2:
  uri: 'https://site2.local'
```

#### Per-site files: `drush/sites/{name}.site.yml`

```
drush/
├── drush.yml
└── sites
    ├── site1.site.yml
    ├── site2.site.yml
    └── site3.site.yml
```

Each `{name}.site.yml` file defines one site alias. For example:

```yaml
# drush/sites/site1.site.yml
site1:
  uri: 'https://site1.local'
```

```yaml
# drush/sites/site2.site.yml
site2:
  uri: 'https://site2.local'
```

The per-site directory format is discovered automatically by scanning `drush/sites/*.site.yml`. The two formats are mutually exclusive — if `drush/sites.yml` exists, the directory is ignored.

The watcher does **not** guess URIs from directory names — a site alias can differ from its directory name. See [Drush site aliases docs](https://www.drush.org/latest/using-drush/site-aliases/) for details.

If multi-site is detected and neither `drush/sites.yml` nor `drush/sites/*.site.yml` exists, the watcher exits with an error and instructions.

### Filtering sites

| Flag | Example | Description |
|---|---|---|
| `--site` | `--site=site1,site2` | Whitelist — only watch these sites |
| `--exclude-site` | `--exclude-site=site3` | Blacklist — watch all except these |
| `--uri` | `--uri=https://site1.local` | Override single-site mode with a specific URI (skips detection) |

### Config persistence

The `Sites` field in `watcher.config.json` persists the site list from a previous run:

```json
{
  "routes": ["docroot/modules/custom"],
  "Sites": ["site1", "site2"]
}
```

When `Sites` is present in the config file, it's auto-resolved against `drush/sites.yml` on startup. This is useful when you always work with the same subset of sites.

### TUI display

Events in the TUI are tagged with the site name:

```
10:00:01  ✔  drush cc plugin [site1] (312ms, exit 0)
10:00:01  ✔  drush cc plugin [site2] (289ms, exit 0)
```

## PID management

The watcher writes a PID file to `~/.cache/drupal-watcher/.drupal-watcher-<project-hash>.pid` (`0600` permissions) to prevent multiple instances. The filename includes a hash of the project root, so you can run the watcher in multiple projects without conflicts. If the process crashes, stale PID files are cleaned up automatically on the next `start`.

## Architecture

The codebase uses a **hexagonal (ports & adapters)** architecture:

```
cmd/
  drupal-watcher/        → Binary (module system with DI container + EventBus)

internal/
  app/
    app.go               → App lifecycle (Start/Stop/Done)
    container.go         → DI container (Set/Get/MustGet)
    module.go            → Module interface
    common/types.go      → ServiceName constants
    eventbus/
      bus.go             → Pub/sub event bus (async, topic-based)
    modules/
      config/            → Config module (loads watcher.config.json, stores in container)
      watcher/           → Watcher module (creates FSNotifyWatcher from config)
      executor/          → Executor module (creates DrushExecutor from config)
      orchestrator/      → Orchestrator module (engine with EventBus, starts in goroutine)
      ui/                → UI module (runs Bubble Tea TUI, blocks until quit)
        providers/tui/   → TUI bridge (EventBus → EngineEvent channel)

  hooks/
    builtin/
      drush_clear.go     → Default post-execution hook
    examples/
      slack.go           → Demo: Slack webhook notifier
  ui/                    → Bubble Tea TUI (model, view, update, styles)
  config/                → Config management (YAML + env vars), Drupal root detection, PID files
  health/                → Liveness check (timestamp file every 30s)
  drush/                 → Drush resolution, execution, health checks
  metrics/               → Runtime statistics (changes, clears, errors per minute)
  training/              → Context-aware training suggestions (training.json)
  validate/              → Config and environment validation (`validate` command)
  xdebug/                → Xdebug detection and disable (Ctrl+X)
  utils/                 → Color helpers, format utilities

pkg/
  core/
    interfaces.go        → Watcher, CommandExecutor, EventFilter, PostProcessor
    models.go            → FileEvent, ExecutionResult, EngineEvent, SiteInfo
  adapters/
    fsnotify_watcher.go  → core.Watcher via fsnotify
    polling_watcher.go   → core.Watcher via periodic file tree scan
    hybrid_watcher.go    → core.Watcher (fsnotify + polling, deduped)
    drush_executor.go    → core.CommandExecutor via drush
    regex_filter.go      → Pattern/Exclude/Dotfile filters
    php_lint.go          → core.LintChecker via php -l
    yaml_lint.go         → core.LintChecker via Go yaml parser
    lint_cache.go        → Caching wrapper for LintChecker (SHA-1, 5min TTL)
    slog_logger.go       → Structured logger factory
```

### Key domain interfaces

```go
// pkg/core/interfaces.go

type Watcher interface {
    Start(ctx context.Context) (<-chan FileEvent, <-chan error)
    Add(path string) error
    Remove(path string) error
    Close() error
}

type CommandExecutor interface {
    Execute(ctx context.Context, commands []string, dir string) ExecutionResult
}

type EventFilter interface {
    ShouldProcess(event FileEvent) bool
}

type PostProcessor interface {
    Name() string
    Process(ctx context.Context, event FileEvent, result ExecutionResult) error
}

type LintChecker interface {
    Lint(filePath string) *LintResult
}

type LintResult struct {
    File  string
    Error string
}
```

### Engine event loop

The orchestrator (`internal/app/modules/orchestrator/engine.go`) runs the central pipeline:

1. Watcher emits raw `FileEvent` on a channel (via fsnotify, polling, or hybrid)
2. All `EventFilter` implementations decide if the event should be processed
3. Debounce timer groups rapid changes into a single batch
4. **Lint check**: each changed file is checked by its `LintChecker` (`.php` → `php -l`, `.yml` → Go yaml parser). If any file fails, the batch is skipped and an `error` EventBus event is published.
5. Matching file extensions are resolved to drush commands via `CommandsPerPattern`
6. `CommandExecutor` runs the resolved commands
7. All `PostProcessor` implementations run sequentially with the execution result
8. `EngineEvent` is published to the EventBus on `file.change` and `cache.clear` topics

---

## Module System (for developers)

### Creating a new module

Implement the `app.Module` interface:

```go
type Module interface {
    Name() string
    DependsOn() []Module
    Init(container *Container) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

- **Name**: unique identifier
- **DependsOn**: list of modules that must init before this one
- **Init**: resolve dependencies from container, register own services
- **Start**: start any background goroutines, or block (UI)
- **Stop**: clean up resources

### Live example: notifications module

```go
package notifications

import (
    "context"
    "fmt"

    "github.com/irving-frias/drupal-watcher/internal/app"
    "github.com/irving-frias/drupal-watcher/internal/app/common"
    "github.com/irving-frias/drupal-watcher/internal/app/eventbus"
)

type Module struct {
    bus *eventbus.EventBus
}

func (m *Module) Name() string { return "notifications" }

func (m *Module) DependsOn() []app.Module { return nil }

func (m *Module) Init(container *app.Container) error {
    m.bus = container.MustGet(common.SvcEventBus).(*eventbus.EventBus)

    // Subscribe to cache clear events
    m.bus.Subscribe(eventbus.TopicCacheClear, func(event any) {
        fmt.Printf("Cache cleared: %+v\n", event)
    })

    return nil
}

func (m *Module) Start(ctx context.Context) error { return nil }
func (m *Module) Stop(ctx context.Context) error  { return nil }
```

### Live example: custom executor

```go
package custom

import (
    "context"
    "github.com/irving-frias/drupal-watcher/internal/app"
    "github.com/irving-frias/drupal-watcher/internal/app/common"
    "github.com/irving-frias/drupal-watcher/pkg/core"
)

type MyExecutor struct{}

func (e *MyExecutor) Execute(ctx context.Context, commands []string, dir string) core.ExecutionResult {
    return core.ExecutionResult{ExitCode: 0, Stdout: "ok"}
}

type Module struct{}

func (m *Module) Name() string { return "custom-executor" }

func (m *Module) DependsOn() []app.Module { return nil }

func (m *Module) Init(container *app.Container) error {
    // Override the default executor
    container.Set(common.SvcExecutor, &MyExecutor{})
    return nil
}

func (m *Module) Start(ctx context.Context) error { return nil }
func (m *Module) Stop(ctx context.Context) error  { return nil }
```

### Registering modules in `cmd/modular/main.go`

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/irving-frias/drupal-watcher/internal/app"
    cfgmodule "github.com/irving-frias/drupal-watcher/internal/app/modules/config"
    execmodule "github.com/irving-frias/drupal-watcher/internal/app/modules/executor"
    orcmodule "github.com/irving-frias/drupal-watcher/internal/app/modules/orchestrator"
    uimodule "github.com/irving-frias/drupal-watcher/internal/app/modules/ui"
    watchermodule "github.com/irving-frias/drupal-watcher/internal/app/modules/watcher"
    "github.com/irving-frias/drupal-watcher/internal/config"
)

func main() {
    root := "."
    if len(os.Args) > 1 {
        root = os.Args[1]
    }

    a := app.New(
        &cfgmodule.Module{WorkDir: root},
        &watchermodule.Module{},
        &execmodule.Module{},
        &orcmodule.Module{},
        &uimodule.Module{},
    )

    if err := config.WritePid(root); err != nil {
        fmt.Fprintf(os.Stderr, "PID: %v\n", err)
        os.Exit(1)
    }
    defer a.Stop(context.Background())

    if err := a.Start(context.Background()); err != nil && err != context.Canceled {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Startup order

```
main()
├── Create cancellable context (ctx)
├── go health.Run(ctx)                  → writes ~/.cache/drupal-watcher/health every 30s
├── app.Start(ctx)
│   ├── Register EventBus in container
│   ├── Init modules (dependency-sorted):
│   │   ├── config     → creates config.Manager, loads configs/config.yaml
│   │   ├── watcher    → creates FSNotifyWatcher from config
│   │   ├── executor   → creates DrushExecutor from config
│   │   ├── orchestrator→ creates Engine with EventBus, watcher, executor
│   │   └── ui         → gets EventBus + workDir (for star banner persistence)
│   ├── Start modules (sequential):
│   │   ├── config     → no-op
│   │   ├── watcher    → no-op
│   │   ├── executor   → no-op
│   │   ├── orchestrator→ starts engine.Run() in a goroutine
│   │   └── ui         → runs TUI (blocks until user quits or ctx cancelled)
│   └── Wait for context cancellation or TUI exit
└── defer app.Stop(ctx)                 → sync.Once, 10s timeout
    ├── Cancel root context → health check stops + engine stops
    └── Stop modules in reverse order
```

### EventBus topics

| Topic | Published by | Event type | Consumers |
|---|---|---|---|
| `file.change` | Orchestrator | `core.EngineEvent` | TUI, notifications |
| `cache.clear` | Orchestrator | `core.EngineEvent` | TUI, notifications |
| `error` | Orchestrator | `core.EngineEvent` | TUI (lint failures, watcher errors) |
| `engine.start` | Orchestrator | (empty) | lifecycle hooks |
| `engine.stop` | Orchestrator | (empty) | lifecycle hooks |

### Extensibility: adding a custom post-processor

The old approach (still works with `cmd/drupal-watcher`):

```go
type SlackNotifier struct {
    WebhookURL string
}

func (s *SlackNotifier) Name() string { return "SlackNotifier" }

func (s *SlackNotifier) Process(ctx context.Context, event core.FileEvent, result core.ExecutionResult) error {
    // your logic here
}
```

Wire it in the module's `Init`:

```go
func (m *Module) Init(container *app.Container) error {
    engine := container.MustGet(common.SvcOrchestrator).(*orchestrator.Engine)
    engine.PostProcessors = append(engine.PostProcessors, &SlackNotifier{...})
    return nil
}
```

### Running the modular binary

```bash
go run ./cmd/modular                          # current directory
go run ./cmd/modular /path/to/drupal/project  # with root

# Commands
go run ./cmd/modular version                  # print version
go run ./cmd/modular help                     # usage

# Build
go build -o modular-watcher ./cmd/modular
./modular-watcher
```

## Development (requires Go 1.25+)

```bash
go test -count=1 ./...
go vet ./...
go build -o drupal-watcher ./cmd/drupal-watcher
```

### Cross-compilation

```bash
GOOS=linux   GOARCH=amd64 go build -o drupal-watcher-linux-amd64   ./cmd/drupal-watcher
GOOS=linux   GOARCH=arm64 go build -o drupal-watcher-linux-arm64   ./cmd/drupal-watcher
GOOS=darwin  GOARCH=amd64 go build -o drupal-watcher-darwin-amd64  ./cmd/drupal-watcher
GOOS=darwin  GOARCH=arm64 go build -o drupal-watcher-darwin-arm64  ./cmd/drupal-watcher
GOOS=windows GOARCH=amd64 go build -o drupal-watcher-windows-amd64.exe ./cmd/drupal-watcher
```

CI builds all platforms automatically on push to `main` and publishes releases.

## License

MIT
