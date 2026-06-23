import { existsSync } from "fs";
import path from "path";
import { RED, GREEN, YELLOW, BLUE, NC, bold, dim, POSSIBLE_DOCROOTS } from "./utils.js";
import { loadConfig, saveConfig, getDefaultConfig, detectDrupalRoot, detectEnvironment, writePid, removePid, checkPid } from "./config.js";
import { getDrushCommand, getDrushSpawnArgs, healthCheck } from "./drush.js";
import { startWatcher, stopWatcher, resetDebounce, printStats, stats, getWatcherHandle } from "./watcher.js";

// --- Help ---
export function cmdHelp(command) {
  if (command === "start") {
    console.log(`${YELLOW}🚀 drupal-watcher start${NC}`);
    console.log(`  Inicia el watcher para vigilar cambios en archivos de Drupal.`);
    console.log();
    console.log(`${BLUE}Flags:${NC}`);
    console.log(`  ${bold("--abort-on-drush-error")}  Termina si Drush no responde`);
    console.log(`  ${bold("--watch=<ruta>")}           Solo vigila una ruta específica`);
    console.log(`  ${bold("--no-watch=<ruta>")}        Excluye una ruta específica`);
    console.log();
    console.log(`${BLUE}Ejemplos:${NC}`);
    console.log(`  vendor/bin/drupal-watcher start`);
    console.log(`  vendor/bin/drupal-watcher start --abort-on-drush-error`);
    console.log(`  vendor/bin/drupal-watcher start --watch modules/mi-modulo`);
    return;
  }

  if (command === "add") {
    console.log(`${YELLOW}🚀 drupal-watcher add${NC}`);
    console.log(`  Añade una ruta para que el watcher la vigile.`);
    console.log();
    console.log(`${BLUE}Uso:${NC}`);
    console.log(`  vendor/bin/drupal-watcher add ${dim("<ruta>")}`);
    console.log();
    console.log(`${BLUE}Ejemplos:${NC}`);
    console.log(`  vendor/bin/drupal-watcher add modules/contrib`);
    console.log(`  vendor/bin/drupal-watcher add docroot/themes/custom/mi-tema`);
    return;
  }

  if (command === "remove") {
    console.log(`${YELLOW}🚀 drupal-watcher remove${NC}`);
    console.log(`  Elimina una ruta de la lista de vigilancia.`);
    console.log();
    console.log(`${BLUE}Uso:${NC}`);
    console.log(`  vendor/bin/drupal-watcher remove ${dim("<ruta>")}`);
    console.log();
    console.log(`${BLUE}Ejemplos:${NC}`);
    console.log(`  vendor/bin/drupal-watcher remove modules/contrib`);
    return;
  }

  if (command === "list") {
    console.log(`${YELLOW}🚀 drupal-watcher list${NC}`);
    console.log(`  Muestra la configuración actual del watcher.`);
    return;
  }

  if (command === "reset") {
    console.log(`${YELLOW}🚀 drupal-watcher reset${NC}`);
    console.log(`  Restablece las rutas a los valores por defecto.`);
    return;
  }

  if (command === "status") {
    console.log(`${YELLOW}🚀 drupal-watcher status${NC}`);
    console.log(`  Muestra si el watcher está corriendo y su PID.`);
    return;
  }

  // General help
  console.log(`${YELLOW}🚀 Drupal Watcher${NC} — Vigila archivos de Drupal y ejecuta drush cr automáticamente`);
  console.log();
  console.log(`${BLUE}Uso:${NC}`);
  console.log(`  vendor/bin/drupal-watcher ${bold("<comando>")} ${dim("[argumentos]")}`);
  console.log();
  console.log(`${BLUE}Comandos:${NC}`);
  console.log(`  ${bold("start")}                   Inicia el watcher`);
  console.log(`  ${bold("status")}                  Muestra si el watcher está corriendo`);
  console.log(`  ${bold("list")}                    Muestra la configuración actual`);
  console.log(`  ${bold("add")}    ${dim("<ruta>")}          Añade una ruta a vigilar`);
  console.log(`  ${bold("remove")} ${dim("<ruta>")}          Elimina una ruta`);
  console.log(`  ${bold("reset")}                   Restablece las rutas por defecto`);
  console.log(`  ${bold("help")}   ${dim("[comando]")}       Muestra ayuda detallada de un comando`);
  console.log(`  ${bold("help, -h, --help")}        Muestra esta ayuda`);
  console.log();
  console.log(`${BLUE}Flags globales:${NC}`);
  console.log(`  ${bold("--abort-on-drush-error")}  Termina si Drush no responde (con start)`);
  console.log(`  ${bold("--watch=<ruta>")}           Solo vigila una ruta específica`);
  console.log(`  ${bold("--no-watch=<ruta>")}        Excluye una ruta específica`);
  console.log();
  console.log(`${BLUE}Ejemplos:${NC}`);
  console.log(`  vendor/bin/drupal-watcher start`);
  console.log(`  vendor/bin/drupal-watcher add docroot/modules/contrib`);
  console.log(`  vendor/bin/drupal-watcher remove docroot/modules/contrib`);
  console.log(`  vendor/bin/drupal-watcher list`);
  console.log(`  vendor/bin/drupal-watcher help start`);
  console.log();
  console.log(`${BLUE}Configuración:${NC} watcher.config.json`);
  console.log(`${BLUE}Documentación:${NC} README.md`);
}

