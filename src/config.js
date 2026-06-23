import { existsSync } from "fs";
import path from "path";
import { P_ERROR, P_WARN, P_INFO, P_SUCCESS, POSSIBLE_DOCROOTS, cyan, yellow } from "./utils.js";

function getRoot(r) {
  return r || process.cwd();
}

const _rootCache = new Map();
let _customConfigPath = null;

export function setCustomConfigPath(p) {
  _customConfigPath = p;
}

function configPath(root) {
  if (_customConfigPath) return _customConfigPath;
  return path.join(getRoot(root), "watcher.config.json");
}

function pidPath(root) {
  return path.join(getRoot(root), ".drupal-watcher.pid");
}

function starttimePath(root) {
  return path.join(getRoot(root), ".drupal-watcher.starttime");
}

export function detectDrupalRoot(root) {
  const r = getRoot(root);
  const cached = _rootCache.get(r);
  if (cached && "root" in cached) return cached.root;

  for (const dir of POSSIBLE_DOCROOTS) {
    const fullPath = path.join(r, dir);
    if (!existsSync(fullPath)) continue;
    if (
      existsSync(path.join(fullPath, "core")) ||
      existsSync(path.join(fullPath, "modules")) ||
      existsSync(path.join(fullPath, "themes")) ||
      existsSync(path.join(fullPath, "index.php"))
    ) {
      const entry = _rootCache.get(r) || {};
      entry.root = dir;
      _rootCache.set(r, entry);
      return dir;
    }
  }

  const entry = _rootCache.get(r) || {};
  entry.root = null;
  _rootCache.set(r, entry);
  return null;
}

export function getDefaultConfig(root) {
  const drupalRoot = detectDrupalRoot(root) || "docroot";
  return {
    routes: [`${drupalRoot}/modules/custom`, `${drupalRoot}/themes/custom`],
    patterns: [".html.twig", ".inc", ".yml", ".module", ".theme", ".php", ".info.yml", ".services.yml"],
    excludePatterns: [],
    debounce: 800,
    drushCmd: null,
    drushCommand: "cr",
    drushArgs: [],
    postClearCommands: [],
    commandsPerPattern: {},
    drupalRoot,
  };
}

export async function loadConfig(root) {
  const r = getRoot(root);
  const cached = _rootCache.get(r);
  if (cached && cached.config) return cached.config;

  const file = Bun.file(configPath(r));

  if (!(await file.exists())) {
    const def = getDefaultConfig(r);
    await Bun.write(configPath(r), JSON.stringify(def, null, 2));
    console.log(`${P_INFO} Created ${cyan("watcher.config.json")} with defaults.`);
    const entry = _rootCache.get(r) || {};
    entry.config = { ...def };
    _rootCache.set(r, entry);
    return entry.config;
  }

  try {
    const raw = await file.text();
    const parsed = JSON.parse(raw);

    if (!parsed.drupalRoot) {
      const detected = detectDrupalRoot(r);
      if (detected) {
        parsed.drupalRoot = detected;
        parsed.routes = parsed.routes.map(route => {
          const parts = route.split("/");
          if (parts.length > 0 && POSSIBLE_DOCROOTS.includes(parts[0]) && parts[0] !== detected) {
            return route.replace(parts[0], detected);
          }
          return route;
        });
        await saveConfig(parsed, r);
      }
    }

    const entry = _rootCache.get(r) || {};
    entry.config = validateConfig({ ...getDefaultConfig(r), ...parsed }, r);
    _rootCache.set(r, entry);
    return entry.config;
  } catch {
    console.error(`${P_ERROR} Failed to read ${cyan("watcher.config.json")}. The file may be corrupted. Using default configuration.`);
    const def = getDefaultConfig(r);
    const entry = _rootCache.get(r) || {};
    entry.config = { ...def };
    _rootCache.set(r, entry);
    return entry.config;
  }
}

export function validateConfig(config, root) {
  const defaults = getDefaultConfig(root);
  if (!Array.isArray(config.routes)) config.routes = defaults.routes;
  if (!Array.isArray(config.patterns)) config.patterns = defaults.patterns;
  if (!Array.isArray(config.excludePatterns)) config.excludePatterns = defaults.excludePatterns;
  if (typeof config.debounce !== "number" || config.debounce <= 0) config.debounce = defaults.debounce;
  if (typeof config.drushCmd !== "string" && config.drushCmd !== null) config.drushCmd = defaults.drushCmd;
  if (typeof config.drushCommand !== "string") config.drushCommand = defaults.drushCommand;
  if (!Array.isArray(config.drushArgs)) config.drushArgs = defaults.drushArgs;
  if (!Array.isArray(config.postClearCommands)) config.postClearCommands = defaults.postClearCommands;
  if (typeof config.commandsPerPattern !== "object" || Array.isArray(config.commandsPerPattern)) config.commandsPerPattern = defaults.commandsPerPattern;
  if (typeof config.drupalRoot !== "string") config.drupalRoot = defaults.drupalRoot;
  config.routes = config.routes.map(r => path.normalize(r).replace(/\/+$/, ""));
  return config;
}

export async function saveConfig(config, root) {
  const r = getRoot(root);
  await Bun.write(configPath(r), JSON.stringify(config, null, 2));
  const cached = _rootCache.get(r);
  if (cached) cached.config = null;
}

// --- PID file ---
export function getPidFile(root) {
  return pidPath(root);
}

export async function writePid(root) {
  await Bun.write(pidPath(root), String(process.pid));
  await writeStarttime(root);
}

export async function removePid(root) {
  try {
    await Bun.write(pidPath(root), "");
    const { rm } = await import("fs/promises");
    await rm(pidPath(root), { force: true });
  } catch (e) {
    console.warn(`${P_WARN} Failed to remove PID file: ${e.message}`);
  }
  await removeStarttime(root);
}

// --- Start time file (for uptime) ---
export async function writeStarttime(root) {
  const r = getRoot(root);
  await Bun.write(starttimePath(r), String(Date.now()));
}

export async function getStarttime(root) {
  const r = getRoot(root);
  const file = Bun.file(starttimePath(r));
  if (!(await file.exists())) return null;
  const t = (await file.text()).trim();
  return t ? parseInt(t, 10) : null;
}

export async function removeStarttime(root) {
  const r = getRoot(root);
  try {
    const { rm } = await import("fs/promises");
    await rm(starttimePath(r), { force: true });
  } catch (e) {
    console.warn(`${P_WARN} Failed to remove starttime file: ${e.message}`);
  }
}

export async function checkPid(root) {
  const file = Bun.file(pidPath(root));
  if (!(await file.exists())) return null;
  const pid = (await file.text()).trim();
  if (!pid) return null;
  try {
    process.kill(parseInt(pid, 10), 0);
    return pid;
  } catch {
    return "stale";
  }
}

export function invalidateConfigCache(root) {
  const r = getRoot(root);
  _rootCache.delete(r);
}
