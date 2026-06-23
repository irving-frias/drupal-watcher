import { existsSync } from "fs";
import path from "path";
import {
  RED, GREEN, YELLOW, BLUE, CYAN, NC, ERROR, WARN, INFO, SUCCESS,
  bold, dim, POSSIBLE_DOCROOTS,
} from "./utils.js";
import {
  loadConfig, saveConfig, getDefaultConfig, detectDrupalRoot,
  writePid, removePid, checkPid,
} from "./config.js";
import { getDrushCommand, getDrushSpawnArgs, healthCheck } from "./drush.js";
import { startWatcher, stopWatcher, resetDebounce, printStats, stats } from "./watcher.js";

// ── Help helpers ──────────────────────────────────────────

function printHeader(title) {
  console.log(`${YELLOW}${title}${NC}`);
}

function printSection(heading, items) {
  console.log(`\n${BLUE}${heading}:${NC}`);
  for (const item of items) {
    const [label, desc] = Array.isArray(item) ? item : [item, ""];
    if (desc) {
      console.log(`  ${bold(label)}  ${desc}`);
    } else {
      console.log(`  ${label}`);
    }
  }
}

const COMMANDS = [
  ["start", "Start the file watcher"],
  ["status", "Show whether the watcher is running"],
  ["list", "Display current configuration"],
  ["add    <path>", "Add a route to watch"],
  ["remove <path>", "Remove a watched route"],
  ["reset", "Reset routes to defaults"],
  ["help   [command]", "Show detailed help for a command"],
];

const GLOBAL_FLAGS = [
  ["--abort-on-drush-error", "Exit if Drush does not respond"],
  ["--watch=<path>", "Watch only a specific route"],
  ["--no-watch=<path>", "Exclude a specific route"],
];

const EXAMPLES = [
  "vendor/bin/drupal-watcher start",
  "vendor/bin/drupal-watcher add docroot/modules/contrib",
  "vendor/bin/drupal-watcher remove docroot/modules/contrib",
  "vendor/bin/drupal-watcher list",
  "vendor/bin/drupal-watcher help start",
];

// ── Help ──────────────────────────────────────────────────

export function cmdHelp(command) {
  if (command === "start") {
    printHeader("🚀 drupal-watcher start");
    console.log("  Start the watcher to monitor Drupal file changes.");
    printSection("Flags", [
      ["--abort-on-drush-error", "Exit if Drush does not respond"],
      ["--watch=<path>", "Watch only a specific route"],
      ["--no-watch=<path>", "Exclude a specific route"],
    ]);
    printSection("Examples", [
      "vendor/bin/drupal-watcher start",
      "vendor/bin/drupal-watcher start --abort-on-drush-error",
      "vendor/bin/drupal-watcher start --watch modules/my-module",
    ]);
    return;
  }

  if (command === "add") {
    printHeader("🚀 drupal-watcher add");
    console.log("  Add a route for the watcher to monitor.");
    printSection("Usage", [`vendor/bin/drupal-watcher add ${dim("<path>")}`]);
    printSection("Examples", [
      "vendor/bin/drupal-watcher add modules/contrib",
      "vendor/bin/drupal-watcher add docroot/themes/custom/my-theme",
    ]);
    return;
  }

  if (command === "remove") {
    printHeader("🚀 drupal-watcher remove");
    console.log("  Remove a route from the watch list.");
    printSection("Usage", [`vendor/bin/drupal-watcher remove ${dim("<path>")}`]);
    printSection("Examples", ["vendor/bin/drupal-watcher remove modules/contrib"]);
    return;
  }

  if (command === "list") {
    printHeader("🚀 drupal-watcher list");
    console.log("  Display the current watcher configuration.");
    return;
  }

  if (command === "reset") {
    printHeader("🚀 drupal-watcher reset");
    console.log("  Reset routes to default values.");
    return;
  }

  if (command === "status") {
    printHeader("🚀 drupal-watcher status");
    console.log("  Show whether the watcher is running and its PID.");
    return;
  }

  // General help
  console.log(`${YELLOW}🚀 Drupal Watcher${NC} — Watch Drupal files and auto-run drush cr`);
  console.log();
  printSection("Usage", [`vendor/bin/drupal-watcher ${bold("<command>")} ${dim("[arguments]")}`]);
  printSection("Commands", COMMANDS);
  printSection("Global flags", GLOBAL_FLAGS);
  printSection("Examples", EXAMPLES);
  console.log(`\n${BLUE}Config:${NC} watcher.config.json`);
  console.log(`${BLUE}Docs:${NC} README.md`);
}