// --- List ---
export async function cmdList() {
  const config = await loadConfig();
  const drushSpawn = getDrushSpawnArgs(config);
  const drupalRoot = config.drupalRoot || detectDrupalRoot() || "no detectado";

  console.log(`${BLUE}📋 Configuración del watcher:${NC}`);
  console.log(`  ${YELLOW}Directorio raíz:${NC} ${drupalRoot}`);
  console.log(`  ${YELLOW}Entorno:${NC} ${detectEnvironment()}`);
  console.log(`  ${YELLOW}Rutas vigiladas:${NC}`);
  config.routes.forEach(r => console.log(`    - ${r}`));
  console.log(`  ${YELLOW}Patrones:${NC} ${config.patterns.join(", ")}`);
  if (config.excludePatterns?.length > 0) {
    console.log(`  ${YELLOW}Excluidos:${NC} ${config.excludePatterns.join(", ")}`);
  }
  console.log(`  ${YELLOW}Debounce:${NC} ${config.debounce}ms`);
  console.log(`  ${YELLOW}Drush:${NC} ${drushSpawn.cmd} ${drushSpawn.args.join(" ")}`);
  if (config.postClearCommands?.length > 0) {
    console.log(`  ${YELLOW}Post-clear:${NC} ${config.postClearCommands.join("; ")}`);
  }
}

// --- Status ---
export async function cmdStatus() {
  const pid = await checkPid();
  if (pid === null) {
    console.log(`${YELLOW}⚠️ Watcher no está corriendo.${NC}`);
    console.log(`   Ejecuta ${bold("vendor/bin/drupal-watcher start")} para iniciarlo.`);
  } else if (pid === "stale") {
    console.log(`${YELLOW}⚠️ PID file obsoleto. El watcher no está corriendo.${NC}`);
    console.log(`   Ejecuta ${bold("vendor/bin/drupal-watcher start")} para iniciarlo.`);
    await removePid();
  } else {
    const result = Bun.spawnSync(["ps", "-p", pid, "-o", "etime="]);
    const elapsed = result.stdout.toString().trim();
    console.log(`${GREEN}✅ Watcher activo${NC}`);
    console.log(`  ${YELLOW}PID:${NC} ${pid}`);
    if (elapsed) console.log(`  ${YELLOW}Tiempo:${NC} ${elapsed}`);
  }
}

