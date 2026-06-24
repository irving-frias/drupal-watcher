export const RED = "\x1b[31m";
export const GREEN = "\x1b[32m";
export const YELLOW = "\x1b[33m";
export const BLUE = "\x1b[34m";
export const CYAN = "\x1b[36m";
export const NC = "\x1b[0m";
export const BOLD = "\x1b[1m";
export const DIM = "\x1b[2m";

let _colorsEnabled = true;
export function setColorsEnabled(v: boolean) { _colorsEnabled = v; }
export function colorsEnabled() { return _colorsEnabled; }

function c(code: string, s: string) {
  return _colorsEnabled ? `${code}${s}${NC}` : s;
}

export function red(s: string) { return c(RED, s); }
export function green(s: string) { return c(GREEN, s); }
export function yellow(s: string) { return c(YELLOW, s); }
export function blue(s: string) { return c(BLUE, s); }
export function cyan(s: string) { return c(CYAN, s); }
export function bold(s: string) { return _colorsEnabled ? `${BOLD}${s}${NC}` : s; }
export function dim(s: string) { return _colorsEnabled ? `${DIM}${s}${NC}` : s; }

export const P_ERROR = `${c(RED, "✖")}`;
export const P_WARN = `${c(YELLOW, "⚠")}`;
export const P_INFO = `${c(BLUE, "ℹ")}`;
export const P_SUCCESS = `${c(GREEN, "✔")}`;

export function timestamp() {
  const d = new Date();
  const hh = String(d.getHours()).padStart(2, "0");
  const mm = String(d.getMinutes()).padStart(2, "0");
  const ss = String(d.getSeconds()).padStart(2, "0");
  return cyan(`[${hh}:${mm}:${ss}]`);
}

export const POSSIBLE_DOCROOTS = ["docroot", "web", "html", "public", "drupal"];
export const EXCLUDED_DIRS = ["node_modules", ".git", "files"];

export const DEFAULT_PATTERNS = [
  ".html.twig", ".inc", ".yml", ".module", ".theme",
  ".php", ".info.yml", ".services.yml",
];

export function printHeader(title: string) {
  console.log(`${yellow(title)}`);
}

type SectionItem = string | [string, string];

export function printSection(heading: string, items: SectionItem[]) {
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
