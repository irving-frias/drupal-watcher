//go:build worldcup

package worldcup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const ofBaseURL = "https://raw.githubusercontent.com/openfootball/worldcup.json/master"

type OFClient struct {
	http  *http.Client
	mu    sync.RWMutex
	cache map[string]*cacheEntry
}

func NewOFClient() *OFClient {
	return &OFClient{
		http:  &http.Client{Timeout: 15 * time.Second},
		cache: make(map[string]*cacheEntry),
	}
}

var teamHosts = map[string][]string{
	"Uruguay":               {"1930"},
	"Italy":                 {"1934", "1990"},
	"France":                {"1938", "1998"},
	"Brazil":                {"1950", "2014"},
	"Switzerland":           {"1954"},
	"Sweden":                {"1958"},
	"Chile":                 {"1962"},
	"England":               {"1966"},
	"Mexico":                {"1970", "1986"},
	"West Germany":          {"1974"},
	"Germany":               {"2006"},
	"Argentina":             {"1978"},
	"Spain":                 {"1982"},
	"USA":                   {"1994"},
	"South Korea":           {"2002"},
	"Japan":                 {"2002"},
	"South Africa":          {"2010"},
	"Russia":                {"2018"},
	"Qatar":                 {"2022"},
	"United States":         {"2026"},
	"Canada":                {"2026"},
}

func (c *OFClient) ofGet(path string, dst interface{}) error {
	c.mu.RLock()
	entry, ok := c.cache[path]
	c.mu.RUnlock()
	if ok && time.Now().Before(entry.expiresAt) {
		return json.Unmarshal(entry.data.([]byte), dst)
	}

	url := fmt.Sprintf("%s/%s/worldcup.json", ofBaseURL, path)
	resp, err := c.http.Get(url)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch %s returned %s", url, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	c.mu.Lock()
	c.cache[path] = &cacheEntry{data: body, expiresAt: time.Now().Add(30 * time.Minute)}
	c.mu.Unlock()

	return json.Unmarshal(body, dst)
}

