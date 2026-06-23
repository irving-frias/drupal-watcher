import { existsSync } from "fs";
import path from "path";
import {
  P_ERROR, P_WARN, P_INFO, P_SUCCESS,
  green, yellow, blue, cyan, bold, dim, red,
  printHeader, printSection, POSSIBLE_DOCROOTS,
} from "./utils.js";
import {
  loadConfig, saveConfig, getDefaultConfig, detectDrupalRoot,
  writePid, removePid, checkPid, getStarttime,
} from "./config.js";
import { getDrushCommand, getDrushSpawnArgs, healthCheck } from "./drush.js";
import { startWatcher, stopWatcher, resetDebounce, printStats, getWatcherHandle } from "./watcher.js";
import { readFileSync } from "fs";

const BIN = "vendor/bin/drupal-watcher";

function pkgVersion() {
  try {
    const p = JSON.parse(readFileSync(new URL("../package.json", import.meta.url), "utf-8"));
    return p.version || "unknown";
  } catch {
    return "unknown";
  }
}

const COMMANDS = [
  ["start", "Start the file watcher"],
  ["status", "Show whether the watcher is running"],
  ["list", "Display current configuration"],
  ["add    <path>", "Add a route to watch"],
  ["remove <path>", "Remove a watched route"],
  ["reset", "Reset routes to defaults"],
  ["restart", "Restart the file watcher"],
  ["help   [command]", "Show detailed help for a command"],
];

const GLOBAL_FLAGS = [
  ["--abort-on-drush-error", "Exit if Drush does not respond"],
  ["--watch=<path>", "Watch only a specific route (substring match)"],
  ["--no-watch=<path>", "Exclude a specific route (substring match)"],
  ["--dry-run", "Show what would happen without starting the watcher"],
  ["--verbose, -v", "Show full Drush output"],
  ["--no-colors", "Disable colored output"],
  ["--debounce=<ms>", "Override debounce interval"],
  ["--no-dotfiles", "Ignore dotfiles (.*)"],
  ["--log-file=<path>", "Save output to a file"],
  ["--config=<path>", "Use a custom config file"],
  ["--version, -V", "Show version number"],
];

const EXAMPLES = [
  `${BIN} start`,
  `${BIN} add docroot/modules/contrib`,
  `${BIN} remove docroot/modules/contrib`,
  `${BIN} list`,
  `${BIN} start --dry-run`,
  `${BIN} help start`,
];

// ── Help ──────────────────────────────────────────────────

export function cmdHelp(command) {
  if (command === "version" || command === "--version" || command === "-V") {
    console.log(`drupal-watcher v${pkgVersion()}`);
    return;
  }

  if (command === "start") {
    printHeader("drupal-watcher start");
    console.log("  Start the watcher to monitor Drupal file changes.");
    printSection("Flags", [
      ["--abort-on-drush-error", "Exit if Drush does not respond"],
      ["--watch=<path>", "Watch only a specific route (substring match)"],
      ["--no-watch=<path>", "Exclude a specific route (substring match)"],
      ["--dry-run", "Show what would happen without starting"],
      ["--verbose, -v", "Show full Drush output"],
    ]);
    printSection("Examples", [
      `${BIN} start`,
      `${BIN} start --abort-on-drush-error`,
      `${BIN} start --watch modules/my-module`,
      `${BIN} start --dry-run`,
    ]);
    return;
  }

  if (command === "add") {
    printHeader("drupal-watcher add");
    console.log("  Add a route for the watcher to monitor.");
    printSection("Usage", [`${BIN} add ${dim("<path>")}`]);
    printSection("Examples", [
      `${BIN} add modules/contrib`,
      `${BIN} add docroot/themes/custom/my-theme`,
    ]);
    return;
  }

  if (command === "remove") {
    printHeader("drupal-watcher remove");
    console.log("  Remove a route from the watch list.");
    printSection("Usage", [`${BIN} remove ${dim("<path>")}`]);
    printSection("Examples", [`${BIN} remove modules/contrib`]);
    return;
  }

  if (command === "list") {
    printHeader("drupal-watcher list");
    console.log("  Display the current watcher configuration.");
    return;
  }

  if (command === "reset") {
    printHeader("drupal-watcher reset");
    console.log("  Reset routes to default values.");
    return;
  }

  if (command === "status") {
    printHeader("drupal-watcher status");
    console.log("  Show whether the watcher is running and its PID.");
    return;
  }

  if (command === "restart") {
    printHeader("drupal-watcher restart");
    console.log("  Restart the file watcher (stop + start).");
    return;
  }

  // General help
  console.log(`${yellow("Drupal Watcher")} — Watch Drupal files and auto-run drush cr`);
  console.log();
  printSection("Usage", [`${BIN} ${green("<command>")} ${dim("[arguments]")}`]);
  printSection("Commands", COMMANDS);
  printSection("Global flags", GLOBAL_FLAGS);
  printSection("Examples", EXAMPLES);
  console.log(`\n${blue("Config:")} watcher.config.json`);
  console.log(`${blue("Docs:")} README.md`);
}

