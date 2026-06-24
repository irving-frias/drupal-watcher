package cli

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/internal/drush"
	"github.com/irving-frias/drupal-watcher/internal/utils"
	"github.com/irving-frias/drupal-watcher/internal/watcher"
	"golang.org/x/sys/unix"
)

var Version = "0.1.0" // overridden via ldflags at build time

func PkgVersion() string { return Version }

func CmdStart(root string, flags map[string]interface{}, mgr *config.Manager) {
	cfg, err := mgr.LoadConfig(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", utils.P_ERROR, err)
		os.Exit(1)
	}

	// Override debounce from flag
	if d, ok := flags["debounce"].(int); ok && d > 0 {
		cfg.Debounce = d
	}

	// Check PID
	pidStatus, err := config.CheckPid(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to check PID: %v\n", utils.P_WARN, err)
	}
	if pidStatus == "stale" {
		config.RemovePid(root)
	} else if pidStatus != nil {
		fmt.Printf("%s Already running (PID %s). Use %s to stop first.\n",
			utils.P_WARN, utils.Cyan(fmt.Sprintf("%v", pidStatus)), utils.Green("drupal-watcher reset"))
		os.Exit(0)
	}

	// Drush health check
	if !drush.HealthCheck(cfg) {
		fmt.Printf("%s Drush is not available. Starting without health check.\n", utils.P_WARN)
	}

	// Write PID
	if err := config.WritePid(root); err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to write PID: %v\n", utils.P_ERROR, err)
		os.Exit(1)
	}
	defer config.RemovePid(root)

	// Determine log file
	var logFile *os.File
	if lf, ok := flags["log-file"].(string); ok && lf != "" {
		f, err := os.OpenFile(lf, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to open log file: %v\n", utils.P_WARN, err)
		} else {
			logFile = f
			fmt.Printf("%s Logging to %s\n", utils.Timestamp(), utils.Cyan(lf))
		}
	}
	if logFile != nil {
		defer logFile.Close()
	}

	// Set up multi-writer for structured output if log file
	// (In Go, we can't easily intercept fmt.Print. For now, log file is separate.)

	// Handle dotfiles
	noDotfiles := false
	if nd, ok := flags["no-dotfiles"].(bool); ok && nd {
		noDotfiles = true
	}
	if noDotfiles {
		// TODO: implement dotfile exclusion via watcher config
	}

	// Start watcher
	h, err := watcher.Start(cfg, logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to start watcher: %v\n", utils.P_ERROR, err)
		os.Exit(1)
	}

	fmt.Printf("%s Watcher started. PID %d. %s to stop.\n",
		utils.Timestamp(), os.Getpid(), utils.Green("Ctrl+C"))

	// Signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-sigCh:
		fmt.Printf("\n%s Received %s, stopping...\n", utils.Timestamp(), utils.Red(sig.String()))
	case <-h.StopCh:
	}

	watcher.Stop(h)

	// Print stats
	uptime := time.Since(h.Stats.StartTime)
	fmt.Printf("\n%s Watcher stopped. %d changes, %d cache clears, uptime %v\n",
		utils.P_INFO, h.Stats.Changes.Load(), h.Stats.Clears.Load(), uptime)
}

func CmdList(root string, mgr *config.Manager) {
	cfg, err := mgr.LoadConfig(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to load config: %v\n", utils.P_ERROR, err)
		os.Exit(1)
	}

	utils.PrintHeader("Active Drupal Watcher")

	var items []utils.SectionItem
	items = append(items, [2]string{"Routes:", strings.Join(cfg.Routes, ", ")})
	items = append(items, [2]string{"Patterns:", strings.Join(cfg.Patterns, ", ")})
	items = append(items, [2]string{"Debounce:", fmt.Sprintf("%dms", cfg.Debounce)})
	items = append(items, [2]string{"Drush:", cfg.DrushCommand})
	if len(cfg.CommandsPerPattern) > 0 {
		var patterns []string
		for k := range cfg.CommandsPerPattern {
			patterns = append(patterns, k)
		}
		sort.Strings(patterns)
		for _, p := range patterns {
			items = append(items, [2]string{fmt.Sprintf("  %s", p), cfg.CommandsPerPattern[p]})
		}
	}
	utils.PrintSection("Configuration", items)
}

