package adapters_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/irving-frias/drupal-watcher/pkg/adapters"
)

func TestYamlLintChecker_ValidYAML(t *testing.T) {
	checker := adapters.NewYamlLintChecker()

	dir := t.TempDir()
	file := filepath.Join(dir, "test.yml")
	os.WriteFile(file, []byte("key: value\nlist:\n  - one\n  - two\n"), 0644)

	result := checker.Lint(file)
	if result != nil {
		t.Errorf("expected nil for valid YAML, got %v", result)
	}
}

func TestYamlLintChecker_ValidYAML_YamlExtension(t *testing.T) {
	checker := adapters.NewYamlLintChecker()

	dir := t.TempDir()
	file := filepath.Join(dir, "test.yaml")
	os.WriteFile(file, []byte("foo: bar\n"), 0644)

	result := checker.Lint(file)
	if result != nil {
		t.Errorf("expected nil for valid YAML, got %v", result)
	}
}

func TestYamlLintChecker_InvalidYAML(t *testing.T) {
	checker := adapters.NewYamlLintChecker()

	dir := t.TempDir()
	file := filepath.Join(dir, "invalid.yml")
	os.WriteFile(file, []byte("key: [value\n"), 0644)

	result := checker.Lint(file)
	if result == nil {
		t.Fatal("expected lint error for invalid YAML (unmatched bracket)")
	}
	if result.File != file {
		t.Errorf("expected file %s, got %s", file, result.File)
	}
	if result.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestYamlLintChecker_NonExistentFile(t *testing.T) {
	checker := adapters.NewYamlLintChecker()
	result := checker.Lint("/nonexistent/config.yml")
	if result == nil {
		t.Fatal("expected lint error for nonexistent file")
	}
}
