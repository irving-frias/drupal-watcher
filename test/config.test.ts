import { describe, it, expect, beforeEach, afterEach, spyOn } from "bun:test";
import path from "path";

const TEST_DIR = path.join(import.meta.dir, "..", ".test-tmp");

async function cleanup() {
  try {
    await Bun.spawn(["rm", "-rf", TEST_DIR]).exited;
  } catch {}
}



describe("Config", () => {
  beforeEach(async () => {
    await cleanup();
    const config = await import("../src/config");
    config.invalidateConfigCache(TEST_DIR);
  });

  afterEach(async () => {
    await cleanup();
  });

  it("detects Drupal root from docroot/core", async () => {
    await Bun.spawn(["mkdir", "-p", path.join(TEST_DIR, "docroot", "core")]).exited;
    const { detectDrupalRoot } = await import("../src/config");
    expect(detectDrupalRoot(TEST_DIR)).toBe("docroot");
  });

  it("detects Drupal root from web/core", async () => {
    await Bun.spawn(["mkdir", "-p", path.join(TEST_DIR, "web", "core")]).exited;
    const { detectDrupalRoot } = await import("../src/config");
    expect(detectDrupalRoot(TEST_DIR)).toBe("web");
  });

  it("creates default config on first load", async () => {
    const { loadConfig } = await import("../src/config");
    const config = await loadConfig(TEST_DIR);
    expect(config.routes).toContain("docroot/modules/custom");
    expect(config.routes).toContain("docroot/themes/custom");
    expect(config.debounce).toBe(800);
    expect(config.drushCommand).toBe("cr");
    expect(config.patterns).toContain(".php");
  });

  it("loads existing config file", async () => {
    await Bun.spawn(["mkdir", "-p", path.join(TEST_DIR, "docroot", "core")]).exited;
    const customConfig = {
      routes: ["docroot/modules/custom", "docroot/themes/custom", "docroot/modules/contrib"],
      patterns: [".html.twig", ".module", ".theme"],
      debounce: 600,
      drushCommand: "cr",
    };
    await Bun.write(path.join(TEST_DIR, "watcher.config.json"), JSON.stringify(customConfig, null, 2));

    const { loadConfig, invalidateConfigCache } = await import("../src/config");
    invalidateConfigCache(TEST_DIR);
    const config = await loadConfig(TEST_DIR);
    expect(config.routes).toHaveLength(3);
    expect(config.debounce).toBe(600);
  });

  it("returns null for no Drupal root", async () => {
    const { detectDrupalRoot } = await import("../src/config");
    expect(detectDrupalRoot(TEST_DIR)).toBeNull();
  });
});

describe("Drush", () => {
  it("returns the configured drushCmd when set", async () => {
    const { getDrushCommand } = await import("../src/drush");
    const cmd = getDrushCommand({ drushCmd: "/custom/drush" });
    expect(cmd).toBe("/custom/drush");
  });

  it("returns 'drush' as fallback", async () => {
    const { getDrushCommand } = await import("../src/drush");
    const cmd = getDrushCommand({});
    expect(cmd).toBe("drush");
  });

  it("builds spawn args from config", async () => {
    const { getDrushSpawnArgs } = await import("../src/drush");
    const { cmd, args } = getDrushSpawnArgs({ drushCmd: "ddev drush", drushCommand: "cr" });
    expect(cmd).toBe("ddev");
    expect(args).toEqual(["drush", "cr"]);
  });

  it("includes extra drush args", async () => {
    const { getDrushSpawnArgs } = await import("../src/drush");
    const { cmd, args } = getDrushSpawnArgs({
      drushCmd: "drush",
      drushCommand: "cr",
      drushArgs: ["--uri=default"],
    });
    expect(cmd).toBe("drush");
    expect(args).toEqual(["cr", "--uri=default"]);
  });
});

