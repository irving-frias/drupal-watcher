import { existsSync } from "fs";
import path from "path";
import { P_ERROR, P_INFO, POSSIBLE_DOCROOTS, cyan } from "./utils.js";

function getRoot(r) {
  return r || process.cwd();
}

function configPath(root) {
  return path.join(getRoot(root), "watcher.config.json");
}

function pidPath(root) {
  return path.join(getRoot(root), ".drupal-watcher.pid");
}

const _rootCache = new Map();

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
    const config = JSON.parse(raw);

    if (!config.drupalRoot) {
      const detected = detectDrupalRoot(r);
      if (detected) {
        config.drupalRoot = detected;
        config.routes = config.routes.map(route => {
          const parts = route.split("/");
          if (parts.length > 0 && POSSIBLE_DOCROOTS.includes(parts[0]) && parts[0] !== detected) {
            return route.replace(parts[0], detected);
          }
          return route;
        });
        await saveConfig(config, r);
      }
    }

    const entry = _rootCache.get(r) || {};
    entry.config = { ...getDefaultConfig(r), ...config };
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
}

export async function removePid(root) {
  try {
    await Bun.write(pidPath(root), "");
    const { rm } = await import("fs/promises");
    await rm(pidPath(root), { force: true });
  } catch {}
}

export async function checkPid(root) {
  const file = Bun.file(pidPath(root));
  if (!(await file.exists())) return null;
  const pid = (await file.text()).trim();
  if (!pid) return null;
  const result = Bun.spawnSync(["ps", "-p", pid, "-o", "pid="]);
  return result.exitCode === 0 ? pid : "stale";
}

export function invalidateConfigCache(root) {
  const r = getRoot(root);
  _rootCache.delete(r);
}
