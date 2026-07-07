# Drupal Watcher — Project Guide

<!-- codebase-memory-mcp:start -->
# Codebase Knowledge Graph (codebase-memory-mcp)

This project uses codebase-memory-mcp to maintain a knowledge graph of the codebase.
ALWAYS prefer MCP graph tools over grep/glob/file-search for code discovery.

## Priority Order
1. `search_graph` — find functions, classes, routes, variables by pattern
2. `trace_path` — trace who calls a function or what it calls
3. `get_code_snippet` — read specific function/class source code
4. `query_graph` — run Cypher queries for complex patterns
5. `get_architecture` — high-level project summary

## When to fall back to grep/glob
- Searching for string literals, error messages, config values
- Searching non-code files (Dockerfiles, shell scripts, configs)
- When MCP tools return insufficient results

## Examples
- Find a handler: `search_graph(name_pattern=".*OrderHandler.*")`
- Who calls it: `trace_path(function_name="OrderHandler", direction="inbound")`
- Read source: `get_code_snippet(qualified_name="pkg/orders.OrderHandler")`
<!-- codebase-memory-mcp:end -->

## Commands
- **Test all**: `go test ./...`
- **Test source only**: `go test ./internal/... ./pkg/...`
- **Test tests/ only**: `go test ./tests/...`
- **Fresh test**: `go test -count=1 ./...`
- **Vet**: `go vet ./...`
- **Build**: `go build -o drupal-watcher ./cmd/drupal-watcher`
- **Run**: `./bin/drupal-watcher <command>` or `go run ./cmd/drupal-watcher <command>`

## Project map
- `project_map.md` — Full dependency graph, data flow, file tree, and exported functions index. **AI agents should read this first** for project orientation
- `scripts/project_map.sh` — Regenerates `project_map.md` from source

## Project structure
- `bin/drupal-watcher` — PHP launcher (`#!/usr/bin/env php`), calls `install.php` then execs Go binary
- `bin/install.php` — Binary downloader (vendor/bin entry managed by Composer via the `bin` field in `composer.json`)
- `cmd/drupal-watcher/main.go` — Modular entry point with DI container + module system
- `internal/app/` — DI setup (`Setup()`/`Shutdown()` via `samber/do/v2`), typed wrappers (`WorkDir`, `DrupalRoot`), `EventBus`
- `internal/app/modules/` — Built-in modules (config, watcher, executor, orchestrator, ui)
- `internal/config/config.go` — `Manager` struct with per-root cache, config load/save, Drupal root detection, PID management
- `internal/drush/drush.go` — Drush command resolution and execution, `DrushConfig` interface
- `internal/app/modules/orchestrator/engine.go` — Engine with EventBus, `EngineConfig`, `NewEngine`
- `internal/ui/` — Bubble Tea TUI (model, view, update, styles, messages, powermode, drupal_logo)
- `pkg/core/` — Domain interfaces (`Watcher`, `CommandExecutor`, `EventFilter`, `LintChecker`)
- `pkg/adapters/` — Adapter implementations (fsnotify, polling_watcher, hybrid_watcher, drush, regex filters, php_lint, yaml_lint, logger)

## Test structure
- `tests/` — External test packages (black-box tests, `package X_test`)
  - `tests/pkg/adapters/` — Adapter tests (drush_executor, fsnotify, php_lint, yaml_lint)
  - `tests/internal/config/` — Config manager tests
  - `tests/internal/drush/` — Drush command tests
  - `tests/internal/health/` — Health check tests
  - `tests/internal/metrics/` — Metrics collection tests
  - `tests/internal/training/` — Training suggestions tests
  - `tests/internal/utils/` — Utility function tests
- White-box tests (access unexported identifiers) stay colocated with source:
  - `pkg/adapters/lint_cache_test.go` — uses unexported `fileChecksum()`
  - `internal/app/modules/orchestrator/engine_test.go` — uses 9 unexported fields/methods
  - `internal/ui/powermode_test.go` — uses 10 unexported fields/methods
  - `internal/validate/validate_test.go` — uses unexported `validateCommand()`, `findPHPCS()`
- Placeholder directories with doc.go templates:
  - `internal/app/modules/executor/adapters/` — future executor adapters (Docker, SSH)
  - `internal/app/modules/watcher/adapters/` — future watcher adapters (inotify, kqueue)
  - `internal/app/modules/ui/providers/cli/` — future non-interactive CLI mode

