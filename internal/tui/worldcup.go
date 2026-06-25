//go:build worldcup

package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	wc "github.com/irving-frias/drupal-watcher/internal/worldcup"
)

var bold = lipgloss.NewStyle().Bold(true)

func (m *Model) handleWorldcup(subview string) tea.Cmd {
	if !wc.Enabled() {
		m.pushEvent(eventLine{
			Timestamp: "worldcup",
			Content:   "Set DRUPAL_WATCHER_WORLDCUP=1 to enable the World Cup feature.",
			Style:     warnStyle,
		})
		return nil
	}

	client := wc.NewClient()

	if err := client.EnsureTeams(); err != nil {
		m.pushEvent(eventLine{
			Timestamp: "worldcup",
			Content:   fmt.Sprintf("API error: %v", err),
			Style:     errorStyle,
		})
		return nil
	}

	m.worldcupMode = true
	m.worldcupView = subview

	var b strings.Builder

	switch subview {
	case "live", "":
		games, err := client.FetchGames()
		if err != nil {
			m.pushEvent(eventLine{Timestamp: "worldcup", Content: fmt.Sprintf("API error: %v", err), Style: errorStyle})
			m.worldcupMode = false
			return nil
		}
		buildSidebarLive(&b, client, games)
	case "groups":
		groups, err := client.FetchGroups()
		if err != nil {
			m.pushEvent(eventLine{Timestamp: "worldcup", Content: fmt.Sprintf("API error: %v", err), Style: errorStyle})
			m.worldcupMode = false
			return nil
		}
		buildSidebarGroups(&b, client, groups)
	case "schedule":
		games, err := client.FetchGames()
		if err != nil {
			m.pushEvent(eventLine{Timestamp: "worldcup", Content: fmt.Sprintf("API error: %v", err), Style: errorStyle})
			m.worldcupMode = false
			return nil
		}
		buildSidebarSchedule(&b, client, games)
	default:
		m.worldcupMode = false
		m.pushEvent(eventLine{
			Timestamp: "worldcup",
			Content:   "Usage: worldcup [live|groups|schedule] or :back to return",
			Style:     infoStyle,
		})
		return nil
	}

	m.worldcupSidebar = b.String()
	return nil
}

func buildSidebarLive(b *strings.Builder, client *wc.Client, games []wc.Game) {
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render(" ⚽ World Cup 2026 "))
	b.WriteString("\n" + dim.Render(" ─────────────────"))
	b.WriteString("\n" + yellow.Render(" ● LIVE") + " — " + time.Now().Format("Jan 2"))
	b.WriteString("\n")

	shown := 0
	for _, g := range games {
		if !isSidebarToday(g.LocalDate) {
			continue
		}
		if shown >= 6 {
			b.WriteString(dim.Render("   ..."))
			break
		}
		home := g.HomeTeamNameEn
		away := g.AwayTeamNameEn
		if home == "" {
			home = g.HomeTeamLabel
		}
		if away == "" {
			away = g.AwayTeamLabel
		}

		hs := g.HomeScore
		as := g.AwayScore
		if hs == "null" || hs == "" {
			hs = "-"
		}
		if as == "null" || as == "" {
			as = "-"
		}

		b.WriteString("\n  " + dim.Render(home))
		b.WriteString(fmt.Sprintf("\n  %s %s-%s", bold.Render(hs), bold.Render(as), dim.Render(away)))
		b.WriteString("\n")
		shown++
	}
	if shown == 0 {
		b.WriteString(dim.Render("  No matches today"))
	}

	b.WriteString("\n" + dim.Render(" ─────────────────"))
	b.WriteString("\n" + dim.Render(" :worldcup groups — groups"))
	b.WriteString("\n" + dim.Render(" :worldcup schedule — schedule"))
	b.WriteString("\n" + dim.Render(" :back — close"))
}

func buildSidebarGroups(b *strings.Builder, client *wc.Client, groups []wc.Group) {
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render(" ⚽ World Cup 2026 "))
	b.WriteString("\n" + dim.Render(" ─────────────────"))

	maxGroups := 4
	for i, grp := range groups {
		if i >= maxGroups {
			b.WriteString("\n" + dim.Render(fmt.Sprintf("  ... and %d more", len(groups)-maxGroups)))
			break
		}

		top2 := make([]wc.GroupStanding, len(grp.Teams))
		copy(top2, grp.Teams)
		for i := 0; i < len(top2); i++ {
			for j := i + 1; j < len(top2); j++ {
				pi, _ := sidebarToInt(top2[i].Pts)
				pj, _ := sidebarToInt(top2[j].Pts)
				if pj > pi {
					top2[i], top2[j] = top2[j], top2[i]
				}
			}
		}

		b.WriteString(fmt.Sprintf("\n  %s %s", bold.Render("Group "+grp.Name), dim.Render("Pts")))
		for _, t := range top2 {
			n := client.TeamName(t.TeamID)
			ptsStyle := dim
			if t.Pts != "0" {
				ptsStyle = bold
			}
			b.WriteString(fmt.Sprintf("\n  %s %s",
				dim.Render(truncateStr(n, 16)), ptsStyle.Render(t.Pts)))
		}
	}

	b.WriteString("\n" + dim.Render(" ─────────────────"))
	b.WriteString("\n" + dim.Render(" :worldcup live — live"))
	b.WriteString("\n" + dim.Render(" :worldcup schedule — schedule"))
	b.WriteString("\n" + dim.Render(" :back — close"))
}

func buildSidebarSchedule(b *strings.Builder, client *wc.Client, games []wc.Game) {
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render(" ⚽ World Cup 2026 "))
	b.WriteString("\n" + dim.Render(" ─────────────────"))
	b.WriteString("\n" + yellow.Render(" UPCOMING"))
	b.WriteString("\n")

	now := time.Now().Local()
	shown := 0
	for _, g := range games {
		datePart := strings.TrimSpace(strings.SplitN(g.LocalDate, " ", 2)[0])
		t, err := time.Parse("01/02/2006", datePart)
		if err != nil || t.Before(now.Truncate(24*time.Hour)) {
			continue
		}
		if strings.ToLower(g.TimeElapsed) != "notstarted" {
			continue
		}
		if shown >= 5 {
			b.WriteString(dim.Render("   ..."))
			break
		}

		home := g.HomeTeamNameEn
		away := g.AwayTeamNameEn
		if home == "" {
			home = g.HomeTeamLabel
		}
		if away == "" {
			away = g.AwayTeamLabel
		}

		b.WriteString(fmt.Sprintf("  %s\n  %s vs %s\n",
			dim.Render(datePart), dim.Render(home), dim.Render(away)))
		shown++
	}
	if shown == 0 {
		b.WriteString(dim.Render("  No upcoming"))
	}

	b.WriteString("\n" + dim.Render(" ─────────────────"))
	b.WriteString("\n" + dim.Render(" :worldcup live — live"))
	b.WriteString("\n" + dim.Render(" :worldcup groups — groups"))
	b.WriteString("\n" + dim.Render(" :back — close"))
}

func isSidebarToday(dateStr string) bool {
	parts := strings.SplitN(strings.TrimSpace(dateStr), " ", 2)
	if len(parts) == 0 {
		return false
	}
	return parts[0] == time.Now().Local().Format("01/02/2006")
}

func sidebarToInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
