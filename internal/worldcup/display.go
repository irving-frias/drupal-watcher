//go:build worldcup

package worldcup

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func isToday(dateStr string) bool {
	dateStr = strings.TrimSpace(dateStr)
	parts := strings.SplitN(dateStr, " ", 2)
	if len(parts) == 0 {
		return false
	}
	return parts[0] == time.Now().Local().Format("01/02/2006")
}

func statusColor(status string) string {
	switch strings.ToLower(status) {
	case "finished":
		return "\x1b[2m"
	case "notstarted":
		return "\x1b[90m"
	}
	return "\x1b[33m"
}

func scoreColor(home, away string) string {
	if home == "" || away == "" || home == "null" || away == "null" {
		return "\x1b[90m"
	}
	return "\x1b[97m"
}

func PrintLiveGames(client *Client, games []Game) {
	FprintLiveGames(os.Stdout, client, games)
}

func FprintLiveGames(w io.Writer, client *Client, games []Game) {
	now := time.Now().Local()
	today := now.Format("01/02/2006")

	var live, todayFinished, upcoming []Game
	for _, g := range games {
		if !isToday(g.LocalDate) {
			continue
		}
		switch strings.ToLower(g.TimeElapsed) {
		case "notstarted":
			upcoming = append(upcoming, g)
		case "finished":
			todayFinished = append(todayFinished, g)
		default:
			live = append(live, g)
		}
	}

	if len(live) > 0 {
		fmt.Fprintf(w, "\n  \x1b[33mLIVE\x1b[0m — %s\n", today)
		for _, g := range live {
			fprintGame(w, client, g)
		}
	}

	if len(upcoming) > 0 {
		fmt.Fprintf(w, "\n  \x1b[90mUPCOMING\x1b[0m — %s\n", today)
		for _, g := range upcoming {
			fprintGame(w, client, g)
		}
	}

	if len(todayFinished) > 0 {
		fmt.Fprintf(w, "\n  \x1b[2mFINISHED\x1b[0m — %s\n", today)
		for _, g := range todayFinished {
			fprintGame(w, client, g)
		}
	}

	if len(live)+len(upcoming)+len(todayFinished) == 0 {
		fmt.Fprintf(w, "\n  \x1b[90mNo matches today (%s)\x1b[0m\n", today)
	}
}

func fprintGame(w io.Writer, client *Client, g Game) {
	home := g.HomeTeamNameEn
	away := g.AwayTeamNameEn
	if home == "" {
		home = g.HomeTeamLabel
	}
	if away == "" {
		away = g.AwayTeamLabel
	}

	home = fmt.Sprintf("%-24s", home)
	away = fmt.Sprintf("%-24s", away)

	status := g.TimeElapsed
	if status == "Finished" || status == "finished" {
		status = "FT"
	} else if status == "notstarted" {
		status = fmt.Sprintf("%s", g.LocalDate)
	}

	sc := scoreColor(g.HomeScore, g.AwayScore)
	hs := g.HomeScore
	as := g.AwayScore
	if hs == "null" || hs == "" {
		hs = "-"
	}
	if as == "null" || as == "" {
		as = "-"
	}

	fmt.Fprintf(w, "  %s %s %s%s\x1b[0m \x1b[33m-\x1b[0m %s%s\x1b[0m %s   \x1b[90m[%s]\x1b[0m\n",
		statusColor(status), home, sc, hs, sc, as, away, status)
}

func PrintGroups(client *Client, groups []Group) {
	FprintGroups(os.Stdout, client, groups)
}

func FprintGroups(w io.Writer, client *Client, groups []Group) {
	for _, grp := range groups {
		fmt.Fprintf(w, "\n  \x1b[1mGroup %s\x1b[0m\n", grp.Name)
		fmt.Fprintf(w, "  \x1b[90m%-24s %2s %2s %2s %2s  %2s  %2s  %3s  %2s\x1b[0m\n",
			"Team", "MP", "W", "D", "L", "GF", "GA", "GD", "Pts")
		for _, t := range grp.Teams {
			name := client.TeamName(t.TeamID)
			ptsColor := "\x1b[97m"
			if t.Pts == "0" {
				ptsColor = "\x1b[90m"
			}
			fmt.Fprintf(w, "  %s%-24s\x1b[0m %2s %2s %2s %2s  %2s  %2s  %3s  %s%2s\x1b[0m\n",
				"\x1b[97m", name, t.MP, t.W, t.D, t.L, t.GF, t.GA, t.GD, ptsColor, t.Pts)
		}
	}
}

func PrintSchedule(client *Client, games []Game) {
	FprintSchedule(os.Stdout, client, games)
}

