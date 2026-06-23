# Drupal Watcher - Project Guide

## Commands
- **Test**: `bun test`
- **Watch mode**: `bun run test:watch`
- **Run**: `vendor/bin/drupal-watcher <command>` or `bun run bin/drupal-watcher <command>`
- **Build standalone**: `bun run build`

## Project structure
- `bin/drupal-watcher` — Thin entry point (shebang + import main)
- `src/main.js` — Arg parsing and dispatch
- `src/config.js` — Config load/save, Drupal root and environment detection
- `src/commands.js` — All CLI commands (start, list, status, add, remove, reset, help)
- `src/drush.js` — Drush command resolution and execution
- `src/watcher.js` — File watcher, debounce, PID management, stats
- `src/utils.js` — Color constants, shared helpers
- `test/config.test.ts` — Unit tests for config module

## Guidelines
- Keep JS (no TypeScript) to avoid build step
- Functions accept optional `root` parameter for testability (defaults to `process.cwd()`)
- Use `import { detectDrupalRoot } from "../src/config.js"` style imports (with `.js` extension)
- Caches are per-root via `_rootCache` Map; use `invalidateConfigCache(root)` to reset
- All paths relative to project root
