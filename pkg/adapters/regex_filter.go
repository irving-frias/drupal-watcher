package adapters

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type PatternFilter struct {
	patterns []string
}

func NewPatternFilter(patterns []string) *PatternFilter {
	return &PatternFilter{patterns: patterns}
}

func (f *PatternFilter) ShouldProcess(event core.FileEvent) bool {
	if len(f.patterns) == 0 {
		return false
	}
	ext := filepath.Ext(event.Path)
	if ext == "" {
		return false
	}
	for _, p := range f.patterns {
		if strings.HasSuffix(event.Path, p) {
			return true
		}
		if ext == p {
			return true
		}
	}
	return false
}

type ExcludeFilter struct {
	re *regexp.Regexp
}

func NewExcludeFilter(excludes []string) *ExcludeFilter {
	if len(excludes) == 0 {
		return &ExcludeFilter{re: regexp.MustCompile(`^$`)}
	}
	parts := make([]string, 0, len(excludes))
	for _, e := range excludes {
		if e == "" {
			continue
		}
		parts = append(parts, regexp.QuoteMeta(e))
	}
	pattern := strings.Join(parts, "|")
	return &ExcludeFilter{re: regexp.MustCompile(pattern)}
}

func (f *ExcludeFilter) ShouldProcess(event core.FileEvent) bool {
	return !f.re.MatchString(event.Path)
}

type DotfileFilter struct{}

func NewDotfileFilter() *DotfileFilter {
	return &DotfileFilter{}
}

func (f *DotfileFilter) ShouldProcess(event core.FileEvent) bool {
	return !strings.Contains(event.Path, "/.")
}