func FprintSchedule(w io.Writer, client *Client, games []Game) {
	now := time.Now().Local()
	today := now.Format("01/02/2006")

	fmt.Fprintf(w, "\n  \x1b[1mUpcoming matches\x1b[0m\n")

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
		if shown >= 20 {
			fmt.Fprintf(w, "  \x1b[90m... and more\x1b[0m\n")
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

		dateDisplay := datePart
		if datePart == today {
			dateDisplay = "\x1b[33mTODAY\x1b[0m"
		} else {
			dateDisplay = fmt.Sprintf("\x1b[90m%s\x1b[0m", datePart)
		}

		fmt.Fprintf(w, "  %s  %-24s vs %-24s  [%s]\n",
			dateDisplay, home, away, g.Group)
		shown++
	}

	if shown == 0 {
		fmt.Fprintf(w, "  \x1b[90mNo upcoming matches found\x1b[0m\n")
	}
}

// ─── Openfootball display ───

func PrintOFHistory(h *OFTeamHistory) {
	FprintOFHistory(os.Stdout, h)
}

func FprintOFHistory(w io.Writer, h *OFTeamHistory) {
	fmt.Fprintf(w, "\n  \x1b[1m%s\x1b[0m\n", h.TeamName)
	fmt.Fprintf(w, "  \x1b[90m%d appearances  |  %d MP  %d W  %d D  %d L  |  GF %d  GA %d  |  Best: %s\x1b[0m\n\n",
		h.Appearances, h.TotalMP, h.TotalW, h.TotalD, h.TotalL, h.TotalGF, h.TotalGA, h.BestResult)

	for _, e := range h.Entries {
		yearStr := e.Year
		if e.Host {
			yearStr = fmt.Sprintf("\x1b[33m%s\x1b[0m \x1b[90m⭐\x1b[0m", e.Year)
		}
		fmt.Fprintf(w, "  %s \x1b[1m%s\x1b[0m\n", yearStr, e.Tournament)
		fmt.Fprintf(w, "  \x1b[90m    %d MP  %d W  %d D  %d L  GF %d  GA %d  — %s\x1b[0m\n",
			e.MatchesPlayed, e.Wins, e.Draws, e.Losses, e.GoalsFor, e.GoalsAgainst, e.Result)
	}
}

func PrintOFYear(t *OFTournament, year string) {
	FprintOFYear(os.Stdout, t, year)
}

func FprintOFYear(w io.Writer, t *OFTournament, year string) {
	fmt.Fprintf(w, "\n  \x1b[1m%s\x1b[0m\n\n", t.Name)

	byRound := make(map[string][]OFMatch)
	for _, m := range t.Matches {
		r := m.Round
		byRound[r] = append(byRound[r], m)
	}

	rounds := []string{"Group A", "Group B", "Group C", "Group D", "Group E", "Group F", "Group G", "Group H",
		"Matchday 1", "Matchday 2", "Matchday 3",
		"Round of 16", "Quarter-finals", "Semi-finals", "Match for third place", "Final"}
	shown := make(map[string]bool)

	for _, r := range rounds {
		matches, ok := byRound[r]
		if !ok {
			continue
		}
		shown[r] = true
		fmt.Fprintf(w, "  \x1b[1m%s\x1b[0m\n", r)
		for _, m := range matches {
			score := "- vs -"
			if m.Score != nil && len(m.Score.FT) == 2 {
				score = fmt.Sprintf("\x1b[97m%d\x1b[0m-\x1b[97m%d\x1b[0m", m.Score.FT[0], m.Score.FT[1])
			}
			dt := m.Date
			if m.Time != "" {
				dt = fmt.Sprintf("%s %s", m.Date, m.Time)
			}
			fmt.Fprintf(w, "    \x1b[90m%s\x1b[0m  %s  %s  %s\n",
				dt, padName(m.Team1, 28), score, padName(m.Team2, 28))
		}
		fmt.Fprintln(w)
	}

	for r, matches := range byRound {
		if shown[r] {
			continue
		}
		fmt.Fprintf(w, "  \x1b[1m%s\x1b[0m\n", r)
		for _, m := range matches {
			score := "- vs -"
			if m.Score != nil && len(m.Score.FT) == 2 {
				score = fmt.Sprintf("\x1b[97m%d\x1b[0m-\x1b[97m%d\x1b[0m", m.Score.FT[0], m.Score.FT[1])
			}
			fmt.Fprintf(w, "    \x1b[90m%s\x1b[0m  %s  %s  %s\n",
				m.Date, padName(m.Team1, 28), score, padName(m.Team2, 28))
		}
		fmt.Fprintln(w)
	}
}

func padName(s string, n int) string {
	if len(s) > n {
		return s[:n-1] + "…"
	}
	return s + strings.Repeat(" ", n-len(s))
}
