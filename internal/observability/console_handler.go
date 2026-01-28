package observability

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

// ANSI color codes
const (
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BrightBlack = "\033[90m" // Gray
)

// ConsoleHandler is a custom slog.Handler that outputs colored logs for development
type ConsoleHandler struct {
	opts   slog.HandlerOptions
	w      io.Writer
	mu     *sync.Mutex
	attrs  []slog.Attr
	groups []string
}

// NewConsoleHandler creates a new ConsoleHandler
func NewConsoleHandler(w io.Writer, opts *slog.HandlerOptions) *ConsoleHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &ConsoleHandler{
		opts:  *opts,
		w:     w,
		mu:    &sync.Mutex{},
		attrs: []slog.Attr{},
	}
}

// Enabled returns true if the level is enabled
func (h *ConsoleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

// Handle handles the record and outputs a formatted log line
func (h *ConsoleHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Get color for level
	levelColor := Reset
	levelName := r.Level.String()

	switch r.Level {
	case slog.LevelDebug:
		levelColor = Magenta
		levelName = "DEBUG"
	case slog.LevelInfo:
		levelColor = Green
		levelName = "INFO "
	case slog.LevelWarn:
		levelColor = Yellow
		levelName = "WARN "
	case slog.LevelError:
		levelColor = Red
		levelName = "ERROR"
	}

	// Format: TIME [LEVEL] MSG
	// Time format: 15:04:05.000 (HH:MM:SS.ms)
	timeStr := r.Time.Format("15:04:05.000")

	// Print base message
	fmt.Fprintf(h.w, "%s%s%s %s[%s]%s %s",
		BrightBlack, timeStr, Reset,
		levelColor, levelName, Reset,
		r.Message,
	)

	// Print stored attributes (context)
	for _, a := range h.attrs {
		h.printAttr(a)
	}

	// Print record attributes
	r.Attrs(func(a slog.Attr) bool {
		h.printAttr(a)
		return true
	})

	// Print newline
	fmt.Fprintln(h.w)

	return nil
}

func (h *ConsoleHandler) printAttr(a slog.Attr) {
	// Skip empty attributes
	if a.Equal(slog.Attr{}) {
		return
	}

	value := a.Value.String()

	// Special handling for common keys
	keyColor := Cyan
	if a.Key == "error" || a.Key == "err" {
		keyColor = Red
	} else if a.Key == "method" {
		keyColor = Yellow
	} else if a.Key == "path" {
		keyColor = Blue
	}

	fmt.Fprintf(h.w, " %s%s=%s\"%s\"", keyColor, a.Key, Reset, value)
}

// WithAttrs returns a new handler with the given attributes
func (h *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	newAttrs := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	newAttrs = append(newAttrs, h.attrs...)
	newAttrs = append(newAttrs, attrs...)

	return &ConsoleHandler{
		opts:   h.opts,
		w:      h.w,
		mu:     h.mu, // Share mutex for writing to same output
		attrs:  newAttrs,
		groups: h.groups,
	}
}

// WithGroup returns a new handler with the given group
func (h *ConsoleHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	newGroups := make([]string, 0, len(h.groups)+1)
	newGroups = append(newGroups, h.groups...)
	newGroups = append(newGroups, name)

	return &ConsoleHandler{
		opts:   h.opts,
		w:      h.w,
		mu:     h.mu,
		attrs:  h.attrs,
		groups: newGroups,
	}
}
