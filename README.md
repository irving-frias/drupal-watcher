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

On first run, a `watcher.config.json` is auto-created with defaults. Edit it to customize routes, patterns, and cache clear commands.

The TUI opens automatically. Events appear in real-time, and you can type commands at the prompt:

```
  ● drupal-watcher  PID: 12345  Uptime: 5m
  Memory: 2.1 MB  |  Kernel watches: 630  |  Changes: 14  |  Clears: 3

  ┌──────────────────────────────────────────────────────────────┐
  │ 10:00:01  ℹ  Waiting for file changes...                     │
  │ 10:02:15  ℹ  Change detected: docroot/modules/custom/foo.module │
  │ 10:02:16  ✔  drush cc plugin (312ms, exit 0)                │
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
| `tui`                      | Terminal UI (experimental)             |
| `status`                   | Show running status and uptime         |
| `list` / `config`          | Display current configuration          |
| `add` <route> [pattern]    | Add route and/or pattern to watch      |
| `remove` / `rm` <route>    | Remove route and/or pattern            |
| `restart`                  | Restart the watcher                    |
| `stop` / `reset`           | Stop the watcher and clear PID         |
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

**Per OS:**

| OS | Method | Requirements |
|---|---|---|
| **macOS** | `osascript` | None (built-in) |
| **Linux** (native) | `notify-send` | `libnotify-bin` (Debian/Ubuntu) or `libnotify` (Fedora/Arch) |
| **WSL** | `powershell.exe` → Windows Toast | None (calls host PowerShell) |
| **Windows** | `powershell` → ToastNotificationManager | None (built-in) |

WSL is auto-detected by reading `/proc/sys/kernel/osrelease` and uses `powershell.exe` to show the native Windows 10/11 Toast. No configuration required.

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

| Command                | Description                            |
|------------------------|----------------------------------------|
| `status`               | Show stats, memory, and kernel watches |
| `help`                 | Show available commands                |
| `stop` / `quit` / `exit` | Stop the watcher                     |

Press `Ctrl+C` or `Ctrl+D` to quit at any time.

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

`watcher.config.json` sits in your Drupal project root:

```json
{
  "routes": ["docroot/modules/custom", "docroot/themes/custom"],
  "patterns": [".php", ".module", ".inc", ".yml", ".html.twig", ".twig", ".css", ".js"],
  "debounce": 800,
  "commandsPerPattern": {
    ".html.twig": "cc render",
    ".twig": "cc render",
    ".theme": "cc theme-registry",
    ".module": "cc plugin",
    ".inc": "cc plugin",
    ".php": "cc plugin",
    ".yml": "cc plugin",
    ".info.yml": "cr",
    ".services.yml": "cr",
    ".routing.yml": "cr",
    ".permissions.yml": "cr",
    ".links.menu.yml": "cr",
    ".css": "cc css-js",
    ".js": "cc css-js"
  },
  "postClearCommands": []
}
```

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

**commandsPerPattern** maps file extensions to drush commands. The most specific match wins (e.g., `.info.yml` matches before `.yml`). Falls back to `cr` if no pattern matches.

## How it works

1. `drupal-watcher start` loads config, detects the Drupal docroot, and writes a PID file
2. Uses `fsnotify` to watch all subdirectories under the configured routes
3. When files change, debounces (default 800ms) collecting all changes into a batch
4. Compatible cache clear commands are merged into a single `drush` call (e.g. `drush cc render,plugin,css-js`)
5. If any change requires a full rebuild (`cr`), it overrides all other commands
6. Drush output and post-clear commands are displayed in the TUI or printed to the terminal
7. `Ctrl+C` stops the watcher, removes the PID file, and prints stats

### Drush optimizations

The watcher applies several optimizations to minimize overhead:

| Optimization | Description |
|---|---|
| **Binary caching** | Resolved `drush` path cached after first lookup, avoids repeated `$PATH` scans |
| **Batch cache clears** | Multiple `cc <type>` commands merged into a single `drush cc type1,type2,...` call |
| **`cr` overrides** | If any change requires `drush cr`, all other commands are skipped |
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

The watcher writes a PID file (`watcher.pid`) to prevent multiple instances. If the process crashes, `drupal-watcher stop` or `drupal-watcher reset` cleans up stale PID files.

## Architecture

The codebase uses a plugin-style architecture. Core interfaces live in `pkg/core/`, implementations in `pkg/adapters/`, and the event loop orchestrator in `internal/orchestrator/`. Adding new features (Slack, Redis, custom scripts) requires zero changes to the orchestrator — just implement an interface.

```
cmd/
  drupal-watcher/        → CLI entry point (flag parsing, command dispatch)
  watcher/               → Thin DI entry point (pure dependency injection)

