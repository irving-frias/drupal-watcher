package adapters

import (
	"os/exec"
	"strings"

	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type PhpLintChecker struct{}

func NewPhpLintChecker() *PhpLintChecker {
	return &PhpLintChecker{}
}

func (c *PhpLintChecker) Lint(filePath string) *core.LintResult {
	cmd := exec.Command("php", "-l", filePath)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	msg := strings.TrimSpace(string(out))
	if msg == "" {
		msg = err.Error()
	}
	return &core.LintResult{File: filePath, Error: msg}
}
