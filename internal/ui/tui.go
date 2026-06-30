package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/irving-frias/drupal-watcher/pkg/core"
)

func Run(eventChan <-chan core.EngineEvent, info EngineInfo, root string, gifPath string) error {
	m := NewModel(eventChan, info, root, gifPath)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

func RunContext(ctx context.Context, eventChan <-chan core.EngineEvent, info EngineInfo, root string, gifPath string) error {
	m := NewModel(eventChan, info, root, gifPath)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	go func() {
		<-ctx.Done()
		p.Quit()
	}()

	_, err := p.Run()
	return err
}
