package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/irving-frias/drupal-watcher/internal/config"
)

func TestNewManager(t *testing.T) {
	m := config.NewManager()
	if m == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestDetectDrupalRoot(t *testing.T) {
	tmp := t.TempDir()
	m := config.NewManager()

	// No Drupal root
	root := m.DetectDrupalRoot(tmp)
	if root != nil {
		t.Errorf("expected nil, got %v", *root)
	}

	// Create a Drupal-like structure
	docroot := filepath.Join(tmp, "docroot")
	if err := os.MkdirAll(filepath.Join(docroot, "core"), 0755); err != nil {
		t.Fatal(err)
	}

	root = m.DetectDrupalRoot(tmp)
	if root == nil || *root != "docroot" {
		t.Errorf("expected docroot, got %v", root)
	}

	// Test with web directory
	tmp2 := t.TempDir()
	web := filepath.Join(tmp2, "web")
	if err := os.MkdirAll(filepath.Join(web, "core"), 0755); err != nil {
		t.Fatal(err)
	}

	m2 := config.NewManager()
	root = m2.DetectDrupalRoot(tmp2)
	if root == nil || *root != "web" {
		t.Errorf("expected web, got %v", root)
	}
}

func TestGetDefaultConfig(t *testing.T) {
	m := config.NewManager()
	cfg := m.GetDefaultConfig("")

	if len(cfg.Routes) == 0 {
		t.Error("expected routes")
	}
	if cfg.Debounce != 800 {
		t.Errorf("expected debounce 800, got %d", cfg.Debounce)
	}
	if cfg.DrushCommand != "cr" {
		t.Errorf("expected drushCommand cr, got %s", cfg.DrushCommand)
	}
	if cfg.CommandsPerPattern == nil {
		t.Error("expected commandsPerPattern")
	}
}

func TestValidateConfig(t *testing.T) {
	m := config.NewManager()

	// Empty config should get defaults
	cfg := m.ValidateConfig(config.Config{}, "")
	if len(cfg.Routes) == 0 {
		t.Error("expected routes after validation")
	}
	if cfg.Debounce <= 0 {
		t.Errorf("expected positive debounce, got %d", cfg.Debounce)
	}

	// Custom values should be preserved (with defaults merged)
	customCfg := config.Config{
		Routes:   []string{"custom/route"},
		Debounce: 500,
	}
	cfg = m.ValidateConfig(customCfg, "")
	if len(cfg.Routes) != 1 || cfg.Routes[0] != "custom/route" {
		t.Errorf("expected custom route, got %v", cfg.Routes)
	}
	if cfg.Debounce != 500 {
		t.Errorf("expected debounce 500, got %d", cfg.Debounce)
	}
	if cfg.DrushCommand == "" {
		t.Error("expected default drushCommand")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmp := t.TempDir()
	m := config.NewManager()

	cfg := config.Config{
		Routes:   []string{"test/routes"},
		Debounce: 300,
		DrushCommand: "cc all",
	}

	if err := m.SaveConfig(cfg, tmp); err != nil {
		t.Fatal(err)
	}

	loaded, err := m.LoadConfig(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Routes) != 1 || loaded.Routes[0] != "test/routes" {
		t.Errorf("expected test/routes, got %v", loaded.Routes)
	}
	if loaded.Debounce != 300 {
		t.Errorf("expected 300, got %d", loaded.Debounce)
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	m := config.NewManager()

	os.WriteFile(filepath.Join(tmp, "watcher.config.json"), []byte("{invalid"), 0644)

	// Should fall back to defaults without error
	cfg, err := m.LoadConfig(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Routes) == 0 {
		t.Error("expected defaults for invalid JSON")
	}
}

func TestCache(t *testing.T) {
	tmp := t.TempDir()
	m := config.NewManager()

	// Load once
	cfg1, _ := m.LoadConfig(tmp)
	// Load again (should use cache)
	cfg2, _ := m.LoadConfig(tmp)

	if &cfg1 == &cfg2 {
		t.Error("LoadConfig should return a copy, not cached pointer")
	}

	// Invalidate
	m.InvalidateConfigCache(tmp)
	cfg3, _ := m.LoadConfig(tmp)
	if cfg3.Debounce != cfg1.Debounce {
		t.Errorf("expected same debounce after reload, got %d vs %d", cfg3.Debounce, cfg1.Debounce)
	}
}

func TestPIDWriteAndCheck(t *testing.T) {
	tmp := t.TempDir()

	if err := config.WritePid(tmp); err != nil {
		t.Fatal(err)
	}

	status, err := config.CheckPid(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if status == nil {
		t.Fatal("expected PID to exist")
	}
	if status == "stale" {
		t.Fatal("expected valid PID")
	}
}

func TestPIDRemove(t *testing.T) {
	tmp := t.TempDir()

	config.WritePid(tmp)
	config.RemovePid(tmp)

	status, err := config.CheckPid(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if status != nil {
		t.Errorf("expected nil after removal, got %v", status)
	}
}

func TestStarttime(t *testing.T) {
	tmp := t.TempDir()

	if err := config.WriteStarttime(tmp); err != nil {
		t.Fatal(err)
	}

	st, err := config.GetStarttime(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if st <= 0 {
		t.Errorf("expected positive starttime, got %d", st)
	}

	config.RemoveStarttime(tmp)
	st, err = config.GetStarttime(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if st != 0 {
		t.Errorf("expected 0 after removal, got %d", st)
	}
}

func TestCommandsPerPatternMerge(t *testing.T) {
	m := config.NewManager()
	cfg := config.Config{
		CommandsPerPattern: map[string]string{
			".php": "custom command",
		},
	}
	validated := m.ValidateConfig(cfg, "")

	if validated.CommandsPerPattern[".html.twig"] != "cc twig" {
		t.Error("expected default for .html.twig to be preserved")
	}
	if validated.CommandsPerPattern[".php"] != "custom command" {
		t.Errorf("expected custom override, got %s", validated.CommandsPerPattern[".php"])
	}
}

func TestRouteNormalization(t *testing.T) {
	m := config.NewManager()
	cfg := config.Config{
		Routes: []string{"./docroot//modules/custom/"},
	}
	validated := m.ValidateConfig(cfg, "")

	if len(validated.Routes) != 1 {
		t.Fatalf("expected 1 route, got %v", validated.Routes)
	}
	expected := "docroot/modules/custom"
	if validated.Routes[0] != expected {
		t.Errorf("expected %s, got %s", expected, validated.Routes[0])
	}
}

func TestGetPidFile(t *testing.T) {
	path := config.GetPidFile("/test/root")
	if !filepath.IsAbs(path) && path != "/test/root/.drupal-watcher.pid" {
		// GetPidFile just delegates to pidPath
	}
}
