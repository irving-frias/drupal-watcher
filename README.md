# Drupal Watcher

File watcher for Drupal development. Monitors custom modules and themes and auto-runs `drush` cache clears whenever a file changes. Works with DDEV, Lando, and any local environment.

## Requirements

- **PHP 8.1+** (for Composer-based install)
- **Drush** installed in your Drupal project

> Go is **not required**. The binary is auto-downloaded during `composer install`.

> **Development tool only.** This package is meant for local development and should not be required in production environments. Install it with `composer require --dev irving-frias/drupal-watcher`.

## Installation

```bash
composer require --dev irving-frias/drupal-watcher
vendor/bin/drupal-watcher help
```

> If `vendor/bin/drupal-watcher` is unavailable after install, use the full path: `vendor/irving-frias/drupal-watcher/bin/drupal-watcher`

On `composer install` the correct binary for your OS/architecture is downloaded from GitHub Releases and placed at `vendor/irving-frias/drupal-watcher/bin/drupal-watcher-go`. No compilation needed.

> Install with `--dev` to exclude it from production deployments via `composer install --no-dev`.

### Manual download

Download the binary for your platform from [GitHub Releases](https://github.com/irving-frias/drupal-watcher/releases), make it executable, and run it from your Drupal project root.

## Quick Start

```bash
cd /path/to/drupal/project
vendor/bin/drupal-watcher start
```

A `watcher.config.json` is auto-created with sensible defaults. Edit it to customize routes, patterns, and cache clear commands.

## Commands

| Command       | Description                            |
|---------------|----------------------------------------|
| start         | Start watching file changes            |
| status        | Show running status and uptime         |
| monitor (m)   | Auto-refresh status every 2 seconds    |
| list          | Display current configuration          |
| add           | Add route and/or pattern to watch      |
| remove        | Remove route and/or pattern            |
| restart       | Restart the watcher                    |
| stop          | Stop the watcher and clear PID         |
| reset         | Clear stale PID (if process crashed)   |
| help          | Show usage information                 |

## Options

| Flag                    | Description                        |
|-------------------------|------------------------------------|
| `--debounce <ms>`       | Debounce interval (default: 800ms) |
| `--log-file <path>`     | Write logs to file                 |
| `--config <path>`       | Custom config file path            |
| `--no-dotfiles`         | Ignore dotfiles                    |

## Configuration

`watcher.config.json` sits in your Drupal project root:

```json
{
  "routes": ["docroot/modules/custom", "docroot/themes/custom"],
  "patterns": [".php", ".module", ".inc", ".yml", ".html.twig"],
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

**commandsPerPattern** maps file extensions to drush commands. The most specific match wins (e.g., `.info.yml` matches before `.yml`). Falls back to `cr` if no pattern matches.

## How it works

1. `drupal-watcher start` loads config, detects the Drupal docroot, and writes a PID file
2. Uses `fsnotify` to watch all subdirectories under the configured routes
3. When files change, debounces (default 800ms) collecting all changes into a batch
4. Compatible cache clear commands are merged into a single `drush` call (e.g. `drush cc render,plugin,css-js`)
5. If any change requires a full rebuild (`cr`), it overrides all other commands
6. Drush output and post-clear commands are printed to the terminal
7. `Ctrl+C` stops the watcher, removes the PID file, and prints stats

## Architecture

```
bin/drupal-watcher      → Shell launcher (downloads binary if missing)
cmd/drupal-watcher/     → Entry point, flag parsing, command dispatch
internal/
  config/               → Config management, Drupal root detection, PID files
  drush/                → Drush resolution, execution, health checks
  watcher/              → fsnotify file watcher with debounce
  cli/                  → Command implementations
  utils/                → Color helpers and shared utilities
```

## Drupal root detection

The watcher scans for `docroot/`, `web/`, `public/`, or `html/` directories containing `core/`, `modules/`, `themes/`, or `index.php`. The detected root is stored in the config file.

## Cache clear per pattern

Different file types run different drush commands:

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

## Interactive commands

While the watcher is running, type commands at the prompt:

| Command                | Description                            |
|------------------------|----------------------------------------|
| `status`               | Show stats, memory, and kernel watches |
| `monitor` / `m`        | Auto-refresh status every 2 seconds    |
| `list` / `config`      | Show current configuration             |
| `stats`                | Show runtime statistics and memory     |
| `add <route>`          | Add a route and restart watcher        |
| `remove <route>`       | Remove a route and restart watcher     |
| `reload`               | Reload config from file                |
| `help`                 | Show available commands                |
| `stop` / `quit` / `exit` | Stop the watcher                     |

### Monitor mode

`monitor` refreshes the status pane every 2 seconds:

```
[15:30:00] ● Watcher running. PID 12345
  Changes: 142  Clears: 18  Uptime: 30m
  Memory: 28.4 MB  |  Kernel watches: 47
  Monitor mode (press Enter to stop)...
```

Press Enter to exit monitor mode.

## Drush optimizations

The watcher applies several optimizations to minimize overhead:

| Optimization | Description |
|---|---|
| **Binary caching** | Resolved `drush` path cached after first lookup, avoids repeated `$PATH` scans |
| **Batch cache clears** | Multiple `cc <type>` commands merged into a single `drush cc type1,type2,...` call |
| **`cr` overrides** | If any change requires `drush cr`, all other commands are skipped |
| **Quiet mode** | Drush runs with `--quiet --no-ansi` by default for minimal output overhead |

On large debounce windows, rapid changes to different file types share a single PHP bootstrap instead of spawning separate processes.

### Pre-warming caches (Drush 13+)

`drush cache:warm` pre-builds caches so the first request isn't slow after a rebuild.
This is optional — add it to `postClearCommands` if you want automatic warming:

```json
"postClearCommands": ["drush cache:warm"]
```

> **Note:** Warming can be slow on large sites. Not recommended during active development.

## Development (requires Go)

```bash
go test -count=1 ./...
go vet ./...
go build -o drupal-watcher ./cmd/drupal-watcher
```

### Cross-compilation

```bash
GOOS=linux   GOARCH=amd64 go build -o drupal-watcher-linux-amd64   ./cmd/drupal-watcher
GOOS=darwin  GOARCH=arm64 go build -o drupal-watcher-darwin-arm64  ./cmd/drupal-watcher
GOOS=windows GOARCH=amd64 go build -o drupal-watcher-windows-amd64.exe ./cmd/drupal-watcher
```

## License

MIT