func CmdStatus(root string, mgr *config.Manager) {
	pidStatus, err := config.CheckPid(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", utils.P_ERROR, err)
		os.Exit(1)
	}

	if pidStatus == nil {
		fmt.Printf("%s Drupal Watcher is not running.\n", utils.Yellow("●"))
		os.Exit(0)
	}

	if pidStatus == "stale" {
		fmt.Printf("%s Drupal Watcher is stopped (stale PID). Run %s to clean up.\n",
			utils.Red("●"), utils.Green("drupal-watcher reset"))
		os.Exit(0)
	}

	pidStr := fmt.Sprintf("%v", pidStatus)
	starttime, _ := config.GetStarttime(root)
	pid := 0
	fmt.Sscanf(pidStr, "%d", &pid)
	running := IsPidRunning(pid)

	if running && starttime > 0 {
		uptime := time.Since(time.UnixMilli(starttime))
		fmt.Printf("%s Drupal Watcher is running (PID %s, uptime %v).\n",
			utils.Green("●"), utils.Cyan(pidStr), utils.Green(FormatDuration(uptime)))
	} else if running {
		fmt.Printf("%s Drupal Watcher is running (PID %s).\n",
			utils.Green("●"), utils.Cyan(pidStr))
	} else {
		fmt.Printf("%s Drupal Watcher is stopped (stale PID). Run %s to clean up.\n",
			utils.Red("●"), utils.Green("drupal-watcher reset"))
	}
}

func CmdAdd(root string, args []string, mgr *config.Manager) {
	cfg, err := mgr.LoadConfig(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", utils.P_ERROR, err)
		os.Exit(1)
	}

	// Parse route:pattern pairs or just routes
	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 2)
		route := parts[0]
		if route == "" {
			continue
		}
		cfg.Routes = append(cfg.Routes, route)
		fmt.Printf("%s Added route: %s\n", utils.P_SUCCESS, utils.Cyan(route))

		if len(parts) > 1 {
			pattern := parts[1]
			cfg.Patterns = append(cfg.Patterns, pattern)
			fmt.Printf("%s Added pattern: %s\n", utils.P_SUCCESS, utils.Cyan(pattern))
		}
	}

	if err := mgr.SaveConfig(cfg, root); err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to save config: %v\n", utils.P_ERROR, err)
		os.Exit(1)
	}
}

func CmdRemove(root string, args []string, mgr *config.Manager) {
	cfg, err := mgr.LoadConfig(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", utils.P_ERROR, err)
		os.Exit(1)
	}

	removedRoutes := 0
	removedPatterns := 0

	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 2)
		route := parts[0]

		// Remove route
		newRoutes := make([]string, 0, len(cfg.Routes))
		for _, r := range cfg.Routes {
			if r != route {
				newRoutes = append(newRoutes, r)
			}
		}
		if len(newRoutes) != len(cfg.Routes) {
			removedRoutes++
		}
		cfg.Routes = newRoutes

		// Remove pattern if specified
		if len(parts) > 1 {
			pattern := parts[1]
			newPatterns := make([]string, 0, len(cfg.Patterns))
			for _, p := range cfg.Patterns {
				if p != pattern {
					newPatterns = append(newPatterns, p)
				}
			}
			if len(newPatterns) != len(cfg.Patterns) {
				removedPatterns++
			}
			cfg.Patterns = newPatterns
		}
	}

	if removedRoutes > 0 {
		fmt.Printf("%s Removed %d route(s)\n", utils.P_SUCCESS, removedRoutes)
	}
	if removedPatterns > 0 {
		fmt.Printf("%s Removed %d pattern(s)\n", utils.P_SUCCESS, removedPatterns)
	}

	if err := mgr.SaveConfig(cfg, root); err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to save config: %v\n", utils.P_ERROR, err)
		os.Exit(1)
	}
}

func CmdReset(root string, mgr *config.Manager) {
	pidStatus, err := config.CheckPid(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", utils.P_ERROR, err)
		os.Exit(1)
	}

	if pidStatus != nil { // PID exists (running or stale)
		pidStr := fmt.Sprintf("%v", pidStatus)
		var pid int
		fmt.Sscanf(pidStr, "%d", &pid)

		if IsPidRunning(pid) {
			proc, err := os.FindProcess(pid)
			if err == nil {
				if err := proc.Signal(os.Interrupt); err != nil {
					proc.Kill()
				}
				time.Sleep(500 * time.Millisecond)
			}
		}
	}

	config.RemovePid(root)
	mgr.InvalidateConfigCache(root)
	fmt.Printf("%s Reset complete. PID and config cache cleared.\n", utils.P_SUCCESS)
}

func CmdRestart(root string, flags map[string]interface{}, mgr *config.Manager) {
	CmdReset(root, mgr)
	time.Sleep(500 * time.Millisecond)
	CmdStart(root, flags, mgr)
}

func CmdHelp() {
	fmt.Printf(`Usage: drupal-watcher <command> [options]

Commands:
  start       Start watching file changes
  stop        Stop the watcher (alias: reset)
  restart     Restart the watcher
  status      Show watcher status
  list        Show current configuration
  add         Add route and/or pattern
  remove      Remove route and/or pattern
  reset       Stop watcher and reset PID
  help        Show this help

Options:
  --debounce <ms>        Debounce interval (default: 800)
  --no-dotfiles          Ignore dotfiles
  --log-file <path>      Write logs to file
  --commands-per-pattern <json>  Override pattern commands
`)
}

func IsPidRunning(pid int) bool {
	// On Unix, sending signal 0 checks if the process exists
	err := unix.Kill(pid, unix.Signal(0))
	return err == nil
}

func FormatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
