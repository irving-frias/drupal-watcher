package tui

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func sparkline(vals []float64, max int) string {
	if len(vals) == 0 {
		return ""
	}
	chars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	mx := 0.0
	for _, v := range vals {
		if v > mx {
			mx = v
		}
	}
	if mx == 0 {
		mx = 1
	}
	n := len(vals)
	if n > max {
		n = max
		vals = vals[len(vals)-max:]
	}
	var b strings.Builder
	for _, v := range vals {
		idx := int(math.Round((v / mx) * 7))
		if idx < 0 {
			idx = 0
		}
		if idx > 7 {
			idx = 7
		}
		b.WriteString(chars[idx])
	}
	return b.String()
}

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

	spark := sparkline(m.memHistory, sparklineSize)
	sparkStr := ""
	if spark != "" {
		sparkStr = "  " + dim.Render(spark)
	}

	memLine := fmt.Sprintf("Memory: %s%s  |  Kernel watches: %d  |  Changes: %d  |  Clears: %d",
		memColor.Render(fmt.Sprintf("%.1f MB", s.AllocMB)),
		sparkStr, s.WatchCount, s.Changes, s.Clears)

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
		content := e.Content
		if e.Count > 1 {
			content = fmt.Sprintf("%dx %s", e.Count, e.Content)
		}
		lines = append(lines, fmt.Sprintf("%s %s %s", ts, icon, content))
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

func (m *Model) renderHelp() string {
	var b strings.Builder
	b.WriteString(bold.Render("  drupal-watcher — Commands"))
	b.WriteString("\n" + dim.Render("  ───────────────────────────────────"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("\n  %s    Show watcher status and runtime stats", green.Render("status")))
	b.WriteString(fmt.Sprintf("\n  %s      Show clear counts per site", green.Render("stats")))
	b.WriteString(fmt.Sprintf("\n  %s  <site>  Filter events by site name", green.Render("filter")))
	b.WriteString(fmt.Sprintf("\n  %s         Show this help", green.Render("help")))
	b.WriteString(fmt.Sprintf("\n  %s       Stop the watcher and exit", green.Render("stop")))
	b.WriteString("\n")
	b.WriteString("\n" + dim.Render("  Keys"))
	b.WriteString("\n" + dim.Render("  ───────────────────────────────────"))
	b.WriteString(fmt.Sprintf("\n  %s    Quit", dim.Render("ctrl+c / ctrl+d")))
	b.WriteString(fmt.Sprintf("\n  %s   Page up/down in event log", dim.Render("pgup / pgdn")))
	b.WriteString(fmt.Sprintf("\n  %s    Scroll to top", dim.Render("home")))
	b.WriteString(fmt.Sprintf("\n  %s   Toggle auto-scroll", dim.Render("end")))
	b.WriteString(fmt.Sprintf("\n  %s         Toggle this help", dim.Render("?")))
	b.WriteString(fmt.Sprintf("\n  %s    Navigate command history", dim.Render("↑ / ↓")))
	b.WriteString(fmt.Sprintf("\n  %s         Mouse wheel to scroll", dim.Render("scroll")))
	b.WriteString(fmt.Sprintf("\n  %s Tab   Complete commands / site names", dim.Render("tab")))
	b.WriteString("\n\n" + dim.Render("  Press ? or Esc to close help"))
	return b.String()
}

func (m *Model) View() string {
	status := statusStyle.Render(m.renderStatus())

	if m.showHelp {
		content := helpStyle.Render(m.renderHelp())
		return lipgloss.JoinVertical(lipgloss.Left, status, content)
	}

	events := m.viewport.View()
	events = eventsStyle.Render(events)

	input := cmdStyle.Render(m.renderInput())

	return lipgloss.JoinVertical(lipgloss.Left, status, events, input)
}
