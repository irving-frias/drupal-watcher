package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/internal/drush"
	"github.com/irving-frias/drupal-watcher/internal/tui"
	"github.com/irving-frias/drupal-watcher/internal/utils"
	"github.com/irving-frias/drupal-watcher/internal/watcher"
)

var Version = "0.1.0" // overridden via ldflags at build time

func PkgVersion() string { return Version }

func CmdStart(ctx context.Context, root string, flags map[string]interface{}, mgr *config.Manager) error {
	cfg, err := mgr.LoadConfig(root)
	if err != nil {
		return fmt.Errorf("%w", err)
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
		return nil
	}

	// Drush health check
	if !drush.HealthCheck(cfg) {
		fmt.Printf("%s Drush is not available. Starting without health check.\n", utils.P_WARN)
	}

	// Write PID
	if err := config.WritePid(root); err != nil {
		return fmt.Errorf("failed to write PID: %w", err)
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

	// Handle dotfiles
	if nd, ok := flags["no-dotfiles"].(bool); ok && nd {
		cfg.ExcludePatterns = append(cfg.ExcludePatterns, "/.")
	}

	// Handle notify
	if n, ok := flags["notify"].(bool); ok && n {
		cfg.Notify = true
	}

	// Site resolution
	drupalRoot := root
	if cfg.DrupalRoot != nil {
		drupalRoot = filepath.Join(root, *cfg.DrupalRoot)
	}

	if uri, ok := flags["uri"].(string); ok && uri != "" {
		// Explicit --uri: single site mode
		cfg.DrushArgs = append([]string{"--uri=" + uri}, cfg.DrushArgs...)
	} else if include, ok := flags["site"].([]string); ok && len(include) > 0 {
		// --site flag: load only specified sites from sites.yml
		allSites, err := drush.LoadSitesYml(drupalRoot, root)
		if err != nil {
			return fmt.Errorf("%s %v", utils.P_ERROR, err)
		}
		filtered := drush.FilterSites(allSites, include, nil)
		if len(filtered) == 0 {
			return fmt.Errorf("%s No matching sites found in drush/sites.yml", utils.P_ERROR)
		}
		var siteList []watcher.SiteInfo
		for _, s := range filtered {
			siteList = append(siteList, watcher.SiteInfo{Name: s.Name, URI: s.URI})
		}
		cfg.SetResolvedSites(siteList)
		fmt.Printf("%s Watching sites: %s\n", utils.P_INFO, utils.Cyan(drush.PrintSiteList(filtered)))
	} else if exclude, ok := flags["exclude-site"].([]string); ok && len(exclude) > 0 {
		// --exclude-site: load all from sites.yml, exclude some
		allSites, err := drush.LoadSitesYml(drupalRoot, root)
		if err != nil {
			return fmt.Errorf("%s %v", utils.P_ERROR, err)
		}
		filtered := drush.FilterSites(allSites, nil, exclude)
		if len(filtered) == 0 {
			return fmt.Errorf("%s All sites excluded. Nothing to watch.", utils.P_ERROR)
		}
		var siteList []watcher.SiteInfo
		for _, s := range filtered {
			siteList = append(siteList, watcher.SiteInfo{Name: s.Name, URI: s.URI})
		}
		cfg.SetResolvedSites(siteList)
		fmt.Printf("%s Watching sites: %s\n", utils.P_INFO, utils.Cyan(drush.PrintSiteList(filtered)))
	} else if len(cfg.Sites) > 0 {
		// Persisted sites from config file
		allSites, err := drush.LoadSitesYml(drupalRoot, root)
		if err != nil {
			return fmt.Errorf("%s %v", utils.P_ERROR, err)
		}
		filtered := drush.FilterSites(allSites, cfg.Sites, nil)
		if len(filtered) == 0 {
			return fmt.Errorf("%s No matching sites in drush/sites.yml for config sites", utils.P_ERROR)
		}
		var siteList []watcher.SiteInfo
		for _, s := range filtered {
			siteList = append(siteList, watcher.SiteInfo{Name: s.Name, URI: s.URI})
		}
		cfg.SetResolvedSites(siteList)
		fmt.Printf("%s Watching sites: %s (from config)\n", utils.P_INFO, utils.Cyan(drush.PrintSiteList(filtered)))
	} else if drush.HasMultiSite(drupalRoot) {
		// Multi-site detected, try to load sites.yml
		allSites, err := drush.LoadSitesYml(drupalRoot, root)
		if err != nil {
			return fmt.Errorf("%s Multi-site detected in sites/. Create drush/sites.yml or drush/sites/*.site.yml to use drupal-watcher.\n  See: https://www.drush.org/latest/using-drush/site-aliases/", utils.P_ERROR)
		}
		var siteList []watcher.SiteInfo
		for _, s := range allSites {
			siteList = append(siteList, watcher.SiteInfo{Name: s.Name, URI: s.URI})
		}
		cfg.SetResolvedSites(siteList)
		fmt.Printf("%s Watching all %d sites: %s\n", utils.P_INFO, len(siteList), utils.Cyan(drush.PrintSiteList(allSites)))
	}

	// Start watcher
	h, err := watcher.Start(cfg, logFile)
	if err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	// Show startup info
	utils.PrintMemStats(utils.GetMemStats(h.WatchCount.Load()))
	fmt.Printf("  Routes: %s\n", utils.Cyan(strings.Join(cfg.Routes, ", ")))
	fmt.Printf("  Patterns: %s\n", utils.Cyan(strings.Join(cfg.Patterns, ", ")))
	fmt.Printf("%s Watcher started. PID %d.\n",
		utils.Timestamp(), os.Getpid())

	// Try TUI, fall back to interactive stdin if no TTY
	isTUI := true
	if nt, ok := flags["no-tui"].(bool); ok && nt {
		isTUI = false
	}

	if isTUI && isatty() {
		eventCh := make(chan watcher.EventMsg, 100)
		h.EventCh = eventCh

		p := tea.NewProgram(tui.NewModel(h), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("%s TUI error: %v\n", utils.P_WARN, err)
		}
	} else {
		if !isTUI {
			fmt.Printf("  %s TUI disabled. Type %s for commands.\n", utils.Timestamp(), utils.Green("help"))
		}
		cmdCh := make(chan string)
		go stdinReader(ctx, cmdCh)

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

		stopped := false
		for !stopped {
			select {
			case sig := <-sigCh:
				fmt.Printf("\n%s Received %s, stopping...\n", utils.Timestamp(), utils.Red(sig.String()))
				stopped = true

			case <-h.StopCh:
				stopped = true

			case <-ctx.Done():
				stopped = true

			case input := <-cmdCh:
				parts := strings.Fields(input)
				if len(parts) == 0 {
					continue
				}
				switch parts[0] {
				case "status":
					printInteractiveStatus(h)
				case "list", "config":
					printInteractiveConfig(cfg)
				case "stats":
					printInteractiveStatus(h)
				case "add":
					if len(parts) < 2 {
						fmt.Println("  Usage: add <route> [pattern]")
						break
					}
					route := sanitizeRoute(parts[1])
					if route == "" {
						fmt.Printf("  %s Invalid route.\n", utils.P_WARN)
						break
					}
					pattern := ""
					if len(parts) > 2 {
						pattern = parts[2]
					}
					cfg.Routes = append(cfg.Routes, route)
					if pattern != "" {
						cfg.Patterns = append(cfg.Patterns, pattern)
					}
					if err := mgr.SaveConfig(cfg, root); err != nil {
						fmt.Printf("  %s Failed to save config: %v\n", utils.P_ERROR, err)
					} else {
						fmt.Printf("  %s Added route: %s\n", utils.P_SUCCESS, utils.Cyan(route))
						h = restartWatcher(h, cfg, logFile)
					}
				case "remove", "rm":
					if len(parts) < 2 {
						fmt.Println("  Usage: remove <route>")
						break
					}
					route := sanitizeRoute(parts[1])
					if route == "" {
						fmt.Printf("  %s Invalid route.\n", utils.P_WARN)
						break
					}
					newRoutes := make([]string, 0, len(cfg.Routes))
					for _, r := range cfg.Routes {
						if r != route {
							newRoutes = append(newRoutes, r)
						}
					}
					if len(newRoutes) == len(cfg.Routes) {
						fmt.Printf("  %s Route not found: %s\n", utils.P_WARN, utils.Cyan(route))
					} else {
						cfg.Routes = newRoutes
						if err := mgr.SaveConfig(cfg, root); err != nil {
							fmt.Printf("  %s Failed to save config: %v\n", utils.P_ERROR, err)
						} else {
							fmt.Printf("  %s Removed route: %s\n", utils.P_SUCCESS, utils.Cyan(route))
							h = restartWatcher(h, cfg, logFile)
						}
					}
				case "reload":
					newCfg, err := mgr.LoadConfig(root)
					if err != nil {
						fmt.Printf("  %s Failed to reload config: %v\n", utils.P_ERROR, err)
					} else {
						cfg = newCfg
						h = restartWatcher(h, cfg, logFile)
						fmt.Printf("  %s Config reloaded.\n", utils.P_SUCCESS)
					}
				case "stop", "quit", "exit":
					fmt.Println("  Stopping watcher...")
					stopped = true
				case "monitor", "m":
					printInteractiveStatus(h)
					monitorLoop(ctx, h)
				case "help":
					printInteractiveHelp()
				default:
					fmt.Printf("  Unknown command: %s. Type %s.\n", parts[0], utils.Green("help"))
				}
			}
		}
	}

	watcher.Stop(h)

	uptime := time.Since(h.Stats.StartTime)
	changes := h.Stats.Changes.Load()
	clears := h.Stats.Clears.Load()

	statusIcon := utils.P_SUCCESS
	if changes > 0 && clears == 0 {
		statusIcon = utils.P_WARN
	}
	fmt.Printf("\n%s Watcher stopped. %d changes, %d clears, uptime %v\n",
		statusIcon, changes, clears, FormatDuration(uptime))
	utils.PrintMemStats(utils.GetMemStats(h.WatchCount.Load()))
	return nil
}

func stdinReader(ctx context.Context, cmdCh chan<- string) {
	scanner := bufio.NewScanner(os.Stdin)
	result := make(chan string, 1)
	defer close(result)
	for {
		go func() {
			if scanner.Scan() {
				select {
				case result <- strings.TrimSpace(scanner.Text()):
				case <-ctx.Done():
				}
			}
		}()
		select {
		case <-ctx.Done():
			return
		case line := <-result:
			cmdCh <- line
		}
	}
}

func sanitizeRoute(route string) string {
	cleaned := filepath.Clean(route)
	if !filepath.IsAbs(cleaned) {
		return ""
	}
	if info, err := os.Stat(cleaned); err != nil || !info.IsDir() {
		return ""
	}
	return cleaned
}

func CmdList(root string, mgr *config.Manager) error {
	cfg, err := mgr.LoadConfig(root)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
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
	return nil
}

func CmdMonitor(root string, mgr *config.Manager) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigCh)

	done := make(chan struct{}, 1)

	// Print initial status, then refresh every 1s
	printMonitorStatus(root, mgr)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Print("\033[5A\033[J")
			printMonitorStatus(root, mgr)
		case <-sigCh:
			fmt.Print("\033[5A\033[J")
			printMonitorStatus(root, mgr)
			fmt.Println("  Monitoring stopped.")
			close(done)
			return nil
		case <-done:
			return nil
		}
	}
}