func (c *OFClient) FetchTournament(year string) (*OFTournament, error) {
	var t OFTournament
	if err := c.ofGet(year, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func roundResult(round string) string {
	r := strings.ToLower(round)
	switch {
	case strings.Contains(r, "final") && !strings.Contains(r, "semi") && !strings.Contains(r, "quarter") && !strings.Contains(r, "third"):
		return "Final"
	case strings.Contains(r, "semi"):
		return "Semi-finals"
	case strings.Contains(r, "quarter"):
		return "Quarter-finals"
	case strings.Contains(r, "round of 16"):
		return "Round of 16"
	case strings.Contains(r, "group") || strings.HasPrefix(r, "matchday") || strings.HasPrefix(r, "group"):
		return "Group stage"
	default:
		return round
	}
}

func (c *OFClient) isHost(team, year string) bool {
	for _, y := range teamHosts[team] {
		if y == year {
			return true
		}
	}
	return false
}

func scoreWinner(score *OFScore, isTeam1 bool) *bool {
	if score == nil {
		return nil
	}
	// Check ET first (accounts for extra time wins)
	if len(score.ET) == 2 && score.ET[0] != score.ET[1] {
		won := (isTeam1 && score.ET[0] > score.ET[1]) || (!isTeam1 && score.ET[1] > score.ET[0])
		return &won
	}
	// Check FT
	if len(score.FT) == 2 && score.FT[0] != score.FT[1] {
		won := (isTeam1 && score.FT[0] > score.FT[1]) || (!isTeam1 && score.FT[1] > score.FT[0])
		return &won
	}
	// Penalty shootout
	if len(score.P) == 2 && score.P[0] != score.P[1] {
		won := (isTeam1 && score.P[0] > score.P[1]) || (!isTeam1 && score.P[1] > score.P[0])
		return &won
	}
	return nil
}

func effectiveGoals(score *OFScore, isTeam1 bool) (gf, ga int) {
	if score == nil {
		return 0, 0
	}
	// ET is cumulative (includes FT goals) — use it when available
	if len(score.ET) == 2 {
		if isTeam1 {
			return score.ET[0], score.ET[1]
		}
		return score.ET[1], score.ET[0]
	}
	if len(score.FT) == 2 {
		if isTeam1 {
			return score.FT[0], score.FT[1]
		}
		return score.FT[1], score.FT[0]
	}
	return 0, 0
}

func (c *OFClient) computeResultForTeam(team string, matches []OFMatch, year string) string {
	best := "Group stage"
	var playedFinal bool
	var wonFinal bool

	for _, m := range matches {
		matchRound := roundResult(m.Round)
		isTeam1 := strings.EqualFold(m.Team1, team)
		isTeam2 := strings.EqualFold(m.Team2, team)
		if !isTeam1 && !isTeam2 {
			continue
		}
		if m.Score == nil {
			continue
		}

		switch matchRound {
		case "Final":
			playedFinal = true
			w := scoreWinner(m.Score, isTeam1)
			if w != nil && *w {
				wonFinal = true
			}
		case "Semi-finals":
			best = "Semi-finals"
		case "Quarter-finals":
			if best != "Semi-finals" && best != "Final" {
				best = "Quarter-finals"
			}
		case "Round of 16":
			if best != "Semi-finals" && best != "Quarter-finals" && best != "Final" {
				best = "Round of 16"
			}
		}
	}

	if playedFinal {
		if wonFinal {
			return "Champion"
		}
		return "Runner-up"
	}
	return best
}

func (c *OFClient) TeamHistory(team string) (*OFTeamHistory, error) {
	history := &OFTeamHistory{
		TeamName: team,
	}

	for _, y := range KnownWorldCupYears {
		t, err := c.FetchTournament(y.Year)
		if err != nil {
			continue
		}

		var matches []OFMatch
		for _, m := range t.Matches {
			isTeam1 := strings.EqualFold(m.Team1, team)
			isTeam2 := strings.EqualFold(m.Team2, team)
			if !isTeam1 && !isTeam2 {
				continue
			}
			if m.Score == nil || len(m.Score.FT) < 2 {
				continue
			}
			matches = append(matches, m)
		}
		if len(matches) == 0 {
			continue
		}

		entry := OFTeamHistoryEntry{
			Year:       y.Year,
			Tournament: y.Name,
			Host:       c.isHost(team, y.Year),
		}

		for _, m := range matches {
			isTeam1 := strings.EqualFold(m.Team1, team)
			gf, ga := effectiveGoals(m.Score, isTeam1)
			w := scoreWinner(m.Score, isTeam1)

			entry.MatchesPlayed++
			entry.GoalsFor += gf
			entry.GoalsAgainst += ga
			if w != nil && *w {
				entry.Wins++
			} else if w != nil && !*w {
				entry.Losses++
			} else {
				entry.Draws++
			}
		}

		entry.Result = c.computeResultForTeam(team, t.Matches, y.Year)

		history.Entries = append(history.Entries, entry)
		history.Appearances++
		history.TotalMP += entry.MatchesPlayed
		history.TotalW += entry.Wins
		history.TotalD += entry.Draws
		history.TotalL += entry.Losses
		history.TotalGF += entry.GoalsFor
		history.TotalGA += entry.GoalsAgainst

		if entry.Result == "Champion" {
			history.BestResult = "Champion"
		} else if entry.Result == "Runner-up" && history.BestResult != "Champion" {
			history.BestResult = "Runner-up"
		} else if entry.Result == "Semi-finals" && history.BestResult != "Champion" && history.BestResult != "Runner-up" {
			history.BestResult = "Semi-finals"
		} else if (entry.Result == "Quarter-finals" || entry.Result == "Round of 16" || entry.Result == "Group stage") && history.BestResult == "" {
			if entry.Result == "Quarter-finals" {
				history.BestResult = "Quarter-finals"
			} else if entry.Result == "Round of 16" && history.BestResult != "Quarter-finals" {
				history.BestResult = "Round of 16"
			} else if entry.Result == "Group stage" && history.BestResult == "" {
				history.BestResult = "Group stage"
			}
		}
	}

	if history.Appearances == 0 {
		return nil, fmt.Errorf("no World Cup history found for '%s'", team)
	}
	return history, nil
}

func (c *OFClient) AllTeams() ([]string, error) {
	seen := make(map[string]bool)
	for _, y := range KnownWorldCupYears {
		t, err := c.FetchTournament(y.Year)
		if err != nil {
			continue
		}
		for _, m := range t.Matches {
			n1 := strings.TrimSpace(m.Team1)
			n2 := strings.TrimSpace(m.Team2)
			if n1 != "" && !seen[n1] {
				seen[n1] = true
			}
			if n2 != "" && !seen[n2] {
				seen[n2] = true
			}
		}
	}

	var teams []string
	for name := range seen {
		teams = append(teams, name)
	}
	return teams, nil
}
