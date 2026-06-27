# Drupal Watcher — Project Guide

## Commands
- **Test**: `go test ./...`
- **Fresh test**: `go test -count=1 ./...`
- **Vet**: `go vet ./...`
- **Build**: `go build -o drupal-watcher ./cmd/drupal-watcher`
- **Run**: `./bin/drupal-watcher <command>` or `go run ./cmd/drupal-watcher <command>`

## Project structure
- `bin/drupal-watcher` — PHP launcher (`#!/usr/bin/env php`), calls `install.php` then execs Go binary
- `bin/install.php` — Binary downloader (vendor/bin entry managed by Composer via the `bin` field in `composer.json`)
- `cmd/drupal-watcher/main.go` — Modular entry point with DI container + module system
- `internal/app/` — Module system (`Container`, `Module` interface, `App` lifecycle, `EventBus`)
- `internal/app/modules/` — Built-in modules (config, watcher, executor, orchestrator, ui)
- `internal/config/config.go` — `Manager` struct with per-root cache, config load/save, Drupal root detection, PID management
- `internal/drush/drush.go` — Drush command resolution and execution, `DrushConfig` interface
- `internal/app/modules/orchestrator/engine.go` — Engine with EventBus
- `internal/ui/` — Bubble Tea TUI (model, view, update, styles, messages)
- `pkg/core/` — Domain interfaces (`Watcher`, `CommandExecutor`, `EventFilter`, `PostProcessor`, `LintChecker`)
- `pkg/adapters/` — Adapter implementations (fsnotify, polling_watcher, hybrid_watcher, drush, regex filters, php_lint, yaml_lint, logger)

## Guidelines
- **All user-facing messages in English**
- Functions accept optional `root` parameter for testability (defaults to `os.Getwd()`)
- Caches are per-root via `map[string]*cacheEntry`; use `InvalidateConfigCache(root)` to reset
- Use `Get*` prefix for interface methods to avoid naming conflict with struct fields
- New features should use the module system (`internal/app/`) with `app.Module` interface
- Modules register services in the `Container` via `Init()`; services are identified by `common.ServiceName`
- EventBus (`internal/app/eventbus/`) decouples modules — new consumers subscribe to topics
- PID/starttime files stored in `~/.cache/drupal-watcher/.drupal-watcher-<hash>.pid` with `0600` perms (hash based on project absolute path, supports multiple projects)
- **Releases**: before each release, update `composer.json` → `extra.drupal-watcher-version`. The `build.yml` workflow reads that version, creates the tag, and auto-bumps the patch post-release

## Service names (common.ServiceName)
- `SvcConfig` — `*config.Config`
- `SvcWatcher` — `core.Watcher` (FSNotifyWatcher)
- `SvcExecutor` — `core.CommandExecutor` (DrushExecutor)
- `SvcOrchestrator` — `*orchestrator.Engine`
- `SvcEventBus` — `*eventbus.EventBus`
- `SvcWorkDir` — `string` (project root)
- `SvcDrupalRoot` — `string` (absolute Drupal root path)

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
- `config.Config` — Main configuration struct with all watcher settings (includes `SkipLint`, `LintCommands`, `PhpCsStandard`, `WatchMode`, `PollInterval`)
- `config.Manager` — Config cache and file operations
- `drush.DrushConfig` — Interface for drush operations (satisfied by config.Config)
- `drush.DrushResult` — Result of a drush command execution
- `core.EngineConfig` — Dependency injection struct for the engine (includes `LintCheckers`, `SkipLint`)
- `core.EngineEvent` — Event emitted on file changes / cache clears
- `core.LintChecker` — Interface for syntax checking before cache clear
- `core.LintResult` — Result of a lint check (file path + error)

## Migration notes (from TS to Go)
- Replaced `bun:test` with `go test`
- Replaced `fs.watch`/`fs.watchFile` with `fsnotify`
- Replaced `exec`/child_process with `os/exec`
- PID files use `syscall.Kill(pid, 0)` for process check
- Colors use pterm (replaced raw ANSI escape codes)
- Interface methods use `Get*` prefix to avoid Go field/method name conflicts
- Notifications use `beeep` (cross-platform Go library)
- Launcher is PHP (not bash) to avoid `/bin/bash` dependency on servers
