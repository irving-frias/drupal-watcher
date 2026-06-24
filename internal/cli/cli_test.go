package cli_test

import (
	"testing"
	"time"

	"github.com/irving-frias/drupal-watcher/internal/cli"
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
		result := cli.FormatDuration(tt.d)
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
