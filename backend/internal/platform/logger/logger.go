// Package logger builds the logger of the application.
package logger

import (
	"io"
	"log/slog"
)

// New builds a logger that writes JSON lines.
func New(level slog.Level, w io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}))
}
