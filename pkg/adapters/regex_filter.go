package adapters

import (
	"path/filepath"
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
	excludes []string
}

func NewExcludeFilter(excludes []string) *ExcludeFilter {
	return &ExcludeFilter{excludes: excludes}
}

func (f *ExcludeFilter) ShouldProcess(event core.FileEvent) bool {
	for _, e := range f.excludes {
		if strings.Contains(event.Path, e) {
			return false
		}
	}
	return true
}

type DotfileFilter struct{}

func NewDotfileFilter() *DotfileFilter {
	return &DotfileFilter{}
}

func (f *DotfileFilter) ShouldProcess(event core.FileEvent) bool {
	return !strings.Contains(event.Path, "/.")
}
