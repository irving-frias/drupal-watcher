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
    // Invalidate all caches
    const config = await import("../src/config");
    config.invalidateConfigCache(TEST_DIR);
  });

  afterEach(async () => {
    await cleanup();
  });

  it("should detect drupal root from docroot", async () => {
    await Bun.spawn(["mkdir", "-p", path.join(TEST_DIR, "docroot", "core")]).exited;
    const { detectDrupalRoot } = await import("../src/config");
    expect(detectDrupalRoot(TEST_DIR)).toBe("docroot");
  });

  it("should detect drupal root from web", async () => {
    await Bun.spawn(["mkdir", "-p", path.join(TEST_DIR, "web", "core")]).exited;
    const { detectDrupalRoot } = await import("../src/config");
    expect(detectDrupalRoot(TEST_DIR)).toBe("web");
  });

  it("should detect ddev environment", async () => {
    await Bun.spawn(["mkdir", "-p", path.join(TEST_DIR, ".ddev")]).exited;
    const { detectEnvironment } = await import("../src/config");
    expect(detectEnvironment(TEST_DIR)).toBe("ddev");
  });

  it("should detect lando environment", async () => {
    await Bun.spawn(["mkdir", "-p", path.join(TEST_DIR, ".lando")]).exited;
    const { detectEnvironment } = await import("../src/config");
    expect(detectEnvironment(TEST_DIR)).toBe("lando");
  });

  it("should return local for no env", async () => {
    const { detectEnvironment } = await import("../src/config");
    expect(detectEnvironment(TEST_DIR)).toBe("local");
  });

  it("should create default config on first load", async () => {
    const { loadConfig } = await import("../src/config");
    const config = await loadConfig(TEST_DIR);
    expect(config.routes).toContain("docroot/modules/custom");
    expect(config.routes).toContain("docroot/themes/custom");
    expect(config.debounce).toBe(800);
    expect(config.drushCommand).toBe("cr");
  });

  it("should load existing config", async () => {
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
});
