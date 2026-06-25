package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	GetNotify() bool
}

type Stats struct {
	Changes         atomic.Int64
	Clears          atomic.Int64
	StartTime       time.Time
	TotalDebounceMs atomic.Int64
}

type EventType int

const (
	EventChange EventType = iota
	EventDrush
	EventError
)

type EventMsg struct {
	Type      EventType
	Timestamp time.Time
	File      string
	Changes   int
	Commands  string
	ExitCode  int
	Duration  time.Duration
	Error     error
}

type Handle struct {
	Watcher    *fsnotify.Watcher
	StopCh     chan struct{}
	EventCh    chan EventMsg
	LogFile    *os.File
	Stats      *Stats
	Config     Config
	WatchCount atomic.Int64
	wg         sync.WaitGroup
}

var (
	mu           sync.Mutex
	pending      atomic.Bool
	changedFiles map[string]struct{}
	lastFile     string // kept for the "last change" display message
)

// dirsSkippedByDefault are known high-cardinality directories unlikely to contain
// files that need drush cache clears. Skipping them reduces inotify watches.
var dirsSkippedByDefault = []string{
	"node_modules", ".git", ".svn", ".hg",
	"contrib",      // drupal contrib modules (massive)
	"vendor",       // composer deps
	"bower_components",
	"files",        // drupal files dir
	"css",          // compiled assets (not source)
	"js",           // compiled/minified js
	"images",
	"fonts",
}

func StartWithEvents(cfg Config, eventCh chan EventMsg) (*Handle, error) {
	return start(cfg, nil, eventCh)
}

func Start(cfg Config, logFile *os.File) (*Handle, error) {
	return start(cfg, logFile, nil)
}

func start(cfg Config, logFile *os.File, eventCh chan EventMsg) (*Handle, error) {
	fsnWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	h := &Handle{
		Watcher: fsnWatcher,
		StopCh:  make(chan struct{}),
		EventCh: eventCh,
		LogFile: logFile,
		Stats:   &Stats{StartTime: time.Now()},
		Config:  cfg,
	}

	routes := cfg.GetRoutes()
	patterns := cfg.GetPatterns()
	exclude := cfg.GetExcludePatterns()
	skipDirs := append(append([]string{}, dirsSkippedByDefault...), exclude...)

	// On macOS FSEvents watches recursively by default (one kernel watch per route).
	// On Linux inotify requires every directory to be added individually.
	recursive := runtime.GOOS == "darwin"

	watchDirs := gatherDirs(routes, skipDirs)
	for _, dir := range watchDirs {
		if err := fsnWatcher.Add(dir); err != nil {
			fmt.Printf("%s Failed to watch %s: %v\n", utils.P_WARN, utils.Cyan(dir), err)
		}
	}

	watchCount := len(watchDirs)
	if recursive {
		watchCount = len(routes) // FSEvents tracks entire trees as 1 watch each
	}

	debounceMs := cfg.GetDebounce()
	if debounceMs <= 0 {
		debounceMs = 800
	}
	debounce := time.Duration(debounceMs) * time.Millisecond

	h.WatchCount.Store(int64(watchCount))
	fmt.Printf("%s Watching %d directories (%d kernel watches), debounce %v, %d patterns\n",
		utils.Timestamp(), len(routes), watchCount, debounce, len(patterns))

	var timer *time.Timer
	var debounceMu sync.Mutex

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		for {
			select {
			case event, ok := <-fsnWatcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Chmod) == 0 {
					continue
				}

				// On Linux, add newly created directories to the watcher
				if event.Op&fsnotify.Create != 0 && runtime.GOOS == "linux" {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						base := filepath.Base(event.Name)
						if !isSkippedDir(base, skipDirs) {
							if err := fsnWatcher.Add(event.Name); err == nil {
								watchCount++
							}
						}
					}
				}

				if !matchPattern(event.Name, patterns) || matchExclude(event.Name, exclude) {
					continue
				}

				mu.Lock()
				lastFile = event.Name
				if changedFiles == nil {
					changedFiles = make(map[string]struct{})
				}
				changedFiles[event.Name] = struct{}{}
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
	h.wg.Wait()
}

func processChange(h *Handle) {
	mu.Lock()
	files := changedFiles
	changedFiles = nil
	dispFile := lastFile
	mu.Unlock()

	if len(files) == 0 {
		return
	}

	changes := int64(len(files))
	h.Stats.Changes.Add(changes)

	commandsPerPattern := h.Config.GetCommandsPerPattern()
	seen := make(map[string]struct{})
	var cmds []string
	for f := range files {
		args := getCacheClearArgs(f, commandsPerPattern)
		cmdStr := strings.Join(args, " ")
		if _, ok := seen[cmdStr]; !ok {
			seen[cmdStr] = struct{}{}
			cmds = append(cmds, cmdStr)
		}
	}

	msg := fmt.Sprintf("Change detected: %s", utils.Dim(dispFile))
	if changes > 1 {
		msg = fmt.Sprintf("%d changes detected (last: %s)", changes, utils.Dim(dispFile))
	}
	fmt.Printf("%s %s\n", utils.Timestamp(), msg)

	// Send event to TUI if connected
	if h.EventCh != nil {
		select {
		case h.EventCh <- EventMsg{
			Type:      EventChange,
			Timestamp: time.Now(),
			File:      dispFile,
			Changes:   int(changes),
		}:
		default:
		}
	}

	cmdStr := strings.Join(cmds, " + ")

	// Batch all cache clears into a single drush invocation
	result := drush.RunCacheClears(h.Config, cmds)
	h.Stats.Clears.Add(1)

	status := utils.P_SUCCESS
	if result.ExitCode != 0 {
		status = utils.P_ERROR
	}
	fmt.Printf("%s %s drush %s (%v, exit %d)\n",
		utils.Timestamp(), status, cmdStr, result.Duration, result.ExitCode)

	if result.Stderr != "" {
		fmt.Fprintf(os.Stderr, "  %s\n", utils.Dim(strings.TrimSpace(result.Stderr)))
	}
	if result.Stdout != "" && result.Stdout != "{}" {
		fmt.Printf("  %s\n", utils.Dim(strings.TrimSpace(result.Stdout)))
	}

	// Send event to TUI if connected
	if h.EventCh != nil {
		select {
		case h.EventCh <- EventMsg{
			Type:      EventDrush,
			Timestamp: time.Now(),
			File:      dispFile,
			Changes:   int(changes),
			Commands:  cmdStr,
			ExitCode:  result.ExitCode,
			Duration:  result.Duration,
		}:
		default:
		}
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

func isSkippedDir(name string, skipDirs []string) bool {
	for _, s := range skipDirs {
		if name == s {
			return true
		}
	}
	return false
}

// gatherDirs returns directories to watch.
// On macOS (FSEvents) returns only the top-level routes (recursive by default).
// On Linux (inotify) walks each route and adds every subdirectory.
func gatherDirs(routes []string, skipDirs []string) []string {
	recursive := runtime.GOOS == "darwin"
	dirSet := make(map[string]bool, len(routes))

	for _, route := range routes {
		abs, err := filepath.Abs(route)
		if err != nil {
			continue
		}

		if recursive {
			dirSet[abs] = true
			continue
		}

		filepath.WalkDir(abs, func(path string, d os.DirEntry, err error) error {
			if err != nil || !d.IsDir() {
				return nil
			}
			if isSkippedDir(d.Name(), skipDirs) {
				return filepath.SkipDir
			}
			dirSet[path] = true
			return nil
		})
	}

	dirs := make([]string, 0, len(dirSet))
	for d := range dirSet {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	return dirs
}
