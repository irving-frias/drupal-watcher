package adapters

import (
	"fmt"
	"os"

	"github.com/irving-frias/drupal-watcher/pkg/core"
	"gopkg.in/yaml.v3"
)

type YamlLintChecker struct{}

func NewYamlLintChecker() *YamlLintChecker {
	return &YamlLintChecker{}
}

func (c *YamlLintChecker) Lint(filePath string) *core.LintResult {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return &core.LintResult{File: filePath, Error: fmt.Sprintf("read: %v", err)}
	}
	var out interface{}
	if err := yaml.Unmarshal(data, &out); err != nil {
		return &core.LintResult{File: filePath, Error: fmt.Sprintf("yaml: %v", err)}
	}
	return nil
}
