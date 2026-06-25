package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/irving-frias/drupal-watcher/internal/cli"
	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/pterm/pterm"
)

func main() {
	// Handle panic gracefully
	defer func() {
		if r := recover(); r != nil {
			pterm.Error.Printfln("Panic: %v", r)
			os.Exit(1)
		}
	}()

	mgr := config.NewManager()
	args := os.Args[1:]
	command, flags, extraArgs := parseFlags(args)

	// Handle --config globally
	if cfgPath, ok := flags["config"].(string); ok && cfgPath != "" {
		mgr.SetCustomConfigPath(cfgPath)
	}

	// Handle --version globally before command dispatch
	if _, ok := flags["version"]; ok {
		fmt.Printf("drupal-watcher %s (go %s)\n", cli.PkgVersion(), strings.TrimPrefix(runtime.Version(), "go"))
		return
	}

	switch command {
	case "start":
		root := ""
		if len(extraArgs) > 0 {
			root = extraArgs[0]
		}
		if err := cli.CmdStart(context.Background(), root, flags, mgr); err != nil {
			pterm.Error.Printfln("%v", err)
			os.Exit(1)
		}

	case "stop", "reset":
		root := ""
		if len(extraArgs) > 0 {
			root = extraArgs[0]
		}
		if err := cli.CmdReset(root, mgr); err != nil {
			pterm.Error.Printfln("%v", err)
			os.Exit(1)
		}

	case "restart":
		root := ""
		if len(extraArgs) > 0 {
			root = extraArgs[0]
		}
		if err := cli.CmdRestart(root, flags, mgr); err != nil {
			pterm.Error.Printfln("%v", err)
			os.Exit(1)
		}

	case "status":
		root := ""
		if len(extraArgs) > 0 {
			root = extraArgs[0]
		}
		if err := cli.CmdStatus(root, mgr); err != nil {
			pterm.Error.Printfln("%v", err)
			os.Exit(1)
		}

	case "list", "config":
		root := ""
		if len(extraArgs) > 0 {
			root = extraArgs[0]
		}
		if err := cli.CmdList(root, mgr); err != nil {
			pterm.Error.Printfln("%v", err)
			os.Exit(1)
		}

	case "add":
		root := ""
		routeArgs := extraArgs
		if len(extraArgs) > 0 && !strings.Contains(extraArgs[0], ":") && !strings.HasPrefix(extraArgs[0], "/") {
			root = extraArgs[0]
			routeArgs = extraArgs[1:]
		}
		if err := cli.CmdAdd(root, routeArgs, mgr); err != nil {
			pterm.Error.Printfln("%v", err)
			os.Exit(1)
		}

	case "remove", "rm":
		root := ""
		routeArgs := extraArgs
		if len(extraArgs) > 0 && !strings.Contains(extraArgs[0], ":") && !strings.HasPrefix(extraArgs[0], "/") {
			root = extraArgs[0]
			routeArgs = extraArgs[1:]
		}
		if err := cli.CmdRemove(root, routeArgs, mgr); err != nil {
			pterm.Error.Printfln("%v", err)
			os.Exit(1)
		}

	case "tui":
		root := getRootFlag(flags, extraArgs)
		if err := cli.CmdTui(root, mgr); err != nil {
			pterm.Error.Printfln("%v", err)
			os.Exit(1)
		}

	case "help", "":
		cli.CmdHelp()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		cli.CmdHelp()
		os.Exit(1)
	}
}

func getRootFlag(flags map[string]interface{}, extraArgs []string) string {
	if r, ok := flags["root"].(string); ok && r != "" {
		return r
	}
	if len(extraArgs) > 0 {
		return extraArgs[0]
	}
	return ""
}

func parseFlags(args []string) (command string, flags map[string]interface{}, extraArgs []string) {
	flags = make(map[string]interface{})
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--no-dotfiles":
			flags["no-dotfiles"] = true

		case arg == "--notify":
			flags["notify"] = true

		case arg == "--no-tui":
			flags["no-tui"] = true

		case arg == "--help" || arg == "-h":
			flags["help"] = true
			return "help", flags, nil

		case arg == "--version" || arg == "-V":
			flags["version"] = true

		case strings.HasPrefix(arg, "--debounce="):
			val := strings.TrimPrefix(arg, "--debounce=")
			var d int
			fmt.Sscanf(val, "%d", &d)
			if d > 0 {
				flags["debounce"] = d
			}

		case arg == "--debounce" && i+1 < len(args):
			i++
			var d int
			fmt.Sscanf(args[i], "%d", &d)
			if d > 0 {
				flags["debounce"] = d
			}

		case strings.HasPrefix(arg, "--log-file="):
			flags["log-file"] = strings.TrimPrefix(arg, "--log-file=")

		case arg == "--log-file" && i+1 < len(args):
			i++
			flags["log-file"] = args[i]

		case strings.HasPrefix(arg, "--config="):
			flags["config"] = strings.TrimPrefix(arg, "--config=")

		case arg == "--config" && i+1 < len(args):
			i++
			flags["config"] = args[i]

		case strings.HasPrefix(arg, "--commands-per-pattern="):
			val := strings.TrimPrefix(arg, "--commands-per-pattern=")
			var cpp map[string]string
			if json.Unmarshal([]byte(val), &cpp) == nil {
				flags["commands-per-pattern"] = cpp
			}

		case arg == "--commands-per-pattern" && i+1 < len(args):
			i++
			var cpp map[string]string
			if json.Unmarshal([]byte(args[i]), &cpp) == nil {
				flags["commands-per-pattern"] = cpp
			}

		case strings.HasPrefix(arg, "--site="):
			val := strings.TrimPrefix(arg, "--site=")
			flags["site"] = strings.Split(val, ",")

		case arg == "--site" && i+1 < len(args):
			i++
			flags["site"] = strings.Split(args[i], ",")

		case strings.HasPrefix(arg, "--exclude-site="):
			val := strings.TrimPrefix(arg, "--exclude-site=")
			flags["exclude-site"] = strings.Split(val, ",")

		case arg == "--exclude-site" && i+1 < len(args):
			i++
			flags["exclude-site"] = strings.Split(args[i], ",")

		case strings.HasPrefix(arg, "--uri="):
			flags["uri"] = strings.TrimPrefix(arg, "--uri=")

		case arg == "--uri" && i+1 < len(args):
			i++
			flags["uri"] = args[i]

		case strings.HasPrefix(arg, "--root="):
			flags["root"] = strings.TrimPrefix(arg, "--root=")

		case arg == "--root" && i+1 < len(args):
			i++
			flags["root"] = args[i]

		default:
			positional = append(positional, arg)
		}
	}

	if len(positional) > 0 {
		command = positional[0]
		extraArgs = positional[1:]
	}

	return command, flags, extraArgs
}
