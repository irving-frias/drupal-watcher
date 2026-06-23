import path from "path";
import { RED, GREEN, YELLOW, BLUE, NC, EXCLUDED_DIRS } from "./utils.js";
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
  console.log(`${YELLOW}🔄 Clearing cache...${suffix}${NC}`);

  const { exitCode, stderr } = await runDrush(drushBase, drushArgsArray);

  if (exitCode === 0) {
    console.log(`${GREEN}✔ Cache cleared.${NC}`);
    stats.clears++;
  } else {
    console.error(`${RED}✖ Drush error (exit ${exitCode}):${NC} ${stderr || "drush unavailable"}`);
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

  console.log(`\n${BLUE}📊 Watcher stats:${NC}`);
  console.log(`  ${YELLOW}Uptime:${NC} ${mins}m ${secs}s`);
  console.log(`  ${YELLOW}Changes detected:${NC} ${stats.changes}`);
  console.log(`  ${YELLOW}Unique files:${NC} ${stats.filesChanged.size}`);
  console.log(`  ${YELLOW}Cache clears:${NC} ${stats.clears}`);
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
      const displayName = path.basename(changePath);
      const total = changeAccumulator.size + 1;
      console.log(`${GREEN}📝 ${displayName}${total > 1 ? ` (${total})` : ""}${NC}`);
      stats.changes++;
      stats.filesChanged.add(changePath);
    }
    changeAccumulator.set(changePath, Date.now());
    scheduleCacheClear(config);
  }

  function onError(err) {
    console.error(`${RED}✖ Watcher error:${NC} ${err?.message || err}`);
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
