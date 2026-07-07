package metrics_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/irving-frias/drupal-watcher/internal/metrics"
)

func TestInit(t *testing.T) {
	metrics.Init()
	snap := metrics.Snapshot()
	if snap.TotalChanges != 0 {
		t.Errorf("expected 0 changes after init, got %d", snap.TotalChanges)
	}
}

func TestRecordChange(t *testing.T) {
	metrics.Init()
	metrics.RecordChange()
	metrics.RecordChange()
	snap := metrics.Snapshot()
	if snap.TotalChanges != 2 {
		t.Errorf("expected 2 changes, got %d", snap.TotalChanges)
	}
}

func TestRecordClear(t *testing.T) {
	metrics.Init()
	metrics.RecordClear("site1")
	metrics.RecordClear("site2")
	metrics.RecordClear("site1")
	snap := metrics.Snapshot()
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
	metrics.Init()
	metrics.RecordError()
	snap := metrics.Snapshot()
	if snap.Errors != 1 {
		t.Errorf("expected 1 error, got %d", snap.Errors)
	}
}

func TestSave(t *testing.T) {
	metrics.Init()
	metrics.RecordChange()
	metrics.RecordClear("default")
	path := filepath.Join(t.TempDir(), "stats.json")
	if err := metrics.Save(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("saved file not found: %v", err)
	}
}
