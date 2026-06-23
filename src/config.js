import { existsSync } from "fs";
import path from "path";
import { RED, NC, POSSIBLE_DOCROOTS } from "./utils.js";

function getRoot(r) {
  return r || process.cwd();
}

function getConfigFile(root) {
  return path.join(getRoot(root), "watcher.config.json");
}

function getPidFilePath(root) {
  return path.join(getRoot(root), ".drupal-watcher.pid");
}

// --- Memoized caches (per-root) ---
const _rootCache = new Map();

export function detectDrupalRoot(root) {
  const r = getRoot(root);
  if (_rootCache.has(r) && _rootCache.get(r).root) return _rootCache.get(r).root;
  for (const dir of POSSIBLE_DOCROOTS) {
    const fullPath = path.join(r, dir);
    if (!existsSync(fullPath)) continue;
    if (existsSync(path.join(fullPath, "core")) ||
        existsSync(path.join(fullPath, "modules")) ||
        existsSync(path.join(fullPath, "themes"))) {
      if (!_rootCache.has(r)) _rootCache.set(r, {});
      _rootCache.get(r).root = dir;
      return dir;
    }
  }
  if (!_rootCache.has(r)) _rootCache.set(r, {});
  _rootCache.get(r).root = null;
  return null;
}

export function detectEnvironment(root) {
  const r = getRoot(root);
  if (_rootCache.has(r) && _rootCache.get(r).env) return _rootCache.get(r).env;
  let env;
  if (existsSync(path.join(r, ".ddev"))) env = "ddev";
  else if (existsSync(path.join(r, ".lando.yml")) || existsSync(path.join(r, ".lando"))) env = "lando";
  else env = "local";
  if (!_rootCache.has(r)) _rootCache.set(r, {});
  _rootCache.get(r).env = env;
  return env;
}

export function getDefaultConfig(root) {
  const drupalRoot = detectDrupalRoot(root) || "docroot";
  return {
    routes: [`${drupalRoot}/modules/custom`, `${drupalRoot}/themes/custom`],
    patterns: [".html.twig", ".inc", ".yml", ".module", ".theme"],
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
  if (_rootCache.has(r) && _rootCache.get(r).config) return _rootCache.get(r).config;

  const configFile = Bun.file(getConfigFile(r));
  if (!(await configFile.exists())) {
    const def = getDefaultConfig(r);
    await Bun.write(getConfigFile(r), JSON.stringify(def, null, 2));
    if (!_rootCache.has(r)) _rootCache.set(r, {});
    _rootCache.get(r).config = { ...def };
    return _rootCache.get(r).config;
  }

  try {
    const raw = await configFile.text();
    const config = JSON.parse(raw);

    if (!config.drupalRoot) {
      const detected = detectDrupalRoot(r);
      if (detected) {
        config.drupalRoot = detected;
        config.routes = config.routes.map(route => {
          const p = route.split("/");
          if (p.length > 0 && POSSIBLE_DOCROOTS.includes(p[0]) && p[0] !== detected) {
            return route.replace(p[0], detected);
          }
          return route;
        });
        await saveConfig(config, r);
      }
    }

    if (!_rootCache.has(r)) _rootCache.set(r, {});
    _rootCache.get(r).config = { ...getDefaultConfig(r), ...config };
    return _rootCache.get(r).config;
  } catch {
    console.error(`${RED}Error al leer configuración. Usando defaults.${NC}`);
    const def = getDefaultConfig(r);
    if (!_rootCache.has(r)) _rootCache.set(r, {});
    _rootCache.get(r).config = { ...def };
    return _rootCache.get(r).config;
  }
}

export async function saveConfig(config, root) {
  const r = getRoot(root);
  await Bun.write(getConfigFile(r), JSON.stringify(config, null, 2));
  if (_rootCache.has(r)) _rootCache.get(r).config = null;
}

// --- PID file ---
export function getPidFile(root) {
  return getPidFilePath(root);
}

export async function writePid(root) {
  await Bun.write(getPidFilePath(root), String(process.pid));
}

export async function removePid(root) {
  try {
    await Bun.write(getPidFilePath(root), "");
    const { rm } = await import("fs/promises");
    await rm(getPidFilePath(root), { force: true });
  } catch {}
}

export async function checkPid(root) {
  const pidFile = Bun.file(getPidFilePath(root));
  if (!(await pidFile.exists())) return null;
  const pid = (await pidFile.text()).trim();
  if (!pid) return null;
  const result = Bun.spawnSync(["ps", "-p", pid, "-o", "pid="]);
  return result.exitCode === 0 ? pid : "stale";
}

export function invalidateConfigCache(root) {
  const r = getRoot(root);
  _rootCache.delete(r);
}
