package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/irving-frias/drupal-watcher/internal/watcher"
)

func testHandle() *watcher.Handle {
	return &watcher.Handle{
		StopCh:  make(chan struct{}),
		EventCh: make(chan watcher.EventMsg, 100),
		Stats:   &watcher.Stats{StartTime: time.Now()},
	}
}

func TestNewModel(t *testing.T) {
	m := NewModel(testHandle())
	if m == nil {
		t.Fatal("expected non-nil model")
	}
	if m.eventCap != eventBufferSize {
		t.Errorf("expected eventCap %d, got %d", eventBufferSize, m.eventCap)
	}
	if m.autoScroll != true {
		t.Error("expected autoScroll true")
	}
	if len(m.memHistory) != 0 {
		t.Errorf("expected empty memHistory, got %d", len(m.memHistory))
	}
}

func TestPushEvent(t *testing.T) {
	m := NewModel(testHandle())

	m.pushEvent(eventLine{Timestamp: "00:00:01", Content: "test", Style: infoStyle})
	if len(m.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(m.events))
	}
	if m.events[0].Count != 1 {
		t.Errorf("expected Count 1, got %d", m.events[0].Count)
	}

	// Same content should collapse
	m.pushEvent(eventLine{Timestamp: "00:00:02", Content: "test", Style: infoStyle})
	if len(m.events) != 1 {
		t.Fatalf("expected 1 event after collapse, got %d", len(m.events))
	}
	if m.events[0].Count != 2 {
		t.Errorf("expected Count 2, got %d", m.events[0].Count)
	}

	// Different content should append
	m.pushEvent(eventLine{Timestamp: "00:00:03", Content: "other", Style: infoStyle})
	if len(m.events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(m.events))
	}
	if m.events[1].Count != 1 {
		t.Errorf("expected Count 1, got %d", m.events[1].Count)
	}

	// Overflow should drop oldest
	m.eventCap = 2
	m.pushEvent(eventLine{Timestamp: "00:00:04", Content: "third", Style: infoStyle})
	if len(m.events) != 2 {
		t.Fatalf("expected 2 events after overflow, got %d", len(m.events))
	}
	if m.events[0].Content != "other" {
		t.Errorf("expected 'other' as first event after overflow, got %s", m.events[0].Content)
	}
}

func TestAddToHistory(t *testing.T) {
	m := NewModel(testHandle())

	m.addToHistory("status")
	if len(m.history) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(m.history))
	}

	// Duplicate should be skipped
	m.addToHistory("status")
	if len(m.history) != 1 {
		t.Errorf("expected 1 history entry after duplicate, got %d", len(m.history))
	}

	// Cap: fill with many entries until it exceeds maxHistory
	for i := 0; i < maxHistory+10; i++ {
		m.addToHistory(fmt.Sprintf("cmd-%d", i))
	}
	if len(m.history) != maxHistory {
		t.Errorf("expected %d history entries after cap, got %d", maxHistory, len(m.history))
	}
	// With cap=100 and 110 entries, oldest should be cmd-10
	if m.history[0] != "cmd-10" {
		t.Errorf("expected 'cmd-10' as oldest, got %s", m.history[0])
	}
	if m.history[maxHistory-1] != "cmd-109" {
		t.Errorf("expected 'cmd-109' as newest, got %s", m.history[maxHistory-1])
	}
}

func TestCompleteInputEmpty(t *testing.T) {
	m := NewModel(testHandle())
	m.completeInput()
	val := m.input.Value()
	if val == "" {
		t.Error("expected input to be filled after completion")
	}
	if !strings.HasSuffix(val, " ") {
		t.Error("expected trailing space after command completion")
	}
}

func TestCompleteInputCommand(t *testing.T) {
	m := NewModel(testHandle())

	m.input.SetValue("st")
	m.input.CursorEnd()
	m.completeInput()
	val := m.input.Value()
	if val != "stats " && val != "status " {
		t.Errorf("expected 'stats ' or 'status ', got %q", val)
	}
}

func TestCompleteInputFilter(t *testing.T) {
	m := NewModel(testHandle())
	m.siteClears["mysite"] = 5

	m.input.SetValue("filter my")
	m.input.CursorEnd()
	m.completeInput()
	val := m.input.Value()
	if val != "filter mysite" {
		t.Errorf("expected 'filter mysite', got %q", val)
	}
}

func TestCompleteInputCycles(t *testing.T) {
	m := NewModel(testHandle())
	m.input.SetValue("")
	m.completeInput()
	first := m.input.Value()
	m.completeInput()
	second := m.input.Value()
	if first == second {
		t.Errorf("expected different completions on second call, got %q both times", first)
	}
}

func TestRenderEventsEmpty(t *testing.T) {
	m := NewModel(testHandle())
	out := m.renderEvents()
	if !strings.Contains(out, "Waiting") {
		t.Errorf("expected 'Waiting' message, got %q", out)
	}
}