func printMonitorStatus(root string, mgr *config.Manager) {
	pidStatus, err := config.CheckPid(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", utils.P_ERROR, err)
		return
	}

	if pidStatus == nil {
		fmt.Printf("%s Drupal Watcher is not running.\n", utils.Yellow("●"))
		fmt.Println()
		fmt.Println("  Press Ctrl+C to stop monitoring.")
		return
	}

	if pidStatus == "stale" {
		fmt.Printf("%s Drupal Watcher is stopped (stale PID).\n", utils.Red("●"))
		fmt.Println()
		fmt.Println("  Press Ctrl+C to stop monitoring.")
		return
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
		fmt.Printf("%s Drupal Watcher is stopped (stale PID).\n", utils.Red("●"))
	}
	fmt.Println()
	fmt.Println("  Press Ctrl+C to stop monitoring.")
}

func CmdStatus(root string, mgr *config.Manager) error {
	pidStatus, err := config.CheckPid(root)
	if err != nil {
		return fmt.Errorf("failed to check PID: %w", err)
	}

	if pidStatus == nil {
		fmt.Printf("%s Drupal Watcher is not running.\n", utils.Yellow("●"))
		return nil
	}

	if pidStatus == "stale" {
		fmt.Printf("%s Drupal Watcher is stopped (stale PID). Run %s to clean up.\n",
			utils.Red("●"), utils.Green("drupal-watcher reset"))
		return nil
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
		fmt.Printf("  Memory: %s\n", utils.Cyan("see 'stats' at runtime"))
	} else if running {
		fmt.Printf("%s Drupal Watcher is running (PID %s).\n",
			utils.Green("●"), utils.Cyan(pidStr))
	} else {
		fmt.Printf("%s Drupal Watcher is stopped (stale PID). Run %s to clean up.\n",
			utils.Red("●"), utils.Green("drupal-watcher reset"))
	}
	return nil
}

