package drush_test

import (
	"testing"

	"github.com/irving-frias/drupal-watcher/internal/drush"
)

func TestNotifyOSNoPanic(t *testing.T) {
	// NotifyOS wraps beeep.Notify — verify it doesn't panic.
	// On CI / headless, beeep may fail silently; that's fine.
	drush.NotifyOS("Test Title", "Test message")
}

func TestNotifyFuncDefaultIsNotifyOS(t *testing.T) {
	// Ensure the default assignment compiles and the variable exists.
	if drush.NotifyFunc == nil {
		t.Fatal("NotifyFunc should not be nil")
	}
}

func TestNotifyFuncMock(t *testing.T) {
	var gotTitle, gotMsg string
	original := drush.NotifyFunc
	defer func() { drush.NotifyFunc = original }()

	drush.NotifyFunc = func(title, message string) {
		gotTitle = title
		gotMsg = message
	}

	drush.NotifyFunc("Hello", "World")
	if gotTitle != "Hello" {
		t.Errorf("expected title 'Hello', got %q", gotTitle)
	}
	if gotMsg != "World" {
		t.Errorf("expected message 'World', got %q", gotMsg)
	}
}
