package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	logoPanelWidth  = 32
	logoFrameHeight = 11
)

// Static "Drupal Watcher" text in Mathematical Bold Fraktur (Pricedown Bold style).
var logoFrame = []string{
	"",
	"",
	"   · · · · · · · · · ·     ",
	"",
	"      𝕯𝖗𝖚𝖕𝖆𝖑 𝖂𝖆𝖙𝖈𝖍𝖊𝖗        ",
	"",
	"",
	"   · · · · · · · · · ·     ",
	"",
	"",
	"",
}

type DrupalLogo struct{}

func NewDrupalLogo() *DrupalLogo {
	return &DrupalLogo{}
}

func (d *DrupalLogo) Tick() {}

func (d *DrupalLogo) Render(width, height int) string {
	vPad := 0
	if height > logoFrameHeight {
		vPad = (height - logoFrameHeight) / 2
	}

	var lines []string
	for i := 0; i < height; i++ {
		var line string
		switch {
		case i < vPad:
			line = ""
		case i-vPad < logoFrameHeight:
			line = logoFrame[i-vPad]
		default:
			line = ""
		}
		lines = append(lines, lipgloss.PlaceHorizontal(width, lipgloss.Center, line))
	}

	return strings.Join(lines, "\n")
}
