package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/irving-frias/drupal-watcher/internal/drush"
	"github.com/irving-frias/drupal-watcher/internal/utils"
)

type Config interface {
	GetRoutes() []string
	GetPatterns() []string
	GetExcludePatterns() []string
	GetDebounce() int
	GetDrushCmd() *string
	GetDrushCommand() string
	GetDrushArgs() []string
	GetPostClearCommands() []string
	GetCommandsPerPattern() map[string]string
	GetDrupalRoot() *string
}

type Stats struct {
	Changes    atomic.Int64
	Clears     atomic.Int64
	StartTime  time.Time
	TotalDebounceMs atomic.Int64
}

type Handle struct {
	Watcher *fsnotify.Watcher
	StopCh  chan struct{}
	LogFile *os.File
	Stats   *Stats
	Config  Config
}

var (
	mu       sync.Mutex
	pending  atomic.Bool
	lastFile string
)

func Start(cfg Config, logFile *os.File) (*Handle, error) {
	fsnWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	h := &Handle{
		Watcher: fsnWatcher,
		StopCh:  make(chan struct{}),
		LogFile: logFile,
		Stats:   &Stats{StartTime: time.Now()},
		Config:  cfg,
	}

	routes := cfg.GetRoutes()
	patterns := cfg.GetPatterns()
	exclude := cfg.GetExcludePatterns()

	watchDirs := gatherDirs(routes, exclude)
	for _, dir := range watchDirs {
		if err := fsnWatcher.Add(dir); err != nil {
			fmt.Printf("%s Failed to watch %s: %v\n", utils.P_WARN, utils.Cyan(dir), err)
		}
	}

	debounceMs := cfg.GetDebounce()
	if debounceMs <= 0 {
		debounceMs = 800
	}
	debounce := time.Duration(debounceMs) * time.Millisecond

	fmt.Printf("%s Watching %d directories, debounce %v, %d patterns\n",
		utils.Timestamp(), len(watchDirs), debounce, len(patterns))

	var timer *time.Timer
	var debounceMu sync.Mutex

	go func() {
		for {
			select {
			case event, ok := <-fsnWatcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Chmod) == 0 {
					continue
				}
				if !matchPattern(event.Name, patterns) || matchExclude(event.Name, exclude) {
					continue
				}

				mu.Lock()
				lastFile = event.Name
				mu.Unlock()
				pending.Store(true)

				debounceMu.Lock()
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(debounce, func() {
					if !pending.Load() {
						return
					}
					pending.Store(false)
					processChange(h)
				})
				debounceMu.Unlock()

			case err, ok := <-fsnWatcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("%s Watcher error: %v\n", utils.P_WARN, err)
			}
		}
	}()

	return h, nil
}

func Stop(h *Handle) {
	close(h.StopCh)
	h.Watcher.Close()
	if h.LogFile != nil {
		h.LogFile.Close()
	}
}

func processChange(h *Handle) {
	mu.Lock()
	file := lastFile
	mu.Unlock()

	h.Stats.Changes.Add(1)

	fmt.Printf("%s Change detected: %s\n", utils.Timestamp(), utils.Dim(file))

	commandsPerPattern := h.Config.GetCommandsPerPattern()
	args := getCacheClearArgs(file, commandsPerPattern)

	drushBase := drush.GetCmd(h.Config)
	result := drush.Run(drushBase, args)
	h.Stats.Clears.Add(1)

	status := utils.P_SUCCESS
	if result.ExitCode != 0 {
		status = utils.P_ERROR
	}
	fmt.Printf("%s %s drush %s (%v, exit %d)\n",
		utils.Timestamp(), status, strings.Join(args, " "), result.Duration, result.ExitCode)

	if result.Stderr != "" {
		fmt.Fprintf(os.Stderr, "  %s\n", utils.Dim(strings.TrimSpace(result.Stderr)))
	}
	if result.Stdout != "" && result.Stdout != "{}" {
		fmt.Printf("  %s\n", utils.Dim(strings.TrimSpace(result.Stdout)))
	}

	postClear := h.Config.GetPostClearCommands()
	if len(postClear) > 0 {
		drush.RunPostClearCommands(postClear)
	}
}

func getCacheClearArgs(file string, commandsPerPattern map[string]string) []string {
	type kv struct {
		pattern string
		command string
	}
	var sorted []kv
	for k, v := range commandsPerPattern {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return len(sorted[i].pattern) > len(sorted[j].pattern)
	})

	for _, kv := range sorted {
		if strings.HasSuffix(file, kv.pattern) {
			return strings.Fields(kv.command)
		}
	}
	return []string{"cr"}
}

func matchPattern(name string, patterns []string) bool {
	ext := filepath.Ext(name)
	if ext == "" {
		return false
	}
	for _, p := range patterns {
		if strings.HasSuffix(name, p) {
			return true
		}
		if ext == p {
			return true
		}
	}
	return false
}

func matchExclude(name string, excludes []string) bool {
	for _, e := range excludes {
		if strings.Contains(name, e) {
			return true
		}
	}
	return false
}

func gatherDirs(routes []string, excludes []string) []string {
	dirSet := make(map[string]bool)
	for _, route := range routes {
		abs, err := filepath.Abs(route)
		if err != nil {
			continue
		}
		filepath.WalkDir(abs, func(path string, d os.DirEntry, err error) error {
			if err != nil || !d.IsDir() {
				return nil
			}
			base := d.Name()
			for _, ex := range excludes {
				if base == ex || strings.Contains(path, "/"+ex+"/") || strings.HasSuffix(path, "/"+ex) {
					return filepath.SkipDir
				}
			}
			dirSet[path] = true
			return nil
		})
	}
	var dirs []string
	for d := range dirSet {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	return dirs
}
