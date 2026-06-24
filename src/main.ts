import { P_ERROR, P_WARN, yellow, setColorsEnabled } from "./utils";
import {
  cmdStart, cmdList, cmdStatus, cmdAdd, cmdRemove, cmdReset, cmdRestart, cmdHelp,
} from "./commands";
import { setCustomConfigPath, invalidateConfigCache, removePid } from "./config";

const BIN = "vendor/bin/drupal-watcher";

export interface ParseFlagsResult {
  flags: {
    abortOnDrushError: boolean
    watchRoutes: string[]
    noWatchRoutes: string[]
    dryRun: boolean
    verbose: boolean
    noColors: boolean
    debounce: number | null
    noDotfiles: boolean
    logFile: string | null
    commandsPerPattern: Record<string, string>
  }
  extra: string[]
}

export function parseFlags(argv: string[]): ParseFlagsResult {
  const flags = {
    abortOnDrushError: false,
    watchRoutes: [] as string[],
    noWatchRoutes: [] as string[],
    dryRun: false,
    verbose: false,
    noColors: false,
    debounce: null as number | null,
    noDotfiles: false,
    logFile: null as string | null,
    commandsPerPattern: {} as Record<string, string>,
  };
  const extra: string[] = [];

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
    } else if (arg.startsWith("--commands-per-pattern=")) {
      const kv = arg.slice(22);
      const eqIdx = kv.indexOf("=");
      if (eqIdx === -1) {
        console.error(`${P_ERROR} ${yellow("--commands-per-pattern")} requires pattern=command. Use ${yellow("--commands-per-pattern=.html.twig=cc\\ twig")}.`);
        process.exit(1);
      }
      flags.commandsPerPattern[kv.slice(0, eqIdx)] = kv.slice(eqIdx + 1);
    } else if (arg.startsWith("--log-file=")) {
      flags.logFile = arg.slice(11);
    } else if (arg.startsWith("--config=")) {
      extra.push(arg);
    } else {
      extra.push(arg);
    }
  }

  return { flags, extra };
}

export async function main() {
  process.on("unhandledRejection", (reason: unknown) => {
    console.error(`${P_ERROR} Unhandled rejection:`, reason);
  });
  process.on("uncaughtException", (err: Error) => {
    console.error(`${P_ERROR} Uncaught exception:`, err.message);
    removePid();
    process.exit(1);
  });

  const args = process.argv.slice(2);

  if (args.includes("--no-colors")) {
    setColorsEnabled(false);
  }

  const configIdx = args.findIndex(a => a.startsWith("--config="));
  if (configIdx !== -1) {
    const p = args[configIdx].slice(9);
    setCustomConfigPath(p);
    invalidateConfigCache();
  }

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
  } catch (err: unknown) {
    console.error(`${P_ERROR} Unexpected error: ${err instanceof Error ? err.message : err}`);
    process.exit(1);
  }
}
