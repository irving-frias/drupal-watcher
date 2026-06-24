package watcher

import (
	"testing"
)

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		patterns []string
		want     bool
	}{
		{".php match", "/path/to/file.php", []string{".php"}, true},
		{".module match", "/path/to/file.module", []string{".module", ".php"}, true},
		{"no match", "/path/to/file.txt", []string{".php"}, false},
		{"double extension", "/path/to/file.html.twig", []string{".twig"}, true},
		{"no extension", "/path/to/README", []string{".php"}, false},
		{"empty patterns", "/path/to/file.php", []string{}, false},
		{"yml vs info.yml", "/path/to/file.info.yml", []string{".info.yml", ".yml"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchPattern(tt.file, tt.patterns)
			if got != tt.want {
				t.Errorf("matchPattern(%q, %v) = %v, want %v", tt.file, tt.patterns, got, tt.want)
			}
		})
	}
}

func TestMatchExclude(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		excludes []string
		want     bool
	}{
		{"excluded by dir", "/path/contrib/file.php", []string{"contrib"}, true},
		{"excluded by vendor", "/path/vendor/file.php", []string{"vendor"}, true},
		{"not excluded", "/path/custom/file.php", []string{"contrib", "vendor"}, false},
		{"excluded by dot", "/path/.hidden/file.php", []string{"/."}, true},
		{"empty excludes", "/path/file.php", []string{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchExclude(tt.file, tt.excludes)
			if got != tt.want {
				t.Errorf("matchExclude(%q, %v) = %v, want %v", tt.file, tt.excludes, got, tt.want)
			}
		})
	}
}

func TestGetCacheClearArgs(t *testing.T) {
	cmds := map[string]string{
		".php":             "cc plugin",
		".info.yml":        "cr",
		".html.twig":       "cc render",
		".css":             "cc css-js",
	}
	tests := []struct {
		name string
		file string
		want []string
	}{
		{".php file", "/path/file.php", []string{"cc", "plugin"}},
		{".info.yml file", "/path/file.info.yml", []string{"cr"}},
		{".html.twig file", "/path/file.html.twig", []string{"cc", "render"}},
		{".css file", "/path/file.css", []string{"cc", "css-js"}},
		{"unknown pattern falls to cr", "/path/file.txt", []string{"cr"}},
		{"longest pattern wins", "/path/file.info.yml", []string{"cr"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCacheClearArgs(tt.file, cmds)
			if len(got) != len(tt.want) {
				t.Fatalf("getCacheClearArgs(%q) = %v, want %v", tt.file, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("getCacheClearArgs(%q)[%d] = %q, want %q", tt.file, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestIsSkippedDir(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		skip []string
		want bool
	}{
		{"node_modules skipped", "node_modules", []string{"node_modules", ".git"}, true},
		{".git skipped", ".git", []string{"node_modules", ".git"}, true},
		{"custom not skipped", "custom", []string{"node_modules", ".git"}, false},
		{"empty skip list", "anything", []string{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSkippedDir(tt.dir, tt.skip)
			if got != tt.want {
				t.Errorf("isSkippedDir(%q, %v) = %v, want %v", tt.dir, tt.skip, got, tt.want)
			}
		})
	}
}
