// Package adapters provides alternative file watcher implementations.
//
// This directory is a placeholder for future watcher adapters (e.g., inotify, kqueue, FSEvents).
// Current watchers are implemented in pkg/adapters/ (fsnotify, polling, hybrid).
//
// To add a new adapter:
//  1. Create a file here (e.g., inotify_watcher.go)
//  2. Implement core.Watcher interface
//  3. Register it in watcher/register.go based on config
package adapters
