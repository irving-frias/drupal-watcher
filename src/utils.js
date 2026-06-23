export const RED = "\x1b[31m";
export const GREEN = "\x1b[32m";
export const YELLOW = "\x1b[33m";
export const BLUE = "\x1b[34m";
export const NC = "\x1b[0m";

export const POSSIBLE_DOCROOTS = ["docroot", "web", "html", "public", "drupal"];
export const EXCLUDED_DIRS = ["node_modules", ".git", "files"];

export function bold(s) {
  return `${GREEN}${s}${NC}`;
}

export function dim(s) {
  return `${YELLOW}${s}${NC}`;
}
