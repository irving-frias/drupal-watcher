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
| `monitor` / `m`            | Auto-refresh status every 2 seconds    |
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

**Per-SO:**

| SO | Método | Requisitos |
|---|---|---|
| **macOS** | `osascript` | Ninguno (built-in) |
| **Linux** (nativo) | `notify-send` | `libnotify-bin` (Debian/Ubuntu) o `libnotify` (Fedora/Arch) |
| **WSL** | `powershell.exe` → Toast de Windows | Ninguno (llama al PowerShell del host) |
| **Windows** | `powershell` → ToastNotificationManager | Ninguno (built-in) |

En WSL se detecta automáticamente leyendo `/proc/sys/kernel/osrelease` y usa `powershell.exe` para mostrar el Toast nativo de Windows 10/11. No requiere configurar nada.

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

### `drush/sites.yml` (required for multi-site)

When multiple sites are detected, you **must** configure site aliases. The watcher supports both formats Drush accepts:

### Single file: `drush/sites.yml`

```yaml
# docroot/drush/sites.yml
site1:
  uri: 'https://site1.local'
site2:
  uri: 'https://site2.local'
```

### Per-site files: `drush/sites/{name}.site.yml`

```yaml
# docroot/drush/sites/egade.site.yml
egade:
  uri: 'https://egade.local'
```

```yaml
# docroot/drush/sites/prepa.site.yml
prepa:
  uri: 'https://prepa.local'
```

The per-site directory format (`drush/sites/*.site.yml`) is discovered automatically. The two formats are mutually exclusive — if `drush/sites.yml` exists, the directory is ignored.

The watcher does not guess URIs from directory names — a site alias can differ from its directory name. See [Drush site aliases docs](https://www.drush.org/latest/using-drush/site-aliases/) for details.

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

```
bin/drupal-watcher      → Shell launcher (downloads binary if missing)
bin/install.php         → Composer hook, downloads binary for current OS/arch
cmd/drupal-watcher/     → Entry point, flag parsing, command dispatch
internal/
  config/               → Config management, Drupal root detection, PID files
  drush/                → Drush resolution, execution, health checks, OS notifications
  watcher/              → fsnotify file watcher with debounce, atomic stats
  cli/                  → All CLI and TUI command implementations
  tui/                  → BubbleTea terminal UI (model, view, update, styles)
  utils/                → Color helpers, format utilities, shared helpers
```

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
