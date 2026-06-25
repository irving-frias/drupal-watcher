//go:build worldcup

package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	wc "github.com/irving-frias/drupal-watcher/internal/worldcup"
)

var bold = lipgloss.NewStyle().Bold(true)

func (m *Model) refreshWorldcup() {
	client := wc.NewClient()

	if !wc.Enabled() {
		return
	}

	if err := client.EnsureTeams(); err != nil {
		return
	}

	games, err := client.FetchGames()
	if err != nil {
		return
	}
	groups, err := client.FetchGroups()
	if err != nil {
		groups = nil
	}

	var b strings.Builder
	buildSidebar(&b, client, games, groups)
	m.worldcupSidebar = b.String()
	m.worldcupLastRefresh = time.Now()
}

func buildSidebar(b *strings.Builder, client *wc.Client, games []wc.Game, groups []wc.Group) {
	now := time.Now().Local()
	todayStr := now.Format("Jan 2")

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render(" ⚽ World Cup 2026 "))
	b.WriteString("\n" + dim.Render("────────────────────"))

	// ── Today's matches ──
	b.WriteString("\n" + yellow.Render(" TODAY ") + dim.Render(todayStr))

	shown := 0
	for _, g := range games {
		if !isSidebarToday(g.LocalDate) {
			continue
		}
		if shown >= 4 {
			b.WriteString("\n" + dim.Render("  ..."))
			break
		}
		home := teamName(g)
		away := teamAway(g)

		hs := g.HomeScore
		as := g.AwayScore
		if hs == "null" || hs == "" {
			hs = "-"
		}
		if as == "null" || as == "" {
			as = "-"
		}

		status := statusSuffix(g.TimeElapsed)
		home = truncateStr(home, 14)
		away = truncateStr(away, 14)
		b.WriteString(fmt.Sprintf("\n  %s %s-%s %s%s",
			dim.Render(home), bold.Render(hs), bold.Render(as), dim.Render(away), status))
		shown++
	}
	if shown == 0 {
		b.WriteString(dim.Render("\n  No matches today"))
	}

	// ── Group standings ──
	maxGroups := 4
	if groups != nil {
		for i, grp := range groups {
			if i >= maxGroups {
				break
			}

			b.WriteString("\n" + dim.Render("────────────────────"))
			b.WriteString("\n  " + bold.Render("Group "+grp.Name))

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

			for _, t := range top2[:min(2, len(top2))] {
				n := client.TeamName(t.TeamID)
				ptsStyle := dim
				if t.Pts != "0" {
					ptsStyle = bold
				}
				b.WriteString(fmt.Sprintf("\n  %s %s",
					dim.Render(truncateStr(n, 14)), ptsStyle.Render(t.Pts)))
			}
		}
	}

	// ── Upcoming matches ──
	upcoming := nextMatches(games, now, 3)
	if len(upcoming) > 0 {
		b.WriteString("\n" + dim.Render("────────────────────"))
		b.WriteString("\n" + yellow.Render(" NEXT"))
		for _, g := range upcoming {
			home := teamName(g)
			away := teamAway(g)
			home = truncateStr(home, 12)
			away = truncateStr(away, 12)
			b.WriteString(fmt.Sprintf("\n  %s", dim.Render(home+" vs "+away)))
		}
	}

	b.WriteString("\n" + dim.Render("────────────────────"))
	b.WriteString(dim.Render("\n  :refresh — update"))
}

func teamName(g wc.Game) string {
	if g.HomeTeamNameEn != "" {
		return g.HomeTeamNameEn
	}
	return g.HomeTeamLabel
}

func teamAway(g wc.Game) string {
	if g.AwayTeamNameEn != "" {
		return g.AwayTeamNameEn
	}
	return g.AwayTeamLabel
}

func statusSuffix(elapsed string) string {
	switch strings.ToLower(elapsed) {
	case "finished":
		return dim.Render(" FT")
	case "notstarted":
		return ""
	default:
		return yellow.Render(" " + elapsed)
	}
}

func nextMatches(games []wc.Game, now time.Time, limit int) []wc.Game {
	var result []wc.Game
	for _, g := range games {
		if len(result) >= limit {
			break
		}
		datePart := strings.TrimSpace(strings.SplitN(g.LocalDate, " ", 2)[0])
		t, err := time.Parse("01/02/2006", datePart)
		if err != nil || t.Before(now.Truncate(24*time.Hour)) {
			continue
		}
		if strings.ToLower(g.TimeElapsed) != "notstarted" {
			continue
		}
		result = append(result, g)
	}
	return result
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
