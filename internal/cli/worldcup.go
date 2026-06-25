//go:build worldcup

package cli

import (
	"fmt"
	"os"
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
	if len(extraArgs) > 0 {
		root = extraArgs[0]
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
		help()
		return nil
	}
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

func help() {
	fmt.Println("Usage: drupal-watcher worldcup <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  live        Today's matches with live scores")
	fmt.Println("  groups      Group standings")
	fmt.Println("  schedule    Upcoming fixtures")
	fmt.Println("  teams       All 48 teams")
	fmt.Println()
	fmt.Println("Requires DRUPAL_WATCHER_WORLDCUP=1 environment variable.")
	fmt.Fprintf(os.Stderr, "  \x1b[2mData source: worldcup26.ir (free, no API key)\x1b[0m\n")
}

func worldcupHelp() string {
	return `  worldcup    World Cup 2026 live scores and standings`
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
