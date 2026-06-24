package main

import (
	"testing"
)

func TestParseFlagsNoArgs(t *testing.T) {
	cmd, flags, extra := parseFlags([]string{})
	if cmd != "" {
		t.Errorf("expected empty command, got %s", cmd)
	}
	if len(flags) != 0 {
		t.Errorf("expected no flags, got %v", flags)
	}
	if len(extra) != 0 {
		t.Errorf("expected no extra args, got %v", extra)
	}
}

func TestParseFlagsCommand(t *testing.T) {
	cmd, flags, extra := parseFlags([]string{"start"})
	if cmd != "start" {
		t.Errorf("expected start, got %s", cmd)
	}
	if len(extra) != 0 {
		t.Errorf("expected no extra args, got %v", extra)
	}
	_ = flags
}

func TestParseFlagsWithFlags(t *testing.T) {
	cmd, flags, extra := parseFlags([]string{"start", "--debounce", "500", "--no-dotfiles", "myroot"})
	if cmd != "start" {
		t.Errorf("expected start, got %s", cmd)
	}
	if d, ok := flags["debounce"].(int); !ok || d != 500 {
		t.Errorf("expected debounce 500, got %v", flags["debounce"])
	}
	if nd, ok := flags["no-dotfiles"].(bool); !ok || nd != true {
		t.Errorf("expected no-dotfiles true, got %v", flags["no-dotfiles"])
	}
	if len(extra) != 1 || extra[0] != "myroot" {
		t.Errorf("expected [myroot], got %v", extra)
	}
}

func TestParseFlagsDebounceEquals(t *testing.T) {
	cmd, flags, _ := parseFlags([]string{"start", "--debounce=300"})
	if cmd != "start" {
		t.Errorf("expected start, got %s", cmd)
	}
	if d, ok := flags["debounce"].(int); !ok || d != 300 {
		t.Errorf("expected debounce 300, got %v", flags["debounce"])
	}
}

func TestParseFlagsLogFile(t *testing.T) {
	_, flags, _ := parseFlags([]string{"start", "--log-file=/tmp/test.log"})
	if lf, ok := flags["log-file"].(string); !ok || lf != "/tmp/test.log" {
		t.Errorf("expected /tmp/test.log, got %v", flags["log-file"])
	}
}

func TestParseFlagsLogFileSeparate(t *testing.T) {
	_, flags, _ := parseFlags([]string{"start", "--log-file", "/tmp/test.log"})
	if lf, ok := flags["log-file"].(string); !ok || lf != "/tmp/test.log" {
		t.Errorf("expected /tmp/test.log, got %v", flags["log-file"])
	}
}

func TestParseFlagsConfigPath(t *testing.T) {
	_, flags, _ := parseFlags([]string{"start", "--config=/custom/path.json"})
	if cfg, ok := flags["config"].(string); !ok || cfg != "/custom/path.json" {
		t.Errorf("expected /custom/path.json, got %v", flags["config"])
	}
}

func TestParseFlagsCommandsPerPattern(t *testing.T) {
	_, flags, _ := parseFlags([]string{"start", `--commands-per-pattern={"test":"val"}`})
	if cpp, ok := flags["commands-per-pattern"].(map[string]string); !ok || cpp["test"] != "val" {
		t.Errorf("expected map[test:val], got %v", flags["commands-per-pattern"])
	}
}

func TestParseFlagsHelp(t *testing.T) {
	cmd, flags, _ := parseFlags([]string{"--help"})
	if cmd != "help" {
		t.Errorf("expected help, got %s", cmd)
	}
	if _, ok := flags["help"]; !ok {
		t.Error("expected help flag")
	}
}

func TestParseFlagsVersion(t *testing.T) {
	cmd, flags, _ := parseFlags([]string{"status", "-V"})
	if cmd != "status" {
		t.Errorf("expected status, got %s", cmd)
	}
	if _, ok := flags["version"]; !ok {
		t.Error("expected version flag")
	}
}