// ── List ──────────────────────────────────────────────────

export async function cmdList() {
  const config = await loadConfig();
  const drushSpawn = getDrushSpawnArgs(config);
  const drupalRoot = config.drupalRoot || detectDrupalRoot() || "not detected";

  printHeader("Watcher configuration");
  console.log(`  ${yellow("Drupal root:")} ${drupalRoot}`);
  console.log(`  ${yellow("Watched routes:")}`);
  config.routes.forEach(r => console.log(`    - ${r}`));
  console.log(`  ${yellow("Patterns:")} ${config.patterns.join(", ")}`);
  if (config.excludePatterns?.length > 0) {
    console.log(`  ${yellow("Excluded:")} ${config.excludePatterns.join(", ")}`);
  }
  console.log(`  ${yellow("Debounce:")} ${config.debounce}ms`);
  console.log(`  ${yellow("Drush:")} ${drushSpawn.cmd} ${drushSpawn.args.join(" ")}`);
  if (config.postClearCommands?.length > 0) {
    console.log(`  ${yellow("Post-clear:")} ${config.postClearCommands.join("; ")}`);
  }
}

// ── Status ────────────────────────────────────────────────

export async function cmdStatus() {
  const pid = await checkPid();

  if (pid === null) {
    console.log(`${P_WARN} Watcher is not running.`);
    console.log(`   Run ${green(`${BIN} start`)} to start it.`);
  } else if (pid === "stale") {
    console.log(`${P_WARN} Stale PID file. Watcher is not running.`);
    console.log(`   Run ${green(`${BIN} start`)} to start it.`);
    await removePid();
  } else {
    console.log(`${P_SUCCESS} Watcher is active`);
    console.log(`  ${yellow("PID:")} ${pid}`);
    const starttime = await getStarttime();
    if (starttime) {
      const elapsed = Math.round((Date.now() - starttime) / 1000);
      const mins = Math.floor(elapsed / 60);
      const secs = elapsed % 60;
      console.log(`  ${yellow("Uptime:")} ${mins}m ${secs}s`);
    }
  }
}

// ── Add ───────────────────────────────────────────────────

