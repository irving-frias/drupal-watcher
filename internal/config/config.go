package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/irving-frias/drupal-watcher/internal/utils"
)

func nowMs() int64 { return time.Now().UnixMilli() }

type Config struct {
	Routes              []string          `json:"routes"`
	Patterns            []string          `json:"patterns"`
	ExcludePatterns     []string          `json:"excludePatterns"`
	Debounce            int               `json:"debounce"`
	DrushCmd            *string           `json:"drushCmd"`
	DrushCommand        string            `json:"drushCommand"`
	DrushArgs           []string          `json:"drushArgs"`
	PostClearCommands   []string          `json:"postClearCommands"`
	CommandsPerPattern  map[string]string `json:"commandsPerPattern"`
	DrupalRoot          *string           `json:"drupalRoot"`
}

func (c Config) GetDrushCmd() *string                 { return c.DrushCmd }
func (c Config) GetDrushCommand() string              { return c.DrushCommand }
func (c Config) GetDrushArgs() []string               { return c.DrushArgs }
func (c Config) GetDrupalRoot() *string               { return c.DrupalRoot }
func (c Config) GetRoutes() []string                  { return c.Routes }
func (c Config) GetPatterns() []string                { return c.Patterns }
func (c Config) GetExcludePatterns() []string         { return c.ExcludePatterns }
func (c Config) GetDebounce() int                     { return c.Debounce }
func (c Config) GetCommandsPerPattern() map[string]string { return c.CommandsPerPattern }
func (c Config) GetPostClearCommands() []string       { return c.PostClearCommands }

type Manager struct {
	cache           map[string]*cacheEntry
	customConfigPath string
}

type cacheEntry struct {
	Root   *string `json:"root,omitempty"`
	Config *Config `json:"config,omitempty"`
}

func NewManager() *Manager {
	return &Manager{cache: make(map[string]*cacheEntry)}
}

