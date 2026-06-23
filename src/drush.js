import { existsSync, accessSync, constants } from "fs";
import path from "path";
import { P_INFO, P_SUCCESS, green, yellow, cyan } from "./utils.js";

export function getDrushCommand(config, options = {}) {
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

export function getDrushSpawnArgs(config) {
  const cmdStr = getDrushCommand(config);
  const parts = cmdStr.split(/\s+/);
  const args = [...parts.slice(1)];
  if (config.drushCommand) args.push(config.drushCommand);
  if (config.drushArgs?.length > 0) args.push(...config.drushArgs);
  return { cmd: parts[0], args };
}

export async function healthCheck(config) {
  const cmdStr = getDrushCommand(config);
  const parts = [...cmdStr.split(/\s+/), "status", "--format=json"];
  const proc = Bun.spawn(parts, { stdout: "pipe", stderr: "pipe" });
  const exitCode = await proc.exited;
  return exitCode === 0;
}

export async function runDrush(drushBase, drushArgsArray) {
  const t0 = Date.now();
  const proc = Bun.spawn([drushBase, ...drushArgsArray], { stdout: "pipe", stderr: "pipe" });
  const exitCode = await proc.exited;
  const stderrText = await new Response(proc.stderr).text();
  const duration = ((Date.now() - t0) / 1000).toFixed(1);
  return { exitCode, stderr: stderrText, duration };
}

export async function runPostClearCommands(commands) {
  for (const cmd of commands) {
    const t0 = Date.now();
    console.log(`  ${yellow("⚡")} ${cyan(cmd)}`);
    const proc = Bun.spawn(["sh", "-c", cmd], { stdio: ["inherit", "inherit", "inherit"] });
    const exitCode = await proc.exited;
    const duration = ((Date.now() - t0) / 1000).toFixed(1);
    if (exitCode === 0) {
      console.log(`  ${P_SUCCESS} Post-clear done in ${duration}s`);
    } else {
      console.error(`  ${P_ERROR} Post-clear failed (exit ${exitCode}) in ${duration}s`);
    }
  }
}
