package adapters_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/irving-frias/drupal-watcher/pkg/adapters"
)

func TestPhpLintChecker_ValidPHP(t *testing.T) {
	if !phpAvailable() {
		t.Skip("php not available")
	}

	checker := adapters.NewPhpLintChecker()

	dir := t.TempDir()
	file := filepath.Join(dir, "test.php")
	os.WriteFile(file, []byte("<?php\n$foo = 'bar';\n"), 0644)

	result := checker.Lint(file)
	if result != nil {
		t.Errorf("expected nil for valid PHP, got %v", result)
	}
}

func phpAvailable() bool {
	_, err := os.Stat("/usr/bin/php")
	if err != nil {
		return os.Getenv("PHP_PATH") != ""
	}
	return true
}

func TestPhpLintChecker_InvalidPHP(t *testing.T) {
	checker := adapters.NewPhpLintChecker()

	dir := t.TempDir()
	file := filepath.Join(dir, "invalid.php")
	os.WriteFile(file, []byte("<?php\n$foo = ;\n"), 0644)

	result := checker.Lint(file)
	if result == nil {
		t.Fatal("expected lint error for invalid PHP")
	}
	if result.File != file {
		t.Errorf("expected file %s, got %s", file, result.File)
	}
	if result.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestPhpLintChecker_NonExistentFile(t *testing.T) {
	checker := adapters.NewPhpLintChecker()
	result := checker.Lint("/nonexistent/path.php")
	if result == nil {
		t.Fatal("expected lint error for nonexistent file")
	}
}

func TestPhpLintChecker_EmptyFile(t *testing.T) {
	if !phpAvailable() {
		t.Skip("php not available")
	}

	checker := adapters.NewPhpLintChecker()

	dir := t.TempDir()
	file := filepath.Join(dir, "empty.php")
	os.WriteFile(file, nil, 0644)

	result := checker.Lint(file)
	// php -l may accept empty files as valid; we just verify no crash
	if result != nil && result.File != file {
		t.Errorf("expected file %s, got %s", file, result.File)
	}
}
