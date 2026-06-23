import { RED, YELLOW, NC } from "./utils.js";
import { cmdStart } from "./commands.js";
import { cmdList } from "./commands.js";
import { cmdStatus } from "./commands.js";
import { cmdAdd, cmdRemove, cmdReset, cmdHelp } from "./commands.js";

function parseFlags(argv) {
  const flags = {
    abortOnDrushError: false,
    watchRoutes: [],
    noWatchRoutes: [],
  };
  const extra = [];
  for (const arg of argv) {
    if (arg === "--abort-on-drush-error") flags.abortOnDrushError = true;
    else if (arg.startsWith("--watch=")) flags.watchRoutes.push(arg.slice(8));
    else if (arg.startsWith("--no-watch=")) flags.noWatchRoutes.push(arg.slice(11));
    else extra.push(arg);
  }
  return { flags, extra };
}

export async function main() {
  const args = process.argv.slice(2);
  if (args.length === 0) {
    cmdHelp();
    return;
  }

  const command = args[0];

  if (command === "help" || command === "-h" || command === "--help") {
    cmdHelp(args[1]);
    return;
  }

  if (command === "start") {
    const { flags } = parseFlags(args.slice(1));
    await cmdStart(flags);
    return;
  }

  if (command === "list") {
    await cmdList();
    return;
  }

  if (command === "status") {
    await cmdStatus();
    return;
  }

  if (command === "add") {
    await cmdAdd(args[1]);
    return;
  }

  if (command === "remove") {
    await cmdRemove(args[1]);
    return;
  }

  if (command === "reset") {
    await cmdReset();
    return;
  }

  console.error(`${RED}✗ Comando desconocido: ${command}${NC}`);
  console.log(`  ${YELLOW}Ejecuta 'drupal-watcher help' para ver los comandos disponibles.${NC}`);
  process.exit(1);
}
