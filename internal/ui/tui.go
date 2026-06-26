package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/irving-frias/drupal-watcher/pkg/core"
)

func Run(eventChan <-chan core.EngineEvent, info EngineInfo) error {
	m := NewModel(eventChan, info)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
