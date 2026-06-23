# Drupal Watcher

> 🚀 A smart file watcher for Drupal that monitors your custom modules and themes, automatically running `drush cr` when changes are detected.

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Bun](https://img.shields.io/badge/Bun-1.3+-black.svg)](https://bun.sh)
[![Composer](https://img.shields.io/badge/Composer-ready-brightgreen.svg)](https://getcomposer.org)

## Table of Contents

- [What it does](#what-it-does)
- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Architecture](#architecture)
- [Examples](#examples)
- [Per-pattern cache commands](#per-pattern-cache-commands)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Quick Start

```bash
# 1. Install in your Drupal project
composer require irving-frias/drupal-watcher

# 2. Start watching (local/Lando)
vendor/bin/drupal-watcher start

# Or with DDEV
ddev drupal-watcher start

# 3. Edit a file — drush cr runs automatically!
📝 my-module.module
🔄 Clearing cache...
✔ Cache cleared in 2.3s
```

## What it does

Forget running `drush cr` manually every time you edit a file. **Drupal Watcher**:

- **Watches** your custom module and theme files in real time
- **Auto-detects** changes in `.html.twig`, `.inc`, `.yml`, `.module`, `.theme`, `.php`, `.info.yml`, and `.services.yml` files
- **Runs `drush cr`** intelligently with debounce to avoid saturating your system
- **Compatible** with DDEV, Lando, and local environments (use `ddev drupal-watcher` in DDEV)
- **Persists** your custom routes in a configuration file

## Features

### Route management
- Add, remove, list, and reset watched routes on the fly
- Filter with `--watch=<path>` and `--no-watch=<path>` at startup
- Persistence in `watcher.config.json`
- Existence validation for watched folders

### Drupal-optimized
- Auto-detects Drupal docroot (`docroot`, `web`, `html`, `public`, `drupal`)
- Smart debounce (800ms default)
- DDEV-ready: run `ddev drupal-watcher <command>`

### Blazing fast
- Bun-powered installation (10-30× faster than npm)
- Cold start in ~8ms
- Low memory footprint

### Built with Bun
- Modular JavaScript architecture (`src/`)
- Zero external dependencies (Bun only)
- Compilable to standalone binary
- Singleton PID file prevents duplicate watchers
- Real-time stats on stop

## Requirements

- **Bun** (installed globally: `curl -fsSL https://bun.sh/install | bash`)
- **Composer** (PHP dependency manager)
- **Drupal** with Drush installed

## Installation

### From Packagist (recommended)

```bash
composer require irving-frias/drupal-watcher
```

### From local repository

1. Clone or download the package to `packages/drupal-watcher/`
2. Add the repository to your `composer.json`:

```json
"repositories": [
    {
        "type": "path",
        "url": "packages/drupal-watcher"
    }
]
```

3. Install:

```bash
composer require irving-frias/drupal-watcher:@dev
```

### From ZIP

1. Download the [latest ZIP](https://github.com/irving-frias/drupal-watcher/archive/refs/heads/main.zip)
2. Extract to `packages/drupal-watcher/`
3. Follow steps from local repository method

## Usage

All commands run from your Drupal project root.

### Start the watcher

```bash
# Local or Lando
vendor/bin/drupal-watcher start

# DDEV
ddev drupal-watcher start
```

Example output:

```
🚀 Starting Drupal Watcher
📁 Drupal root: docroot
🔧 Drush: drush cr
👀 Watching routes:
  - docroot/modules/custom
  - docroot/themes/custom
✔ Watcher active. Waiting for changes... (Ctrl+C to stop)
```

### List configured routes

```bash
vendor/bin/drupal-watcher list
```

Shows current routes, patterns, debounce, and drush command.

### Add a route

```bash
vendor/bin/drupal-watcher add docroot/modules/contrib
```

### Remove a route

```bash
vendor/bin/drupal-watcher remove docroot/modules/contrib
```

### Reset routes to defaults

```bash
vendor/bin/drupal-watcher reset
```

### Check watcher status

```bash
vendor/bin/drupal-watcher status
```

Shows PID and uptime if running.

### Filter routes at startup

```bash
# Watch only a specific route (substring match)
vendor/bin/drupal-watcher start --watch=modules/my-module

# Exclude a specific route (substring match)
vendor/bin/drupal-watcher start --no-watch=modules/contrib

# Abort if Drush is not responding
vendor/bin/drupal-watcher start --abort-on-drush-error

# Dry run — preview what would happen without starting
vendor/bin/drupal-watcher start --dry-run
```

### Global flags

| Flag | Description |
| :--- | :--- |
| `--version`, `-V` | Show version number |
| `--no-colors` | Disable colored output (useful for CI/logs) |
| `--verbose`, `-v` | Show full Drush stdout/stderr output |

### Runtime features

- **Timestamps**: Every change and cache clear is prefixed with `[HH:MM:SS]`
- **Pending counter**: Shows how many unique files are queued in the current debounce window, e.g. `📝 file.php (3 pending)`
- **Cache clear duration**: Shows how long `drush cr` took, e.g. `✔ Cache cleared in 2.3s`
- **Post-clear feedback**: Each post-clear command shows success/failure and duration
- **Shutdown stats**: On Ctrl+C, shows uptime, total changes, unique files, and cache clears
- **SIGTERM handling**: Also responds to `SIGTERM` for clean Docker/process manager shutdowns
- **Reset confirmation**: `vendor/bin/drupal-watcher reset` asks for confirmation before removing custom routes

### Composer script aliases (optional)

Add to your root `composer.json`:

```json
"scripts": {
    "watcher:start": "vendor/bin/drupal-watcher start",
    "watcher:list": "vendor/bin/drupal-watcher list",
    "watcher:status": "vendor/bin/drupal-watcher status",
    "watcher:add": "vendor/bin/drupal-watcher add",
    "watcher:remove": "vendor/bin/drupal-watcher remove",
    "watcher:reset": "vendor/bin/drupal-watcher reset"
}
```

Run with: `composer watcher:start`

## Configuration

The `watcher.config.json` file is auto-created in your project root.

### File structure

```json
{
  "routes": [
    "docroot/modules/custom",
    "docroot/themes/custom"
  ],
  "patterns": [
    ".html.twig", ".inc", ".yml", ".module",
    ".theme", ".php", ".info.yml", ".services.yml"
  ],
  "excludePatterns": [],
  "debounce": 800,
  "drushCmd": null,
  "drushCommand": "cr",
  "drushArgs": [],
  "postClearCommands": [],
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

### Options

| Option | Description | Default |
| :--- | :--- | :--- |
| `routes` | Paths to watch | `["docroot/modules/custom", "docroot/themes/custom"]` |
| `patterns` | File extensions to watch | `[".html.twig", ".inc", ".yml", ".module", ".theme", ".php", ".info.yml", ".services.yml"]` |
| `excludePatterns` | Extensions to ignore (e.g. `[".css", ".js"]`) | `[]` |
| `debounce` | Wait time (ms) before running drush | `800` |
| `drushCmd` | Custom Drush command | `null` (auto-detects) |
| `drushCommand` | Drush subcommand to run | `"cr"` |
| `drushArgs` | Extra arguments for Drush | `[]` |
| `postClearCommands` | Shell commands to run after each cache clear | `[]` |
| `commandsPerPattern` | Map file patterns to specific cache clear commands | see below |

### Notes

- **Patterns**: Add or remove extensions as needed. `.php` is included by default since Drupal modules contain PHP hooks, forms, and controllers.
- **Exclude patterns**: Ignore files like `.css` or `.js` to avoid unnecessary cache clears.
- **Debounce**: Increase for large projects (e.g. `1200` for 1.2s).
- **Custom Drush**: Set `drushCmd` to a specific binary path if needed. In DDEV, run `ddev drupal-watcher` instead.
- **drushCommand**: Use `"cc bin"` for faster partial cache clears instead of full `"cr"`.
- **postClearCommands**: Array of shell commands run after each cache clear (e.g. `["drush cex"]`).
- **commandsPerPattern**: Map file patterns to specific cache clear commands. See [Per-pattern cache commands](#per-pattern-cache-commands).

### Per-pattern cache commands

Instead of running `drush cr` for every change, you can map file patterns to specific, lighter cache clears:

| Pattern | Command | Effect |
| :--- | :--- | :--- |
| `.html.twig` | `cc twig` | Clears Twig template cache |
| `.theme` | `cc theme-registry` | Clears theme registry |
| `.module` | `cc plugin` | Clears plugin/hook discovery cache |
| `.inc` | `cc plugin` | Clears plugin/hook discovery cache |
| `.yml` | `cc plugin` | Clears plugin/hook discovery cache |
| `.php` | `cc plugin` | Clears plugin/hook discovery cache |
| `.info.yml` | `cr` | Full rebuild (module info changes) |
| `.services.yml` | `cr` | Full rebuild (container changes) |

When multiple changed files match different patterns with different commands, the watcher falls back to `drush cr` to ensure everything is refreshed.

Only `.info.yml` and `.services.yml` trigger a full `cr`. Everything else uses `cc plugin` (clears `cache.discovery` — plugin definitions, hooks, services) or specific clears like `cc twig`. This is **significantly faster** than a full rebuild on every change.

Override or extend via `--commands-per-pattern`:

```bash
vendor/bin/drupal-watcher start --commands-per-pattern=.module=cr
```

Or in `watcher.config.json`:

```json
{
  "commandsPerPattern": {
    ".html.twig": "cc twig",
    ".module": "cr"
  }
}
```

> **Note**: User-supplied values are **merged** with defaults — you only need to specify the patterns you want to override.

## Architecture

```
bin/drupal-watcher    # Thin entry point (shebang + import main)
src/
  main.js            # Argument parsing and dispatch
  config.js          # Config load/save, Drupal root detection, PID management
  commands.js        # All CLI commands (start, list, status, add, remove, reset, help)
  drush.js           # Drush resolution, health check, execution
  watcher.js         # File watching, debounce, PID enforcement, stats
  utils.js           # Color constants, Drupal paths, helpers
test/
  config.test.ts     # Unit tests (21 tests, see below)
```

### Key design decisions

- **`root` parameter**: Every function accepts an optional `root` for testability (defaults to `process.cwd()`)
- **Per-root cache**: `_rootCache` Map caches detection results; use `invalidateConfigCache(root)` to reset
- **PID singleton**: `.drupal-watcher.pid` prevents multiple watcher instances
- **No TypeScript**: Pure JS avoids a build step for the Composer distribution
- **import with `.js` extension**: ESM requires explicit file extensions

## Examples

### Example 1: Basic watcher

```bash
# Install
composer require irving-frias/drupal-watcher

# Start (local or Lando)
vendor/bin/drupal-watcher start

# In DDEV
ddev drupal-watcher start

# Edit a .twig file...
📝 my-template.html.twig
🔄 Clearing cache...
✔ Cache cleared.
```

### Example 2: Add contrib modules

```bash
vendor/bin/drupal-watcher add docroot/modules/contrib
vendor/bin/drupal-watcher list
```

### Example 3: Custom cache clear per file type

Use `cc twig` for `.html.twig` and `cr` for everything else:

```bash
vendor/bin/drupal-watcher start --commands-per-pattern=.html.twig="cc twig" --commands-per-pattern=.module=cr
```

Or in `watcher.config.json`:

```json
{
  "commandsPerPattern": {
    ".html.twig": "cc twig",
    ".module": "cr"
  }
}
```

The watcher runs the matching command per changed file. If multiple files match different commands, it falls back to `drush cr`.

### Example 4: Post-clear commands

Automatically run `drush cex` after each change:

```json
{
  "postClearCommands": ["drush cex"]
}
```

### Example 5: Standalone binary

```bash
# From Composer package
bun build --compile ./vendor/irving-frias/drupal-watcher/bin/drupal-watcher --outfile ./drupal-watcher

# Or from local repo
bun run build          # current platform
bun run build:mac      # macOS ARM64
bun run build:linux    # Linux x64
bun run build:win      # Windows x64

./drupal-watcher start
```

### Example 6: Run tests

```bash
bun test              # run once
bun run test:watch    # watch mode
```

### Example 7: Advanced start flags

```bash
vendor/bin/drupal-watcher start --abort-on-drush-error
vendor/bin/drupal-watcher start --watch=docroot/modules/custom/my-module
vendor/bin/drupal-watcher start --no-watch=docroot/modules/contrib
vendor/bin/drupal-watcher start --dry-run
vendor/bin/drupal-watcher start --no-colors
vendor/bin/drupal-watcher start --verbose
```

### Example 8: Display version

```bash
vendor/bin/drupal-watcher --version
# → drupal-watcher v0.3.0

vendor/bin/drupal-watcher -V
```

## Troubleshooting

### ❌ `command not found: bun`

Bun is not installed globally.

1. Check installation: `bun --version`
2. Install: `curl -fsSL https://bun.sh/install | bash`

### ❌ `Drush not found`

The watcher looks for Drush at:
- `vendor/bin/drush` (Drupal project)
- `bin/drush` (alternative)

Verify that:
1. Drush is installed: `composer require drush/drush`
2. You are running from the Drupal project root
3. In DDEV, use `ddev drupal-watcher <command>` instead of `vendor/bin/drupal-watcher <command>`

### ❌ `None of the configured routes exist`

Ensure that:
1. `docroot/modules/custom` and `docroot/themes/custom` exist
2. Or add valid routes with `vendor/bin/drupal-watcher add`

### ❌ Watcher does not detect changes

Verify that:
1. You are editing files with correct extensions (`.html.twig`, `.inc`, `.yml`, `.module`, `.theme`, `.php`, `.info.yml`, `.services.yml`)
2. Files are inside configured routes
3. On large projects, the watcher may take time to initialize

### ❌ Cache is cleared too frequently

Increase `debounce` in `watcher.config.json`:

```json
{
  "debounce": 1200
}
```

## FAQ

### Why Bun instead of Node.js?

Bun is **10-30× faster** for installs and **8× faster** for cold starts. It's an all-in-one runtime, package manager, and bundler, reducing dependencies and complexity.

### Can I use it without Composer?

Yes. Compile to a standalone binary:

```bash
bun build --compile ./bin/drupal-watcher --outfile ./drupal-watcher
./drupal-watcher start
```

### Does it work on Windows?

Yes. Bun is cross-platform. Install from [bun.sh](https://bun.sh).

### Can I watch multiple projects?

Not directly. Each project has its own watcher and configuration. Run from each project's root.

### What files does it watch?

Default patterns: `.html.twig`, `.inc`, `.yml`, `.module`, `.theme`, `.php`, `.info.yml`, `.services.yml`. Add more in `watcher.config.json`.

## Development

### Setup

```bash
git clone <repo>
cd drupal-watcher
bun install
```

### Testing

```bash
bun test              # 21 tests
bun run test:watch    # watch mode
```

### Building

```bash
bun run build          # current platform
bun run build:linux    # Linux x64
bun run build:mac      # macOS ARM64
bun run build:win      # Windows x64
```

### Auto-tagging

The repository includes a GitHub Action (`.github/workflows/tag.yml`) that automatically creates a new tag on push to `main`. It analyzes commits since the last tag:

| Commit type | Bump |
|---|---|
| `BREAKING CHANGE` or `feat!:...` | **major** |
| `feat:...` | **minor** |
| `fix:`, `refactor:`, `chore:`, `ci:`, `docs:`, etc | **patch** |

### Commands reference

| Command | Description |
| :--- | :--- |
| `start` | Start the file watcher |
| `stop` (Ctrl+C / SIGTERM) | Stop the watcher (prints stats) |
| `status` | Show PID and uptime if running |
| `list` | Display current configuration |
| `add <path>` | Add a route to watch |
| `remove <path>` | Remove a watched route |
| `reset` | Reset routes to defaults (with confirmation) |
| `help [command]` | Show detailed help |

### Flags reference

| Flag | Applies to | Description |
| :--- | :--- | :--- |
| `--abort-on-drush-error` | `start` | Exit if Drush health check fails |
| `--watch=<path>` | `start` | Filter routes to those containing `<path>` (substring) |
| `--no-watch=<path>` | `start` | Exclude routes containing `<path>` (substring) |
| `--dry-run` | `start` | Preview configuration without starting the watcher |
| `--verbose`, `-v` | `start` | Show full Drush output |
| `--no-colors` | all | Disable ANSI colors |
| `--version`, `-V` | all | Show version number |
| `--help`, `-h` | all | Show help |
| `--commands-per-pattern=<p>=<c>` | `start` | Map pattern to drush command (repeatable) |

## Contributing

Contributions are welcome!

1. Fork the project
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Reporting issues

Use the [issue tracker](https://github.com/irving-frias/drupal-watcher/issues) for bugs or feature requests.

## License

MIT. See [LICENSE](LICENSE).

---

## Acknowledgments

- [Bun](https://bun.sh) — For incredible speed
- [Drupal](https://www.drupal.org) — The best CMS

---

*Built with ❤️ by [Irving Frías](https://github.com/irving-frias)*