export async function cmdAdd(newRoute) {
  if (!newRoute) {
    console.error(`${P_ERROR} Specify a path to add.`);
    console.log(`   ${yellow("Examples:")}`);
    console.log(`     ${BIN} add modules/contrib`);
    console.log(`     ${BIN} add docroot/themes/custom/my-theme`);
    process.exit(1);
  }

  const config = await loadConfig();
  const drupalRoot = config.drupalRoot || detectDrupalRoot();

  if (!drupalRoot) {
    console.error(`${P_ERROR} Could not detect Drupal root directory.`);
    console.log(`   ${yellow("Supported structures:")} ${POSSIBLE_DOCROOTS.join(", ")}`);
    process.exit(1);
  }

  let normalized = path.normalize(newRoute).replace(/^\.\//, "");
  if (!normalized.startsWith(drupalRoot) && !POSSIBLE_DOCROOTS.some(r => normalized.startsWith(r))) {
    normalized = path.join(drupalRoot, normalized);
    console.log(`${P_INFO} Normalized path: ${cyan(normalized)}`);
  }

  if (config.routes.includes(normalized)) {
    console.log(`${P_WARN} Path is already in the list.`);
    console.log(`   ${yellow("Current routes:")}`);
    config.routes.forEach(r => console.log(`     - ${r}`));
    return;
  }

  if (!existsSync(path.join(process.cwd(), normalized))) {
    console.error(`${P_ERROR} Path does not exist: ${cyan(normalized)}`);
    console.log(`   ${yellow("Verify the folder exists at:")} ${process.cwd()}/${normalized}`);
    process.exit(1);
  }

  config.routes.push(normalized);
  await saveConfig(config);
  console.log(`${P_SUCCESS} Path added: ${cyan(normalized)}`);
  await cmdList();
}

// ── Remove ────────────────────────────────────────────────

export async function cmdRemove(routeToRemove) {
  if (!routeToRemove) {
    console.error(`${P_ERROR} Specify a path to remove.`);
    console.log(`   ${yellow("Examples:")}`);
    console.log(`     ${BIN} remove modules/contrib`);
    process.exit(1);
  }

  const config = await loadConfig();
  const normalized = path.normalize(routeToRemove).replace(/^\.\//, "");
  const index = config.routes.indexOf(normalized);

  if (index === -1) {
    console.error(`${P_ERROR} Path is not in the list.`);
    console.log(`   ${yellow("Current routes:")}`);
    config.routes.forEach(r => console.log(`     - ${r}`));
    console.log(`   ${yellow("To add a path:")} ${BIN} add <path>`);
    process.exit(1);
  }

  config.routes.splice(index, 1);
  await saveConfig(config);
  console.log(`${P_SUCCESS} Path removed: ${cyan(normalized)}`);
  await cmdList();
}

// ── Reset ─────────────────────────────────────────────────

export async function cmdReset() {
  const root = process.cwd();
  const config = await loadConfig(root);
  const def = getDefaultConfig(root);

  if (config.routes.length > 0) {
    console.log(`${P_WARN} This will reset routes to defaults:`);
    console.log(`   ${yellow("Current:")} ${config.routes.join(", ")}`);
    console.log(`   ${yellow("Default:")} ${def.routes.join(", ")}`);
    const { confirm } = await ask(` ${P_WARN} Continue? [y/N] `);
    if (!confirm) {
      console.log("  Cancelled.");
      return;
    }
  }

  config.routes = [...def.routes];
  config.drupalRoot = def.drupalRoot;
  await saveConfig(config, root);
  console.log(`${P_SUCCESS} Routes reset to defaults.`);
  await cmdList();
}

// ── Restart ────────────────────────────────────────────────

export async function cmdRestart() {
  const pid = await checkPid();
  if (pid && pid !== "stale") {
    console.log(`${P_INFO} Stopping watcher (PID: ${pid})...`);
    process.kill(parseInt(pid, 10), "SIGTERM");
    await removePid();
    await new Promise(r => setTimeout(r, 1000));
  } else {
    console.log(`${P_INFO} No active watcher to stop. Starting fresh.`);
  }
  await cmdStart();
}

async function ask(prompt) {
  const buf = new Uint8Array(1024);
  await Bun.write(Bun.stdout, Buffer.from(prompt));
  const n = await Bun.stdin.read(buf);
  const answer = new TextDecoder().decode(buf.subarray(0, n)).trim().toLowerCase();
  return { confirm: answer === "y" || answer === "yes" };
}

// ── Start ─────────────────────────────────────────────────

export async function cmdStart(flags = {}) {
  const {
    abortOnDrushError = false, watchRoutes = [], noWatchRoutes = [],
    dryRun = false, verbose = false, debounce, noDotfiles = false, logFile,
  } = flags;

  if (dryRun) console.log(`${cyan("🏁")} Dry run mode — no watcher will be started\n`);

  console.log(`${yellow("🚀 Starting Drupal Watcher")}`);

  const existingPid = await checkPid();
  if (existingPid && existingPid !== "stale") {
    console.error(`${P_ERROR} Watcher is already running (PID: ${existingPid}).`);
    console.log(`   Run ${green(`${BIN} status`)} for details.`);
    process.exit(1);
  }

  const config = await loadConfig();
  if (debounce) config.debounce = debounce;
  if (noDotfiles) {
    config.excludePatterns = [...(config.excludePatterns || []), ".*"];
  }
  const drupalRoot = config.drupalRoot || detectDrupalRoot();
  const drushPath = getDrushCommand(config);

  if (!drupalRoot) {
    console.error(`${P_ERROR} Could not detect Drupal root directory.`);
    console.error(`   ${yellow("Supported structures:")} ${POSSIBLE_DOCROOTS.join(", ")}`);
    process.exit(1);
  }

  const { cmd: drushBase, args: drushArgs } = getDrushSpawnArgs(config);
  const drushCmdStr = getDrushCommand(config);

  console.log(`  ${green("Drupal root:")} ${drupalRoot}`);
  console.log(`  ${green("Drush:")} ${cyan(drushPath)} ${config.drushCommand}`);

  if (!dryRun) {
    const drushOk = await healthCheck(config);
    if (!drushOk) {
      if (abortOnDrushError) {
        console.error(`${P_ERROR} Drush is not responding. Aborting.`);
        process.exit(1);
      }
      console.log(`  ${P_WARN} Drush is not responding. Watcher will start anyway.`);
    } else {
      console.log(`  ${P_SUCCESS} Drush is responding.`);
    }
  } else {
    console.log(`  ${P_INFO} Skipping health check (dry run).`);
  }

  let routes = [...config.routes];
  if (watchRoutes.length > 0) {
    routes = routes.filter(r => watchRoutes.some(w => r.includes(w)));
    if (routes.length === 0) {
      console.error(`${P_ERROR} No routes match the --watch filter.`);
      process.exit(1);
    }
  }
  if (noWatchRoutes.length > 0) {
    routes = routes.filter(r => !noWatchRoutes.some(n => r.includes(n)));
  }

  console.log(`  ${green("Watching routes:")}`);
  let hasValidRoute = false;
  routes.forEach(route => {
    if (!existsSync(path.join(process.cwd(), route))) {
      console.log(`    ${P_WARN} Path does not exist (skipping): ${route}`);
      return;
    }
    hasValidRoute = true;
    console.log(`    - ${route}`);
  });

  if (!hasValidRoute) {
    console.error(`${P_ERROR} None of the configured routes exist.`);
    console.error(`   ${yellow(`Add routes with '${BIN} add'`)}`);
    process.exit(1);
  }

  if (dryRun) {
    console.log(`\n${cyan("🏁")} Dry run complete. Pass no flags to start the watcher.`);
    return;
  }

  config.routes = routes;

  if (logFile) {
    const writer = Bun.file(logFile).writer();
    const origLog = console.log;
    const origError = console.error;
    console.log = (...args) => {
      origLog(...args);
      writer.write(args.map(String).join(" ") + "\n");
    };
    console.error = (...args) => {
      origError(...args);
      writer.write(args.map(String).join(" ") + "\n");
    };
  }

  await startWatcher(config);

  if (!getWatcherHandle()) {
    console.error(`${P_ERROR} Failed to start watcher.`);
    process.exit(1);
  }

  await writePid();

  console.log(`\n${P_SUCCESS} Watcher active. Waiting for changes... (Ctrl+C to stop)`);

  function shutdown(signal) {
    console.log(`\n${yellow("🛑")} Stopping watcher (${signal})...`);
    resetDebounce();
    stopWatcher();
    removePid();
    printStats();
    process.exit(0);
  }

  process.on("SIGINT", () => shutdown("SIGINT"));
  process.on("SIGTERM", () => shutdown("SIGTERM"));
}
