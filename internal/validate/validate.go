package validate

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/irving-frias/drupal-watcher/internal/config"
)

type Result struct {
	Pass    bool
	Entries []Entry
}

type Entry struct {
	Field   string
	Message string
	OK      bool
}

func Validate(root string) Result {
	res := Result{Pass: true}
	mgr := config.NewManager()

	cfg, err := mgr.LoadConfig(root)
	if err != nil {
		res.Entries = append(res.Entries, Entry{Field: "config", Message: fmt.Sprintf("Failed to load config: %v", err)})
		res.Pass = false
		return res
	}

	res.Entries = append(res.Entries, Entry{Field: "config", Message: "Config loaded successfully", OK: true})

	// Validate YAML config
	yp := filepath.Join(root, "configs", "config.yaml")
	if _, err := os.Stat(yp); err == nil {
		res.Entries = append(res.Entries, Entry{Field: "config.yaml", Message: "Found", OK: true})
	} else {
		res.Entries = append(res.Entries, Entry{Field: "config.yaml", Message: "Not found (using JSON legacy)", OK: true})
	}

	// Validate Drupal root
	dr := root
	if cfg.DrupalRoot != nil && *cfg.DrupalRoot != "" {
		dr = filepath.Join(root, *cfg.DrupalRoot)
	}
	if info, err := os.Stat(dr); err == nil && info.IsDir() {
		res.Entries = append(res.Entries, Entry{Field: "drupalRoot", Message: fmt.Sprintf("Found: %s", dr), OK: true})
	} else {
		res.Entries = append(res.Entries, Entry{Field: "drupalRoot", Message: fmt.Sprintf("Not found: %s", dr)})
		res.Pass = false
	}

	// Validate routes
	for i, r := range cfg.Routes {
		full := filepath.Join(root, r)
		if info, err := os.Stat(full); err == nil && info.IsDir() {
			res.Entries = append(res.Entries, Entry{Field: fmt.Sprintf("routes[%d]", i), Message: fmt.Sprintf("Found: %s", r), OK: true})
		} else {
			res.Entries = append(res.Entries, Entry{Field: fmt.Sprintf("routes[%d]", i), Message: fmt.Sprintf("Not found: %s", r)})
			res.Pass = false
		}
	}

	// Validate patterns
	if len(cfg.Patterns) > 0 {
		res.Entries = append(res.Entries, Entry{Field: "patterns", Message: fmt.Sprintf("%d patterns configured", len(cfg.Patterns)), OK: true})
	}

	// Validate drush
	cmd := exec.Command("drush", "--version")
	drushOut, drushErr := cmd.Output()
	if drushErr == nil {
		res.Entries = append(res.Entries, Entry{Field: "drush", Message: fmt.Sprintf("Found: drush (%s)", strings.TrimSpace(string(drushOut))), OK: true})
	} else {
		res.Entries = append(res.Entries, Entry{Field: "drush", Message: fmt.Sprintf("Not found or not working: %v", drushErr)})
		res.Pass = false
	}

	// Validate PHPCS
	if cfg.PhpCsStandard != "" {
		phpcsPath := findPHPCS(root)
		if phpcsPath != "" {
			res.Entries = append(res.Entries, Entry{Field: "phpcs", Message: fmt.Sprintf("Found: %s", phpcsPath), OK: true})
		} else {
			res.Entries = append(res.Entries, Entry{Field: "phpcs", Message: "phpCsStandard configured but phpcs not found"})
			res.Pass = false
		}
	}

	// Validate sites
	for _, name := range cfg.Sites {
		entries := validateSite(root, name)
		res.Entries = append(res.Entries, entries...)
	}

	// Validate CommandsPerPattern
	for pattern, cmd := range cfg.CommandsPerPattern {
		if ok := validateCommand(cmd); !ok {
			res.Entries = append(res.Entries, Entry{Field: fmt.Sprintf("commandsPerPattern[%s]", pattern), Message: fmt.Sprintf("Unknown command: %s", cmd)})
			res.Pass = false
		} else {
			res.Entries = append(res.Entries, Entry{Field: fmt.Sprintf("commandsPerPattern[%s]", pattern), Message: cmd, OK: true})
		}
	}

	if res.Pass {
		res.Entries = append(res.Entries, Entry{Field: "summary", Message: "All checks passed", OK: true})
	} else {
		res.Entries = append(res.Entries, Entry{Field: "summary", Message: "Some checks failed"})
	}

	return res
}

var knownCommands = map[string]bool{
	"cr": true, "cache:rebuild": true,
	"cc": true, "cache:clear": true,
	"cc render": true, "cc theme-registry": true,
	"cc plugin": true, "cc css-js": true,
	"cc router": true, "cc entity": true,
}

func validateCommand(cmd string) bool {
	if knownCommands[cmd] {
		return true
	}
	return strings.HasPrefix(cmd, "php ") || cmd == "php -l" || cmd == "yaml"
}

func findPHPCS(root string) string {
	candidates := []string{
		filepath.Join(root, "vendor", "bin", "phpcs"),
		filepath.Join(root, "..", "vendor", "bin", "phpcs"),
	}
	for _, c := range candidates {
		if p, _ := exec.LookPath(c); p != "" {
			return p
		}
	}
	if p, _ := exec.LookPath("phpcs"); p != "" {
		return p
	}
	return ""
}

func validateSite(root, name string) []Entry {
	var entries []Entry
	sitePath := filepath.Join(root, "sites", name)
	if info, err := os.Stat(sitePath); err == nil && info.IsDir() {
		settings := filepath.Join(sitePath, "settings.php")
		if _, err := os.Stat(settings); err == nil {
			entries = append(entries, Entry{Field: fmt.Sprintf("sites/%s", name), Message: "settings.php found", OK: true})
		} else {
			entries = append(entries, Entry{Field: fmt.Sprintf("sites/%s", name), Message: "settings.php not found"})
		}
	} else {
		entries = append(entries, Entry{Field: fmt.Sprintf("sites/%s", name), Message: "Directory not found"})
	}
	return entries
}

// ErrValidation is returned by RunValidate on failure.
var ErrValidation = errors.New("validation failed")
