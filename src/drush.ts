import { existsSync, accessSync, constants } from "fs";
import path from "path";
import { P_ERROR, P_WARN, P_SUCCESS, yellow, cyan } from "./utils";
import type { WatcherConfig, DrushResult, DrushSpawnArgs } from "./types";

export function getDrushCommand(config: Partial<WatcherConfig>): string {
  if (config.drushCmd) return config.drushCmd;

  const vendorDrush = path.join(process.cwd(), "vendor/bin/drush");
  const binDrush = path.join(process.cwd(), "bin/drush");

  if (existsSync(vendorDrush)) {
    try { accessSync(vendorDrush, constants.X_OK); return vendorDrush; } catch {}
  }
  if (existsSync(binDrush)) {
    try { accessSync(binDrush, constants.X_OK); return binDrush; } catch {}
  }
  return "drush";
}

export function getDrushSpawnArgs(config: Partial<WatcherConfig>): DrushSpawnArgs {
  const cmdStr = getDrushCommand(config);
  const parts = cmdStr.split(/\s+/);
  const args = [...parts.slice(1)];
  if (config.drushCommand) args.push(config.drushCommand);
  const drushArgs = config.drushArgs;
  if (drushArgs?.length) args.push(...drushArgs);
  return { cmd: parts[0], args };
}

export async function healthCheck(config: Partial<WatcherConfig>): Promise<boolean> {
  try {
    const cmdStr = getDrushCommand(config);
    const parts = [...cmdStr.split(/\s+/), "status", "--format=json"];
    const proc = Bun.spawn(parts, { stdout: "pipe", stderr: "pipe" });
    const exitCode = await proc.exited;
    return exitCode === 0;
  } catch (e: unknown) {
    console.warn(`${P_WARN} Health check failed: ${e instanceof Error ? e.message : e}`);
    return false;
  }
}

export async function runDrush(drushBase: string, drushArgsArray: string[]): Promise<DrushResult> {
  try {
    const t0 = Date.now();
    const proc = Bun.spawn([drushBase, ...drushArgsArray], { stdout: "pipe", stderr: "pipe" });
    const exitCode = await proc.exited;
    const stderrText = await new Response(proc.stderr).text();
    const duration = ((Date.now() - t0) / 1000).toFixed(1);
    return { exitCode, stderr: stderrText, duration };
  } catch (e: unknown) {
    return { exitCode: -1, stderr: e instanceof Error ? e.message : String(e), duration: "0.0" };
  }
}

export async function runPostClearCommands(commands: string[]) {
  for (const cmd of commands) {
    const t0 = Date.now();
    console.log(`  ${yellow("⚡")} ${cyan(cmd)}`);
    const shell = process.platform === "win32" ? ["cmd", "/c"] : ["sh", "-c"];
    const proc = Bun.spawn([...shell, cmd], { stdio: ["inherit", "inherit", "inherit"] });
    const exitCode = await proc.exited;
    const duration = ((Date.now() - t0) / 1000).toFixed(1);
    if (exitCode === 0) {
      console.log(`  ${P_SUCCESS} Post-clear done in ${duration}s`);
    } else {
      console.error(`  ${P_ERROR} Post-clear failed (exit ${exitCode}) in ${duration}s`);
    }
  }
}