## Guidelines
- **All user-facing messages in English**
- Functions accept optional `root` parameter for testability (defaults to `os.Getwd()`)
- Caches are per-root via `map[string]*cacheEntry`; use `InvalidateConfigCache(root)` to reset
- Use `Get*` prefix for interface methods to avoid naming conflict with struct fields
- New features should register services via `do.ProvideValue` / `do.Provide` in module `Register(i do.Injector)` functions
- DI uses `samber/do/v2` — type-safe, no reflection, no codegen. `do.Injector` is the container interface
- EventBus (`internal/app/eventbus/`) decouples modules — new consumers subscribe to topics
- PID/starttime files stored in `~/.cache/drupal-watcher/.drupal-watcher-<hash>.pid` with `0600` perms (hash based on project absolute path, supports multiple projects)
- **Releases**: before each release, update `composer.json` → `extra.drupal-watcher-version`. The `build.yml` workflow reads that version, creates the tag, and auto-bumps the patch post-release

## DI services (samber/do/v2 types)
- `*config.Config` — Config struct (provided by config module)
- `core.Watcher` — FS/Polling/Hybrid watcher (provided by watcher module)
- `core.CommandExecutor` — Drush executor (provided by executor module)
- `*orchestrator.Engine` — Engine with EventBus (provided by orchestrator module, implements `Shutdowner`)
- `*eventbus.EventBus` — Event bus (provided at app setup, before modules)
- `common.WorkDir` — `string` wrapper (project root)
- `common.DrupalRoot` — `string` wrapper (absolute Drupal root path)

## phpcs linting
- `PhpCsLintChecker` in `pkg/adapters/phpcs_lint.go` — runs `phpcs` with Drupal standards
- Config field `phpCsStandard`: empty = `php -l`, `"auto"` = detect Drupal 11/10, `"Drupal"` / `"DrupalStrict"` = explicit standard
- Auto-detection reads `composer.json` `require.drupal/core` version constraint
- Falls back to `PhpLintChecker` (`php -l`) if `phpCsStandard` is empty
- PHPCS binary found at `vendor/bin/phpcs`, `../vendor/bin/phpcs`, or `$PATH`

## EventBus topics
- `file.change` — File change detected
- `cache.clear` — Drush cache clear executed
- `error` — Watcher or engine error

## Key types
- `config.Config` — Main configuration struct with all watcher settings (includes `SkipLint`, `LintCommands`, `PhpCsStandard`, `WatchMode`, `PollInterval`, `ShowLogo`)
- `config.Manager` — Config cache and file operations
- `drush.DrushConfig` — Interface for drush operations (satisfied by config.Config)
- `drush.DrushResult` — Result of a drush command execution
- `orchestrator.EngineConfig` — Dependency injection struct for the engine (includes `LintCheckers`, `SkipLint`)
- `core.EngineEvent` — Event emitted on file changes / cache clears (includes `Changes int` field for batch size, used by PowerMode skull detection)
- `core.LintChecker` — Interface for syntax checking before cache clear
- `core.LintResult` — Result of a lint check (file path + error)
- `DrupalLogo` in `internal/ui/drupal_logo.go` — Static "Drupal Watcher" text in Mathematical Bold Fraktur (Pricedown Bold style)
- `PowerMode` in `internal/ui/powermode.go` — Combo counter, energy bar, particle system, overheating levels (Normal→Warm→Hot→Power), cooldown smoke, skull of death at 50+ changes

## Particle types (powermode.go)
- `Spark` — `✦ ⚡ ★` — fast, short life (explosions)
- `Fire` — `🔥 💥 ⚡` — wavy, medium life (overheat)
- `Smoke` — `· ‧ ∘ ° ≈` — slow, long life (cooldown)

## PowerMode levels
- `Normal` (combo 0-2) — idle, no effects
- `Warm` (combo 3-5) — light particles, yellow tint
- `Hot` (combo 6-10) — fire particles, orange border
- `Power` (combo 11+) — explosions, red border
- `Cooldown` — 3s idle → blue border, ❄ icon, smoke rises
- `Skull` — 50+ changes/batch → 💀 icon, white/red flash, skull particles
