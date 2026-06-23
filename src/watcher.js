import path from "path";
import { RED, GREEN, YELLOW, BLUE, NC, EXCLUDED_DIRS } from "./utils.js";
import { getDrushSpawnArgs } from "./drush.js";
import { runDrush, runPostClearCommands } from "./drush.js";

// --- Stats ---
export const stats = {
  changes: 0,
  clears: 0,
  startTime: null,
  filesChanged: new Set(),
};

// --- Adaptive debounce ---
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

  const suffix = files.length > 1 ? ` (${files.length} archivos)` : "";
  console.log(`${YELLOW}🔄 Limpiando caché...${suffix}${NC}`);

  const { exitCode, stderr } = await runDrush(drushBase, drushArgsArray);

  if (exitCode === 0) {
    console.log(`${GREEN}✅ Caché limpiada correctamente.${NC}`);
    stats.clears++;
  } else {
    console.error(`${RED}❌ Error drush (exit ${exitCode}):${NC} ${stderr || "drush no disponible"}`);
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
  console.log(`\n${BLUE}📊 Estadísticas del watcher:${NC}`);
  console.log(`  ${YELLOW}Tiempo activo:${NC} ${mins}m ${secs}s`);
  console.log(`  ${YELLOW}Cambios detectados:${NC} ${stats.changes}`);
  console.log(`  ${YELLOW}Archivos únicos:${NC} ${stats.filesChanged.size}`);
  console.log(`  ${YELLOW}Limpiezas de caché:${NC} ${stats.clears}`);
}

// --- Watcher creation ---
let watcherHandle = null;

export function getWatcherHandle() {
  return watcherHandle;
}

export async function startWatcher(config) {
  const drupalRoot = config.drupalRoot;
  const rootPath = path.join(process.cwd(), drupalRoot);

  const routeSuffixes = config.routes.map(r => r.replace(`${drupalRoot}/`, "").replace(drupalRoot, ""));

  const onChange = (info) => {
    if (!info?.path) return;
    const normalized = path.normalize(info.path);

    // Ignore system directories
    if (EXCLUDED_DIRS.some(dir => normalized.startsWith(dir + "/") || normalized === dir)) return;

    // Check if in watched routes
    if (!routeSuffixes.some(suffix => normalized.startsWith(suffix))) return;

    // Check pattern match
    if (!config.patterns.some(p => info.path.endsWith(p))) return;

    // Check exclude patterns
    if (config.excludePatterns?.some(p => info.path.endsWith(p))) return;

    // Track change
    if (!changeAccumulator.has(info.path)) {
      const displayName = path.basename(info.path);
      const total = changeAccumulator.size + 1;
      console.log(`${GREEN}📝 ${displayName}${total > 1 ? ` (${total})` : ""}${NC}`);
      stats.changes++;
      stats.filesChanged.add(info.path);
    }
    changeAccumulator.set(info.path, Date.now());
    scheduleCacheClear(config);
  };

  const onError = (err) => {
    console.error(`${RED}❌ Error en watcher:${NC} ${err?.message || err}`);
  };

  try {
    watcherHandle = Bun.watch({
      paths: [rootPath],
      recursive: true,
      onChange,
      onError,
    });
  } catch {
    const { watch: fsWatch } = await import("fs");
    watcherHandle = fsWatch(rootPath, { recursive: true }, (eventType, filename) => {
      if (!filename) return;
      onChange({ path: filename, type: eventType });
    });
    watcherHandle.on("error", onError);
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
