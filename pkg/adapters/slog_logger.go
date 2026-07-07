package adapters

import (
	"io"
	"log/slog"
)

func NewDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}
