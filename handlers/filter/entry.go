package filter

import (
	"fmt"
	"log"
	"os"

	"github.com/darvaza-proxy/slog"
)

var (
	_ slog.Logger = (*LogEntry)(nil)
)

// LogEntry implements a level filtered logger
type LogEntry struct {
	logger *Logger

	level slog.LogLevel
	entry slog.Logger
}

// Enabled tells this logger would record logs
func (l *LogEntry) Enabled() bool {
	if l == nil || l.logger == nil {
		return false
	} else if l.level <= slog.UndefinedLevel || l.level > l.logger.Threshold {
		return false
	} else if l.entry != nil {
		return l.entry.Enabled()
	} else {
		return l.level == slog.Fatal
	}
}

// WithEnabled returns itself and if it's enabled
func (l *LogEntry) WithEnabled() (slog.Logger, bool) {
	return l, l.Enabled()
}

// Print would, if conditions are met, add a log entry with the arguments in the manner of fmt.Print
func (l *LogEntry) Print(args ...any) {
	if l.Enabled() {
		l.print(fmt.Sprint(args...))
	}
}

// Println would, if conditions are met, add a log entry with the arguments in the manner of fmt.Println
func (l *LogEntry) Println(args ...any) {
	if l.Enabled() {
		l.print(fmt.Sprintln(args...))
	}
}

// Printf would, if conditions are met, add a log entry with the arguments in the manner of fmt.Printf
func (l *LogEntry) Printf(format string, args ...any) {
	if l.Enabled() {
		l.print(fmt.Sprintf(format, args...))
	}
}

// print applies MessageFilter before sending the message to
// the parent Logger
func (l *LogEntry) print(msg string) {
	if fn := l.logger.MessageFilter; fn != nil {
		var ok bool

		msg, ok = fn(msg)
		if !ok {
			return
		}
	}

	if l.entry == nil {
		// parentless is either Fatal or Panic
		log.Output(3, msg)

		if l.level == slog.Fatal {
			os.Exit(1)
		} else {
			panic(msg)
		}

		// unreachable
	}

	l.entry.Print(msg)
}

// Debug creates a new filtered logger on level slog.Debug
func (l *LogEntry) Debug() slog.Logger {
	return l.logger.WithLevel(slog.Debug)
}

// Info creates a new filtered logger on level slog.Info
func (l *LogEntry) Info() slog.Logger {
	return l.logger.WithLevel(slog.Info)
}

// Warn creates a new filtered logger on level slog.Warn
func (l *LogEntry) Warn() slog.Logger {
	return l.logger.WithLevel(slog.Warn)
}

// Error creates a new filtered logger on level slog.Error
func (l *LogEntry) Error() slog.Logger {
	return l.logger.WithLevel(slog.Error)
}

// Fatal creates a new filtered logger on level slog.Fatal
func (l *LogEntry) Fatal() slog.Logger {
	return l.logger.WithLevel(slog.Fatal)
}

// Panic creates a new filtered logger on level slog.Panic
func (l *LogEntry) Panic() slog.Logger {
	return l.logger.WithLevel(slog.Panic)
}

// WithLevel creates a new filtered logger on the given level
func (l *LogEntry) WithLevel(level slog.LogLevel) slog.Logger {
	return l.logger.WithLevel(level)
}

// WithStack would, if conditions are met, attach a call stack to the log entry
func (l *LogEntry) WithStack(skip int) slog.Logger {
	if l.Enabled() && l.entry != nil {
		l.entry.WithStack(skip + 1)
	}
	return l
}

// WithField would, if conditions are met, attach a field to the log entry. This
// field could be altered if a FieldFilter is used
func (l *LogEntry) WithField(label string, value any) slog.Logger {
	if l.Enabled() {
		if l.entry == nil {
			// parentless doesn't support fields
		} else if fn := l.logger.FieldOverride; fn != nil {
			// intercepted
			fn(l.entry, label, value)
		} else if fn := l.logger.FieldsOverride; fn != nil {
			// intercepted
			fn(l.entry, slog.Fields{label: value})
		} else if fn := l.logger.FieldFilter; fn == nil {
			// as-is
			l.entry.WithField(label, value)
		} else if label, value, ok := fn(label, value); ok {
			// modified
			l.entry.WithField(label, value)
		}
	}
	return l
}

// WithFields would, if conditions are met, attach fields to the log entry.
// These fields could be altered if a FieldFilter is used
func (l *LogEntry) WithFields(fields map[string]any) slog.Logger {
	if count := len(fields); count == 0 {
		// skip empty
	} else if !l.Enabled() {
		// skip disabled
	} else if l.entry == nil {
		// parentless doesn't support fields
	} else if fn := l.logger.FieldsOverride; fn != nil {
		// intercepted
		fn(l.entry, fields)
	} else if fn := l.logger.FieldOverride; fn != nil {
		// intercepted
		for label, value := range fields {
			fn(l.entry, label, value)
		}
	} else if fn := l.logger.FieldFilter; fn == nil {
		// as-is
		l.entry.WithFields(fields)
	} else {
		// modified
		m := make(map[string]any, count)
		for k, v := range fields {
			if k, v, ok := fn(k, v); ok {
				m[k] = v
			}
		}
		if len(m) > 0 {
			l.entry.WithFields(m)
		}
	}
	return l
}