// --- Add ---
export async function cmdAdd(newRoute) {
  if (!newRoute) {
    console.error(`${RED}❌ Especifica una ruta a añadir.${NC}`);
    console.log(`   ${YELLOW}Ejemplos:${NC}`);
    console.log(`     vendor/bin/drupal-watcher add modules/contrib`);
    console.log(`     vendor/bin/drupal-watcher add docroot/themes/custom/mi-tema`);
    process.exit(1);
  }
  const config = await loadConfig();
  const drupalRoot = config.drupalRoot || detectDrupalRoot();
  if (!drupalRoot) {
    console.error(`${RED}❌ No se pudo detectar el directorio raíz de Drupal.${NC}`);
    console.log(`   ${YELLOW}Estructuras soportadas:${NC} ${POSSIBLE_DOCROOTS.join(", ")}`);
    process.exit(1);
  }

  let normalized = path.normalize(newRoute).replace(/^\.\//, "");
  if (!normalized.startsWith(drupalRoot) && !POSSIBLE_DOCROOTS.some(r => normalized.startsWith(r))) {
    normalized = path.join(drupalRoot, normalized);
    console.log(`${YELLOW}ℹ️  Ruta normalizada a:${NC} ${normalized}`);
  }

  if (config.routes.includes(normalized)) {
    console.log(`${YELLOW}⚠️ La ruta ya está en la lista.${NC}`);
    console.log(`   ${YELLOW}Rutas actuales:${NC}`);
    config.routes.forEach(r => console.log(`     - ${r}`));
    return;
  }

  if (!existsSync(path.join(process.cwd(), normalized))) {
    console.error(`${RED}❌ La ruta no existe: ${normalized}${NC}`);
    console.log(`   ${YELLOW}Verifica que la carpeta existe en:${NC} ${process.cwd()}/${normalized}`);
    process.exit(1);
  }

  config.routes.push(normalized);
  await saveConfig(config);
  console.log(`${GREEN}✅ Ruta añadida: ${normalized}${NC}`);
  await cmdList();
}

// --- Remove ---
export async function cmdRemove(routeToRemove) {
  if (!routeToRemove) {
    console.error(`${RED}❌ Especifica una ruta a eliminar.${NC}`);
    console.log(`   ${YELLOW}Ejemplos:${NC}`);
    console.log(`     vendor/bin/drupal-watcher remove modules/contrib`);
    process.exit(1);
  }
  const config = await loadConfig();
  const normalized = path.normalize(routeToRemove).replace(/^\.\//, "");
  const index = config.routes.indexOf(normalized);
  if (index === -1) {
    console.error(`${RED}❌ La ruta no está en la lista.${NC}`);
    console.log(`   ${YELLOW}Rutas actuales:${NC}`);
    config.routes.forEach(r => console.log(`     - ${r}`));
    console.log(`   ${YELLOW}Para añadir una ruta:${NC} vendor/bin/drupal-watcher add <ruta>`);
    process.exit(1);
  }
  config.routes.splice(index, 1);
  await saveConfig(config);
  console.log(`${GREEN}✅ Ruta eliminada: ${normalized}${NC}`);
  await cmdList();
}

// --- Reset ---
export async function cmdReset() {
  const config = await loadConfig();
  const def = getDefaultConfig();
  config.routes = [...def.routes];
  config.drupalRoot = def.drupalRoot;
  await saveConfig(config);
  console.log(`${GREEN}✅ Rutas restablecidas a valores por defecto.${NC}`);
  await cmdList();
}

// --- Start ---
export async function cmdStart(flags = {}) {
  const { abortOnDrushError = false, watchRoutes = [], noWatchRoutes = [] } = flags;

  console.log(`${YELLOW}🚀 Iniciando Drupal Watcher${NC}`);

  // Check PID
  const existingPid = await checkPid();
  if (existingPid && existingPid !== "stale") {
    console.error(`${RED}❌ El watcher ya está corriendo (PID: ${existingPid}).${NC}`);
    console.log(`   Ejecuta ${bold("vendor/bin/drupal-watcher status")} para más información.`);
    process.exit(1);
  }

  const config = await loadConfig();
  const drupalRoot = config.drupalRoot || detectDrupalRoot();

  if (!drupalRoot) {
    console.error(`${RED}❌ No se pudo detectar el directorio raíz de Drupal.${NC}`);
    console.error(`${YELLOW}   Estructuras soportadas: ${POSSIBLE_DOCROOTS.join(", ")}${NC}`);
    process.exit(1);
  }

  const { cmd: drushBase, args: drushArgs } = getDrushSpawnArgs(config);
  const drushCmdStr = getDrushCommand(config);

  console.log(`${GREEN}📁 Directorio raíz:${NC} ${drupalRoot}`);
  console.log(`${GREEN}🌐 Entorno:${NC} ${detectEnvironment()}`);
  console.log(`${GREEN}🔧 Drush:${NC} ${drushCmdStr} ${config.drushCommand}`);

  // Health check
  const drushOk = await healthCheck(config);
  if (!drushOk) {
    if (abortOnDrushError) {
      console.error(`${RED}❌ Drush no responde. Abortando.${NC}`);
      console.log(`   Usa ${bold("--abort-on-drush-error")} para evitar este comportamiento.`);
      process.exit(1);
    }
    console.log(`${YELLOW}⚠️ Drush no responde. El watcher iniciará igualmente.${NC}`);
  } else {
    console.log(`${GREEN}✅ Drush responde correctamente.${NC}`);
  }

  // Filter routes
  let routes = [...config.routes];
  if (watchRoutes.length > 0) {
    routes = routes.filter(r => watchRoutes.some(w => r.includes(w)));
    if (routes.length === 0) {
      console.error(`${RED}❌ Ninguna de las rutas coincide con --watch.${NC}`);
      process.exit(1);
    }
  }
  if (noWatchRoutes.length > 0) {
    routes = routes.filter(r => !noWatchRoutes.some(n => r.includes(n)));
  }

  // Show routes
  console.log(`${GREEN}👀 Vigilando rutas:${NC}`);
  let hasValidRoute = false;
  routes.forEach(route => {
    if (!existsSync(path.join(process.cwd(), route))) {
      console.log(`${YELLOW}⚠️ Ruta no existe (se omite):${NC} ${route}`);
      return;
    }
    hasValidRoute = true;
    console.log(`  - ${route}`);
  });

  if (!hasValidRoute) {
    console.error(`${RED}❌ Ninguna de las rutas configuradas existe.${NC}`);
    console.error(`${YELLOW}   Añade rutas con 'vendor/bin/drupal-watcher add'${NC}`);
    process.exit(1);
  }

  // Write PID
  await writePid();

  // Start watcher
  config.routes = routes;
  const watcher = await startWatcher(config);

  console.log(`${GREEN}✅ Watcher activo. Esperando cambios... (Ctrl+C para salir)${NC}`);

  // SIGINT handler
  process.on("SIGINT", async () => {
    console.log(`\n${YELLOW}🛑 Deteniendo watcher...${NC}`);
    resetDebounce();
    stopWatcher();
    await removePid();
    printStats();
    process.exit(0);
  });
}