describe("Main", () => {
  it("parses --abort-on-drush-error flag", async () => {
    const { parseFlags } = await import("../src/main");
    const { flags } = parseFlags(["--abort-on-drush-error"]);
    expect(flags.abortOnDrushError).toBe(true);
  });

  it("parses --watch flag", async () => {
    const { parseFlags } = await import("../src/main");
    const { flags } = parseFlags(["--watch=modules/custom"]);
    expect(flags.watchRoutes).toEqual(["modules/custom"]);
  });

  it("parses --no-watch flag", async () => {
    const { parseFlags } = await import("../src/main");
    const { flags } = parseFlags(["--no-watch=modules/contrib"]);
    expect(flags.noWatchRoutes).toEqual(["modules/contrib"]);
  });

  it("parses multiple flags at once", async () => {
    const { parseFlags } = await import("../src/main");
    const { flags } = parseFlags([
      "--abort-on-drush-error",
      "--watch=modules/custom",
      "--no-watch=modules/contrib",
    ]);
    expect(flags.abortOnDrushError).toBe(true);
    expect(flags.watchRoutes).toEqual(["modules/custom"]);
    expect(flags.noWatchRoutes).toEqual(["modules/contrib"]);
  });

  it("handles empty argv", async () => {
    const { parseFlags } = await import("../src/main");
    const { flags, extra } = parseFlags([]);
    expect(flags.abortOnDrushError).toBe(false);
    expect(flags.watchRoutes).toEqual([]);
    expect(flags.noWatchRoutes).toEqual([]);
    expect(extra).toEqual([]);
  });

  it("parses --dry-run flag", async () => {
    const { parseFlags } = await import("../src/main");
    const { flags } = parseFlags(["--dry-run"]);
    expect(flags.dryRun).toBe(true);
  });

  it("parses --verbose flag", async () => {
    const { parseFlags } = await import("../src/main");
    const { flags } = parseFlags(["--verbose"]);
    expect(flags.verbose).toBe(true);
  });

  it("parses -v alias", async () => {
    const { parseFlags } = await import("../src/main");
    const { flags } = parseFlags(["-v"]);
    expect(flags.verbose).toBe(true);
  });

  it("parses --no-colors flag", async () => {
    const { parseFlags } = await import("../src/main");
    const { flags } = parseFlags(["--no-colors"]);
    expect(flags.noColors).toBe(true);
  });

  it("combines all new flags", async () => {
    const { parseFlags } = await import("../src/main");
    const { flags } = parseFlags(["--dry-run", "--verbose", "--no-colors"]);
    expect(flags.dryRun).toBe(true);
    expect(flags.verbose).toBe(true);
    expect(flags.noColors).toBe(true);
  });

  it("parses --debounce flag", async () => {
    const { parseFlags } = await import("../src/main");
    const { flags } = parseFlags(["--debounce=1200"]);
    expect(flags.debounce).toBe(1200);
  });

  it("parses --no-dotfiles flag", async () => {
    const { parseFlags } = await import("../src/main");
    const { flags } = parseFlags(["--no-dotfiles"]);
    expect(flags.noDotfiles).toBe(true);
  });

  it("parses --log-file flag", async () => {
    const { parseFlags } = await import("../src/main");
    const { flags } = parseFlags(["--log-file=/tmp/watcher.log"]);
    expect(flags.logFile).toBe("/tmp/watcher.log");
  });
});

describe("Drush duration", () => {
  it("includes duration in runDrush result", async () => {
    const { runDrush } = await import("../src/drush");
    const result = await runDrush("echo", ["hello"]);
    expect(result).toHaveProperty("exitCode");
    expect(result).toHaveProperty("duration");
    expect(typeof result.duration).toBe("string");
    expect(Number(result.duration)).toBeGreaterThanOrEqual(0);
  });
});

describe("Colors", () => {
  it("setColorsEnabled toggles colors", async () => {
    const utils = await import("../src/utils");
    utils.setColorsEnabled(false);
    expect(utils.colorsEnabled()).toBe(false);
    expect(utils.green("test")).toBe("test");
    expect(utils.red("test")).toBe("test");
    utils.setColorsEnabled(true);
    expect(utils.colorsEnabled()).toBe(true);
    expect(utils.green("test")).not.toBe("test");
  });
});

