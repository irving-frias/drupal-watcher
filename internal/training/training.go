package training

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"sync"

	"github.com/irving-frias/drupal-watcher/internal/metrics"
)

type Suggestion struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
	Condition   string `json:"condition,omitempty"`
}

var lock sync.RWMutex
var suggestions []Suggestion

func Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			suggestions = defaultSuggestions()
			return saveDefaults(path)
		}
		return err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&suggestions); err != nil {
		suggestions = defaultSuggestions()
		return nil
	}
	return nil
}

func Get() []Suggestion {
	lock.RLock()
	defer lock.RUnlock()
	if len(suggestions) == 0 {
		return defaultSuggestions()
	}
	return suggestions
}

func Random() *Suggestion {
	snapshot := metrics.Snapshot()
	all := Get()

	var filtered []Suggestion
	for _, s := range all {
		if matchCondition(s, snapshot) {
			filtered = append(filtered, s)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return &filtered[rand.Intn(len(filtered))]
}

func matchCondition(s Suggestion, snap metrics.SnapshotData) bool {
	switch s.Condition {
	case "high_errors":
		return snap.Errors > 5
	case "many_changes":
		return snap.TotalChanges > 100
	case "many_clears":
		return snap.TotalClears > 50
	case "high_memory":
		return false
	default:
		return true
	}
}

func defaultSuggestions() []Suggestion {
	return []Suggestion{
		{
			Title:       "Check Drush status",
			Description: "Run drush status to verify Drupal is healthy",
			Command:     "drush status",
		},
		{
			Title:       "Clear specific cache",
			Description: "Use 'cc render' or 'cc plugin' instead of full 'cr' for faster clears",
			Condition:   "many_clears",
		},
		{
			Title:       "Check error logs",
			Description: "High error count — check watchdog and PHP error logs",
			Condition:   "high_errors",
		},
		{
			Title:       "Review recent changes",
			Description: "Many file changes detected — review your work with git diff",
			Condition:   "many_changes",
		},
		{
			Title:       "YAML lint",
			Description: "Run yamllint on your .yml files for syntax issues",
			Command:     "yamllint .",
		},
		{
			Title:       "PHP syntax check",
			Description: "Run php -l on all .php files to check syntax",
			Command:     "find . -name '*.php' -exec php -l {} \\;",
		},
		{
			Title:       "Optimize CSS/JS",
			Description: "Aggregate CSS/JS with drush cc css-js for frontend changes",
			Command:     "drush cc css-js",
		},
		{
			Title:       "Rebuild cache",
			Description: "Full cache rebuild with drush cr",
			Command:     "drush cr",
			Condition:   "many_clears",
		},
	}
}

func saveDefaults(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(suggestions)
}
