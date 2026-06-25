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

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render(" ⚽ World Cup 2026 "))
	b.WriteString("\n" + dim.Render("────────────────────"))

	// ── Today's matches ──
	todayCount := 0
	for _, g := range games {
		if !isSidebarToday(g.LocalDate) {
			continue
		}
		home := realName(g.HomeTeamNameEn, g.HomeTeamLabel)
		away := realName(g.AwayTeamNameEn, g.AwayTeamLabel)
		if home == "" || away == "" {
			continue
		}
		if todayCount == 0 {
			b.WriteString("\n" + yellow.Render(" TODAY ") + dim.Render(now.Format("Jan 2")))
		}
		if todayCount >= 3 {
			if todayCount == 3 {
				b.WriteString("\n" + dim.Render("  ..."))
			}
			todayCount++
			continue
		}

		hs := g.HomeScore
		as := g.AwayScore
		if hs == "null" || hs == "" {
			hs = "-"
		}
		if as == "null" || as == "" {
			as = "-"
		}

		status := statusSuffix(g.TimeElapsed)
		b.WriteString(fmt.Sprintf("\n  %s %s-%s %s%s",
			dim.Render(truncateStr(home, 14)),
			bold.Render(hs), bold.Render(as),
			dim.Render(truncateStr(away, 14)), status))
		todayCount++
	}
	if todayCount == 0 {
		b.WriteString("\n" + yellow.Render(" TODAY ") + dim.Render(now.Format("Jan 2")))
		b.WriteString(dim.Render("\n  No matches today"))
	}

	// ── Group standings (top 2 groups) ──
	if groups != nil {
		for i, grp := range groups {
			if i >= 2 {
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

			b.WriteString("\n" + dim.Render("────────────────────"))
			b.WriteString("\n  " + bold.Render("Group "+grp.Name))
			limit := min(2, len(top2))
			for _, t := range top2[:limit] {
				n := client.TeamName(t.TeamID)
				if n == "" {
					n = t.TeamID
				}
				ptsStyle := dim
				if t.Pts != "0" {
					ptsStyle = bold
				}
				b.WriteString(fmt.Sprintf("\n  %s %s",
					dim.Render(truncateStr(n, 14)), ptsStyle.Render(t.Pts)))
			}
		}
	}

	// ── Upcoming matches (real teams only) ──
	b.WriteString("\n" + dim.Render("────────────────────"))
	b.WriteString("\n" + yellow.Render(" NEXT"))
	shownNext := 0
	for _, g := range games {
		if shownNext >= 3 {
			break
		}
		home := realName(g.HomeTeamNameEn, g.HomeTeamLabel)
		away := realName(g.AwayTeamNameEn, g.AwayTeamLabel)
		if home == "" || away == "" {
			continue
		}
		datePart := strings.TrimSpace(strings.SplitN(g.LocalDate, " ", 2)[0])
		t, err := time.Parse("01/02/2006", datePart)
		if err != nil || t.Before(now.Truncate(24*time.Hour)) {
			continue
		}
		if strings.ToLower(g.TimeElapsed) != "notstarted" {
			continue
		}
		b.WriteString(fmt.Sprintf("\n  %s vs %s",
			dim.Render(truncateStr(home, 11)), dim.Render(truncateStr(away, 11))))
		shownNext++
	}
	if shownNext == 0 {
		b.WriteString(dim.Render("\n  —"))
	}

	b.WriteString(dim.Render("\n  :refresh — update"))
}

func realName(nameEn, label string) string {
	n := nameEn
	if n == "" {
		n = label
	}
	if n == "" || strings.HasPrefix(n, "Winner") || strings.HasPrefix(n, "Runner-up") || strings.HasPrefix(n, "Loser") || strings.HasPrefix(n, "3rd") {
		return ""
	}
	return n
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
