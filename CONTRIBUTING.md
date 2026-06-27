# Contributing

Thanks for your interest in contributing to Drupal Watcher.

## Getting started

Requires Go 1.25+. Clone the repo and run:

```bash
go test -count=1 ./...
go vet ./...
go build -o drupal-watcher ./cmd/drupal-watcher
```

## Architecture

See [Architecture section in README.md](README.md#architecture) for the hexagonal (ports & adapters) layout and the module system.

Key points:

- `cmd/drupal-watcher/` — production binary entry point
- `cmd/modular/` — modular dev binary with DI container + module system
- `pkg/core/` — domain interfaces (`Watcher`, `CommandExecutor`, etc.)
- `pkg/adapters/` — adapter implementations
- `internal/app/` — module system (`Container`, `Module` interface, lifecycle)
- `internal/app/modules/` — built-in modules
- `internal/app/eventbus/` — pub/sub event bus for decoupling modules

## Adding a feature

New functionality should use the module system (`internal/app/`) with the `app.Module` interface when possible:

```go
type Module interface {
    Name() string
    DependsOn() []Module
    Init(container *Container) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

See the [Module System docs in README.md](README.md#module-system-for-developers) for examples (notifications module, custom executor, etc.).

## Guidelines

- **All user-facing messages in English**
- Functions accept optional `root` parameter for testability (defaults to `os.Getwd()`)
- Use `Get*` prefix for interface methods to avoid naming conflict with struct fields
- Caches are per-root via `map[string]*cacheEntry`; use `InvalidateConfigCache(root)` to reset
- Interface method names use `Get*` prefix
- New modules subscribe to EventBus topics rather than depending on other modules directly

## Before submitting

1. Run `go vet ./...` — no warnings
2. Run `go test -count=1 ./...` — all tests pass
3. If you added a flag, update the CLI docs in README.md
4. If you added a config field, update the config docs in README.md

## Commit style

Use conventional commit messages when possible:

```
feat: add twig debug mode toggling
fix: handle empty routes in config
docs: update watch mode documentation
refactor: extract drush path resolution
```

## Code of Conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md). By participating, you agree to uphold its guidelines.
