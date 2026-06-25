package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/irving-frias/drupal-watcher/internal/watcher"
)

const eventBufferSize = 100
const maxHistory = 100
const sparklineSize = 20

type eventLine struct {
	Timestamp string
	Content   string
	Style     lipgloss.Style
	Count     int
}

type Model struct {
	Watcher  *watcher.Handle
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
	memHistory   []float64
	completions  []string
	completionIdx int
}

type statusLine struct {
	PID        int
	Uptime     string
	Changes    int64
	Clears     int64
	WatchCount int64
	AllocMB    float64
	Running    bool
}

func NewModel(w *watcher.Handle) *Model {
	ti := textinput.New()
	ti.Placeholder = "type help to see commands"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return &Model{
		Watcher:    w,
		events:     make([]eventLine, 0, eventBufferSize),
		eventCap:   eventBufferSize,
		viewport:   viewport.New(78, 10),
		input:      ti,
		width:      80,
		historyIdx: -1,
		autoScroll: true,
		siteClears: make(map[string]int64),
		memHistory: make([]float64, 0, sparklineSize),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		listenForEvents(m.Watcher),
		textinput.Blink,
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func listenForEvents(w *watcher.Handle) tea.Cmd {
	return func() tea.Msg {
		if w.EventCh == nil {
			return nil
		}
		evt, ok := <-w.EventCh
		if !ok {
			return nil
		}
		return watcherEventMsg{Event: evt}
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

var commands = []string{"status", "stats", "filter", "help", "stop", "quit", "exit"}