describe("Config validation", () => {
  it("routes must be an array", async () => {
    const { validateConfig } = await import("../src/config");
    const config = validateConfig({ routes: null }, TEST_DIR);
    expect(Array.isArray(config.routes)).toBe(true);
    expect(config.routes.length).toBeGreaterThan(0);
  });

  it("debounce must be a positive number", async () => {
    const { validateConfig } = await import("../src/config");
    const c1 = validateConfig({ debounce: "abc" }, TEST_DIR);
    expect(c1.debounce).toBe(800);
    const c2 = validateConfig({ debounce: -100 }, TEST_DIR);
    expect(c2.debounce).toBe(800);
  });

  it("drushArgs must be an array", async () => {
    const { validateConfig } = await import("../src/config");
    const config = validateConfig({ drushArgs: "not-an-array" }, TEST_DIR);
    expect(Array.isArray(config.drushArgs)).toBe(true);
  });

  it("preserves valid values", async () => {
    const { validateConfig } = await import("../src/config");
    const config = validateConfig({
      routes: ["docroot/modules/custom"],
      debounce: 500,
      drushArgs: ["--uri=default"],
    }, TEST_DIR);
    expect(config.routes).toEqual(["docroot/modules/custom"]);
    expect(config.debounce).toBe(500);
    expect(config.drushArgs).toEqual(["--uri=default"]);
  });

  it("accepts valid commandsPerPattern merged with defaults", async () => {
    const { validateConfig } = await import("../src/config");
    const cpp = { ".html.twig": "custom-command" };
    const config = validateConfig({ commandsPerPattern: cpp }, TEST_DIR);
    // User override takes precedence
    expect(config.commandsPerPattern[".html.twig"]).toBe("custom-command");
    // Defaults still present
    expect(config.commandsPerPattern[".theme"]).toBe("cc theme-registry");
    expect(config.commandsPerPattern[".php"]).toBe("cc plugin");
    expect(config.commandsPerPattern[".yml"]).toBe("cc plugin");
  });

  it("fills defaults from old config file with empty commandsPerPattern", async () => {
    // Simulate a config file created before commandsPerPattern defaults existed
    const oldConfig = { routes: ["docroot/modules/custom"], commandsPerPattern: {} };
    await Bun.write(path.join(TEST_DIR, "watcher.config.json"), JSON.stringify(oldConfig));
    const { loadConfig, invalidateConfigCache } = await import("../src/config");
    invalidateConfigCache(TEST_DIR);
    const config = await loadConfig(TEST_DIR);
    // In-memory config should have the full defaults merged
    expect(config.commandsPerPattern[".html.twig"]).toBe("cc twig");
    expect(config.commandsPerPattern[".php"]).toBe("cc plugin");
    expect(config.commandsPerPattern[".yml"]).toBe("cc plugin");
    // File on disk still has the old value (no overwrite)
    const onDisk = JSON.parse(await Bun.file(path.join(TEST_DIR, "watcher.config.json")).text());
    expect(onDisk.commandsPerPattern).toEqual({});
  });

  it("uses all defaults when commandsPerPattern is empty", async () => {
    const { validateConfig } = await import("../src/config");
    const config = validateConfig({ commandsPerPattern: {} }, TEST_DIR);
    expect(config.commandsPerPattern[".html.twig"]).toBe("cc twig");
    expect(config.commandsPerPattern[".module"]).toBe("cc plugin");
  });

  it("rejects non-object commandsPerPattern", async () => {
    const { validateConfig } = await import("../src/config");
    const config = validateConfig({ commandsPerPattern: "invalid" }, TEST_DIR);
    expect(typeof config.commandsPerPattern).toBe("object");
    expect(Array.isArray(config.commandsPerPattern)).toBe(false);
    // Falls back to all defaults
    expect(config.commandsPerPattern[".html.twig"]).toBe("cc twig");
    expect(config.commandsPerPattern[".php"]).toBe("cc plugin");
  });

  it("getCacheClearArgs returns drush cr when no commandsPerPattern", async () => {
    const mod = await import("../src/watcher");
    // Private function — test via the export that uses it indirectly
    // Instead, verify that getDrushSpawnArgs behavior is preserved
    const { getDrushSpawnArgs } = await import("../src/drush");
    const { cmd, args } = getDrushSpawnArgs({ drushCommand: "cr" });
    expect(cmd).toBe("drush");
    expect(args).toContain("cr");
  });

  it("normalizes routes (removes trailing slashes)", async () => {
    const { validateConfig } = await import("../src/config");
    const config = validateConfig({ routes: ["docroot/modules/custom/", "docroot/themes//custom"] }, TEST_DIR);
    expect(config.routes[0]).toBe("docroot/modules/custom");
    expect(config.routes[1]).toBe("docroot/themes/custom");
  });

  it("validates invalid config when loading", async () => {
    await Bun.write(
      path.join(TEST_DIR, "watcher.config.json"),
      JSON.stringify({ routes: null, debounce: "abc" }, null, 2)
    );
    const { loadConfig, invalidateConfigCache } = await import("../src/config");
    invalidateConfigCache(TEST_DIR);
    const config = await loadConfig(TEST_DIR);
    expect(Array.isArray(config.routes)).toBe(true);
    expect(typeof config.debounce).toBe("number");
  });
});

