package utils_test

import (
	"testing"

	"github.com/irving-frias/drupal-watcher/internal/utils"
)

func TestSetColorsEnabled(t *testing.T) {
	utils.SetColorsEnabled(false)
	if utils.ColorsEnabled() {
		t.Error("expected colors disabled")
	}

	utils.SetColorsEnabled(true)
	if !utils.ColorsEnabled() {
		t.Error("expected colors enabled")
	}
}

func TestColorFuncs(t *testing.T) {
	utils.SetColorsEnabled(true)

	if utils.Red("test") == "test" {
		t.Error("expected colored output")
	}
	if utils.Green("test") == "test" {
		t.Error("expected colored output")
	}
	if utils.Bold("test") == "test" {
		t.Error("expected bold output")
	}
	if utils.Dim("test") == "test" {
		t.Error("expected dim output")
	}

	utils.SetColorsEnabled(false)
	if utils.Red("test") != "test" {
		t.Error("expected plain output when colors disabled")
	}
	if utils.Green("test") != "test" {
		t.Error("expected plain output when colors disabled")
	}
}

func TestTimestamp(t *testing.T) {
	ts := utils.Timestamp()
	if len(ts) == 0 {
		t.Error("expected non-empty timestamp")
	}
}

func TestPrintHeader(t *testing.T) {
	// Should not panic
	utils.PrintHeader("test header")
}

func TestPrintSection(t *testing.T) {
	items := []utils.SectionItem{
		[2]string{"Key", "Value"},
		"Plain string",
	}
	utils.PrintSection("Test", items)
}

func TestDefaultPatterns(t *testing.T) {
	patterns := utils.DefaultPatterns
	if len(patterns) == 0 {
		t.Error("expected default patterns")
	}
	expected := map[string]bool{".php": true, ".twig": true, ".css": true, ".js": true}
	for _, p := range patterns {
		delete(expected, p)
	}
	for missing := range expected {
		t.Errorf("expected %s in default patterns", missing)
	}
}

func TestPossibleDocroots(t *testing.T) {
	if len(utils.PossibleDocroots) == 0 {
		t.Error("expected possible docroots")
	}
}

func TestExcludedDirs(t *testing.T) {
	if len(utils.ExcludedDirs) == 0 {
		t.Error("expected excluded dirs")
	}
}

func TestGetMemStats(t *testing.T) {
	s := utils.GetMemStats(42)
	if s.AllocMB <= 0 {
		t.Error("expected positive AllocMB")
	}
	if s.WatchCount != 42 {
		t.Errorf("expected WatchCount 42, got %d", s.WatchCount)
	}
}

func TestPrintMemStats(t *testing.T) {
	// Should not panic
	utils.PrintMemStats(utils.MemStats{AllocMB: 50, WatchCount: 10})
	utils.PrintMemStats(utils.MemStats{AllocMB: 150, WatchCount: 10})
	utils.PrintMemStats(utils.MemStats{AllocMB: 600, WatchCount: 10})
}
