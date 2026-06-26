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

func loadSitesFile(path string) (map[string]Site, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
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
	return sites, nil
}

func loadSitesDir(sitesDir string) (map[string]Site, error) {
	entries, err := os.ReadDir(sitesDir)
	if err != nil {
		return nil, err
	}
	sites := make(map[string]Site)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".site.yml") {
			continue
		}
		fileName := strings.TrimSuffix(e.Name(), ".site.yml")
		path := filepath.Join(sitesDir, e.Name())
		fileSites, err := loadSitesFile(path)
		if err != nil {
			return nil, err
		}

		// Prefer key matching filename; otherwise use filename as site name
		// and take URI from the first entry that has one (group/environment format).
		if s, ok := fileSites[fileName]; ok {
			sites[fileName] = s
		} else {
			for _, s := range fileSites {
				if s.URI != "" {
					sites[fileName] = Site{Name: fileName, URI: s.URI}
					break
				}
			}
		}
	}
	if len(sites) == 0 {
		return nil, fmt.Errorf("no *.site.yml files found in %s", sitesDir)
	}
	return sites, nil
}

func LoadSitesYml(drupalRoot, projectRoot string) (map[string]Site, error) {
	// Search paths: project root first (common for Composer Drupal), then drupal root
	bases := []string{projectRoot, drupalRoot}
	if projectRoot == drupalRoot {
		bases = []string{drupalRoot}
	}

	// Try single drush/sites.yml
	for _, base := range bases {
		path := filepath.Join(base, "drush", "sites.yml")
		sites, err := loadSitesFile(path)
		if err == nil {
			return sites, nil
		}
	}

	// Try drush/sites/*.site.yml
	for _, base := range bases {
		dir := filepath.Join(base, "drush", "sites")
		sites, err := loadSitesDir(dir)
		if err == nil {
			return sites, nil
		}
	}

	return nil, fmt.Errorf("no site aliases found in %s, %s, or drush/sites/*.site.yml under those directories", filepath.Join(projectRoot, "drush", "sites.yml"), filepath.Join(drupalRoot, "drush", "sites.yml"))
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
