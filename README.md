# Drupal Watcher

File watcher for Drupal development. Monitors custom modules and themes, auto-runs `drush` cache clears on file changes. Supports DDEV, Lando, and local environments.

## Requirements

- **Go 1.21+** to compile (or download a [pre-built binary](https://github.com/irving-frias/drupal-watcher/releases))
- **Drush** installed in your Drupal project

## Installation

### Via Composer

```bash
composer require irving-frias/drupal-watcher
vendor/bin/drupal-watcher help
```

The binary compiles automatically on `composer install` (requires Go). If Go is not available, the launcher will compile on first run.

### Standalone binary

```bash
go install github.com/irving-frias/drupal-watcher/cmd/drupal-watcher@latest
```

Or download from [GitHub Releases](https://github.com/irving-frias/drupal-watcher/releases).

### From source

```bash
git clone https://github.com/irving-frias/drupal-watcher.git
cd drupal-watcher
go build -o drupal-watcher ./cmd/drupal-watcher
./drupal-watcher help
```

## Quick Start

```bash
cd /path/to/drupal/project
drupal-watcher start
```

A `watcher.config.json` is auto-generated with sensible defaults. Edit it to customize.

## Commands

| Command   | Description                                   |
|-----------|-----------------------------------------------|
| start     | Start watching file changes                   |
| stop      | Stop the watcher and clear PID                |
| restart   | Restart the watcher                           |
| status    | Show running status and uptime                |
| list      | Display current configuration                 |
| add       | Add route and/or pattern to watch            |
| remove    | Remove route and/or pattern                  |
| reset     | Clear stale PID (if process crashed)          |
| help      | Show usage information                        |

## Options

- `--debounce <ms>` — Debounce interval (default: 800ms)
- `--no-dotfiles` — Ignore dotfiles (not yet implemented)
- `--log-file <path>` — Write logs to file
- `--config <path>` — Custom config file path

## Configuration

The config file `watcher.config.json` is in your project root:

```json
{
  "routes": ["docroot/modules/custom", "docroot/themes/custom"],
  "patterns": [".php", ".module", ".inc", ".yml", ".html.twig"],
  "debounce": 800,
  "commandsPerPattern": {
    ".html.twig": "cc twig",
    ".theme": "cc theme-registry",
    ".module": "cc plugin",
    ".inc": "cc plugin",
    ".yml": "cc plugin",
    ".php": "cc plugin",
    ".info.yml": "cr",
    ".services.yml": "cr"
  }
}
```

## Architecture

```
bin/drupal-watcher     → Shell launcher (builds if needed, then execs)
cmd/drupal-watcher/    → Entry point (flag parsing + dispatch)
internal/
  config/              → Config load/save, Drupal root detection, PID mgmt
  drush/               → Drush command resolution and execution
  watcher/             → fsnotify-based file watcher with debounce
  cli/                 → Command implementations (start, list, status, etc.)
  utils/               → Color helpers, shared constants
```

## Development

```bash
go test ./...
go test -count=1 ./...   # Force re-run (no cache)
go vet ./...
go build -o drupal-watcher ./cmd/drupal-watcher
```

## Cross-compilation

```bash
GOOS=linux   GOARCH=amd64 go build -o drupal-watcher-linux-amd64   ./cmd/drupal-watcher
GOOS=darwin  GOARCH=arm64 go build -o drupal-watcher-darwin-arm64  ./cmd/drupal-watcher
GOOS=windows GOARCH=amd64 go build -o drupal-watcher-windows-amd64.exe ./cmd/drupal-watcher
```

## License

MIT
