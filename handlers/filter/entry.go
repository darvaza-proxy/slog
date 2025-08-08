package filter

import (
	"fmt"
	"log"
	"os"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/internal"
)

var (
	_ slog.Logger = (*LogEntry)(nil)
)

// LogEntry implements a level filtered logger
type LogEntry struct {
	entry  internal.Loglet
	config *Logger
}

// Enabled tells this logger would record logs
func (l *LogEntry) Enabled() bool {
	switch {
	case l == nil:
		return false
	case l.isEnabled():
		return true
	case l.config != nil && l.config.Parent == nil:
		// Parentless Fatal/Panic are enabled for termination (but won't use fields)
		level := l.entry.Level()
		return level <= slog.Fatal
	default:
		return false
	}
}

// isEnabled checks both filter threshold and parent enablement at current level
func (l *LogEntry) isEnabled() bool {
	if l == nil || l.config == nil {
		// uninitialised
		return false
	}

	level := l.entry.Level()
	switch {
	case level <= slog.UndefinedLevel || level > l.config.Threshold:
		// locally disabled
		return false
	case l.config.Parent == nil:
		// parentless
		return false
	default:
		// Check if parent would be enabled at the current level
		parentLogger := l.config.Parent.WithLevel(level)
		return parentLogger.Enabled()
	}
}

// WithEnabled returns itself and if it's enabled
func (l *LogEntry) WithEnabled() (slog.Logger, bool) {
	return l, l.Enabled()
}

// Print would, if conditions are met, add a log entry with the arguments
// in the manner of fmt.Print
func (l *LogEntry) Print(args ...any) {
	if l.Enabled() {
		l.msg(1, fmt.Sprint(args...))
	}
}

// Println would, if conditions are met, add a log entry with the arguments
// in the manner of fmt.Println
func (l *LogEntry) Println(args ...any) {
	if l.Enabled() {
		l.msg(1, fmt.Sprintln(args...))
	}
}

// Printf would, if conditions are met, add a log entry with the arguments
// in the manner of fmt.Printf
func (l *LogEntry) Printf(format string, args ...any) {
	if l.Enabled() {
		l.msg(1, fmt.Sprintf(format, args...))
	}
}

// msg applies MessageFilter before sending the message to
// the parent Logger
func (l *LogEntry) msg(skip int, msg string) {
	msg, ok := applyMessageFilter(l.config, msg)
	if !ok {
		return
	}

	ll, ok := l.makeLogger(skip + 1)
	if !ok {
		// parentless is either Fatal or Panic
		_ = log.Output(3, msg)

		level := l.entry.Level()
		if level != slog.Fatal {
			panic(core.NewPanicError(skip+1, msg))
		}

		// revive:disable:deep-exit
		os.Exit(1)
	}

	ll.Print(msg)
}

func (l *LogEntry) makeLogger(skip int) (slog.Logger, bool) {
	switch {
	case l == nil, l.config == nil, l.config.Parent == nil:
		return nil, false
	}

	// Collect fields and delegate to parent
	fields := l.fieldsMap()
	ll := l.config.Parent.WithLevel(l.entry.Level())
	if l.entry.CallStack() != nil {
		ll = ll.WithStack(skip + 1)
	}
	if internal.HasFields(fields) {
		ll = ll.WithFields(fields)
	}
	return ll, true
}

// fieldsMap collects all fields from the entry Loglet chain
func (l *LogEntry) fieldsMap() map[string]any {
	fields := make(map[string]any)
	iter := l.entry.Fields()
	for iter.Next() {
		k, v := iter.Field()
		fields[k] = v
	}
	return fields
}

// Debug creates a new filtered logger on level slog.Debug
func (l *LogEntry) Debug() slog.Logger { return l.WithLevel(slog.Debug) }

// Info creates a new filtered logger on level slog.Info
func (l *LogEntry) Info() slog.Logger { return l.WithLevel(slog.Info) }

// Warn creates a new filtered logger on level slog.Warn
func (l *LogEntry) Warn() slog.Logger { return l.WithLevel(slog.Warn) }

// Error creates a new filtered logger on level slog.Error
func (l *LogEntry) Error() slog.Logger { return l.WithLevel(slog.Error) }

// Fatal creates a new filtered logger on level slog.Fatal
func (l *LogEntry) Fatal() slog.Logger { return l.WithLevel(slog.Fatal) }

// Panic creates a new filtered logger on level slog.Panic
func (l *LogEntry) Panic() slog.Logger { return l.WithLevel(slog.Panic) }

// WithLevel creates a new filtered logger on the given level
func (l *LogEntry) WithLevel(level slog.LogLevel) slog.Logger {
	if l.isEnabled() {
		return doWithLevel(l.config, &l.entry, level)
	}
	// pass-through
	return l
}

// WithStack would, if conditions are met, attach a call stack to the log entry
func (l *LogEntry) WithStack(skip int) slog.Logger {
	if l.isEnabled() {
		return doWithStack(l.config, &l.entry, skip+1)
	}
	// pass-through
	return l
}

// WithField would, if conditions are met, attach a field to the log entry. This
// field could be altered if a FieldFilter is used
func (l *LogEntry) WithField(label string, value any) slog.Logger {
	if label != "" && l.isEnabled() {
		if out := doWithField(l.config, &l.entry, label, value); out != nil {
			return out
		}
	}
	// pass-through
	return l
}

// WithFields would, if conditions are met, attach fields to the log entry.
// These fields could be altered if a FieldFilter is used
func (l *LogEntry) WithFields(fields map[string]any) slog.Logger {
	if internal.HasFields(fields) && l.isEnabled() {
		if out := doWithFields(l.config, &l.entry, fields); out != nil {
			return out
		}
	}
	// pass-through
	return l
}