// ── List ──────────────────────────────────────────────────

export async function cmdList() {
  const config = await loadConfig();
  const drushSpawn = getDrushSpawnArgs(config);
  const drupalRoot = config.drupalRoot || detectDrupalRoot() || "not detected";

  printHeader("📋 Watcher configuration");
  console.log(`  ${YELLOW}Drupal root:${NC} ${drupalRoot}`);
  console.log(`  ${YELLOW}Watched routes:${NC}`);
  config.routes.forEach(r => console.log(`    - ${r}`));
  console.log(`  ${YELLOW}Patterns:${NC} ${config.patterns.join(", ")}`);
  if (config.excludePatterns?.length > 0) {
    console.log(`  ${YELLOW}Excluded:${NC} ${config.excludePatterns.join(", ")}`);
  }
  console.log(`  ${YELLOW}Debounce:${NC} ${config.debounce}ms`);
  console.log(`  ${YELLOW}Drush:${NC} ${drushSpawn.cmd} ${drushSpawn.args.join(" ")}`);
  if (config.postClearCommands?.length > 0) {
    console.log(`  ${YELLOW}Post-clear:${NC} ${config.postClearCommands.join("; ")}`);
  }
}

// ── Status ────────────────────────────────────────────────

export async function cmdStatus() {
  const pid = await checkPid();

  if (pid === null) {
    console.log(`${WARN} Watcher is not running.`);
    console.log(`   Run ${bold("vendor/bin/drupal-watcher start")} to start it.`);
  } else if (pid === "stale") {
    console.log(`${WARN} Stale PID file. Watcher is not running.`);
    console.log(`   Run ${bold("vendor/bin/drupal-watcher start")} to start it.`);
    await removePid();
  } else {
    const result = Bun.spawnSync(["ps", "-p", pid, "-o", "etime="]);
    const elapsed = result.stdout.toString().trim();
    console.log(`${SUCCESS} Watcher is active`);
    console.log(`  ${YELLOW}PID:${NC} ${pid}`);
    if (elapsed) console.log(`  ${YELLOW}Uptime:${NC} ${elapsed}`);
  }
}

// ── Add ───────────────────────────────────────────────────

