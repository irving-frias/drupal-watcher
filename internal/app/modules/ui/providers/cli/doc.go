// Package cli provides a non-interactive CLI provider for drupal-watcher.
//
// This directory is a placeholder for a future CLI mode that runs without the Bubble Tea TUI.
// The current TUI implementation is in providers/tui/.
//
// To implement:
//  1. Create a file here (e.g., runner.go)
//  2. Accept engine events via a channel
//  3. Print status updates to stdout
//  4. Register it in ui/register.go as an alternative to the TUI
package cli
