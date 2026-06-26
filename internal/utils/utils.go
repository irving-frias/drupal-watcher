package utils

import (
	"fmt"
	"runtime"
	"time"

	"github.com/pterm/pterm"
)

func SetColorsEnabled(v bool) {
	if v {
		pterm.EnableColor()
	} else {
		pterm.DisableColor()
	}
}

func ColorsEnabled() bool { return pterm.PrintColor }

func Red(s string) string    { return pterm.FgRed.Sprint(s) }
func Green(s string) string  { return pterm.FgGreen.Sprint(s) }
func Yellow(s string) string { return pterm.FgYellow.Sprint(s) }
func Blue(s string) string   { return pterm.FgBlue.Sprint(s) }
func Cyan(s string) string   { return pterm.FgCyan.Sprint(s) }
func Bold(s string) string   { return pterm.Bold.Sprint(s) }

func Dim(s string) string {
	return pterm.FgGray.Sprint(s)
}

func Timestamp() string {
	now := time.Now()
	return Cyan(fmt.Sprintf("[%02d:%02d:%02d]", now.Hour(), now.Minute(), now.Second()))
}

var PossibleDocroots = []string{"docroot", "web", "html", "public", "drupal"}
var ExcludedDirs = []string{"node_modules", ".git", "files"}

var DefaultPatterns = []string{
	".html.twig", ".twig", ".inc", ".yml", ".module", ".theme",
	".php", ".info.yml", ".services.yml",
	".routing.yml", ".permissions.yml", ".links.menu.yml",
	".css", ".js",
}

func PrintHeader(title string) {
	pterm.DefaultSection.Println(title)
}

type SectionItem interface{}

func PrintSection(heading string, items []SectionItem) {
	pterm.DefaultSection.Println(heading)
	var data pterm.TableData
	for _, item := range items {
		switch v := item.(type) {
		case [2]string:
			data = append(data, []string{v[0], v[1]})
		case string:
			data = append(data, []string{"", v})
		default:
			data = append(data, []string{"", fmt.Sprintf("%v", v)})
		}
	}
	if len(data) > 0 {
		pterm.DefaultTable.WithHasHeader().WithData(data).Render()
	}
}

type MemStats struct {
	AllocMB    float64
	WatchCount int64
}

func FormatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func GetMemStats(watchCount int64) MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return MemStats{
		AllocMB:    float64(m.Alloc) / 1024 / 1024,
		WatchCount: watchCount,
	}
}

func PrintMemStats(s MemStats) {
	memColor := pterm.FgGreen
	if s.AllocMB >= 500 {
		memColor = pterm.FgRed
	} else if s.AllocMB >= 100 {
		memColor = pterm.FgYellow
	}
	pterm.Info.Printfln("Memory: %s  |  Kernel watches: %d",
		memColor.Sprintf("%.1f MB", s.AllocMB), s.WatchCount)
}
