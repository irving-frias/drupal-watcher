import { existsSync, accessSync, constants } from "fs";
import path from "path";

export function getDrushCommand(config) {
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
  const drushCmdStr = getDrushCommand(config);
  const healthParts = [...drushCmdStr.split(/\s+/), "status", "--format=json"];
  const proc = Bun.spawn(healthParts, { stdout: "pipe", stderr: "pipe" });
  const exitCode = await proc.exited;
  return exitCode === 0;
}

export async function runDrush(drushBase, drushArgsArray) {
  const proc = Bun.spawn([drushBase, ...drushArgsArray], { stdout: "pipe", stderr: "pipe" });
  const exitCode = await proc.exited;
  const stderrText = await new Response(proc.stderr).text();
  return { exitCode, stderr: stderrText };
}

export async function runPostClearCommands(commands) {
  for (const cmd of commands) {
    console.log(`${YELLOW}⚡ ${cmd}${NC}`);
    await Bun.spawn(["sh", "-c", cmd], { stdio: ["inherit", "inherit", "inherit"] }).exited;
  }
}
