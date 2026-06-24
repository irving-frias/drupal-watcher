package drush_test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/irving-frias/drupal-watcher/internal/drush"
)

type mockDrushConfig struct {
	cmd     *string
	command string
	args    []string
	root    *string
}

func (m mockDrushConfig) GetDrushCmd() *string    { return m.cmd }
func (m mockDrushConfig) GetDrushCommand() string  { return m.command }
func (m mockDrushConfig) GetDrushArgs() []string   { return m.args }
func (m mockDrushConfig) GetDrupalRoot() *string   { return m.root }

func TestGetCmd(t *testing.T) {
	cmd := "my-drush"
	cfg := mockDrushConfig{cmd: &cmd}
	result := drush.GetCmd(cfg)
	if result != "my-drush" {
		t.Errorf("expected my-drush, got %s", result)
	}

	cfg2 := mockDrushConfig{}
	result = drush.GetCmd(cfg2)
	if result == "" {
		t.Error("expected resolved drush command")
	}
}

func TestRunWithBasicCommand(t *testing.T) {
	result := drush.Run("echo", []string{"hello"})
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

	result := drush.Run(drushPath, []string{"--version"})
	if result.ExitCode != 0 {
		t.Fatalf("drush --version failed: exit %d, stderr=%s", result.ExitCode, result.Stderr)
	}
	if !strings.Contains(result.Stdout, "Drush") {
		t.Logf("drush --version output: %s", result.Stdout)
	}
}

func TestHealthCheck(t *testing.T) {
	cfg := mockDrushConfig{command: "cr"}
	ok := drush.HealthCheck(cfg)
	t.Logf("HealthCheck returned %v", ok)
}

func TestGetSpawnArgs(t *testing.T) {
	cmd := "/usr/local/bin/drush"
	cfg := mockDrushConfig{
		cmd:     &cmd,
		command: "cc all",
		args:    []string{"--yes"},
	}

	base, args := drush.GetSpawnArgs(cfg)
	if base != "/usr/local/bin/drush" {
		t.Errorf("expected /usr/local/bin/drush, got %s", base)
	}
	if len(args) < 4 {
		t.Errorf("expected at least 4 args, got %d: %v", len(args), args)
	}
}

func TestRunFailsGracefully(t *testing.T) {
	result := drush.Run("nonexistent-command-12345", []string{})
	if result.ExitCode == 0 {
		t.Error("expected non-zero exit for nonexistent command")
	}
}