pkg/
  core/
    interfaces.go        → Watcher, CommandExecutor, EventFilter, PostProcessor
    models.go            → FileEvent, ExecutionResult, EngineEvent, SiteInfo
  adapters/
    fsnotify_watcher.go  → core.Watcher via fsnotify
    drush_executor.go    → core.CommandExecutor via drush
    regex_filter.go      → Pattern/Exclude/Dotfile filters
    slog_logger.go       → Structured logger factory

internal/
  orchestrator/
    engine.go            → Central engine: debounce, filter pipeline, execution, hooks
  hooks/
    builtin/
      drush_clear.go     → Default post-execution hook
    examples/
      slack.go           → Demo: Slack webhook notifier
  ui/
    model.go             → BubbleTea TUI model
    view.go             → TUI rendering
    update.go           → TUI event handling
    styles.go           → TUI lipgloss styles
    tui.go              → Public Run() function
  config/                → Config management, Drupal root detection, PID files
  drush/                 → Drush resolution, execution, health checks, OS notifications
  cli/                   → CLI command implementations (uses orchestrator)
  utils/                 → Color helpers, format utilities, shared helpers
```

### Key interfaces

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
```

### Engine event loop

The orchestrator (`internal/orchestrator/engine.go`) runs the central pipeline:

1. Watcher emits raw `FileEvent` on a channel
2. All `EventFilter` implementations decide if the event should be processed
3. Debounce timer groups rapid changes into a single batch
4. Matching file extensions are resolved to drush commands via `CommandsPerPattern`
5. `CommandExecutor` runs the resolved commands
6. All `PostProcessor` implementations run sequentially with the execution result
7. `EngineEvent` is sent to the TUI channel (if subscribed)

### Extensibility: adding a custom post-processor

To add a new feature (Slack, Redis, metrics, etc.), create a file in `internal/hooks/` or anywhere in the project, implementing `core.PostProcessor`:

```go
package hooks

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "time"

    "github.com/irving-frias/drupal-watcher/pkg/core"
)

type SlackNotifier struct {
    WebhookURL string
    client     *http.Client
}

func (s *SlackNotifier) Name() string { return "SlackNotifier" }

func (s *SlackNotifier) Process(ctx context.Context, event core.FileEvent, result core.ExecutionResult) error {
    payload, _ := json.Marshal(map[string]interface{}{
        "text": "drush " + result.Command + " — exit " + string(rune(result.ExitCode)),
    })
    req, _ := http.NewRequestWithContext(ctx, "POST", s.WebhookURL, bytes.NewReader(payload))
    resp, err := s.client.Do(req)
    if err != nil {
        return err
    }
    resp.Body.Close()
    return nil
}
```

Then wire it in the entry point (zero changes to the orchestrator):

```go
// cmd/watcher/main.go

engineCfg := core.EngineConfig{
    Watcher:  fsnWatcher,
    Executor: drushExec,
    Filters:  []core.EventFilter{patternFilter, excludeFilter},
    PostProcessors: []core.PostProcessor{
        builtin.NewDrushClear(),                               // default
        hooks.NewSlackNotifier("https://hooks.slack.com/..."), // one-liner addition
    },
    // ...
}
```

The `PostProcessor` interface is the extension point. The engine iterates over all registered processors after every cache clear.

### New entry point: `cmd/watcher`

A minimal dependency-injection entry point that wires all components explicitly:

```bash
go run ./cmd/watcher --root /path/to/drupal --debounce 500
```

Flags: `--root`, `--config`, `--debounce`, `--no-tui`. Does **not** include legacy CLI commands (add/remove/list/stop). For those, use `cmd/drupal-watcher`.

## Development (requires Go 1.24+)

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
