// Package filter is a Logger that only allows entries of a given level
package filter

import (
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

	// FieldOverride intercepts calls to WithField() on enabled loggers
	// to let you transform the field
	FieldOverride func(entry slog.Logger, key string, val any)

	// FieldsOverride intercepts calls to WithFields() on enabled loggers
	// to let you transform the fields
	FieldsOverride func(entry slog.Logger, fields map[string]any)

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
	var entry slog.Logger

	if l.Parent != nil {
		entry = l.Parent.WithLevel(level)
	} else if level != slog.Fatal {
		// Parentless non-Fatal, NOOP
		return l
	}

	return &LogEntry{
		logger: l,
		level:  level,
		entry:  entry,
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
// Parentless is treated as `noop`, with Fatal implemented like log.Fatal
func New(parent slog.Logger, threshold slog.LogLevel) slog.Logger {
	if parent == nil {
		threshold = slog.Fatal
	} else if threshold <= slog.UndefinedLevel {
		threshold = slog.Error
	}
	return &Logger{
		Parent:    parent,
		Threshold: threshold,
	}
}

// NewNoop creates a new filtered log factory that only implements Fatal().Print()
func NewNoop() slog.Logger {
	return New(nil, slog.Fatal)
}
