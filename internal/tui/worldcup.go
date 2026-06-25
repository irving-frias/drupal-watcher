//go:build worldcup

package tui

import (
	"fmt"
	"sort"
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

	// ── TODAY ──
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
		if todayCount < 3 {
			hs, as := g.HomeScore, g.AwayScore
			if hs == "null" || hs == "" {
				hs = "-"
			}
			if as == "null" || as == "" {
				as = "-"
			}
			b.WriteString(fmt.Sprintf("\n  %s %s-%s %s%s",
				dim.Render(truncateStr(home, 10)),
				bold.Render(hs), bold.Render(as),
				dim.Render(truncateStr(away, 10)),
				statusSuffix(g.TimeElapsed)))
		} else if todayCount == 3 {
			b.WriteString("\n" + dim.Render("  ..."))
		}
		todayCount++
	}
	if todayCount == 0 {
		b.WriteString("\n" + yellow.Render(" TODAY ") + dim.Render(now.Format("Jan 2")))
		b.WriteString(dim.Render("\n  No matches today"))
	}

	// ── Group A ──
	if groups != nil && len(groups) > 0 {
		b.WriteString("\n\n  " + bold.Render("Group "+groups[0].Name))
		for _, t := range groupTop2(groups[0].Teams) {
			n := client.TeamName(t.TeamID)
			if n == "" {
				n = t.TeamID
			}
			p := t.Pts
			ps := dim
			if p != "0" {
				ps = bold
			}
			b.WriteString(fmt.Sprintf("\n  %s %s", dim.Render(truncateStr(n, 10)), ps.Render(p)))
		}
	}

	// ── NEXT ──
	var upcoming []wc.Game
	for _, g := range games {
		if len(upcoming) >= 2 {
			break
		}
		home := realName(g.HomeTeamNameEn, g.HomeTeamLabel)
		away := realName(g.AwayTeamNameEn, g.AwayTeamLabel)
		if home == "" || away == "" {
			continue
		}
		if !isUpcoming(g, now) {
			continue
		}
		upcoming = append(upcoming, g)
	}
	if len(upcoming) > 0 {
		b.WriteString("\n\n  " + yellow.Render("NEXT"))
		for _, g := range upcoming {
			home := realName(g.HomeTeamNameEn, g.HomeTeamLabel)
			away := realName(g.AwayTeamNameEn, g.AwayTeamLabel)
			b.WriteString(fmt.Sprintf("\n  %s vs %s",
				dim.Render(truncateStr(home, 10)), dim.Render(truncateStr(away, 10))))
		}
	}

	b.WriteString("\n" + dim.Render("────────────────────"))
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

func isUpcoming(g wc.Game, now time.Time) bool {
	if strings.ToLower(g.TimeElapsed) != "notstarted" {
		return false
	}
	datePart := strings.TrimSpace(strings.SplitN(g.LocalDate, " ", 2)[0])
	t, err := time.Parse("01/02/2006", datePart)
	if err != nil {
		return false
	}
	return !t.Before(now.Truncate(24 * time.Hour))
}

func groupTop2(teams []wc.GroupStanding) []wc.GroupStanding {
	top := make([]wc.GroupStanding, len(teams))
	copy(top, teams)
	sort.Slice(top, func(i, j int) bool {
		pi, _ := sidebarToInt(top[i].Pts)
		pj, _ := sidebarToInt(top[j].Pts)
		return pj < pi
	})
	if len(top) > 2 {
		top = top[:2]
	}
	return top
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
