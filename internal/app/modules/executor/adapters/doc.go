// Package adapters provides alternative command executor implementations.
//
// This directory is a placeholder for future executor adapters (e.g., Docker, SSH, remote).
// The default executor is drush, implemented in pkg/adapters/drush_executor.go.
//
// To add a new adapter:
//  1. Create a file here (e.g., docker_executor.go)
//  2. Implement core.CommandExecutor interface
//  3. Register it in executor/register.go based on config
package adapters
