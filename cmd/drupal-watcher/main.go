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
	"github.com/irving-frias/drupal-watcher/internal/health"
	"github.com/irving-frias/drupal-watcher/internal/validate"
	"github.com/pterm/pterm"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		arg := args[0]
		if arg == "validate" {
			root := "."
			if len(args) > 1 {
				root = args[1]
			}
			runValidate(root)
			return
		}
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfgMod := &cfgmodule.Module{WorkDir: root}
	watchMod := &watchermodule.Module{}
	execMod := &execmodule.Module{}
	orcMod := &orcmodule.Module{}
	uiMod := &uimodule.Module{}

	a := app.New(cfgMod, watchMod, execMod, orcMod, uiMod)

	defer func() {
		a.Stop(ctx)
		config.RemovePid(root)
	}()

	if err := config.WritePid(root); err != nil {
		fmt.Fprintf(os.Stderr, "PID: %v\n", err)
		os.Exit(1)
	}

	go health.Run(ctx)

	if err := a.Start(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func isCommand(s string) bool {
	return s == "start" || s == "watch" || s == "validate"
}

func runValidate(root string) {
	result := validate.Validate(root)
	for _, e := range result.Entries {
		if e.OK {
			pterm.Success.Printfln("  %s: %s", e.Field, e.Message)
		} else {
			pterm.Error.Printfln("  %s: %s", e.Field, e.Message)
		}
	}
	if !result.Pass {
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Drupal Watcher — file watcher with auto drush cache clears

Usage:
  modular-watcher [root]       Start watching (default: .)
  modular-watcher validate     Validate configuration and environment
  modular-watcher version      Print version
  modular-watcher help         Show this help

The TUI opens automatically. Press Ctrl+C or Ctrl+D to quit.`)
}
