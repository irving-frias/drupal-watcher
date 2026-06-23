import path from "path";
import { RED, GREEN, YELLOW, BLUE, NC, CYAN, EXCLUDED_DIRS, timestamp, green, cyan, yellow, blue } from "./utils.js";
import { getDrushSpawnArgs, runDrush, runPostClearCommands } from "./drush.js";

// --- Runtime stats ---
export const stats = {
  changes: 0,
  clears: 0,
  startTime: null,
  filesChanged: new Set(),
};

// --- Debounce state ---
let debounceTimer = null;
let changeAccumulator = new Map();

export function resetDebounce() {
  if (debounceTimer) {
    clearTimeout(debounceTimer);
    debounceTimer = null;
  }
  changeAccumulator.clear();
}

async function runCacheClear(drushBase, drushArgsArray, postClearCommands) {
  const files = Array.from(changeAccumulator.keys());
  changeAccumulator.clear();
  if (files.length === 0) return;

  const suffix = files.length > 1 ? ` (${files.length} files)` : "";
  console.log(`${timestamp()} ${yellow("🔄")} Clearing cache...${suffix}`);

  const { exitCode, stderr, duration } = await runDrush(drushBase, drushArgsArray);

  if (exitCode === 0) {
    console.log(`${timestamp()} ${green("✔")} Cache cleared in ${duration}s`);
    stats.clears++;
  } else {
    console.error(`${timestamp()} ${red("✖")} Drush error (exit ${exitCode}): ${stderr || "drush unavailable"}`);
  }

  await runPostClearCommands(postClearCommands);
}

function scheduleCacheClear(config) {
  if (debounceTimer) clearTimeout(debounceTimer);
  const { cmd, args } = getDrushSpawnArgs(config);
  debounceTimer = setTimeout(() => {
    debounceTimer = null;
    runCacheClear(cmd, args, config.postClearCommands || []);
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

function formatChangePath(rootPath, changePath) {
  const rel = path.relative(rootPath, path.resolve(rootPath, changePath));
  if (rel === changePath) return path.basename(changePath);
  if (rel.length < 40) return rel;
  return `...${rel.slice(-37)}`;
}

// --- Watcher ---
let watcherHandle = null;

export function getWatcherHandle() {
  return watcherHandle;
}

export async function startWatcher(config) {
  const drupalRoot = config.drupalRoot;
  const rootPath = path.join(process.cwd(), drupalRoot);
  const routeSuffixes = config.routes.map(r =>
    r.replace(`${drupalRoot}/`, "").replace(drupalRoot, "")
  );

  function onChange(changePath, eventType) {
    if (!changePath) return;

    const normalized = path.normalize(String(changePath));

    if (EXCLUDED_DIRS.some(dir => normalized.startsWith(dir + "/") || normalized === dir)) return;
    if (!routeSuffixes.some(s => normalized.startsWith(s))) return;
    if (!config.patterns.some(p => changePath.endsWith(p))) return;
    if (config.excludePatterns?.some(p => changePath.endsWith(p))) return;

    if (!changeAccumulator.has(changePath)) {
      const displayName = formatChangePath(rootPath, changePath);
      const pending = changeAccumulator.size + 1;
      console.log(`${timestamp()} ${green("📝")} ${displayName} ${cyan(`(${pending} pending)`)}`);
      stats.changes++;
      stats.filesChanged.add(changePath);
    }
    changeAccumulator.set(changePath, Date.now());
    scheduleCacheClear(config);
  }

  function onError(err) {
    console.error(`${timestamp()} ${red("✖")} Watcher error: ${err?.message || err}`);
  }

  try {
    watcherHandle = Bun.watch({
      paths: [rootPath],
      recursive: true,
      onChange: (info) => {
        if (info?.path) onChange(info.path, info.type);
      },
      onError,
    });
  } catch {
    const { watch } = await import("fs");
    console.log(`${timestamp()} ${cyan("ℹ")} Bun.watch unavailable, falling back to fs.watch`);
    watcherHandle = watch(rootPath, { recursive: true }, (eventType, filename) => {
      if (filename) onChange(filename, eventType);
    });
    if (watcherHandle && typeof watcherHandle.on === "function") {
      watcherHandle.on("error", onError);
    }
  }

  stats.startTime = Date.now();
  return watcherHandle;
}

export function stopWatcher() {
  if (watcherHandle) {
    try {
      if (typeof watcherHandle.stop === "function") watcherHandle.stop();
      else if (typeof watcherHandle.close === "function") watcherHandle.close();
    } catch {}
    watcherHandle = null;
  }
}
