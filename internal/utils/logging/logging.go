package logging

import (
	"log/slog"
	"os"
)

// SetupCLILogger configures slog for clean CLI output
// Removes timestamps and log levels for a better user experience
func SetupCLILogger() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Remove time and level from output for cleaner CLI messages
			if a.Key == slog.TimeKey || a.Key == slog.LevelKey {
				return slog.Attr{}
			}
			return a
		},
	}))
	slog.SetDefault(logger)

}
