//go:build !worldcup

package cli

import (
	"fmt"

	"github.com/irving-frias/drupal-watcher/internal/config"
)

func CmdWorldcup(subcommand string, flags map[string]interface{}, extraArgs []string, mgr *config.Manager) error {
	fmt.Println("World Cup feature not available. Build with:")
	fmt.Println("  go build -tags worldcup ./cmd/drupal-watcher")
	fmt.Println()
	fmt.Println("And enable with DRUPAL_WATCHER_WORLDCUP=1")
	return nil
}

func CmdHelpExtra() string {
	return ""
}
