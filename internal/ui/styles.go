package ui

import "github.com/charmbracelet/lipgloss"

var (
	bold = lipgloss.NewStyle().Bold(true)

	statusStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	eventsStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("33")).
			Padding(0, 1)

	starStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("220")).
			Padding(0, 1)

	xdebugStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("214")).
			Padding(0, 1).
			Foreground(lipgloss.Color("196"))



	cmdStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("64")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
			Width(50)

	green  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	red    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	blue   = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	cyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
	dim    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).SetString(" ℹ ")
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString(" ✔ ")
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).SetString(" ⚠ ")
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).SetString(" ✖ ")

	comboNormal = lipgloss.Color("33")
	comboWarm   = lipgloss.Color("214")
	comboHot    = lipgloss.Color("208")
	comboPower  = lipgloss.Color("196")
)
