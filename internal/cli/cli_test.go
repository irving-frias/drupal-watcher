package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/irving-frias/drupal-watcher/internal/cli"
	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/internal/utils"
	"github.com/pterm/pterm"
)

func TestPkgVersion(t *testing.T) {
	v := cli.PkgVersion()
	if v == "" {
		t.Error("expected non-empty version")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d        time.Duration
		expected string
	}{
		{5 * time.Second, "5s"},
		{65 * time.Second, "1m 5s"},
		{3665 * time.Second, "1h 1m 5s"},
		{90061 * time.Second, "1d 1h 1m 1s"},
	}
	for _, tt := range tests {
		result := utils.FormatDuration(tt.d)
		if result != tt.expected {
			t.Errorf("FormatDuration(%v) = %s, want %s", tt.d, result, tt.expected)
		}
	}
}

func TestIsPidRunning(t *testing.T) {
	running := cli.IsPidRunning(0)
	t.Logf("PID 0 running: %v", running)
}

func TestCmdHelp(t *testing.T) {
	cli.CmdHelp()
}

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	pterm.SetDefaultOutput(w)

	f()

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = old
	pterm.SetDefaultOutput(old)
	return buf.String()
}

func TestCmdListWithConfig(t *testing.T) {
	tmp := t.TempDir()
	mgr := config.NewManager()

	// Create a config file
	cfg := config.Config{
		Routes:   []string{tmp},
		Patterns: []string{".php", ".module"},
		Debounce: 500,
	}
	if err := mgr.SaveConfig(cfg, tmp); err != nil {
		t.Fatal(err)
	}

	// Reload from cache-invalidated state to pick up saved config
	mgr.InvalidateConfigCache(tmp)

	out := captureStdout(func() {
		if err := cli.CmdList(tmp, mgr); err != nil {
			t.Errorf("CmdList returned error: %v", err)
		}
	})

	if !strings.Contains(out, "Active Drupal Watcher") {
		t.Error("expected header in output")
	}
	if !strings.Contains(out, ".php") {
		t.Error("expected patterns in output")
	}
}

func TestCmdListError(t *testing.T) {
	mgr := config.NewManager()
	err := cli.CmdList("/nonexistent/path", mgr)
	if err == nil {
		t.Error("expected error for nonexistent root")
	}
}

func TestCmdStatusNotRunning(t *testing.T) {
	tmp := t.TempDir()
	mgr := config.NewManager()

	err := cli.CmdStatus(tmp, mgr)
	if err != nil {
		t.Errorf("expected no error for not-running, got: %v", err)
	}
}

func TestCmdStatusWithError(t *testing.T) {
	mgr := config.NewManager()
	err := cli.CmdStatus("", mgr)
	// Should not error — just reports "not running"
	if err != nil {
		t.Errorf("expected nil, got: %v", err)
	}
}

func TestCmdResetNoPid(t *testing.T) {
	tmp := t.TempDir()
	mgr := config.NewManager()

	out := captureStdout(func() {
		if err := cli.CmdReset(tmp, mgr); err != nil {
			t.Errorf("CmdReset returned error: %v", err)
		}
	})

	if !strings.Contains(out, "Reset complete") {
		t.Errorf("expected reset message, got: %s", out)
	}
}

func TestCmdResetStalePid(t *testing.T) {
	tmp := t.TempDir()
	mgr := config.NewManager()

	// Write a PID file with a nonexistent PID
	pidPath := filepath.Join(tmp, ".drupal-watcher.pid")
	if err := os.WriteFile(pidPath, []byte("999999999"), 0644); err != nil {
		t.Fatal(err)
	}

	out := captureStdout(func() {
		if err := cli.CmdReset(tmp, mgr); err != nil {
			t.Errorf("CmdReset returned error: %v", err)
		}
	})

	if !strings.Contains(out, "Reset complete") {
		t.Errorf("expected reset message, got: %s", out)
	}
}

func TestCmdAddRoute(t *testing.T) {
	tmp := t.TempDir()
	mgr := config.NewManager()

	// Create a subdirectory to add
	routeDir := filepath.Join(tmp, "modules", "custom")
	if err := os.MkdirAll(routeDir, 0755); err != nil {
		t.Fatal(err)
	}

	err := cli.CmdAdd(tmp, []string{routeDir}, mgr)
	if err != nil {
		t.Errorf("CmdAdd returned error: %v", err)
	}

	// Verify route was saved
	cfg, err := mgr.LoadConfig(tmp)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range cfg.Routes {
		if r == routeDir {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected route %s in config, got %v", routeDir, cfg.Routes)
	}
}

func TestCmdAddInvalidRoute(t *testing.T) {
	tmp := t.TempDir()
	mgr := config.NewManager()

	// Create initial config to avoid "created with defaults" noise
	cfg := config.Config{Routes: []string{tmp}}
	mgr.SaveConfig(cfg, tmp)
	mgr.InvalidateConfigCache(tmp)

	out := captureStdout(func() {
		err := cli.CmdAdd(tmp, []string{"/nonexistent/path"}, mgr)
		if err != nil {
			t.Errorf("CmdAdd should not error on invalid route, got: %v", err)
		}
	})

	if !strings.Contains(out, "Invalid route") {
		t.Errorf("expected invalid route message, got: %s", out)
	}
}

func TestCmdRemoveRoute(t *testing.T) {
	tmp := t.TempDir()
	mgr := config.NewManager()

	routeDir := filepath.Join(tmp, "modules", "custom")
	if err := os.MkdirAll(routeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Add first
	if err := cli.CmdAdd(tmp, []string{routeDir}, mgr); err != nil {
		t.Fatal(err)
	}

	// Then remove
	mgr.InvalidateConfigCache(tmp)
	out := captureStdout(func() {
		if err := cli.CmdRemove(tmp, []string{routeDir}, mgr); err != nil {
			t.Errorf("CmdRemove returned error: %v", err)
		}
	})

	if !strings.Contains(out, "Removed") {
		t.Errorf("expected removal message, got: %s", out)
	}

	// Verify route is gone
	cfg, err := mgr.LoadConfig(tmp)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range cfg.Routes {
		if r == routeDir {
			t.Error("route should have been removed")
		}
	}
}

func TestCmdRemoveNonexistent(t *testing.T) {
	tmp := t.TempDir()
	mgr := config.NewManager()

	cfg := config.Config{Routes: []string{tmp}}
	mgr.SaveConfig(cfg, tmp)
	mgr.InvalidateConfigCache(tmp)

	out := captureStdout(func() {
		err := cli.CmdRemove(tmp, []string{"/nonexistent/path"}, mgr)
		if err != nil {
			t.Errorf("CmdRemove should not error on invalid route, got: %v", err)
		}
	})

	if !strings.Contains(out, "Invalid route") {
		t.Errorf("expected invalid route message, got: %s", out)
	}
}
