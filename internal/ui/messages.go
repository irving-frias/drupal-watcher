package ui

import (
	"time"

	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type tickMsg time.Time

type engineEventMsg struct {
	Event core.EngineEvent
}
