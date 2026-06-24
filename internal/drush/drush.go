package drush

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/irving-frias/drupal-watcher/internal/utils"
)

type DrushResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

type DrushConfig interface {
	GetDrushCmd() *string
	GetDrushCommand() string
	GetDrushArgs() []string
	GetDrupalRoot() *string
}

func GetCmd(cfg DrushConfig) string {
	if d := cfg.GetDrushCmd(); d != nil && *d != "" {
		return *d
	}
	// Try to find drush in PATH first
	if path, err := exec.LookPath("drush"); err == nil {
		return path
	}
	// Try vendor/bin/drush relative to Drupal root
	drupalRoot := cfg.GetDrupalRoot()
	if drupalRoot != nil && *drupalRoot != "" {
		vendorDrush := filepath.Join(*drupalRoot, "..", "vendor", "bin", "drush")
		if _, err := os.Stat(vendorDrush); err == nil {
			return vendorDrush
		}
	}
	// Fallback
	return "drush"
}

func GetSpawnArgs(cfg DrushConfig) (string, []string) {
	cmd := GetCmd(cfg)
	drushCommand := cfg.GetDrushCommand()
	drushArgs := cfg.GetDrushArgs()
	args := []string{cmd}
	if drushCommand != "" {
		args = append(args, strings.Fields(drushCommand)...)
	}
	args = append(args, drushArgs...)
	return cmd, args
}

func HealthCheck(cfg DrushConfig) bool {
	cmd := GetCmd(cfg)
	statusArgs := []string{cmd, "--version"}
	start := time.Now()

	out, err := exec.Command(statusArgs[0], statusArgs[1:]...).CombinedOutput()
	elapsed := time.Since(start)
	utils.PrintDrushHealthResult(utils.DrushHealth{Ok: err == nil, Duration: elapsed, Output: string(out)})
	return err == nil
}

func Run(drushBase string, args []string) DrushResult {
	allArgs := strings.Fields(drushBase)
	if len(allArgs) == 0 {
		return DrushResult{ExitCode: 1, Stderr: "no drush command"}
	}
	allArgs = append(allArgs, args...)
	start := time.Now()

	cmd := exec.Command(allArgs[0], allArgs[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return DrushResult{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}
}

func RunPostClearCommands(commands []string) {
	for _, cmdStr := range commands {
		if cmdStr == "" {
			continue
		}
		fmt.Printf("%s Running post-clear command: %s\n", utils.Timestamp(), utils.Dim(cmdStr))

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", cmdStr)
		} else {
			cmd = exec.Command("sh", "-c", cmdStr)
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("%s Post-clear command failed: %v\n", utils.P_WARN, err)
		}
	}
}

func PromptConfirm(prompt string) bool {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
		return answer == "y" || answer == "yes"
	}
	return false
}
