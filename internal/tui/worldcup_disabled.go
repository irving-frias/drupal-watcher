//go:build !worldcup

package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleWorldcup(subview string) tea.Cmd {
	m.pushEvent(eventLine{
		Timestamp: "worldcup",
		Content:   "World Cup feature not available. Build with: go build -tags worldcup ./cmd/drupal-watcher",
		Style:     warnStyle,
	})
	fmt.Println()
	fmt.Println("And enable with DRUPAL_WATCHER_WORLDCUP=1")
	return nil
}