export async function cmdAdd(newRoute) {
  if (!newRoute) {
    console.error(`${ERROR} Specify a path to add.`);
    console.log(`   ${YELLOW}Examples:${NC}`);
    console.log(`     vendor/bin/drupal-watcher add modules/contrib`);
    console.log(`     vendor/bin/drupal-watcher add docroot/themes/custom/my-theme`);
    process.exit(1);
  }

  const config = await loadConfig();
  const drupalRoot = config.drupalRoot || detectDrupalRoot();

  if (!drupalRoot) {
    console.error(`${ERROR} Could not detect Drupal root directory.`);
    console.log(`   ${YELLOW}Supported structures:${NC} ${POSSIBLE_DOCROOTS.join(", ")}`);
    process.exit(1);
  }

  let normalized = path.normalize(newRoute).replace(/^\.\//, "");
  if (!normalized.startsWith(drupalRoot) && !POSSIBLE_DOCROOTS.some(r => normalized.startsWith(r))) {
    normalized = path.join(drupalRoot, normalized);
    console.log(`${INFO} Normalized path: ${normalized}`);
  }

  if (config.routes.includes(normalized)) {
    console.log(`${WARN} Path is already in the list.`);
    printSection("Current routes", config.routes.map(r => `  - ${r}`));
    return;
  }

  if (!existsSync(path.join(process.cwd(), normalized))) {
    console.error(`${ERROR} Path does not exist: ${normalized}`);
    console.log(`   ${YELLOW}Verify the folder exists at:${NC} ${process.cwd()}/${normalized}`);
    process.exit(1);
  }

  config.routes.push(normalized);
  await saveConfig(config);
  console.log(`${SUCCESS} Path added: ${normalized}`);
  await cmdList();
}

// ── Remove ────────────────────────────────────────────────

export async function cmdRemove(routeToRemove) {
  if (!routeToRemove) {
    console.error(`${ERROR} Specify a path to remove.`);
    console.log(`   ${YELLOW}Examples:${NC}`);
    console.log(`     vendor/bin/drupal-watcher remove modules/contrib`);
    process.exit(1);
  }

  const config = await loadConfig();
  const normalized = path.normalize(routeToRemove).replace(/^\.\//, "");
  const index = config.routes.indexOf(normalized);

  if (index === -1) {
    console.error(`${ERROR} Path is not in the list.`);
    printSection("Current routes", config.routes.map(r => `  - ${r}`));
    console.log(`   ${YELLOW}To add a path:${NC} vendor/bin/drupal-watcher add <path>`);
    process.exit(1);
  }

  config.routes.splice(index, 1);
  await saveConfig(config);
  console.log(`${SUCCESS} Path removed: ${normalized}`);
  await cmdList();
}

// ── Reset ─────────────────────────────────────────────────

export async function cmdReset() {
  const root = process.cwd();
  const config = await loadConfig(root);
  const def = getDefaultConfig(root);
  config.routes = [...def.routes];
  config.drupalRoot = def.drupalRoot;
  await saveConfig(config, root);
  console.log(`${SUCCESS} Routes reset to defaults.`);
  await cmdList();
}

// ── Start ─────────────────────────────────────────────────

export async function cmdStart(flags = {}) {
  const { abortOnDrushError = false, watchRoutes = [], noWatchRoutes = [] } = flags;

  console.log(`${YELLOW}🚀 Starting Drupal Watcher${NC}`);

  const existingPid = await checkPid();
  if (existingPid && existingPid !== "stale") {
    console.error(`${ERROR} Watcher is already running (PID: ${existingPid}).`);
    console.log(`   Run ${bold("vendor/bin/drupal-watcher status")} for details.`);
    process.exit(1);
  }

  const config = await loadConfig();
  const drupalRoot = config.drupalRoot || detectDrupalRoot();

  if (!drupalRoot) {
    console.error(`${ERROR} Could not detect Drupal root directory.`);
    console.error(`${YELLOW}   Supported structures: ${POSSIBLE_DOCROOTS.join(", ")}${NC}`);
    process.exit(1);
  }

  const { cmd: drushBase, args: drushArgs } = getDrushSpawnArgs(config);
  const drushCmdStr = getDrushCommand(config);

  console.log(`${GREEN}📁 Drupal root:${NC} ${drupalRoot}`);
  console.log(`${GREEN}🔧 Drush:${NC} ${drushCmdStr} ${config.drushCommand}`);

  const drushOk = await healthCheck(config);
  if (!drushOk) {
    if (abortOnDrushError) {
      console.error(`${ERROR} Drush is not responding. Aborting.`);
      process.exit(1);
    }
    console.log(`${WARN} Drush is not responding. Watcher will start anyway.`);
  } else {
    console.log(`${SUCCESS} Drush is responding.`);
  }

  let routes = [...config.routes];
  if (watchRoutes.length > 0) {
    routes = routes.filter(r => watchRoutes.some(w => r.includes(w)));
    if (routes.length === 0) {
      console.error(`${ERROR} No routes match the --watch filter.`);
      process.exit(1);
    }
  }
  if (noWatchRoutes.length > 0) {
    routes = routes.filter(r => !noWatchRoutes.some(n => r.includes(n)));
  }

  console.log(`${GREEN}👀 Watching routes:${NC}`);
  let hasValidRoute = false;
  routes.forEach(route => {
    if (!existsSync(path.join(process.cwd(), route))) {
      console.log(`${WARN} Path does not exist (skipping):${NC} ${route}`);
      return;
    }
    hasValidRoute = true;
    console.log(`  - ${route}`);
  });

  if (!hasValidRoute) {
    console.error(`${ERROR} None of the configured routes exist.`);
    console.error(`${YELLOW}   Add routes with 'vendor/bin/drupal-watcher add'${NC}`);
    process.exit(1);
  }

  await writePid();
  config.routes = routes;
  await startWatcher(config);

  console.log(`${SUCCESS} Watcher active. Waiting for changes... (Ctrl+C to stop)`);

  process.on("SIGINT", async () => {
    console.log(`\n${YELLOW}🛑 Stopping watcher...${NC}`);
    resetDebounce();
    stopWatcher();
    await removePid();
    printStats();
    process.exit(0);
  });
}
