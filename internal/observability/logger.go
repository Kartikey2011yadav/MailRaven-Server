package observability

import (
	"log/slog"
	"os"
)

// Logger wraps the structured logger for MailRaven
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new structured logger based on configuration
func NewLogger(level, format string) *Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// WithSMTPSession creates a logger with SMTP session context
func (l *Logger) WithSMTPSession(sessionID, remoteIP string) *Logger {
	return &Logger{
		Logger: l.With(
			slog.String("session_id", sessionID),
			slog.String("remote_ip", remoteIP),
			slog.String("component", "smtp"),
		),
	}
}

// WithAPI creates a logger with API request context
func (l *Logger) WithAPI(requestID, method, path string) *Logger {
	return &Logger{
		Logger: l.With(
			slog.String("request_id", requestID),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("component", "api"),
		),
	}
}

// WithStorage creates a logger with storage operation context
func (l *Logger) WithStorage(operation, resource string) *Logger {
	return &Logger{
		Logger: l.With(
			slog.String("operation", operation),
			slog.String("resource", resource),
			slog.String("component", "storage"),
		),
	}
}
