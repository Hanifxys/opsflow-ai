// Package logging provides a pre-configured structured JSON logger using slog.
package logging

import (
	"log/slog"
	"os"
	"strings"
)

// New creates a structured JSON logger at the given level.
// Accepts level strings: "debug", "info", "warn", "error".
func New(level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     lvl,
		AddSource: lvl == slog.LevelDebug,
	})

	return slog.New(handler)
}
