// Package filtered is a Logger that only allows entries of a given level
package filtered

import (
	"errors"

	"github.com/darvaza-proxy/slog"
)

var (
	_ slog.Logger = (*Logger)(nil)
)

// Logger implements a factory for level filtered loggers
type Logger struct {
	// Parent is the Logger to used as backend when conditions are met
	Parent slog.Logger

	// Threshold is the minimum level to be logged
	Threshold slog.LogLevel

	// FieldFilter allows us to modify filters before passing them
	// to the Parent logger
	FieldFilter func(key string, val any) (string, any, bool)

	// MessageFilter allows us to modify Print() messages before passing
	// them to the Parent logger, on completely discard the entry
	MessageFilter func(msg string) (string, bool)
}

// Enabled tells this logger doesn't log anything, but WithLevel() might
func (l *Logger) Enabled() bool {
	return false
}

// WithEnabled tells this logger doesn't log anything, but WithLevel() might
func (l *Logger) WithEnabled() (slog.Logger, bool) {
	return l, false
}

// Print does nothing
func (l *Logger) Print(args ...any) {}

// Println does nothing
func (l *Logger) Println(args ...any) {}

// Printf does nothing
func (l *Logger) Printf(format string, args ...any) {}

// Debug returns a filtered logger on level slog.Debug
func (l *Logger) Debug() slog.Logger { return l.WithLevel(slog.Debug) }

// Info returns a filtered logger on level slog.Info
func (l *Logger) Info() slog.Logger { return l.WithLevel(slog.Info) }

// Warn returns a filtered logger on level slog.Warn
func (l *Logger) Warn() slog.Logger { return l.WithLevel(slog.Warn) }

// Error returns a filtered logger on level slog.Error
func (l *Logger) Error() slog.Logger { return l.WithLevel(slog.Error) }

// Fatal returns a filtered logger on level slog.Fatal
func (l *Logger) Fatal() slog.Logger { return l.WithLevel(slog.Fatal) }

// WithLevel returns a filtered logger set to the given level
func (l *Logger) WithLevel(level slog.LogLevel) slog.Logger {
	return &Loglet{
		logger: l,
		level:  level,
		entry:  l.Parent.WithLevel(level),
	}
}

// WithStack does nothing
func (l *Logger) WithStack(skip int) slog.Logger { return l }

// WithField does nothing
func (l *Logger) WithField(label string, value any) slog.Logger { return l }

// WithFields does nothing
func (l *Logger) WithFields(fields map[string]any) slog.Logger { return l }

// New creates a new filtered log factory at a given level. Logger can be manually
// initialised as well. Defaults filter entries at level slog.Error or higher
func New(parent slog.Logger, threshold slog.LogLevel) slog.Logger {
	if parent == nil {
		panic(errors.New("parent logger missing"))
	}
	if threshold <= slog.UndefinedLevel {
		threshold = slog.Error
	}
	return &Logger{
		Parent:    parent,
		Threshold: threshold,
	}
}
