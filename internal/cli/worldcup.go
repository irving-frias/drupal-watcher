//go:build worldcup

package cli

import (
	"fmt"
	"strings"

	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/internal/utils"
	"github.com/irving-frias/drupal-watcher/internal/worldcup"
)

func CmdWorldcup(subcommand string, flags map[string]interface{}, extraArgs []string, mgr *config.Manager) error {
	if !worldcup.Enabled() {
		fmt.Printf("%s World Cup feature is disabled. Set DRUPAL_WATCHER_WORLDCUP=1 to enable.\n", utils.P_WARN)
		return nil
	}

	root := ""
	if r, ok := flags["root"].(string); ok && r != "" {
		root = r
	}
	cfg, err := mgr.LoadConfig(root)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Worldcup != nil && !cfg.Worldcup.Enabled {
		fmt.Printf("%s World Cup feature is disabled in config (worldcup.enabled: false).\n", utils.P_WARN)
		return nil
	}

	client := worldcup.NewClient()

	switch subcommand {
	case "live", "":
		return cmdWorldcupLive(client)
	case "groups":
		return cmdWorldcupGroups(client)
	case "schedule", "fixtures":
		return cmdWorldcupSchedule(client)
	case "teams":
		return cmdWorldcupTeams(client)
	default:
		return cmdWorldcupHistory(subcommand, extraArgs)
	}
}

func cmdWorldcupHistory(subcommand string, extraArgs []string) error {
	of := worldcup.NewOFClient()

	switch subcommand {
	case "history":
		if len(extraArgs) == 0 {
			return fmt.Errorf("usage: worldcup history <team>")
		}
		team := extraArgs[0]
		h, err := of.TeamHistory(team)
		if err != nil {
			return err
		}
		worldcup.PrintOFHistory(h)
		return nil

	case "year":
		if len(extraArgs) == 0 {
			fmt.Println("Available years:")
			for _, y := range worldcup.KnownWorldCupYears {
				fmt.Printf("  %s  %s\n", y.Year, y.Name)
			}
			return nil
		}
		year := extraArgs[0]
		t, err := of.FetchTournament(year)
		if err != nil {
			return fmt.Errorf("fetch year %s: %w", year, err)
		}
		worldcup.PrintOFYear(t, year)
		return nil

	case "stats":
		if len(extraArgs) == 0 {
			return showAllTeamStats(of)
		}
		team := extraArgs[0]
		h, err := of.TeamHistory(team)
		if err != nil {
			return err
		}
		worldcup.PrintOFHistory(h)
		return nil

	case "teams-all":
		teams, err := of.AllTeams()
		if err != nil {
			return fmt.Errorf("fetch all teams: %w", err)
		}
		fmt.Println()
		for _, t := range teams {
			fmt.Printf("  %s\n", t)
		}
		fmt.Printf("\n  \x1b[90m%d teams\x1b[0m\n", len(teams))
		return nil

	default:
		return help()
	}
}

func showAllTeamStats(of *worldcup.OFClient) error {
	fmt.Print("\n  \x1b[1mTeam Stats (all-time)\x1b[0m\n")
	fmt.Println("  \x1b[90mUse 'worldcup stats <team>' for detailed history\x1b[0m")
	fmt.Println()

	teams, err := of.AllTeams()
	if err != nil {
		return err
	}

	type statLine struct {
		name       string
		apps       int
		mp, w, d, l int
		gf, ga     int
		best       string
	}

	var all []statLine
	for _, team := range teams {
		h, err := of.TeamHistory(team)
		if err != nil {
			continue
		}
		all = append(all, statLine{
			name: h.TeamName, apps: h.Appearances,
			mp: h.TotalMP, w: h.TotalW, d: h.TotalD, l: h.TotalL,
			gf: h.TotalGF, ga: h.TotalGA, best: h.BestResult,
		})
	}

	for _, s := range all {
		fmt.Printf("  %-24s %2d apps  %3d MP  %2d W  %2d D  %2d L  GF %3d  GA %3d  %s\n",
			s.name, s.apps, s.mp, s.w, s.d, s.l, s.gf, s.ga, s.best)
	}
	return nil
}

func cmdWorldcupLive(client *worldcup.Client) error {
	if err := client.EnsureTeams(); err != nil {
		return fmt.Errorf("fetch teams: %w", err)
	}
	games, err := client.FetchGames()
	if err != nil {
		return fmt.Errorf("fetch games: %w", err)
	}
	worldcup.PrintLiveGames(client, games)
	return nil
}

func cmdWorldcupGroups(client *worldcup.Client) error {
	if err := client.EnsureTeams(); err != nil {
		return fmt.Errorf("fetch teams: %w", err)
	}
	groups, err := client.FetchGroups()
	if err != nil {
		return fmt.Errorf("fetch groups: %w", err)
	}
	worldcup.PrintGroups(client, groups)
	return nil
}

func cmdWorldcupSchedule(client *worldcup.Client) error {
	if err := client.EnsureTeams(); err != nil {
		return fmt.Errorf("fetch teams: %w", err)
	}
	games, err := client.FetchGames()
	if err != nil {
		return fmt.Errorf("fetch games: %w", err)
	}
	worldcup.PrintSchedule(client, games)
	return nil
}

func cmdWorldcupTeams(client *worldcup.Client) error {
	teams, err := client.FetchTeams()
	if err != nil {
		return fmt.Errorf("fetch teams: %w", err)
	}
	fmt.Println()
	for _, t := range teams {
		fmt.Printf("  %-24s [%s]  Group %s\n", t.NameEN, t.FifaCode, t.Group)
	}
	return nil
}

func help() error {
	fmt.Println("Usage: drupal-watcher worldcup <command> [args]")
	fmt.Println()
	fmt.Println("Live 2026 (worldcup26.ir):")
	fmt.Println("  live                Today's matches with live scores")
	fmt.Println("  groups              Group standings")
	fmt.Println("  schedule            Upcoming fixtures")
	fmt.Println("  teams               All 48 teams")
	fmt.Println()
	fmt.Println("Historical (openfootball):")
	fmt.Println("  history <team>      World Cup history for a country")
	fmt.Println("  year [year]         Tournament details for a year (lists years if omitted)")
	fmt.Println("  stats [team]        All-time stats (or detailed for a team)")
	fmt.Println("  teams-all           All teams that have played in any World Cup")
	fmt.Println()
	fmt.Println("Requires DRUPAL_WATCHER_WORLDCUP=1 environment variable.")
	return nil
}

func worldcupHelp() string {
	return `  worldcup    World Cup live scores, standings, and history`
}

var extraHelp []func() string

func init() {
	extraHelp = append(extraHelp, worldcupHelp)
}

func CmdHelpExtra() string {
	if len(extraHelp) == 0 {
		return ""
	}
	var b strings.Builder
	for _, fn := range extraHelp {
		b.WriteString(fn())
		b.WriteString("\n")
	}
	return b.String()
}
