# Drupal Watcher — Project Guide (Go Migration)

## Commands
- **Test**: `go test ./...`
- **Fresh test**: `go test -count=1 ./...`
- **Vet**: `go vet ./...`
- **Build**: `go build -o drupal-watcher ./cmd/drupal-watcher`
- **Run**: `./bin/drupal-watcher <command>` or `go run ./cmd/drupal-watcher <command>`

## Project structure
- `bin/drupal-watcher` — Shell launcher (runs `install.php` to download Go binary, then execs it)
- `bin/install.php` — Binary downloader (vendor/bin entry managed by Composer via the `bin` field in `composer.json`)
- `cmd/drupal-watcher/main.go` — Entry point, arg parsing (`parseFlags`), dispatch (switch-based)
- `internal/config/config.go` — `Manager` struct with per-root cache, config load/save, Drupal root detection, PID management
- `internal/drush/drush.go` — Drush command resolution and execution, `DrushConfig` interface
- `internal/watcher/watcher.go` — fsnotify file watcher, debounce, config-aware cache clear args
- `internal/cli/cli.go` — All CLI commands (`CmdStart`, `CmdList`, `CmdStatus`, etc.)
- `internal/utils/utils.go` — Color helpers (pterm-based), shared helpers
- All config package tests in `internal/config/config_test.go`
- All cli package tests in `internal/cli/cli_test.go`
- All drush package tests in `internal/drush/drush_test.go`
- cmd tests in `cmd/drupal-watcher/main_test.go`

## Guidelines
- **All user-facing messages in English**
- Functions accept optional `root` parameter for testability (defaults to `os.Getwd()`)
- Caches are per-root via `map[string]*cacheEntry`; use `InvalidateConfigCache(root)` to reset
- Use `Get*` prefix for interface methods to avoid naming conflict with struct fields
- All paths relative to project root unless otherwise specified
- Convention: `CmdStart`, `CmdList`, `CmdStatus` etc. for command functions
- Config satisfies both `watcher.Config` and `drush.DrushConfig` interfaces via method set
- PID/starttime files stored in project root (`cwd/`) by default, or in `root` if specified
- **Releases**: before each release, update `composer.json` → `extra.drupal-watcher-version`. The `build.yml` workflow reads that version, creates the tag, and auto-bumps the patch post-release

## Key types
- `config.Config` — Main configuration struct with all watcher settings
- `config.Manager` — Config cache and file operations
- `drush.DrushConfig` — Interface for drush operations (satisfied by config.Config)
- `drush.DrushResult` — Result of a drush command execution
- `watcher.Config` — Interface for watcher operations (satisfied by config.Config)
- `watcher.Stats` — Runtime statistics (atomic counters for changes/clears)
- `watcher.Handle` — Watcher handle with stop channel and references

## Migration notes (from TS to Go)
- Replaced `bun:test` with `go test`
- Replaced `fs.watch`/`fs.watchFile` with `fsnotify`
- Replaced `exec`/child_process with `os/exec`
- PID files use `syscall.Kill(pid, 0)` for process check
- Colors use pterm (replaced raw ANSI escape codes)
- Interface methods use `Get*` prefix to avoid Go field/method name conflicts