func getRoot(r string) string {
	if r != "" {
		return r
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

func (m *Manager) SetCustomConfigPath(p string) {
	m.customConfigPath = p
}

func (m *Manager) configPath(root string) string {
	if m.customConfigPath != "" {
		return m.customConfigPath
	}
	return filepath.Join(getRoot(root), "watcher.config.json")
}

func packageDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func pidPath(root string) string {
	if root != "" {
		return filepath.Join(root, ".drupal-watcher.pid")
	}
	return filepath.Join(packageDir(), ".drupal-watcher.pid")
}

func starttimePath(root string) string {
	if root != "" {
		return filepath.Join(root, ".drupal-watcher.starttime")
	}
	return filepath.Join(packageDir(), ".drupal-watcher.starttime")
}

func (m *Manager) DetectDrupalRoot(root string) *string {
	r := getRoot(root)
	if e, ok := m.cache[r]; ok && e.Root != nil {
		return e.Root
	}

	for _, dir := range utils.PossibleDocroots {
		fullPath := filepath.Join(r, dir)
		info, err := os.Stat(fullPath)
		if err != nil || !info.IsDir() {
			continue
		}
		if dirExists(filepath.Join(fullPath, "core")) ||
			dirExists(filepath.Join(fullPath, "modules")) ||
			dirExists(filepath.Join(fullPath, "themes")) ||
			fileExists(filepath.Join(fullPath, "index.php")) {
			m.cache[r] = &cacheEntry{Root: &dir}
			return &dir
		}
	}

	m.cache[r] = &cacheEntry{Root: nil}
	return nil
}

func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func (m *Manager) GetDefaultConfig(root string) Config {
	drupalRoot := "docroot"
	if d := m.DetectDrupalRoot(root); d != nil {
		drupalRoot = *d
	}
	return Config{
		Routes:    []string{drupalRoot + "/modules/custom", drupalRoot + "/themes/custom"},
		Patterns:  utils.DefaultPatterns,
		Debounce:  800,
		DrushCmd:  nil,
		DrushCommand: "cr",
		CommandsPerPattern: map[string]string{
			".html.twig":     "cc bin twig",
			".theme":         "cc theme-registry",
			".module":        "cc plugin",
			".inc":           "cc plugin",
			".yml":           "cc plugin",
			".info.yml":      "cr",
			".services.yml":  "cr",
			".php":           "cc plugin",
		},
		DrupalRoot: &drupalRoot,
	}
}

func (m *Manager) LoadConfig(root string) (Config, error) {
	r := getRoot(root)
	if e, ok := m.cache[r]; ok && e.Config != nil {
		return *e.Config, nil
	}

	cp := m.configPath(r)
	data, err := os.ReadFile(cp)
	if err != nil {
		def := m.GetDefaultConfig(r)
		b, _ := json.MarshalIndent(def, "", "  ")
		if werr := os.WriteFile(cp, b, 0644); werr != nil {
			return def, fmt.Errorf("failed to create config: %w", werr)
		}
		fmt.Printf("%s Created %s with defaults.\n", utils.P_INFO, utils.Cyan("watcher.config.json"))
		m.cache[r] = &cacheEntry{Config: &def}
		return def, nil
	}

	var parsed Config
	if err := json.Unmarshal(data, &parsed); err != nil {
		fmt.Printf("%s Failed to read %s. Using defaults.\n", utils.P_ERROR, utils.Cyan("watcher.config.json"))
		def := m.GetDefaultConfig(r)
		m.cache[r] = &cacheEntry{Config: &def}
		return def, nil
	}

	if parsed.DrupalRoot == nil {
		if d := m.DetectDrupalRoot(r); d != nil {
			parsed.DrupalRoot = d
			// Update routes with new docroot
			for i, route := range parsed.Routes {
				parts := strings.SplitN(route, "/", 2)
				if len(parts) > 0 {
					for _, dr := range utils.PossibleDocroots {
						if parts[0] == dr && parts[0] != *d {
							parsed.Routes[i] = strings.Replace(route, parts[0], *d, 1)
							break
						}
					}
				}
			}
			m.SaveConfig(parsed, r)
		}
	}

	cfg := m.ValidateConfig(parsed, r)
	m.cache[r] = &cacheEntry{Config: &cfg}
	return cfg, nil
}

func (m *Manager) ValidateConfig(cfg Config, root string) Config {
	def := m.GetDefaultConfig(root)
	if len(cfg.Routes) == 0 {
		cfg.Routes = def.Routes
	}
	if len(cfg.Patterns) == 0 {
		cfg.Patterns = def.Patterns
	}
	if cfg.ExcludePatterns == nil {
		cfg.ExcludePatterns = def.ExcludePatterns
	}
	if cfg.Debounce <= 0 {
		cfg.Debounce = def.Debounce
	}
	if cfg.DrushCommand == "" {
		cfg.DrushCommand = def.DrushCommand
	}
	if cfg.DrushArgs == nil {
		cfg.DrushArgs = def.DrushArgs
	}
	if cfg.PostClearCommands == nil {
		cfg.PostClearCommands = def.PostClearCommands
	}
	if cfg.CommandsPerPattern == nil {
		cfg.CommandsPerPattern = def.CommandsPerPattern
	} else {
		// Deep merge: defaults first, then user overrides
		merged := make(map[string]string)
		for k, v := range def.CommandsPerPattern {
			merged[k] = v
		}
		for k, v := range cfg.CommandsPerPattern {
			merged[k] = v
		}
		cfg.CommandsPerPattern = merged
	}
	if cfg.DrupalRoot == nil {
		cfg.DrupalRoot = def.DrupalRoot
	}
	// Normalize routes
	for i, r := range cfg.Routes {
		cfg.Routes[i] = strings.TrimRight(filepath.ToSlash(filepath.Clean(r)), "/")
	}
	return cfg
}

func (m *Manager) SaveConfig(cfg Config, root string) error {
	r := getRoot(root)
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(m.configPath(r), b, 0644); err != nil {
		return err
	}
	m.cache[r] = &cacheEntry{}
	return nil
}

func (m *Manager) InvalidateConfigCache(root string) {
	r := getRoot(root)
	delete(m.cache, r)
}

// --- PID file ---

func GetPidFile(root string) string {
	return pidPath(root)
}

func WritePid(root string) error {
	if err := os.WriteFile(pidPath(root), []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		return err
	}
	return WriteStarttime(root)
}

func RemovePid(root string) error {
	_ = os.WriteFile(pidPath(root), []byte(""), 0644)
	err := os.Remove(pidPath(root))
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "%s Failed to remove PID file: %v\n", utils.P_WARN, err)
	}
	_ = RemoveStarttime(root)
	return nil
}

func CheckPid(root string) (interface{}, error) {
	data, err := os.ReadFile(pidPath(root))
	if err != nil {
		return nil, nil // no PID file
	}
	pidStr := strings.TrimSpace(string(data))
	if pidStr == "" {
		return nil, nil
	}
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return nil, nil
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return "stale", nil
	}
	if err := proc.Signal(os.Signal(syscall.Signal(0))); err != nil {
		return "stale", nil
	}
	return pidStr, nil
}

// --- Start time ---

func WriteStarttime(root string) error {
	return os.WriteFile(starttimePath(root), []byte(strconv.FormatInt(nowMs(), 10)), 0644)
}

func GetStarttime(root string) (int64, error) {
	data, err := os.ReadFile(starttimePath(root))
	if err != nil {
		return 0, nil
	}
	t := strings.TrimSpace(string(data))
	if t == "" {
		return 0, nil
	}
	return strconv.ParseInt(t, 10, 64)
}

func RemoveStarttime(root string) error {
	err := os.Remove(starttimePath(root))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}