func TestRenderEventsCollapsed(t *testing.T) {
	m := NewModel(testHandle())
	m.pushEvent(eventLine{Timestamp: "00:00:01", Content: "test", Style: infoStyle})
	m.pushEvent(eventLine{Timestamp: "00:00:02", Content: "test", Style: infoStyle})
	m.pushEvent(eventLine{Timestamp: "00:00:03", Content: "test", Style: infoStyle})

	out := m.renderEvents()
	if !strings.Contains(out, "3x test") {
		t.Errorf("expected collapsed '3x test', got %q", out)
	}
}

func TestRenderEventsFilter(t *testing.T) {
	m := NewModel(testHandle())
	m.pushEvent(eventLine{Timestamp: "00:00:01", Content: "[site1] drush cr", Style: infoStyle})
	m.pushEvent(eventLine{Timestamp: "00:00:02", Content: "[site2] drush cr", Style: infoStyle})

	m.siteFilter = "site2"
	out := m.renderEvents()
	if !strings.Contains(out, "site2") {
		t.Error("expected site2 event in output")
	}
	if strings.Contains(out, "site1") {
		t.Error("expected site1 event filtered out")
	}
}

func TestRenderInputNormal(t *testing.T) {
	m := NewModel(testHandle())
	m.input.SetValue("test")
	out := m.renderInput()
	if !strings.HasPrefix(out, ">") {
		t.Errorf("expected '>' prefix, got %q", out)
	}
	if !strings.Contains(out, "test") {
		t.Errorf("expected 'test' in input, got %q", out)
	}
}

func TestRenderInputHistory(t *testing.T) {
	m := NewModel(testHandle())
	m.historyIdx = 0
	out := m.renderInput()
	if !strings.Contains(out, "history") {
		t.Errorf("expected 'history' indicator, got %q", out)
	}
}

func TestExecuteCommandStatus(t *testing.T) {
	m := NewModel(testHandle())
	cmd := m.executeCommand("status")
	if cmd != nil {
		t.Error("expected nil cmd for status")
	}
	if len(m.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(m.events))
	}
	if !strings.Contains(m.events[0].Content, "PID") {
		t.Errorf("expected 'PID' in status output, got %q", m.events[0].Content)
	}
}

func TestExecuteCommandHelp(t *testing.T) {
	m := NewModel(testHandle())
	cmd := m.executeCommand("help")
	if cmd != nil {
		t.Error("expected nil cmd for help")
	}
	if !m.showHelp {
		t.Error("expected showHelp true after help command")
	}
}

func TestExecuteCommandFilter(t *testing.T) {
	m := NewModel(testHandle())
	m.executeCommand("filter mysite")
	if m.siteFilter != "mysite" {
		t.Errorf("expected siteFilter 'mysite', got %q", m.siteFilter)
	}

	m.executeCommand("filter")
	if m.siteFilter != "" {
		t.Errorf("expected siteFilter cleared, got %q", m.siteFilter)
	}
}

func TestExecuteCommandStats(t *testing.T) {
	m := NewModel(testHandle())
	m.siteClears["site-a"] = 3

	m.executeCommand("stats")
	if len(m.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(m.events))
	}
	if !strings.Contains(m.events[0].Content, "site-a") {
		t.Errorf("expected 'site-a' in stats, got %q", m.events[0].Content)
	}
	if !strings.Contains(m.events[0].Content, "3") {
		t.Errorf("expected '3' in stats, got %q", m.events[0].Content)
	}
}

func TestExecuteCommandStatsEmpty(t *testing.T) {
	m := NewModel(testHandle())
	m.executeCommand("stats")
	if len(m.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(m.events))
	}
	if !strings.Contains(m.events[0].Content, "No site-specific") {
		t.Errorf("expected 'No site-specific' message, got %q", m.events[0].Content)
	}
}

func TestExecuteCommandQuit(t *testing.T) {
	m := NewModel(testHandle())
	cmd := m.executeCommand("stop")
	if cmd == nil {
		t.Error("expected non-nil cmd for stop")
	}
}

func TestExecuteCommandUnknown(t *testing.T) {
	m := NewModel(testHandle())
	m.executeCommand("foobar")
	if len(m.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(m.events))
	}
	if !strings.Contains(m.events[0].Content, "Unknown") {
		t.Errorf("expected 'Unknown' message, got %q", m.events[0].Content)
	}
}

func TestSparklineEmpty(t *testing.T) {
	out := sparkline(nil, 10)
	if out != "" {
		t.Errorf("expected empty output, got %q", out)
	}
}

func TestSparklineSingleValue(t *testing.T) {
	out := sparkline([]float64{50}, 10)
	if out == "" {
		t.Fatal("expected non-empty output")
	}
	if n := utf8.RuneCountInString(out); n != 1 {
		t.Errorf("expected 1 character, got %d", n)
	}
}

func TestSparklineTruncation(t *testing.T) {
	vals := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	out := sparkline(vals, 3)
	if n := utf8.RuneCountInString(out); n != 3 {
		t.Errorf("expected 3 characters (truncated), got %d", n)
	}
}

func TestSparklineAllEqual(t *testing.T) {
	vals := []float64{5, 5, 5, 5, 5}
	out := sparkline(vals, 10)
	if n := utf8.RuneCountInString(out); n != 5 {
		t.Errorf("expected 5 characters, got %d", n)
	}
}
