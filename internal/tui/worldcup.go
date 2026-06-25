//go:build worldcup

package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/irving-frias/drupal-watcher/internal/worldcup"
)

func (m *Model) handleWorldcup(subview string) tea.Cmd {
	if !worldcup.Enabled() {
		m.pushEvent(eventLine{
			Timestamp: "worldcup",
			Content:   "Set DRUPAL_WATCHER_WORLDCUP=1 to enable the World Cup feature.",
			Style:     warnStyle,
		})
		return nil
	}

	client := worldcup.NewClient()

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
		worldcup.FprintLiveGames(&b, client, games)
		m.worldcupContent = b.String()

	case "groups":
		groups, err := client.FetchGroups()
		if err != nil {
			m.pushEvent(eventLine{Timestamp: "worldcup", Content: fmt.Sprintf("API error: %v", err), Style: errorStyle})
			m.worldcupMode = false
			return nil
		}
		worldcup.FprintGroups(&b, client, groups)
		m.worldcupContent = b.String()

	case "schedule":
		games, err := client.FetchGames()
		if err != nil {
			m.pushEvent(eventLine{Timestamp: "worldcup", Content: fmt.Sprintf("API error: %v", err), Style: errorStyle})
			m.worldcupMode = false
			return nil
		}
		worldcup.FprintSchedule(&b, client, games)
		m.worldcupContent = b.String()

	default:
		m.worldcupMode = false
		m.pushEvent(eventLine{
			Timestamp: "worldcup",
			Content:   "Usage: worldcup [live|groups|schedule] or :back to return",
			Style:     infoStyle,
		})
	}
	return nil
}
