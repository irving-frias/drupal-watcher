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

type statsMsg struct {
	PID        int
	Uptime     time.Duration
	Changes    int64
	Clears     int64
	WatchCount int64
	AllocMB    float64
}
