package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/irving-frias/drupal-watcher/internal/app"
	cfgmodule "github.com/irving-frias/drupal-watcher/internal/app/modules/config"
	execmodule "github.com/irving-frias/drupal-watcher/internal/app/modules/executor"
	orcmodule "github.com/irving-frias/drupal-watcher/internal/app/modules/orchestrator"
	uimodule "github.com/irving-frias/drupal-watcher/internal/app/modules/ui"
	watchermodule "github.com/irving-frias/drupal-watcher/internal/app/modules/watcher"
	"github.com/irving-frias/drupal-watcher/internal/app/common"
	"github.com/irving-frias/drupal-watcher/internal/config"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		arg := args[0]
		if arg == "--version" || arg == "-V" || arg == "version" {
			fmt.Printf("drupal-watcher %s (go %s)\n", common.PkgVersion(), strings.TrimPrefix(runtime.Version(), "go"))
			return
		}
		if arg == "--help" || arg == "-h" || arg == "help" {
			printUsage()
			return
		}
	}

	root := "."
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") && !isCommand(args[0]) {
		root = args[0]
	}

	cfgMod := &cfgmodule.Module{WorkDir: root}
	watchMod := &watchermodule.Module{}
	execMod := &execmodule.Module{}
	orcMod := &orcmodule.Module{}
	uiMod := &uimodule.Module{}

	a := app.New(cfgMod, watchMod, execMod, orcMod, uiMod)

	defer func() {
		a.Stop(context.Background())
		config.RemovePid(root)
	}()

	if err := config.WritePid(root); err != nil {
		fmt.Fprintf(os.Stderr, "PID: %v\n", err)
		os.Exit(1)
	}

	if err := a.Start(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func isCommand(s string) bool {
	return s == "start" || s == "watch"
}

func printUsage() {
	fmt.Println(`Drupal Watcher — file watcher with auto drush cache clears

Usage:
  modular-watcher [root]     Start watching the given Drupal root (default: .)
  modular-watcher version    Print version
  modular-watcher help       Show this help

The TUI opens automatically. Press Ctrl+C or Ctrl+D to quit.`)
}
