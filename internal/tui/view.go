package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderStatus() string {
	s := m.status
	statusDot := green.Render("●")
	if !s.Running {
		statusDot = red.Render("●")
	}

	statusLine := fmt.Sprintf("%s drupal-watcher  PID: %d  Uptime: %s",
		statusDot, s.PID, s.Uptime)

	memColor := green
	if s.AllocMB >= 500 {
		memColor = red
	} else if s.AllocMB >= 100 {
		memColor = yellow
	}

	memLine := fmt.Sprintf("Memory: %s  |  Kernel watches: %d  |  Changes: %d  |  Clears: %d",
		memColor.Render(fmt.Sprintf("%.1f MB", s.AllocMB)),
		s.WatchCount, s.Changes, s.Clears)

	return lipgloss.JoinVertical(lipgloss.Left,
		statusLine,
		memLine,
	)
}

func (m *Model) renderEvents() string {
	if len(m.events) == 0 {
		return dim.Render("Waiting for file changes...")
	}

	var lines []string
	for _, e := range m.events {
		ts := dim.Render(e.Timestamp)
		icon := e.Style.String()
		lines = append(lines, fmt.Sprintf("%s %s %s", ts, icon, e.Content))
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderInput() string {
	return fmt.Sprintf("> %s", m.input.View())
}

func (m *Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	status := statusStyle.Render(m.renderStatus())
	events := m.viewport.View()
	// wrap events in bordered style
	events = eventsStyle.Render(events)
	input := cmdStyle.Render(m.renderInput())

	return lipgloss.JoinVertical(lipgloss.Left, status, events, input)
}
