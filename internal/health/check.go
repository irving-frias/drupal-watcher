package health

import (
	"context"
	"os"
	"path/filepath"
	"time"
)

const healthFile = "health"
const defaultInterval = 30 * time.Second

func Run(ctx context.Context) {
	p := healthPath()
	os.MkdirAll(filepath.Dir(p), 0700)

	ticker := time.NewTicker(defaultInterval)
	defer ticker.Stop()

	write(p)
	for {
		select {
		case <-ctx.Done():
			os.Remove(p)
			return
		case <-ticker.C:
			write(p)
		}
	}
}

func write(p string) {
	os.WriteFile(p, []byte(time.Now().UTC().Format(time.RFC3339)+"\n"), 0644)
}

func healthPath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = "/tmp"
	}
	return filepath.Join(dir, "drupal-watcher", healthFile)
}
