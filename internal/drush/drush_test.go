package drush

import (
	"os/exec"
	"strings"
	"testing"
)

type mockDrushConfig struct {
	cmd     *string
	command string
	args    []string
	root    *string
	notify  bool
}

func (m mockDrushConfig) GetDrushCmd() *string    { return m.cmd }
func (m mockDrushConfig) GetDrushCommand() string  { return m.command }
func (m mockDrushConfig) GetDrushArgs() []string   { return m.args }
func (m mockDrushConfig) GetDrupalRoot() *string   { return m.root }
func (m mockDrushConfig) GetNotify() bool           { return m.notify }

func TestGetCmd(t *testing.T) {
	ResetCmdCache()

	cmd := "my-drush"
	cfg := mockDrushConfig{cmd: &cmd}
	result := GetCmd(cfg)
	if result != "my-drush" {
		t.Errorf("expected my-drush, got %s", result)
	}
}

func TestGetCmdCache(t *testing.T) {
	ResetCmdCache()

	// First call without explicit cmd — resolves drush and caches it
	r1 := GetCmd(mockDrushConfig{})
	if r1 == "" {
		t.Fatal("expected a resolved path")
	}

	// Second call without cmd — should return cached value
	r2 := GetCmd(mockDrushConfig{})
	if r2 != r1 {
		t.Errorf("expected cached (%s), got %s", r1, r2)
	}
}

func TestGetCmdFallback(t *testing.T) {
	ResetCmdCache()

	cfg := mockDrushConfig{}
	result := GetCmd(cfg)
	if result == "" {
		t.Error("expected resolved drush command")
	}
}

func TestResetCmdCache(t *testing.T) {
	ResetCmdCache()

	// Prime the cache
	cmd := "test-cmd"
	cfg := mockDrushConfig{cmd: &cmd}
	GetCmd(cfg)

	ResetCmdCache()

	// After reset, should resolve again
	cfg2 := mockDrushConfig{}
	result := GetCmd(cfg2)
	if result == "" {
		t.Error("expected resolved drush command after reset")
	}
}

func TestRunWithBasicCommand(t *testing.T) {
	result := Run("echo", []string{"hello"})
	if result.ExitCode != 0 {
		t.Errorf("expected exit 0, got %d: %s", result.ExitCode, result.Stderr)
	}
	if result.Duration <= 0 {
		t.Error("expected positive duration")
	}
}

func TestRunWithDrush(t *testing.T) {
	drushPath, err := exec.LookPath("drush")
	if err != nil {
		t.Skip("drush not found in PATH")
	}

	result := Run(drushPath, []string{"--version"})
	if result.ExitCode != 0 {
		t.Fatalf("drush --version failed: exit %d, stderr=%s", result.ExitCode, result.Stderr)
	}
	if !strings.Contains(result.Stdout, "Drush") {
		t.Logf("drush --version output: %s", result.Stdout)
	}
}

func TestHealthCheck(t *testing.T) {
	cfg := mockDrushConfig{command: "cr"}
	ok := HealthCheck(cfg)
	t.Logf("HealthCheck returned %v", ok)
}

func TestRunFailsGracefully(t *testing.T) {
	result := Run("nonexistent-command-12345", []string{})
	if result.ExitCode == 0 {
		t.Error("expected non-zero exit for nonexistent command")
	}
}

func TestRunCacheClearsEmpty(t *testing.T) {
	result := RunCacheClears(nil, nil)
	if result.ExitCode != 0 {
		t.Errorf("expected exit 0, got %d", result.ExitCode)
	}
}

func TestRunCacheClearsCR(t *testing.T) {
	ResetCmdCache()

	cmd := "echo"
	cfg := mockDrushConfig{cmd: &cmd}
	result := RunCacheClears(cfg, []string{"cr"})
	if result.ExitCode != 0 {
		t.Errorf("expected exit 0, got %d: %s", result.ExitCode, result.Stderr)
	}
}

func TestRunCacheClearsBatchesCC(t *testing.T) {
	ResetCmdCache()

	cmd := "echo"
	cfg := mockDrushConfig{cmd: &cmd}
	result := RunCacheClears(cfg, []string{"cc render", "cc plugin", "cc css-js"})
	if result.ExitCode != 0 {
		t.Errorf("expected exit 0, got %d: %s", result.ExitCode, result.Stderr)
	}
	// echo outputs the args; check it contains comma-separated types
	if !strings.Contains(result.Stdout, "render,plugin,css-js") {
		t.Errorf("expected batched cc args, got: %s", result.Stdout)
	}
}

func TestNotifyCalledOnSuccess(t *testing.T) {
	ResetCmdCache()

	var calledTitle, calledMsg string
	NotifyFunc = func(title, message string) {
		calledTitle = title
		calledMsg = message
	}
	defer func() { NotifyFunc = NotifyOS }()

	cmd := "echo"
	cfg := mockDrushConfig{cmd: &cmd, notify: true}
	result := RunCacheClears(cfg, []string{"cr"})
	if result.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d", result.ExitCode)
	}
	if calledTitle != "Drupal Watcher" {
		t.Errorf("expected 'Drupal Watcher', got %q", calledTitle)
	}
	if calledMsg != "drush cr" {
		t.Errorf("expected 'drush cr', got %q", calledMsg)
	}
}

func TestNotifyNotCalledWhenDisabled(t *testing.T) {
	ResetCmdCache()

	var called bool
	NotifyFunc = func(title, message string) {
		called = true
	}
	defer func() { NotifyFunc = NotifyOS }()

	cmd := "echo"
	cfg := mockDrushConfig{cmd: &cmd, notify: false}
	result := RunCacheClears(cfg, []string{"cr"})
	if result.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d", result.ExitCode)
	}
	if called {
		t.Error("expected no notification when notify is disabled")
	}
}

func TestNotifyNotCalledOnError(t *testing.T) {
	ResetCmdCache()

	var called bool
	NotifyFunc = func(title, message string) {
		called = true
	}
	defer func() { NotifyFunc = NotifyOS }()

	cfg := mockDrushConfig{cmd: nil, notify: true}
	result := RunCacheClears(cfg, []string{"cr"})
	if result.ExitCode == 0 {
		t.Skip("unexpected success, skipping error test")
	}
	if called {
		t.Error("expected no notification on failure")
	}
}

func TestRunCacheClearsCRoverridesCC(t *testing.T) {
	ResetCmdCache()

	cmd := "echo"
	cfg := mockDrushConfig{cmd: &cmd}
	result := RunCacheClears(cfg, []string{"cc render", "cr", "cc plugin"})
	if result.ExitCode != 0 {
		t.Errorf("expected exit 0, got %d: %s", result.ExitCode, result.Stderr)
	}
	// Should run "cr" only, not batched cc
	if strings.Contains(result.Stdout, "render") {
		t.Errorf("expected cr to override cc, got: %s", result.Stdout)
	}
}
