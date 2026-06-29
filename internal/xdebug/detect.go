package xdebug

import (
	"os/exec"
	"strings"
)

func Detect() bool {
	out, err := exec.Command("php", "-m").Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.EqualFold(strings.TrimSpace(line), "xdebug") {
			return true
		}
	}
	return false
}

func Disable() error {
	if err := exec.Command("ddev", "xdebug", "off").Run(); err != nil {
		if err := exec.Command("phpenmod", "-v", "all", "-s", "all", "-x", "xdebug").Run(); err != nil {
			return exec.Command("bash", "-c", "echo '; disabled by drupal-watcher' > /dev/null; php -r 'ini_set(\"xdebug.mode\", \"off\");'").Run()
		}
	}
	return nil
}
