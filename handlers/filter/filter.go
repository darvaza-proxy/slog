// Package filter is a Logger that only allows entries of a given level
package filter

import (
	"darvaza.org/slog"
	"darvaza.org/slog/internal"
)

var (
	_ slog.Logger = (*Logger)(nil)
)

// Logger implements a factory for level filtered loggers
type Logger struct {
	root internal.Loglet

	// Parent is the Logger to used as backend when conditions are met
	Parent slog.Logger

	// Threshold is the minimum level to be logged
	Threshold slog.LogLevel

	// FieldFilter allows us to modify single fields before passing them
	// to the Parent logger
	FieldFilter func(key string, val any) (string, any, bool)

	// FieldsFilter allows us to modify field maps before passing them
	// to the Parent logger. Returns filtered fields and whether to continue.
	FieldsFilter func(fields slog.Fields) (slog.Fields, bool)

	// MessageFilter allows us to modify Print() messages before passing
	// them to the Parent logger, on completely discard the entry
	MessageFilter func(msg string) (string, bool)
}

// Enabled tells this logger doesn't log anything, but WithLevel() might
func (*Logger) Enabled() bool {
	return false
}

// isEnabled tells if we should create a LogEntry or not
func (l *Logger) isEnabled() bool {
	switch {
	case l == nil, l.Threshold == slog.UndefinedLevel:
		// uninitialised
		return false
	default:
		// collect fields just in case.
		return true
	}
}

// WithEnabled tells this logger doesn't log anything, but WithLevel() might
func (l *Logger) WithEnabled() (slog.Logger, bool) {
	return l, false
}

// Print does nothing
func (*Logger) Print(...any) {}

// Println does nothing
func (*Logger) Println(...any) {}

// Printf does nothing
func (*Logger) Printf(string, ...any) {}

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

// Panic returns a filtered logger on level slog.Panic
func (l *Logger) Panic() slog.Logger { return l.WithLevel(slog.Panic) }

// WithLevel returns a filtered logger set to the given level
func (l *Logger) WithLevel(level slog.LogLevel) slog.Logger {
	if l.isEnabled() {
		return doWithLevel(l, &l.root, level)
	}
	return l
}

// WithStack creates a LogEntry with stack information
func (l *Logger) WithStack(skip int) slog.Logger {
	if l.isEnabled() {
		return doWithStack(l, &l.root, skip+1)
	}
	return l
}

// WithField creates a LogEntry with the field
func (l *Logger) WithField(label string, value any) slog.Logger {
	if label != "" && l.isEnabled() {
		if out := doWithField(l, &l.root, label, value); out != nil {
			return out
		}
	}
	return l
}

// WithFields creates a LogEntry with the fields
func (l *Logger) WithFields(fields map[string]any) slog.Logger {
	if internal.HasFields(fields) && l.isEnabled() {
		if out := doWithFields(l, &l.root, fields); out != nil {
			return out
		}
	}
	return l
}

// New creates a new filtered log factory at a given level. Logger can be manually
// initialised as well. Defaults filter entries at level slog.Error or higher
// Parentless is treated as `noop`, with Fatal implemented like log.Fatal
func New(parent slog.Logger, threshold slog.LogLevel) slog.Logger {
	switch {
	case parent == nil:
		threshold = slog.Fatal
	case threshold <= slog.UndefinedLevel:
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
