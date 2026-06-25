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

type eventLine struct {
	Timestamp string
	Content   string
	Style     lipgloss.Style
}

type Model struct {
	Watcher  *watcher.Handle
	status   statusLine
	events   []eventLine
	eventCap int
	viewport viewport.Model
	input    textinput.Model
	ready    bool
	width    int

	history    []string
	historyIdx int

	autoScroll bool

	siteFilter string
	siteClears map[string]int64

	worldcupSidebar     string
	worldcupLastRefresh time.Time
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
	ti.Placeholder = "Type a command..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return &Model{
		Watcher:    w,
		events:     make([]eventLine, 0, eventBufferSize),
		eventCap:   eventBufferSize,
		viewport:   viewport.New(78, 10),
		input:      ti,
		ready:      true,
		width:      80,
		historyIdx: -1,
		autoScroll: true,
		siteClears: make(map[string]int64),
	}
}

func (m *Model) Init() tea.Cmd {
	m.refreshWorldcup()
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
	m.events = append(m.events, line)
	if len(m.events) > m.eventCap {
		m.events = m.events[1:]
	}
}
