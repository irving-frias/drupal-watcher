package tui

import "github.com/charmbracelet/lipgloss"

var (
	appStyle = lipgloss.NewStyle().
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1).
			Width(80)

	eventsStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("33")).
			Padding(0, 1).
			Width(80).
			Height(20)

	cmdStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("64")).
			Padding(0, 1).
			Width(80)

	green = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	red   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	blue  = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	dim   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).SetString(" ℹ ")
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString(" ✔ ")
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).SetString(" ⚠ ")
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).SetString(" ✖ ")
)
