package drush

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Site struct {
	Name string
	URI  string
}

type siteAlias struct {
	Root *string `yaml:"root,omitempty"`
	URI  string  `yaml:"uri"`
}

func LoadSitesYml(drupalRoot string) (map[string]Site, error) {
	candidates := []string{
		filepath.Join(drupalRoot, "drush", "sites.yml"),
		filepath.Join(drupalRoot, "drush", "sites", "sites.yml"),
	}

	var data []byte
	var path string
	for _, c := range candidates {
		b, err := os.ReadFile(c)
		if err == nil {
			data = b
			path = c
			break
		}
	}
	if data == nil {
		return nil, fmt.Errorf("drush/sites.yml not found in %s", drupalRoot)
	}

	var raw map[string]siteAlias
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	sites := make(map[string]Site, len(raw))
	for name, alias := range raw {
		if alias.URI == "" {
			continue
		}
		sites[name] = Site{Name: name, URI: alias.URI}
	}

	if len(sites) == 0 {
		return nil, fmt.Errorf("no valid sites found in %s", path)
	}

	return sites, nil
}

func HasMultiSite(drupalRoot string) bool {
	sitesDir := filepath.Join(drupalRoot, "sites")
	entries, err := os.ReadDir(sitesDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "default" {
			continue
		}
		if _, err := os.Stat(filepath.Join(sitesDir, e.Name(), "settings.php")); err == nil {
			return true
		}
	}
	return false
}

func FilterSites(all map[string]Site, include, exclude []string) map[string]Site {
	if len(include) > 0 {
		filtered := make(map[string]Site, len(include))
		for _, name := range include {
			if s, ok := all[name]; ok {
				filtered[name] = s
			}
		}
		return filtered
	}

	if len(exclude) > 0 {
		excludeSet := make(map[string]bool, len(exclude))
		for _, name := range exclude {
			excludeSet[name] = true
		}
		filtered := make(map[string]Site, len(all))
		for name, s := range all {
			if !excludeSet[name] {
				filtered[name] = s
			}
		}
		return filtered
	}

	return all
}

func PrintSiteList(sites map[string]Site) string {
	var names []string
	for name := range sites {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

// SiteDrushConfig wraps a DrushConfig to inject --uri for a specific site.
type SiteDrushConfig struct {
	DrushConfig
	Name string
	URI  string
}

func (c *SiteDrushConfig) GetDrushArgs() []string {
	return append([]string{"--uri=" + c.URI}, c.DrushConfig.GetDrushArgs()...)
}
