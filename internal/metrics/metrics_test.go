package metrics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	Init()
	snap := Snapshot()
	if snap.TotalChanges != 0 {
		t.Errorf("expected 0 changes after init, got %d", snap.TotalChanges)
	}
}

func TestRecordChange(t *testing.T) {
	Init()
	RecordChange()
	RecordChange()
	snap := Snapshot()
	if snap.TotalChanges != 2 {
		t.Errorf("expected 2 changes, got %d", snap.TotalChanges)
	}
}

func TestRecordClear(t *testing.T) {
	Init()
	RecordClear("site1")
	RecordClear("site2")
	RecordClear("site1")
	snap := Snapshot()
	if snap.TotalClears != 3 {
		t.Errorf("expected 3 clears, got %d", snap.TotalClears)
	}
	if snap.ClearsPerSite["site1"] != 2 {
		t.Errorf("expected 2 clears for site1, got %d", snap.ClearsPerSite["site1"])
	}
	if snap.ClearsPerSite["site2"] != 1 {
		t.Errorf("expected 1 clears for site2, got %d", snap.ClearsPerSite["site2"])
	}
}

func TestRecordError(t *testing.T) {
	Init()
	RecordError()
	snap := Snapshot()
	if snap.Errors != 1 {
		t.Errorf("expected 1 error, got %d", snap.Errors)
	}
}

func TestSave(t *testing.T) {
	Init()
	RecordChange()
	RecordClear("default")
	path := filepath.Join(t.TempDir(), "stats.json")
	if err := Save(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("saved file not found: %v", err)
	}
}
