package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

const (
	colorRed    = "\033[31m"
	colorOrange = "\033[33m"
	colorReset  = "\033[0m"
)

// cliHandler is a custom slog handler for clean CLI output
type cliHandler struct {
	w io.Writer
}

func (h *cliHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= slog.LevelInfo
}

func (h *cliHandler) Handle(ctx context.Context, r slog.Record) error {
	// Apply colors based on log level
	switch {
	case r.Level >= slog.LevelError:
		fmt.Fprint(h.w, colorRed)
	case r.Level >= slog.LevelWarn:
		fmt.Fprint(h.w, colorOrange)
	}

	// Print message without "msg=" prefix
	fmt.Fprint(h.w, r.Message)

	// Print attributes as key=value without quotes
	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(h.w, " %s=%v", a.Key, a.Value.Any())
		return true
	})

	// Reset color
	if r.Level >= slog.LevelWarn {
		fmt.Fprint(h.w, colorReset)
	}

	fmt.Fprintln(h.w)
	return nil
}

func (h *cliHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *cliHandler) WithGroup(name string) slog.Handler {
	return h
}

// SetupCLILogger configures slog for clean CLI output
// Removes timestamps and log levels for a better user experience
func SetupCLILogger() {
	logger := slog.New(&cliHandler{w: os.Stderr})
	slog.SetDefault(logger)
}
