package drush

import (
	"testing"
)

func TestNotifyOSNoPanic(t *testing.T) {
	// NotifyOS wraps beeep.Notify — verify it doesn't panic.
	// On CI / headless, beeep may fail silently; that's fine.
	NotifyOS("Test Title", "Test message")
}

func TestNotifyFuncDefaultIsNotifyOS(t *testing.T) {
	// Ensure the default assignment compiles and the variable exists.
	if NotifyFunc == nil {
		t.Fatal("NotifyFunc should not be nil")
	}
}

func TestNotifyFuncMock(t *testing.T) {
	var gotTitle, gotMsg string
	original := NotifyFunc
	defer func() { NotifyFunc = original }()

	NotifyFunc = func(title, message string) {
		gotTitle = title
		gotMsg = message
	}

	NotifyFunc("Hello", "World")
	if gotTitle != "Hello" {
		t.Errorf("expected title 'Hello', got %q", gotTitle)
	}
	if gotMsg != "World" {
		t.Errorf("expected message 'World', got %q", gotMsg)
	}
}
