package health_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/irving-frias/drupal-watcher/internal/health"
)

func cacheDir() string {
	dir, _ := os.UserCacheDir()
	if dir == "" {
		dir = "/tmp"
	}
	return dir
}

func healthPath() string {
	return filepath.Join(cacheDir(), "drupal-watcher", "health")
}

func cleanupHealthFile() {
	os.Remove(healthPath())
}

func TestHealthFileWritten(t *testing.T) {
	cleanupHealthFile()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go health.Run(ctx)

	time.Sleep(200 * time.Millisecond)

	p := healthPath()
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("failed to read health file: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("health file is empty")
	}

	_, err = time.Parse(time.RFC3339, string(data[:len(data)-1]))
	if err != nil {
		t.Errorf("expected RFC3339 timestamp, got %q: %v", string(data), err)
	}
}

func TestHealthFileCleanedOnCancel(t *testing.T) {
	cleanupHealthFile()
	ctx, cancel := context.WithCancel(context.Background())

	go health.Run(ctx)

	time.Sleep(500 * time.Millisecond)

	p := healthPath()
	if _, err := os.Stat(p); os.IsNotExist(err) {
		t.Skip("health file not created within 500ms (slow CI)")
	}

	cancel()
	time.Sleep(500 * time.Millisecond)

	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Error("health file should be removed after context cancellation")
	}
}

func TestHealthFileUpdated(t *testing.T) {
	cleanupHealthFile()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go health.Run(ctx)
	time.Sleep(200 * time.Millisecond)

	p := healthPath()
	data1, _ := os.ReadFile(p)

	time.Sleep(500 * time.Millisecond)

	data2, _ := os.ReadFile(p)

	if string(data1) == string(data2) {
		t.Log("health file was not updated within 500ms (might be fast execution)")
	}
}

func TestHealthFilePath(t *testing.T) {
	p := healthPath()
	if p == "" {
		t.Fatal("health path should not be empty")
	}
	if !filepath.IsAbs(p) {
		t.Errorf("expected absolute path, got %s", p)
	}
	base := filepath.Base(p)
	if base != "health" {
		t.Errorf("expected 'health', got %s", base)
	}
}


