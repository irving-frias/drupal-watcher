package drush

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/irving-frias/drupal-watcher/internal/utils"
	"github.com/pterm/pterm"
)

var isWSL = sync.OnceValue(func() bool {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return false
	}
	lower := strings.ToLower(string(data))
	return strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl")
})

var (
	cachedCmd string
	cmdMu     sync.RWMutex
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
	GetNotify() bool
}

var NotifyFunc = NotifyOS

func GetCmd(cfg DrushConfig) string {
	if d := cfg.GetDrushCmd(); d != nil && *d != "" {
		return *d
	}

	cmdMu.RLock()
	if cachedCmd != "" {
		cmdMu.RUnlock()
		return cachedCmd
	}
	cmdMu.RUnlock()

	cmdMu.Lock()
	defer cmdMu.Unlock()

	if cachedCmd != "" {
		return cachedCmd
	}

	// Try to find drush in PATH first
	if path, err := exec.LookPath("drush"); err == nil {
		cachedCmd = path
		return path
	}
	// Try vendor/bin/drush relative to Drupal root
	drupalRoot := cfg.GetDrupalRoot()
	if drupalRoot != nil && *drupalRoot != "" {
		vendorDrush := filepath.Join(*drupalRoot, "..", "vendor", "bin", "drush")
		if _, err := os.Stat(vendorDrush); err == nil {
			cachedCmd = vendorDrush
			return vendorDrush
		}
	}
	// Fallback
	cachedCmd = "drush"
	return "drush"
}

func ResetCmdCache() {
	cmdMu.Lock()
	defer cmdMu.Unlock()
	cachedCmd = ""
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

	_, err := exec.Command(statusArgs[0], statusArgs[1:]...).CombinedOutput()
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

// RunCacheClears executes multiple cache clear commands in a single drush invocation.
// It batches compatible "cc <type>" commands into one call (e.g. "cc render,plugin,css-js").
// If any command is "cr", it runs "cr" once (covers everything).
func RunCacheClears(cfg DrushConfig, commands []string) DrushResult {
	if len(commands) == 0 {
		return DrushResult{ExitCode: 0}
	}

	drushBase := GetCmd(cfg)
	var hasCR bool
	var ccTypes []string
	for _, cmd := range commands {
		switch {
		case cmd == "cr" || cmd == "cache:rebuild":
			hasCR = true
		case strings.HasPrefix(cmd, "cc ") || strings.HasPrefix(cmd, "cache:clear "):
			parts := strings.Fields(cmd)
			if len(parts) >= 2 {
				for _, t := range parts[1:] {
					if t != "" {
						ccTypes = append(ccTypes, t)
					}
				}
			}
		default:
			// Unknown command, fall back to cr to be safe
			hasCR = true
		}
	}

	var result DrushResult
	drushArgs := cfg.GetDrushArgs()
	if hasCR {
		result = Run(drushBase, append(drushArgs, "cr"))
	} else if len(ccTypes) > 0 {
		ccArgs := append(drushArgs, "cc", strings.Join(ccTypes, ","))
		result = Run(drushBase, ccArgs)
	} else {
		return DrushResult{ExitCode: 0}
	}

	if cfg != nil && cfg.GetNotify() && result.ExitCode == 0 {
		cmdStr := "cr"
		if !hasCR {
			cmdStr = "cc " + strings.Join(ccTypes, ",")
		}
		NotifyFunc("Drupal Watcher", "drush "+cmdStr)
	}

	return result
}

func RunPostClearCommands(commands []string) {
	for _, cmdStr := range commands {
		if cmdStr == "" {
			continue
		}
		pterm.Info.Printfln("Running post-clear command: %s", utils.Dim(cmdStr))

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", cmdStr)
		} else {
			cmd = exec.Command("sh", "-c", cmdStr)
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			pterm.Warning.Printfln("Post-clear command failed: %v", err)
		}
	}
}

func NotifyOS(title, message string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("osascript", "-e",
			fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)).Start()

	case "linux":
		if isWSL() {
			notifyWindowsToast(title, message)
		} else {
			exec.Command("notify-send", title, message).Start()
		}

	case "windows":
		notifyWindowsToast(title, message)
	}
}

func notifyWindowsToast(title, message string) {
	ps := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
$tmpl = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$text = $tmpl.GetElementsByTagName("text")
$text.Item(0).AppendChild($tmpl.CreateTextNode("%s")) | Out-Null
$text.Item(1).AppendChild($tmpl.CreateTextNode("%s")) | Out-Null
$toast = [Windows.UI.Notifications.ToastNotification]::new($tmpl)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("Drupal Watcher").Show($toast)`, title, message)

	cmd := "powershell.exe"
	if runtime.GOOS == "windows" {
		cmd = "powershell"
	}
	exec.Command(cmd, "-NoProfile", "-Command", ps).Start()
}

func PromptConfirm(prompt string) bool {
	result, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).WithDefaultText(prompt).Show()
	return result
}
