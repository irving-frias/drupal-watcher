package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type mockLinter struct {
	called int
}

func (m *mockLinter) Lint(path string) *core.LintResult {
	m.called++
	return nil
}

func TestCachingLintChecker_CachesResult(t *testing.T) {
	inner := &mockLinter{}
	c := NewCachingLintChecker(inner)

	dir := t.TempDir()
	fp := filepath.Join(dir, "test.php")
	if err := os.WriteFile(fp, []byte("<?php\n"), 0644); err != nil {
		t.Fatal(err)
	}

	c.Lint(fp)
	c.Lint(fp)
	c.Lint(fp)

	if inner.called != 1 {
		t.Errorf("expected 1 call to inner linter, got %d", inner.called)
	}
}

func TestCachingLintChecker_Invalidate(t *testing.T) {
	inner := &mockLinter{}
	c := NewCachingLintChecker(inner)

	dir := t.TempDir()
	fp := filepath.Join(dir, "test.php")
	if err := os.WriteFile(fp, []byte("<?php\n"), 0644); err != nil {
		t.Fatal(err)
	}

	c.Lint(fp)
	c.Invalidate(fp)
	c.Lint(fp)

	if inner.called != 2 {
		t.Errorf("expected 2 calls after invalidation, got %d", inner.called)
	}
}

func TestCachingLintChecker_ChangedFile(t *testing.T) {
	inner := &mockLinter{}
	c := NewCachingLintChecker(inner)

	dir := t.TempDir()
	fp := filepath.Join(dir, "test.php")
	if err := os.WriteFile(fp, []byte("<?php\n"), 0644); err != nil {
		t.Fatal(err)
	}

	c.Lint(fp)

	if err := os.WriteFile(fp, []byte("<?php\n// changed\n"), 0644); err != nil {
		t.Fatal(err)
	}

	c.Lint(fp)

	if inner.called != 2 {
		t.Errorf("expected 2 calls after file change, got %d", inner.called)
	}
}

func TestFileChecksum_Error(t *testing.T) {
	_, err := fileChecksum("/nonexistent/file.php")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
