package filtered

import (
	"fmt"

	"github.com/darvaza-proxy/slog"
)

var (
	_ slog.Logger = (*Loglet)(nil)
)

// Loglet implements a level filtered logger
type Loglet struct {
	logger *Logger

	level slog.LogLevel
	entry slog.Logger
}

// Enabled tells this logger would record logs
func (l *Loglet) Enabled() bool {
	if l == nil || l.logger == nil || l.entry == nil {
		return false
	} else if l.level <= slog.UndefinedLevel || l.level > l.logger.Threshold {
		return false
	} else {
		return l.entry.Enabled()
	}
}

// WithEnabled returns itself and if it's enabled
func (l *Loglet) WithEnabled() (slog.Logger, bool) {
	return l, l.Enabled()
}

// Print would, if conditions are met, add a log entry with the arguments in the manner of fmt.Print
func (l *Loglet) Print(args ...any) {
	if l.Enabled() {
		l.print(fmt.Sprint(args...))
	}
}

// Println would, if conditions are met, add a log entry with the arguments in the manner of fmt.Println
func (l *Loglet) Println(args ...any) {
	if l.Enabled() {
		l.print(fmt.Sprintln(args...))
	}
}

// Printf would, if conditions are met, add a log entry with the arguments in the manner of fmt.Printf
func (l *Loglet) Printf(format string, args ...any) {
	if l.Enabled() {
		l.print(fmt.Sprintf(format, args...))
	}
}

// print applies MessageFilter before sending the message to
// the parent Logger
func (l *Loglet) print(msg string) {
	if fn := l.logger.MessageFilter; fn != nil {
		var ok bool

		msg, ok = fn(msg)
		if !ok {
			return
		}
	}
	l.entry.Print(msg)
}

// Debug creates a new filtered logger on level slog.Debug
func (l *Loglet) Debug() slog.Logger {
	return l.logger.WithLevel(slog.Debug)
}

// Info creates a new filtered logger on level slog.Info
func (l *Loglet) Info() slog.Logger {
	return l.logger.WithLevel(slog.Info)
}

// Warn creates a new filtered logger on level slog.Warn
func (l *Loglet) Warn() slog.Logger {
	return l.logger.WithLevel(slog.Warn)
}

// Error creates a new filtered logger on level slog.Error
func (l *Loglet) Error() slog.Logger {
	return l.logger.WithLevel(slog.Error)
}

// Fatal creates a new filtered logger on level slog.Fatal
func (l *Loglet) Fatal() slog.Logger {
	return l.logger.WithLevel(slog.Fatal)
}

// WithLevel creates a new filtered logger on the given level
func (l *Loglet) WithLevel(level slog.LogLevel) slog.Logger {
	return l.logger.WithLevel(level)
}

// WithStack would, if conditions are met, attach a call stack to the log entry
func (l *Loglet) WithStack(skip int) slog.Logger {
	if l.Enabled() {
		l.entry.WithStack(skip + 1)
	}
	return l
}

// WithField would, if conditions are met, attach a field to the log entry. This
// field could be altered if a FieldFilter is used
func (l *Loglet) WithField(label string, value any) slog.Logger {
	if l.Enabled() {
		if fn := l.logger.FieldOverride; fn != nil {
			// intercepted
			fn(l.entry, label, value)
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
func (l *Loglet) WithFields(fields map[string]any) slog.Logger {
	if count := len(fields); count == 0 {
		// skip empty
	} else if !l.Enabled() {
		// skip disabled
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
