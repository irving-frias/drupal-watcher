package ui

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/irving-frias/drupal-watcher/internal/utils"
	"github.com/irving-frias/drupal-watcher/pkg/core"
)

func dismissedPath(root string) string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = "/tmp"
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return ""
	}
	h := md5.Sum([]byte(abs))
	return filepath.Join(dir, "drupal-watcher", fmt.Sprintf(".drupal-watcher-%x.dismissed", h[:8]))
}

func isStarDismissed(root string) bool {
	p := dismissedPath(root)
	if p == "" {
		return false
	}
	_, err := os.Stat(p)
	return err == nil
}

func writeStarDismissed(root string) {
	p := dismissedPath(root)
	if p == "" {
		return
	}
	os.MkdirAll(filepath.Dir(p), 0700)
	os.WriteFile(p, nil, 0644)
}

const eventBufferSize = 100
const maxHistory = 100
const sparklineSize = 20

type eventLine struct {
	Timestamp string
	Content   string
	Style     lipgloss.Style
	Count     int
}

type statusLine struct {
	PID        int
	Uptime     string
	Changes    int64
	Clears     int64
	AllocMB    float64
	Running    bool
}

type EngineInfo interface {
	Stats() (changes, clears int64)
	StartTime() time.Time
}

type Model struct {
	eventChan <-chan core.EngineEvent
	engineInfo EngineInfo

	status   statusLine
	events   []eventLine
	eventCap int
	viewport viewport.Model
	input    textinput.Model
	width    int

	history    []string
	historyIdx int

	autoScroll bool

	siteFilter string
	siteClears map[string]int64

	showHelp     bool
	showStar     bool
	root         string
	memHistory   []float64
	completions  []string
	completionIdx int
}

func NewModel(eventChan <-chan core.EngineEvent, info EngineInfo, root string) *Model {
	ti := textinput.New()
	ti.Placeholder = "type help to see commands"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return &Model{
		eventChan:  eventChan,
		engineInfo: info,
		events:     make([]eventLine, 0, eventBufferSize),
		eventCap:   eventBufferSize,
		viewport:   viewport.New(78, 10),
		input:      ti,
		width:      80,
		historyIdx: -1,
		autoScroll: true,
		showStar:   !isStarDismissed(root),
		root:       root,
		siteClears: make(map[string]int64),
		memHistory: make([]float64, 0, sparklineSize),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		listenForEvents(m.eventChan),
		textinput.Blink,
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func listenForEvents(eventChan <-chan core.EngineEvent) tea.Cmd {
	return func() tea.Msg {
		evt, ok := <-eventChan
		if !ok {
			return nil
		}
		return engineEventMsg{Event: evt}
	}
}

func (m *Model) pushEvent(line eventLine) {
	if len(m.events) > 0 {
		last := &m.events[len(m.events)-1]
		if last.Content == line.Content {
			last.Count++
			return
		}
	}
	line.Count = 1
	m.events = append(m.events, line)
	if len(m.events) > m.eventCap {
		m.events = m.events[1:]
	}
}

func (m *Model) addToHistory(cmd string) {
	if len(m.history) > 0 && m.history[len(m.history)-1] == cmd {
		return
	}
	m.history = append(m.history, cmd)
	if len(m.history) > maxHistory {
		m.history = m.history[1:]
	}
}

var commands = []string{"status", "stats", "filter", "star", "dismiss", "help", "stop", "quit", "exit"}

func (m *Model) updateStatus() {
	uptime := time.Since(m.engineInfo.StartTime())
	var allocMB float64
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	allocMB = float64(mem.Alloc) / 1024 / 1024

	m.memHistory = append(m.memHistory, allocMB)
	if len(m.memHistory) > sparklineSize {
		m.memHistory = m.memHistory[1:]
	}

	changes, clears := m.engineInfo.Stats()
	m.status = statusLine{
		PID:     os.Getpid(),
		Uptime:  utils.FormatDuration(uptime),
		Changes: changes,
		Clears:  clears,
		AllocMB: allocMB,
		Running: true,
	}
}
