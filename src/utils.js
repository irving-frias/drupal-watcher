import { existsSync } from "fs";

// ANSI color codes
export const RED = "\x1b[31m";
export const GREEN = "\x1b[32m";
export const YELLOW = "\x1b[33m";
export const BLUE = "\x1b[34m";
export const CYAN = "\x1b[36m";
export const NC = "\x1b[0m";

// Message prefixes
export const ERROR = `${RED}✖${NC}`;
export const WARN = `${YELLOW}⚠${NC}`;
export const INFO = `${BLUE}ℹ${NC}`;
export const SUCCESS = `${GREEN}✔${NC}`;

// Drupal directory detection
export const POSSIBLE_DOCROOTS = ["docroot", "web", "html", "public", "drupal"];
export const EXCLUDED_DIRS = ["node_modules", ".git", "files"];

// Default watcher patterns
export const DEFAULT_PATTERNS = [
  ".html.twig", ".inc", ".yml", ".module", ".theme",
  ".php", ".info.yml", ".services.yml",
];

export function bold(s) {
  return `${GREEN}${s}${NC}`;
}

export function dim(s) {
  return `${YELLOW}${s}${NC}`;
}

export function pathExists(p) {
  try {
    return existsSync(p);
  } catch {
    return false;
  }
}
