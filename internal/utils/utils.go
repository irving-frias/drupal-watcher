package utils

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

var colorsEnabled = true

func SetColorsEnabled(v bool) { colorsEnabled = v }
func ColorsEnabled() bool     { return colorsEnabled }

const (
	RED    = "\x1b[31m"
	GREEN  = "\x1b[32m"
	YELLOW = "\x1b[33m"
	BLUE   = "\x1b[34m"
	CYAN   = "\x1b[36m"
	NC     = "\x1b[0m"
	BOLD   = "\x1b[1m"
	DIM    = "\x1b[2m"
)

func c(code, s string) string {
	if colorsEnabled {
		return code + s + NC
	}
	return s
}

func Red(s string) string    { return c(RED, s) }
func Green(s string) string  { return c(GREEN, s) }
func Yellow(s string) string { return c(YELLOW, s) }
func Blue(s string) string   { return c(BLUE, s) }
func Cyan(s string) string   { return c(CYAN, s) }
func Bold(s string) string {
	if colorsEnabled {
		return BOLD + s + NC
	}
	return s
}
func Dim(s string) string {
	if colorsEnabled {
		return DIM + s + NC
	}
	return s
}

var (
	P_ERROR   = "✖"
	P_WARN    = "⚠"
	P_INFO    = "ℹ"
	P_SUCCESS = "✔"
)

func P(colored string) string {
	return c(strings.Replace(colored, "\x1b", "", -1), colored)
}

func init() {
	P_ERROR = c("\x1b[31m", "✖")
	P_WARN = c("\x1b[33m", "⚠")
	P_INFO = c("\x1b[34m", "ℹ")
	P_SUCCESS = c("\x1b[32m", "✔")
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
	fmt.Println(Yellow(title))
}

type SectionItem interface{}

func PrintSection(heading string, items []SectionItem) {
	fmt.Printf("\n%s:\n", Blue(heading))
	for _, item := range items {
		switch v := item.(type) {
		case [2]string:
			fmt.Printf("  %s  %s\n", Green(v[0]), v[1])
		case string:
			fmt.Printf("  %s\n", v)
		default:
			fmt.Printf("  %v\n", v)
		}
	}
}

type DrushHealth struct {
	Ok       bool
	Duration time.Duration
	Output   string
}

func PrintDrushHealthResult(h DrushHealth) {
	status := P_SUCCESS
	if !h.Ok {
		status = P_ERROR
	}
	fmt.Printf("%s Drush health check: %s (%v)\n", Timestamp(), status, h.Duration)
	if !h.Ok && h.Output != "" {
		fmt.Printf("  %s\n", Dim(strings.TrimSpace(h.Output)))
	}
}

type MemStats struct {
	AllocMB    float64
	WatchCount int
}

func GetMemStats(watchCount int) MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return MemStats{
		AllocMB:    float64(m.Alloc) / 1024 / 1024,
		WatchCount: watchCount,
	}
}

func PrintMemStats(s MemStats) {
	memColor := GREEN
	if s.AllocMB >= 500 {
		memColor = RED
	} else if s.AllocMB >= 100 {
		memColor = YELLOW
	}
	fmt.Printf("  Memory: %s  |  Kernel watches: %d\n",
		c(memColor, fmt.Sprintf("%.1f MB", s.AllocMB)), s.WatchCount)
}
