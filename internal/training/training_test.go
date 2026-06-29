package training

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NonExistent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	if err := Load(path); err != nil {
		t.Fatal(err)
	}
	if len(Get()) == 0 {
		t.Error("expected default suggestions after loading nonexistent file")
	}
}

func TestLoad_ValidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "training.json")
	data := `[{"title":"Test","description":"A test suggestion"}]`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Load(path); err != nil {
		t.Fatal(err)
	}
	suggestions := Get()
	if len(suggestions) != 1 || suggestions[0].Title != "Test" {
		t.Errorf("expected 1 suggestion with title Test, got %d %+v", len(suggestions), suggestions)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{bad json}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Load(path); err != nil {
		t.Fatal(err)
	}
	if len(Get()) == 0 {
		t.Error("expected fallback to defaults on invalid JSON")
	}
}

func TestRandom(t *testing.T) {
	s := Random()
	if s == nil {
		t.Error("expected non-nil Random suggestion")
	}
	if s.Title == "" {
		t.Error("expected non-empty title")
	}
}