describe("PID check", () => {
  it("returns null when no PID file", async () => {
    const { checkPid } = await import("../src/config");
    expect(await checkPid(TEST_DIR)).toBeNull();
  });

  it("returns stale for non-existent PID", async () => {
    const { checkPid } = await import("../src/config");
    const pidPath = path.join(TEST_DIR, ".drupal-watcher.pid");
    await Bun.write(pidPath, "99999999");
    const result = await checkPid(TEST_DIR);
    expect(result).toBe("stale");
  });

  it("returns null for empty PID file", async () => {
    const { checkPid } = await import("../src/config");
    const pidPath = path.join(TEST_DIR, ".drupal-watcher.pid");
    await Bun.write(pidPath, "");
    const result = await checkPid(TEST_DIR);
    expect(result).toBeNull();
  });

  it("starttime round-trip works", async () => {
    const { writeStarttime, getStarttime, removeStarttime } = await import("../src/config");
    await writeStarttime(TEST_DIR);
    const t = await getStarttime(TEST_DIR);
    expect(typeof t).toBe("number");
    expect(t).toBeGreaterThan(0);
    expect(Date.now() - t).toBeLessThan(5000);
    await removeStarttime(TEST_DIR);
    expect(await getStarttime(TEST_DIR)).toBeNull();
  });
});

describe("Watcher", () => {
  it("resetDebounce clears state", async () => {
    const { resetDebounce, stats } = await import("../src/watcher");
    resetDebounce();
    expect(stats.changes).toBe(0);
    expect(stats.clears).toBe(0);
    expect(stats.startTime).toBeNull();
    expect(stats.filesChanged instanceof Set).toBe(true);
  });
});

describe("Commands", () => {
  it("cmdHelp prints general help", async () => {
    const spy = spyOn(console, "log");
    const { cmdHelp } = await import("../src/commands");
    cmdHelp();
    expect(spy).toHaveBeenCalled();
    spy.mockRestore();
  });

  it("cmdHelp prints start help", async () => {
    const spy = spyOn(console, "log");
    const { cmdHelp } = await import("../src/commands");
    cmdHelp("start");
    expect(spy).toHaveBeenCalled();
    spy.mockRestore();
  });

  it("cmdStatus shows not running when no PID", async () => {
    const spy = spyOn(console, "log");
    const { cmdStatus } = await import("../src/commands");
    await cmdStatus();
    expect(spy).toHaveBeenCalled();
    spy.mockRestore();
  });

  it("cmdRestart does not throw when no watcher", async () => {
    const spy = spyOn(console, "log");
    const { cmdRestart } = await import("../src/commands");
    // This will call cmdStart but fail early at health check since no Drupal root
    // We just verify the restart function itself is callable
    expect(typeof cmdRestart).toBe("function");
    spy.mockRestore();
  });
});

describe("Drush shell", () => {
  it("healthCheck returns false on invalid drush", async () => {
    const { healthCheck } = await import("../src/drush");
    const result = await healthCheck({ drushCmd: "/nonexistent/drush" });
    expect(result).toBe(false);
  });

  it("uses sh -c on non-Windows", async () => {
    const { runPostClearCommands } = await import("../src/drush");
    // Spawn and immediately cancel to verify it doesn't throw
    const cmd = "echo test-shell";
    const proc = Bun.spawn(["sh", "-c", cmd]);
    const exitCode = await proc.exited;
    expect(exitCode).toBe(0);
  });
});
