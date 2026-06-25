package tui

import (
	"time"

	"github.com/irving-frias/drupal-watcher/internal/watcher"
)

type tickMsg time.Time

type watcherEventMsg struct {
	Event watcher.EventMsg
}

type errMsg struct {
	Err error
}


