# Drupal Watcher — Project Guide

## Commands
- **Test**: `bun test` (34 tests)
- **Watch mode**: `bun run test:watch`
- **Run**: `vendor/bin/drupal-watcher <command>` or `bun run bin/drupal-watcher <command>`
- **Build standalone**: `bun run build`

## Project structure
- `bin/drupal-watcher` — Thin entry point (shebang + import main)
- `src/main.js` — Arg parsing (`parseFlags`) and dispatch (switch-based)
- `src/config.js` — Config load/save, Drupal root detection, PID management
- `src/commands.js` — All CLI commands (start, list, status, add, remove, reset, help)
- `src/drush.js` — Drush command resolution and execution
- `src/watcher.js` — File watcher, debounce, PID enforcement, runtime stats
- `src/utils.js` — Color constants, Drupal path lists, shared helpers
- `test/config.test.ts` — Unit tests for config, drush, and main modules

## Guidelines
- Keep JS (no TypeScript) to avoid build step
- All user-facing messages in **English**
- Functions accept optional `root` parameter for testability (defaults to `process.cwd()`)
- Use `import { detectDrupalRoot } from "../src/config.js"` style (with `.js` extension)
- Caches are per-root via `_rootCache` Map; use `invalidateConfigCache(root)` to reset
- All paths relative to project root
- Use `parseFlags()` from `main.js` for CLI flag parsing in tests
- Convention: `cmdHelp`, `cmdStart`, `cmdList` etc. for command functions
