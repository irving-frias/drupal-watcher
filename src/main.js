import { RED, YELLOW, NC } from "./utils.js";
import { cmdStart, cmdList, cmdStatus, cmdAdd, cmdRemove, cmdReset, cmdHelp } from "./commands.js";

export function parseFlags(argv) {
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

  try {
    switch (command) {
      case "help":
      case "-h":
      case "--help":
        cmdHelp(args[1]);
        break;

      case "start": {
        const { flags } = parseFlags(args.slice(1));
        await cmdStart(flags);
        break;
      }

      case "list":
        await cmdList();
        break;

      case "status":
        await cmdStatus();
        break;

      case "add":
        await cmdAdd(args[1]);
        break;

      case "remove":
        await cmdRemove(args[1]);
        break;

      case "reset":
        await cmdReset();
        break;

      default:
        console.error(`${RED}✖ Unknown command: ${command}${NC}`);
        console.log(`  ${YELLOW}Run 'drupal-watcher help' to see available commands.${NC}`);
        process.exit(1);
    }
  } catch (err) {
    console.error(`${RED}✖ Unexpected error:${NC} ${err?.message || err}`);
    process.exit(1);
  }
}
