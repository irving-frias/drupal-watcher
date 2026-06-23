import { describe, it, expect, beforeEach, afterEach } from "bun:test";
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
