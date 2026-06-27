package common

import (
	"encoding/json"
	"os"
	"sync"
)

var (
	versionOnce sync.Once
	version     string
)

func PkgVersion() string {
	versionOnce.Do(func() {
		composer, err := os.ReadFile("composer.json")
		if err != nil {
			version = "0.1.0"
			return
		}
		var meta struct {
			Extra struct {
				Version string `json:"drupal-watcher-version"`
			} `json:"extra"`
		}
		if err := json.Unmarshal(composer, &meta); err != nil {
			version = "0.1.0"
			return
		}
		if meta.Extra.Version != "" {
			version = meta.Extra.Version
		} else {
			version = "0.1.0"
		}
	})
	return version
}
