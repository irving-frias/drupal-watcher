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

	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/internal/drush"
	"github.com/irving-frias/drupal-watcher/internal/orchestrator"
	"github.com/irving-frias/drupal-watcher/internal/ui"
	"github.com/irving-frias/drupal-watcher/internal/utils"
	"github.com/irving-frias/drupal-watcher/pkg/adapters"
	"github.com/irving-frias/drupal-watcher/pkg/core"
	"github.com/pterm/pterm"
)

var Version = "0.1.0"

func PkgVersion() string { return Version }

func CmdStart(ctx context.Context, root string, flags map[string]interface{}, mgr *config.Manager) error {
	cfg, err := mgr.LoadConfig(root)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if d, ok := flags["debounce"].(int); ok && d > 0 {
		cfg.Debounce = d
	}

	pidStatus, err := config.CheckPid(root)
	if err != nil {
		pterm.Warning.Printfln("Failed to check PID: %v", err)
	}
	if pidStatus == "stale" {
		config.RemovePid(root)
	} else if pidStatus != nil {
		pterm.Warning.Printfln("Already running (PID %s). Use %s to stop first.",
			utils.Cyan(fmt.Sprintf("%v", pidStatus)), utils.Green("drupal-watcher reset"))
		return nil
	}

	spinner, _ := pterm.DefaultSpinner.Start("Checking drush health...")
	ok := drush.HealthCheck(cfg)
	if ok {
		spinner.Success("Drush is available")
	} else {
		spinner.Warning("Drush is not available. Starting without health check.")
	}

	if err := config.WritePid(root); err != nil {
		return fmt.Errorf("failed to write PID: %w", err)
	}
	defer config.RemovePid(root)

	if nd, ok := flags["no-dotfiles"].(bool); ok && nd {
		cfg.ExcludePatterns = append(cfg.ExcludePatterns, "/.")
	}

	if n, ok := flags["notify"].(bool); ok && n {
		cfg.Notify = true
	}

	drupalRoot := root
	if cfg.DrupalRoot != nil {
		drupalRoot = filepath.Join(root, *cfg.DrupalRoot)
	}

	if uri, ok := flags["uri"].(string); ok && uri != "" {
		cfg.DrushArgs = append([]string{"--uri=" + uri}, cfg.DrushArgs...)
	} else if include, ok := flags["site"].([]string); ok && len(include) > 0 {
		allSites, err := drush.LoadSitesYml(drupalRoot, root)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		filtered := drush.FilterSites(allSites, include, nil)
		if len(filtered) == 0 {
			return fmt.Errorf("no matching sites found in drush/sites.yml")
		}
		var siteList []core.SiteInfo
		for _, s := range filtered {
			siteList = append(siteList, core.SiteInfo{Name: s.Name, URI: s.URI})
		}
		cfg.SetResolvedSites(siteList)
		pterm.Info.Printfln("Watching sites: %s", utils.Cyan(drush.PrintSiteList(filtered)))
	} else if exclude, ok := flags["exclude-site"].([]string); ok && len(exclude) > 0 {
		allSites, err := drush.LoadSitesYml(drupalRoot, root)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		filtered := drush.FilterSites(allSites, nil, exclude)
		if len(filtered) == 0 {
			return fmt.Errorf("all sites excluded. Nothing to watch.")
		}
		var siteList []core.SiteInfo
		for _, s := range filtered {
			siteList = append(siteList, core.SiteInfo{Name: s.Name, URI: s.URI})
		}
		cfg.SetResolvedSites(siteList)
		pterm.Info.Printfln("Watching sites: %s", utils.Cyan(drush.PrintSiteList(filtered)))
	} else if len(cfg.Sites) > 0 {
		allSites, err := drush.LoadSitesYml(drupalRoot, root)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		filtered := drush.FilterSites(allSites, cfg.Sites, nil)
		if len(filtered) == 0 {
			return fmt.Errorf("no matching sites in drush/sites.yml for config sites")
		}
		var siteList []core.SiteInfo
		for _, s := range filtered {
			siteList = append(siteList, core.SiteInfo{Name: s.Name, URI: s.URI})
		}
		cfg.SetResolvedSites(siteList)
		pterm.Info.Printfln("Watching sites: %s (from config)", utils.Cyan(drush.PrintSiteList(filtered)))
	} else if drush.HasMultiSite(drupalRoot) {
		allSites, err := drush.LoadSitesYml(drupalRoot, root)
		if err != nil {
			return fmt.Errorf("multi-site detected: %v.\n  Create drush/sites.yml or drush/sites/*.site.yml to use drupal-watcher.\n  See: https://www.drush.org/latest/using-drush/site-aliases/", err)
		}
		var siteList []core.SiteInfo
		for _, s := range allSites {
			siteList = append(siteList, core.SiteInfo{Name: s.Name, URI: s.URI})
		}
		cfg.SetResolvedSites(siteList)
		pterm.Info.Printfln("Watching all %d sites: %s", len(siteList), utils.Cyan(drush.PrintSiteList(allSites)))
	}

	if resolved := cfg.GetResolvedSites(); len(resolved) > 0 && cfg.DrupalRoot != nil {
		for _, site := range resolved {
			siteRoot := filepath.Join(*cfg.DrupalRoot, "sites", site.Name)
			for _, sub := range []string{"modules", "themes", "profiles", "custom"} {
				dir := filepath.Join(siteRoot, sub)
				if info, err := os.Stat(dir); err == nil && info.IsDir() {
					cfg.Routes = append(cfg.Routes, dir)
				}
			}
		}
	}

	skipDirs := append(adapters.DefaultSkipDirs(), cfg.ExcludePatterns...)

	fsnWatcher, err := adapters.NewFSNotifyWatcher(cfg.Routes, skipDirs)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer fsnWatcher.Close()

	drushExec := adapters.NewDrushExecutor(cfg)

	patternFilter := adapters.NewPatternFilter(cfg.Patterns)
	excludeFilter := adapters.NewExcludeFilter(cfg.ExcludePatterns)
	filters := []core.EventFilter{patternFilter, excludeFilter}

	var resolvedSites []core.SiteInfo
	for _, s := range cfg.GetResolvedSites() {
		resolvedSites = append(resolvedSites, core.SiteInfo{Name: s.Name, URI: s.URI})
	}

	dr := ""
	if cfg.DrupalRoot != nil {
		dr = filepath.Join(root, *cfg.DrupalRoot)
	}

	isTUI := true
	if nt, ok := flags["no-tui"].(bool); ok && nt {
		isTUI = false
	}

	var eventChan chan core.EngineEvent
	logger := adapters.NewSlogLogger("")
	if isTUI && isatty() {
		eventChan = make(chan core.EngineEvent, 100)
		logger = adapters.NewDiscardLogger()
	}

	engineCfg := core.EngineConfig{
		Watcher:  fsnWatcher,
		Executor: drushExec,
		SiteExecutorFactory: func(site core.SiteInfo) core.CommandExecutor {
			return adapters.NewSiteAwareDrushExecutor(cfg, site.Name, site.URI)
		},
		Filters:            filters,
		PostProcessors:     []core.PostProcessor{},
		EventChan:          eventChan,
		Logger:             logger,
		Debounce:           cfg.Debounce,
		Patterns:           cfg.Patterns,
		ExcludePatterns:    cfg.ExcludePatterns,
		CommandsPerPattern: cfg.CommandsPerPattern,
		ResolvedSites:      resolvedSites,
		DrupalRoot:         dr,
	}

	engine := orchestrator.New(engineCfg)

	pterm.Info.Printfln("Routes: %s", utils.Cyan(strings.Join(cfg.Routes, ", ")))
	pterm.Info.Printfln("Patterns: %s", utils.Cyan(strings.Join(cfg.Patterns, ", ")))
	pterm.Success.Printfln("Watcher started. PID %d.", os.Getpid())

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	if isTUI && isatty() {
		go func() {
			if err := engine.Run(ctx); err != nil && err != context.Canceled {
				pterm.Error.Printfln("Engine error: %v", err)
			}
		}()

		if err := ui.Run(eventChan, engine); err != nil {
			pterm.Warning.Printfln("TUI error: %v", err)
		}
	} else {
		if !isTUI {
			pterm.Info.Println("TUI disabled. Type help for commands.")
		}

		cmdCh := make(chan string)
		go stdinReader(ctx, cmdCh)

		go func() {
			if err := engine.Run(ctx); err != nil && err != context.Canceled {
				pterm.Error.Printfln("Engine error: %v", err)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return nil
			case input := <-cmdCh:
				parts := strings.Fields(input)
				if len(parts) == 0 {
					continue
				}
				switch parts[0] {
				case "status", "stats":
					changes, clears := engine.Stats()
					uptime := time.Since(engine.StartTime())
					pterm.Success.Printfln("Watcher running. PID %d", os.Getpid())
					pterm.Info.Printfln("Changes: %d  Clears: %d  Uptime: %v",
						changes, clears, utils.FormatDuration(uptime))
					var m runtime.MemStats
					runtime.ReadMemStats(&m)
					allocMB := float64(m.Alloc) / 1024 / 1024
					memColor := pterm.FgGreen
					if allocMB >= 500 {
						memColor = pterm.FgRed
					} else if allocMB >= 100 {
						memColor = pterm.FgYellow
					}
					pterm.Info.Printfln("Memory: %s", memColor.Sprintf("%.1f MB", allocMB))
				case "list", "config":
					printInteractiveConfig(cfg)
				case "add":
					pterm.Warning.Println("Use `drupal-watcher add` in a separate terminal, then restart.")
				case "remove", "rm":
					pterm.Warning.Println("Use `drupal-watcher remove` in a separate terminal, then restart.")
				case "reload":
					pterm.Warning.Println("Restart the watcher to reload config.")
				case "stop", "quit", "exit":
					fmt.Println("  Stopping watcher...")
					cancel()
					return nil
				case "help":
					printInteractiveHelp()
				default:
					fmt.Printf("  Unknown command: %s. Type %s.\n", parts[0], utils.Green("help"))
				}
			}
		}
	}

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
		text := fmt.Sprintf("Drupal Watcher is running (PID %s, uptime %s).", utils.Cyan(pidStr), utils.Green(utils.FormatDuration(uptime)))
		fmt.Printf("%s %s\n", utils.Green("●"), text)
		fmt.Printf("  Memory: %s\n", utils.Cyan("see 'stats' at runtime"))
	} else if running {
		text := fmt.Sprintf("Drupal Watcher is running (PID %s).", utils.Cyan(pidStr))
		fmt.Printf("%s %s\n", utils.Green("●"), text)
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

	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 2)
		route := sanitizeRoute(parts[0])
		if route == "" {
			pterm.Warning.Printfln("  Invalid route: %s", parts[0])
			continue
		}
		cfg.Routes = append(cfg.Routes, route)
		pterm.Success.Printfln("Added route: %s", utils.Cyan(route))

		if len(parts) > 1 {
			pattern := parts[1]
			cfg.Patterns = append(cfg.Patterns, pattern)
			pterm.Success.Printfln("Added pattern: %s", utils.Cyan(pattern))
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
			pterm.Warning.Printfln("  Invalid route: %s", parts[0])
			continue
		}

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
		pterm.Success.Printfln("Removed %d route(s)", removedRoutes)
	}
	if removedPatterns > 0 {
		pterm.Success.Printfln("Removed %d pattern(s)", removedPatterns)
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

	if pidStatus != nil {
		spinner, _ := pterm.DefaultSpinner.Start("Stopping watcher...")
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
		spinner.Success("Watcher stopped")
	}

	config.RemovePid(root)
	mgr.InvalidateConfigCache(root)
	pterm.Success.Println("Reset complete. PID and config cache cleared.")
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

	logger := adapters.NewSlogLogger("")
	skipDirs := append(adapters.DefaultSkipDirs(), cfg.ExcludePatterns...)

	fsnWatcher, err := adapters.NewFSNotifyWatcher(cfg.Routes, skipDirs)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer fsnWatcher.Close()

	drushExec := adapters.NewDrushExecutor(cfg)

	patternFilter := adapters.NewPatternFilter(cfg.Patterns)
	excludeFilter := adapters.NewExcludeFilter(cfg.ExcludePatterns)
	filters := []core.EventFilter{patternFilter, excludeFilter}

	var resolvedSites []core.SiteInfo
	for _, s := range cfg.GetResolvedSites() {
		resolvedSites = append(resolvedSites, core.SiteInfo{Name: s.Name, URI: s.URI})
	}

	dr := ""
	if cfg.DrupalRoot != nil {
		dr = filepath.Join(root, *cfg.DrupalRoot)
	}

	eventChan := make(chan core.EngineEvent, 100)

	engineCfg := core.EngineConfig{
		Watcher:  fsnWatcher,
		Executor: drushExec,
		SiteExecutorFactory: func(site core.SiteInfo) core.CommandExecutor {
			return adapters.NewSiteAwareDrushExecutor(cfg, site.Name, site.URI)
		},
		Filters:            filters,
		PostProcessors:     []core.PostProcessor{},
		EventChan:          eventChan,
		Logger:             logger,
		Debounce:           cfg.Debounce,
		Patterns:           cfg.Patterns,
		ExcludePatterns:    cfg.ExcludePatterns,
		CommandsPerPattern: cfg.CommandsPerPattern,
		ResolvedSites:      resolvedSites,
		DrupalRoot:         dr,
	}

	engine := orchestrator.New(engineCfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := engine.Run(ctx); err != nil && err != context.Canceled {
			pterm.Error.Printfln("Engine error: %v", err)
		}
	}()

	return ui.Run(eventChan, engine)
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

func isatty() bool {
	if runtime.GOOS == "windows" {
		return true
	}
	_, err := os.Stat("/dev/tty")
	return err == nil
}

func printInteractiveConfig(cfg config.Config) {
	fmt.Printf("  Routes: %s\n", utils.Cyan(strings.Join(cfg.Routes, ", ")))
	fmt.Printf("  Patterns: %s\n", utils.Cyan(strings.Join(cfg.Patterns, ", ")))
	fmt.Printf("  Debounce: %dms\n", cfg.Debounce)
}

func printInteractiveHelp() {
	fmt.Println("  Commands:")
	fmt.Println("    status              Show watcher status, stats and memory")
	fmt.Println("    list                Show current configuration")
	fmt.Println("    stop/quit/exit      Stop the watcher")
	fmt.Println("    help                Show this help")
	fmt.Println()
	fmt.Println("  Top-level commands (run without start):")
	fmt.Println("    drupal-watcher status   Show watcher status")
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
