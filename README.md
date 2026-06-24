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

| Command   | Description                            |
|-----------|----------------------------------------|
| start     | Start watching file changes            |
| status    | Show running status and uptime         |
| list      | Display current configuration          |
| add       | Add route and/or pattern to watch      |
| remove    | Remove route and/or pattern            |
| restart   | Restart the watcher                    |
| stop      | Stop the watcher and clear PID         |
| reset     | Clear stale PID (if process crashed)   |
| help      | Show usage information                 |

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
  }
}
```

**commandsPerPattern** maps file extensions to drush commands. The most specific match wins (e.g., `.info.yml` matches before `.yml`). Falls back to `cr` if no pattern matches.

## How it works

1. `drupal-watcher start` loads config, detects the Drupal docroot, and writes a PID file
2. Uses `fsnotify` to watch all subdirectories under the configured routes
3. When a file changes, debounces (default 800ms) and runs the corresponding drush command
4. Drush output and post-clear commands are printed to the terminal
5. `Ctrl+C` stops the watcher, removes the PID file, and prints stats

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

| Extension      | Drush command       |
|----------------|---------------------|
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