func CmdAdd(root string, args []string, mgr *config.Manager) error {
	cfg, err := mgr.LoadConfig(root)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// Parse route:pattern pairs or just routes
	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 2)
		route := sanitizeRoute(parts[0])
		if route == "" {
			fmt.Printf("  %s Invalid route: %s\n", utils.P_WARN, parts[0])
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
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

func CmdRemove(root string, args []string, mgr *config.Manager) error {
	cfg, err := mgr.LoadConfig(root)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	removedRoutes := 0
	removedPatterns := 0

	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 2)
		route := sanitizeRoute(parts[0])
		if route == "" {
			fmt.Printf("  %s Invalid route: %s\n", utils.P_WARN, parts[0])
			continue
		}

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
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

func CmdReset(root string, mgr *config.Manager) error {
	pidStatus, err := config.CheckPid(root)
	if err != nil {
		return fmt.Errorf("failed to check PID: %w", err)
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
	return nil
}

func CmdRestart(root string, flags map[string]interface{}, mgr *config.Manager) error {
	if err := CmdReset(root, mgr); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	return CmdStart(context.Background(), root, flags, mgr)
}

func CmdTui(root string, mgr *config.Manager) error {
	cfg, err := mgr.LoadConfig(root)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	cfg.Debounce = 200
	eventCh := make(chan watcher.EventMsg, 100)
	h, err := watcher.StartWithEvents(cfg, eventCh)
	if err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}
	defer watcher.Stop(h)

	p := tea.NewProgram(tui.NewModel(h), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

func CmdHelp() {
	fmt.Printf(`Usage: drupal-watcher <command> [options]

Commands:
  start       Start watching file changes
  stop        Stop the watcher (alias: reset)
  restart     Restart the watcher
  status      Show watcher status
  monitor     Auto-refresh status every 2 seconds
  list        Show current configuration
  add         Add route and/or pattern
  remove      Remove route and/or pattern
  reset       Stop watcher and reset PID
  tui         Terminal UI (experimental)
  help        Show this help

Options:
  --root <path>          Drupal root directory (default: cwd)
  --debounce <ms>        Debounce interval (default: 800)
  --no-dotfiles          Ignore dotfiles
  --no-tui               Disable TUI, use interactive CLI instead
  --notify               Send desktop notification on cache clear
  --log-file <path>      Write logs to file
  --commands-per-pattern <json>  Override pattern commands
`)
}

func restartWatcher(h *watcher.Handle, cfg config.Config, logFile *os.File) *watcher.Handle {
	watcher.Stop(h)
	time.Sleep(200 * time.Millisecond)
	newH, err := watcher.Start(cfg, logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to restart watcher: %v\n", utils.P_ERROR, err)
		return h
	}
	fmt.Printf("  %s Watcher restarted.\n", utils.P_SUCCESS)
	return newH
}

func isatty() bool {
	if runtime.GOOS == "windows" {
		return true
	}
	_, err := os.Stat("/dev/tty")
	return err == nil
}

func printInteractiveStatus(h *watcher.Handle) {
	uptime := time.Since(h.Stats.StartTime)
	fmt.Printf("  %s Watcher running. PID %d\n", utils.Green("●"), os.Getpid())
	fmt.Printf("  Changes: %d  Clears: %d  Uptime: %v\n",
		h.Stats.Changes.Load(), h.Stats.Clears.Load(), FormatDuration(uptime))
	utils.PrintMemStats(utils.GetMemStats(h.WatchCount.Load()))
}

func monitorLoop(ctx context.Context, h *watcher.Handle) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		scanner := bufio.NewScanner(os.Stdin)
		line := make(chan struct{}, 1)
		go func() {
			scanner.Scan()
			close(line)
		}()
		select {
		case <-line:
		case <-ctx.Done():
		}
	}()

	fmt.Println("  Monitor mode (press Enter to stop)...")
	for {
		select {
		case <-ticker.C:
			fmt.Print("\033[4A\033[J") // move up 4 lines, clear to end
			printInteractiveStatus(h)
			fmt.Println("  Monitor mode (press Enter to stop)...")
		case <-done:
			fmt.Print("\033[4A\033[J")
			printInteractiveStatus(h)
			return
		case <-h.StopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

func printInteractiveConfig(cfg config.Config) {
	fmt.Printf("  Routes: %s\n", utils.Cyan(strings.Join(cfg.Routes, ", ")))
	fmt.Printf("  Patterns: %s\n", utils.Cyan(strings.Join(cfg.Patterns, ", ")))
	fmt.Printf("  Debounce: %dms\n", cfg.Debounce)
}

func printInteractiveHelp() {
	fmt.Println("  Commands:")
	fmt.Println("    status              Show watcher status, stats and memory")
	fmt.Println("    monitor (m)         Auto-refresh status every 2s")
	fmt.Println("    list                Show current configuration")
	fmt.Println("    stats               Show runtime statistics")
	fmt.Println("    add <route>         Add a route to watch")
	fmt.Println("    remove <route>      Remove a watched route")
	fmt.Println("    reload              Reload config from file")
	fmt.Println("    stop/quit/exit      Stop the watcher")
	fmt.Println("    help                Show this help")
	fmt.Println()
	fmt.Println("  Top-level commands (run without start):")
	fmt.Println("    drupal-watcher status   Show watcher status")
	fmt.Println("    drupal-watcher monitor  Auto-refresh watcher status")
	fmt.Println("    drupal-watcher list     Show configuration")
}

func IsPidRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	if err := proc.Signal(os.Signal(syscall.Signal(0))); err != nil {
		return false
	}
	return true
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
