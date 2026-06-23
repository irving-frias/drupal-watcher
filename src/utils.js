import { existsSync } from "fs";

// ANSI color codes
export const RED = "\x1b[31m";
export const GREEN = "\x1b[32m";
export const YELLOW = "\x1b[33m";
export const BLUE = "\x1b[34m";
export const CYAN = "\x1b[36m";
export const NC = "\x1b[0m";
export const BOLD = "\x1b[1m";
export const DIM = "\x1b[2m";

// Whether ANSI colors are enabled (can be disabled via --no-colors)
let _colorsEnabled = true;
export function setColorsEnabled(v) { _colorsEnabled = v; }
export function colorsEnabled() { return _colorsEnabled; }

function c(code, s) {
  return _colorsEnabled ? `${code}${s}${NC}` : s;
}

// Color helpers
export function red(s) { return c(RED, s); }
export function green(s) { return c(GREEN, s); }
export function yellow(s) { return c(YELLOW, s); }
export function blue(s) { return c(BLUE, s); }
export function cyan(s) { return c(CYAN, s); }
export function bold(s) { return _colorsEnabled ? `${BOLD}${s}${NC}` : s; }
export function dim(s) { return _colorsEnabled ? `${DIM}${s}${NC}` : s; }

// Message prefixes
export const P_ERROR = `${c(RED, "✖")}`;
export const P_WARN = `${c(YELLOW, "⚠")}`;
export const P_INFO = `${c(BLUE, "ℹ")}`;
export const P_SUCCESS = `${c(GREEN, "✔")}`;

// Timestamp helper
export function timestamp() {
  const d = new Date();
  const hh = String(d.getHours()).padStart(2, "0");
  const mm = String(d.getMinutes()).padStart(2, "0");
  const ss = String(d.getSeconds()).padStart(2, "0");
  return cyan(`[${hh}:${mm}:${ss}]`);
}

// Drupal directory detection
export const POSSIBLE_DOCROOTS = ["docroot", "web", "html", "public", "drupal"];
export const EXCLUDED_DIRS = ["node_modules", ".git", "files"];

// Default watcher patterns
export const DEFAULT_PATTERNS = [
  ".html.twig", ".inc", ".yml", ".module", ".theme",
  ".php", ".info.yml", ".services.yml",
];

// Help display helpers
export function printHeader(title) {
  console.log(`${yellow(title)}`);
}

export function printSection(heading, items) {
  console.log(`\n${blue(`${heading}:`)}`);
  for (const item of items) {
    if (Array.isArray(item)) {
      const [label, desc] = item;
      console.log(`  ${green(label)}  ${desc}`);
    } else {
      console.log(`  ${item}`);
    }
  }
}

export function pathExists(p) {
  try {
    return existsSync(p);
  } catch {
    return false;
  }
}
