package adapters

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type PhpCsLintChecker struct {
	standard string
	phpcsBin string
	once     sync.Once
}

func NewPhpCsLintChecker(standard string) *PhpCsLintChecker {
	return &PhpCsLintChecker{standard: standard}
}

func (c *PhpCsLintChecker) resolve() {
	c.once.Do(func() {
		std := c.standard
		if std == "" || std == "auto" {
			std = detectDrupalStandard()
		}
		c.standard = std

		for _, p := range []string{"vendor/bin/phpcs", "../vendor/bin/phpcs"} {
			if _, err := os.Stat(p); err == nil {
				abs, _ := filepath.Abs(p)
				c.phpcsBin = abs
				return
			}
		}
		if p, err := exec.LookPath("phpcs"); err == nil {
			c.phpcsBin = p
		}
	})
}

func (c *PhpCsLintChecker) Lint(filePath string) *core.LintResult {
	c.resolve()
	if c.phpcsBin == "" {
		return nil
	}

	cmd := exec.Command(c.phpcsBin, "--standard="+c.standard, "--report=full", filePath)
	out, err := cmd.CombinedOutput()
	msg := strings.TrimSpace(string(out))
	if err == nil {
		return nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 2 {
		return nil
	}

	if msg == "" {
		msg = err.Error()
	}

	lines := strings.SplitN(msg, "\n", 2)
	if len(lines) > 0 {
		msg = strings.TrimSpace(lines[0])
	}

	if len(msg) > 250 {
		msg = msg[:250] + "..."
	}

	return &core.LintResult{File: filePath, Error: msg}
}

func detectDrupalStandard() string {
	data, err := os.ReadFile("composer.json")
	if err != nil {
		return "Drupal"
	}
	var meta struct {
		Require    map[string]string `json:"require"`
		RequireDev map[string]string `json:"require-dev"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return "Drupal"
	}
	v := meta.Require["drupal/core"]
	if v == "" {
		v = meta.RequireDev["drupal/core"]
	}
	if v == "" {
		return "Drupal"
	}
	if strings.HasPrefix(v, "^11") || strings.HasPrefix(v, "~11") || strings.HasPrefix(v, ">=11") || strings.HasPrefix(v, "11.") {
		return "DrupalStrict"
	}
	return "Drupal"
}
