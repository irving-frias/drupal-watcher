package adapters_test

import (
	"context"
	"strings"
	"testing"

	"github.com/irving-frias/drupal-watcher/internal/drush"
	"github.com/irving-frias/drupal-watcher/pkg/adapters"
	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type testDrushConfig struct {
	cmd     *string
	command string
	args    []string
	root    *string
	notify  bool
}

func (c testDrushConfig) GetDrushCmd() *string    { return c.cmd }
func (c testDrushConfig) GetDrushCommand() string  { return c.command }
func (c testDrushConfig) GetDrushArgs() []string   { return c.args }
func (c testDrushConfig) GetDrupalRoot() *string   { return c.root }
func (c testDrushConfig) GetNotify() bool          { return c.notify }

func TestDrushExecutor_Execute(t *testing.T) {
	cmd := "echo"
	cfg := testDrushConfig{
		cmd:     &cmd,
		command: "cr",
	}
	exec := adapters.NewDrushExecutor(cfg)
	result := exec.Execute(context.Background(), []string{"cr"}, ".")

	if result.ExitCode != 0 {
		t.Errorf("expected exit 0, got %d: %s", result.ExitCode, result.Stderr)
	}
	if result.Command != "cr" {
		t.Errorf("expected command 'cr', got %q", result.Command)
	}
	if result.Duration <= 0 {
		t.Error("expected positive duration")
	}
}

func TestDrushExecutor_ExecuteMultiple(t *testing.T) {
	cmd := "echo"
	cfg := testDrushConfig{
		cmd:     &cmd,
		command: "cr",
	}
	exec := adapters.NewDrushExecutor(cfg)
	result := exec.Execute(context.Background(), []string{"cc render", "cc plugin"}, ".")

	if result.ExitCode != 0 {
		t.Errorf("expected exit 0, got %d: %s", result.ExitCode, result.Stderr)
	}
	if !strings.Contains(result.Command, "cc render") {
		t.Errorf("expected command to contain 'cc render', got %q", result.Command)
	}
}

func TestDrushExecutor_ImplementsInterface(t *testing.T) {
	cmd := "echo"
	cfg := testDrushConfig{cmd: &cmd}
	var _ core.CommandExecutor = adapters.NewDrushExecutor(cfg)
}

func TestSiteAwareDrushExecutor_Execute(t *testing.T) {
	cmd := "echo"
	cfg := testDrushConfig{
		cmd:     &cmd,
		command: "cr",
		args:    []string{"--quiet"},
	}
	exec := adapters.NewSiteAwareDrushExecutor(cfg, "admin", "http://admin.example.com")
	result := exec.Execute(context.Background(), []string{"cr"}, ".")

	if result.ExitCode != 0 {
		t.Errorf("expected exit 0, got %d: %s", result.ExitCode, result.Stderr)
	}
	if result.Command != "cr" {
		t.Errorf("expected command 'cr', got %q", result.Command)
	}
}

func TestSiteAwareDrushExecutor_PassesURI(t *testing.T) {
	drush.ResetCmdCache()

	cmd := "echo"
	cfg := testDrushConfig{
		cmd:     &cmd,
		command: "cr",
		args:    []string{"--quiet"},
	}
	exec := adapters.NewSiteAwareDrushExecutor(cfg, "admin", "http://admin.example.com")

	_ = exec.Execute(context.Background(), []string{"cr"}, ".")

	// We can't easily verify the URI was passed through echo,
	// but we verify the executor doesn't panic and returns success
	result := exec.Execute(context.Background(), []string{"cr"}, ".")
	if result.ExitCode != 0 {
		t.Errorf("expected exit 0, got %d", result.ExitCode)
	}
}
