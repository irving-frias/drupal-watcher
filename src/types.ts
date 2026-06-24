export interface WatcherConfig {
  routes: string[]
  patterns: string[]
  excludePatterns: string[]
  debounce: number
  drushCmd: string | null
  drushCommand: string
  drushArgs: string[]
  postClearCommands: string[]
  commandsPerPattern: Record<string, string>
  drupalRoot: string | null
}

export interface StartFlags {
  abortOnDrushError: boolean
  watchRoutes: string[]
  noWatchRoutes: string[]
  dryRun: boolean
  verbose: boolean
  noColors: boolean
  debounce: number | null
  noDotfiles: boolean
  logFile: string | null
  commandsPerPattern: Record<string, string>
}

export interface DrushResult {
  exitCode: number
  stderr: string
  duration: string
}

export interface DrushSpawnArgs {
  cmd: string
  args: string[]
}

export interface WatcherHandle {
  stop?: () => void
  close?: () => void
}
