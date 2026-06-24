import path from "path";
import { EXCLUDED_DIRS, timestamp, green, red, cyan, yellow, blue } from "./utils";
import { getDrushCommand, getDrushSpawnArgs, runDrush, runPostClearCommands } from "./drush";
import type { WatcherConfig, WatcherHandle } from "./types";

// --- Runtime stats ---
export const stats = {
  changes: 0,
  clears: 0,
  startTime: null as number | null,
  filesChanged: new Set<string>(),
};

// --- Debounce state ---
let debounceTimer: ReturnType<typeof setTimeout> | null = null;
let changeAccumulator = new Map<string, number>();

export function resetDebounce() {
  if (debounceTimer) {
    clearTimeout(debounceTimer);
    debounceTimer = null;
  }
  changeAccumulator.clear();
}

function getCacheClearArgs(config: WatcherConfig, files: string[]) {
  const commandsPerPattern = config.commandsPerPattern || {};
  const patternEntries = Object.entries(commandsPerPattern).sort((a, b) => b[0].length - a[0].length);
  const matchedCommands = new Set<string>();

  for (const file of files) {
    for (const [pattern, command] of patternEntries) {
      if (file.endsWith(pattern)) {
        matchedCommands.add(command);
        break;
      }
    }
  }

  if (matchedCommands.size === 1) {
    const specificCommand = [...matchedCommands][0];
    const drushCmdStr = getDrushCommand(config);
    const parts = drushCmdStr.split(/\s+/);
    const specificParts = specificCommand.split(/\s+/);
    return {
      cmd: parts[0],
      args: [...parts.slice(1), ...specificParts, ...(config.drushArgs || [])],
    };
  }

  return getDrushSpawnArgs(config);
}

async function runCacheClear(config: WatcherConfig) {
  const pendingChanges = changeAccumulator;
  changeAccumulator = new Map();
  const files = Array.from(pendingChanges.keys());
  if (files.length === 0) return;

  const { cmd, args } = getCacheClearArgs(config, files);
  const suffix = files.length > 1 ? ` (${files.length} files)` : "";
  const cmdLabel = [...new Set([cmd, ...args])].join(" ");
  console.log(`${timestamp()} ${yellow("🔄")} ${cyan(cmdLabel)}${suffix}`);

  const { exitCode, stderr, duration } = await runDrush(cmd, args);

  if (exitCode === 0) {
    console.log(`${timestamp()} ${green("✔")} Cache cleared in ${duration}s`);
    stats.clears++;
  } else {
    console.error(`${timestamp()} ${red("✖")} Drush error (exit ${exitCode}): ${stderr || "drush unavailable"}`);
  }

  await runPostClearCommands(config.postClearCommands || []);
}

function scheduleCacheClear(config: WatcherConfig) {
  if (debounceTimer) clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    debounceTimer = null;
    runCacheClear(config);
  }, config.debounce || 800);
}

export function printStats() {
  if (!stats.startTime) return;
  const elapsed = Math.round((Date.now() - stats.startTime) / 1000);
  const mins = Math.floor(elapsed / 60);
  const secs = elapsed % 60;

  console.log(`\n${blue("📊 Watcher stats:")}`);
  console.log(`  ${yellow("Uptime:")} ${mins}m ${secs}s`);
  console.log(`  ${yellow("Changes detected:")} ${stats.changes}`);
  console.log(`  ${yellow("Unique files:")} ${stats.filesChanged.size}`);
  console.log(`  ${yellow("Cache clears:")} ${stats.clears}`);
}

function formatChangePath(rootPath: string, changePath: string) {
  const rel = path.relative(rootPath, path.resolve(rootPath, changePath));
  if (rel === changePath) return path.basename(changePath);
  if (rel.length < 40) return rel;
  return `...${rel.slice(-37)}`;
}

// --- Watcher ---
let watcherHandle: WatcherHandle | null = null;

export function getWatcherHandle() {
  return watcherHandle;
}

export async function startWatcher(config: WatcherConfig) {
  const watchPaths = config.routes.map(r => path.join(process.cwd(), r));

  function onChange(changePath: string, _eventType?: string) {
    if (!changePath) return;

    const normalized = path.normalize(String(changePath));

    if (EXCLUDED_DIRS.some(dir => normalized.startsWith(dir + "/") || normalized === dir)) return;
    if (!config.patterns.some(p => changePath.endsWith(p))) return;
    if (config.excludePatterns?.some(p => changePath.endsWith(p))) return;

    if (!changeAccumulator.has(changePath)) {
      const displayName = formatChangePath(process.cwd(), changePath);
      const pending = changeAccumulator.size + 1;
      console.log(`${timestamp()} ${green("📝")} ${displayName} ${cyan(`(${pending} pending)`)}`);
      stats.changes++;
      stats.filesChanged.add(changePath);
    }
    changeAccumulator.set(changePath, Date.now());
    scheduleCacheClear(config);
  }

  function onError(err: Error) {
    console.error(`${timestamp()} ${red("✖")} Watcher error: ${err?.message || err}`);
  }

  try {
    watcherHandle = (Bun as any).watch({
      paths: watchPaths,
      recursive: true,
      onChange: (info: { path?: string; type?: string }) => {
        if (info?.path) onChange(info.path, info.type);
      },
      onError,
    }) as WatcherHandle;
  } catch {
    const { watch } = await import("fs");
    console.log(`${timestamp()} ${cyan("ℹ")} Bun.watch unavailable, falling back to fs.watch`);
    const watchers = watchPaths.map(p => watch(p, { recursive: true }, (eventType: string, filename: string | null) => {
      if (filename) onChange(filename, eventType);
    }));
    watcherHandle = {
      close: () => watchers.forEach(w => { try { w.close(); } catch {} }),
    };
  }

  stats.startTime = Date.now();
  return watcherHandle;
}

export function stopWatcher() {
  if (watcherHandle) {
    try {
      if (typeof watcherHandle.stop === "function") watcherHandle.stop();
      else if (typeof watcherHandle.close === "function") watcherHandle.close();
    } catch (e: unknown) {
      console.warn(`Failed to stop watcher: ${e instanceof Error ? e.message : e}`);
    }
    watcherHandle = null;
  }
}
