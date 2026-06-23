import { P_ERROR, P_WARN, yellow, setColorsEnabled } from "./utils.js";
import {
  cmdStart, cmdList, cmdStatus, cmdAdd, cmdRemove, cmdReset, cmdRestart, cmdHelp,
} from "./commands.js";
import { setCustomConfigPath, invalidateConfigCache } from "./config.js";

const BIN = "vendor/bin/drupal-watcher";

export function parseFlags(argv) {
  const flags = {
    abortOnDrushError: false,
    watchRoutes: [],
    noWatchRoutes: [],
    dryRun: false,
    verbose: false,
    noColors: false,
    debounce: null,
    noDotfiles: false,
    logFile: null,
  };
  const extra = [];

  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i];

    if (arg === "--abort-on-drush-error") {
      flags.abortOnDrushError = true;
    } else if (arg.startsWith("--watch=")) {
      flags.watchRoutes.push(arg.slice(8));
    } else if (arg === "--watch") {
      console.error(`${P_ERROR} ${yellow("--watch")} requires a value. Use ${yellow("--watch=<path>")}.`);
      process.exit(1);
    } else if (arg.startsWith("--no-watch=")) {
      flags.noWatchRoutes.push(arg.slice(11));
    } else if (arg === "--no-watch") {
      console.error(`${P_ERROR} ${yellow("--no-watch")} requires a value. Use ${yellow("--no-watch=<path>")}.`);
      process.exit(1);
    } else if (arg === "--dry-run") {
      flags.dryRun = true;
    } else if (arg === "--verbose" || arg === "-v") {
      flags.verbose = true;
    } else if (arg === "--no-colors") {
      flags.noColors = true;
    } else if (arg.startsWith("--debounce=")) {
      flags.debounce = parseInt(arg.slice(11), 10);
      if (isNaN(flags.debounce) || flags.debounce <= 0) {
        console.error(`${P_ERROR} ${yellow("--debounce")} requires a positive number.`);
        process.exit(1);
      }
    } else if (arg === "--no-dotfiles") {
      flags.noDotfiles = true;
    } else if (arg.startsWith("--log-file=")) {
      flags.logFile = arg.slice(11);
    } else if (arg.startsWith("--config=")) {
      // Handled globally in main()
      extra.push(arg);
    } else {
      extra.push(arg);
    }
  }

  return { flags, extra };
}

export async function main() {
  const args = process.argv.slice(2);

  // Handle --no-colors before any output
  if (args.includes("--no-colors")) {
    setColorsEnabled(false);
  }

  // Handle --config before any command
  const configIdx = args.findIndex(a => a.startsWith("--config="));
  if (configIdx !== -1) {
    const p = args[configIdx].slice(9);
    setCustomConfigPath(p);
    invalidateConfigCache();
  }

  // Handle --version and -V as standalone (before command check)
  if (args[0] === "--version" || args[0] === "-V") {
    cmdHelp("version");
    return;
  }

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
        if (flags.noColors) setColorsEnabled(false);
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

      case "restart":
        await cmdRestart();
        break;

      case "--version":
      case "-V":
        cmdHelp("version");
        break;

      default:
        console.error(`${P_ERROR} Unknown command: ${yellow(command)}`);
        console.log(`  ${yellow(`Run '${BIN} help' to see available commands.`)}`);
        process.exit(1);
    }
  } catch (err) {
    console.error(`${P_ERROR} Unexpected error: ${err?.message || err}`);
    process.exit(1);
  }
}
