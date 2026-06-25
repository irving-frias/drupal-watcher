package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderStatus() string {
	s := m.status
	statusDot := green.Render("●")
	if !s.Running {
		statusDot = red.Render("●")
	}

	left := fmt.Sprintf("%s drupal-watcher  PID: %d  Uptime: %s",
		statusDot, s.PID, s.Uptime)

	right := ""
	if !m.autoScroll {
		right = dim.Render(" [paused]")
	}
	if m.siteFilter != "" {
		right += cyan.Render(" [filter: " + m.siteFilter + "]")
	}

	statusLine := left + right

	memColor := green
	if s.AllocMB >= 500 {
		memColor = red
	} else if s.AllocMB >= 100 {
		memColor = yellow
	}

	memLine := fmt.Sprintf("Memory: %s  |  Kernel watches: %d  |  Changes: %d  |  Clears: %d",
		memColor.Render(fmt.Sprintf("%.1f MB", s.AllocMB)),
		s.WatchCount, s.Changes, s.Clears)

	// Per-site clears
	if len(m.siteClears) > 0 {
		var names []string
		for name := range m.siteClears {
			names = append(names, name)
		}
		sort.Strings(names)
		var parts []string
		for _, name := range names {
			parts = append(parts, fmt.Sprintf("%s: %d", cyan.Render(name), m.siteClears[name]))
		}
		memLine += "  |  " + strings.Join(parts, "  ")
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		statusLine,
		memLine,
	)
}

func (m *Model) renderEvents() string {
	if len(m.events) == 0 {
		return dim.Render("Waiting for file changes...")
	}

	filter := m.siteFilter

	var lines []string
	for _, e := range m.events {
		if filter != "" && !strings.Contains(e.Content, "["+filter+"]") && !strings.Contains(e.Content, filter) {
			continue
		}
		ts := dim.Render(e.Timestamp)
		icon := e.Style.String()
		lines = append(lines, fmt.Sprintf("%s %s %s", ts, icon, e.Content))
	}

	if len(lines) == 0 {
		return dim.Render("No events match filter: " + filter)
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderInput() string {
	prefix := "> "
	if m.historyIdx != -1 {
		prefix = dim.Render("(history) > ")
	}
	return fmt.Sprintf("%s%s", prefix, m.input.View())
}

func (m *Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	status := statusStyle.Render(m.renderStatus())
	events := m.viewport.View()
	events = eventsStyle.Render(events)
	input := cmdStyle.Render(m.renderInput())

	return lipgloss.JoinVertical(lipgloss.Left, status, events, input)
}
