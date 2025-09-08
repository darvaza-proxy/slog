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

// LogEntry implements a level-filtered logger.
type LogEntry struct {
	config *Logger
	loglet internal.Loglet
}

// Level returns the current log level. Exposed for testing only.
func (l *LogEntry) Level() slog.LogLevel {
	if l == nil {
		return slog.UndefinedLevel
	}
	return l.loglet.Level()
}

func (l *LogEntry) check() bool {
	if l == nil || l.config == nil || l.config.Threshold == slog.UndefinedLevel {
		return false
	}

	return true
}

// Enabled tells if this logger would record logs.
func (l *LogEntry) Enabled() bool {
	parent, level, enabled := l.getEnabled()
	switch {
	case !enabled:
		return false
	case parent == nil:
		return true // verified by getEnabled
	default:
		// Confirm the parent will accept the messages.
		return parent.WithLevel(level).Enabled()
	}
}

func (l *LogEntry) getEnabled() (parent slog.Logger, level slog.LogLevel, enabled bool) {
	if !l.check() {
		return nil, slog.UndefinedLevel, false
	}

	level = l.loglet.Level()
	if level == slog.UndefinedLevel {
		// No level set - not enabled.
		return nil, level, false
	}

	parent = l.config.Parent
	if parent == nil {
		// For parentless, only Fatal/Panic are enabled for termination.
		enabled = level == slog.Fatal || level == slog.Panic
	} else {
		// Below threshold, enabled.
		enabled = level <= l.config.Threshold
	}

	return parent, level, enabled
}

// WithEnabled returns itself and whether it's enabled.
func (l *LogEntry) WithEnabled() (slog.Logger, bool) {
	return l, l.Enabled() // skipcq: GO-W4006
}

// Print would, if conditions are met, add a log entry with the arguments
// in the manner of fmt.Print.
func (l *LogEntry) Print(args ...any) {
	if _, _, ok := l.getEnabled(); ok {
		l.msg(1, fmt.Sprint(args...))
	}
}

// Println would, if conditions are met, add a log entry with the arguments
// in the manner of fmt.Println.
func (l *LogEntry) Println(args ...any) {
	if _, _, ok := l.getEnabled(); ok {
		l.msg(1, fmt.Sprintln(args...))
	}
}

// Printf would, if conditions are met, add a log entry with the arguments
// in the manner of fmt.Printf.
func (l *LogEntry) Printf(format string, args ...any) {
	if _, _, ok := l.getEnabled(); ok {
		l.msg(1, fmt.Sprintf(format, args...))
	}
}

// msg applies MessageFilter before sending the message to
// the parent Logger.
func (l *LogEntry) msg(skip int, msg string) {
	msg, ok := applyMessageFilter(l.config, msg)
	if !ok {
		return
	}

	if l.config.Parent == nil {
		// Parentless is either Fatal or Panic.
		_ = log.Output(skip+2, msg)

		switch l.loglet.Level() {
		case slog.Panic:
			panic(core.NewPanicError(skip+1, msg))
		case slog.Fatal:
			// revive:disable:deep-exit
			os.Exit(1)
		}

		// Unreachable.
		return
	}

	if ll, ok := l.makeLogger(skip + 1); ok {
		ll.Print(msg)
	}
}

func (l *LogEntry) makeLogger(skip int) (slog.Logger, bool) {
	ll := l.config.Parent.WithLevel(l.loglet.Level())
	if !ll.Enabled() {
		// Level disabled for the parent.
		return ll, false
	}

	// Stack.
	if l.loglet.CallStack() != nil {
		ll = ll.WithStack(skip + 1)
	}

	// Collect fields.
	fields := l.loglet.FieldsMap()
	if internal.HasFields(fields) {
		ll = ll.WithFields(fields)
	}

	return ll, true
}

// Debug creates a new filtered logger on level slog.Debug.
func (l *LogEntry) Debug() slog.Logger {
	return l.WithLevel(slog.Debug)
}

// Info creates a new filtered logger on level slog.Info.
func (l *LogEntry) Info() slog.Logger {
	return l.WithLevel(slog.Info)
}

// Warn creates a new filtered logger on level slog.Warn.
func (l *LogEntry) Warn() slog.Logger {
	return l.WithLevel(slog.Warn)
}

// Error creates a new filtered logger on level slog.Error.
func (l *LogEntry) Error() slog.Logger {
	return l.WithLevel(slog.Error)
}

// Fatal creates a new filtered logger on level slog.Fatal.
func (l *LogEntry) Fatal() slog.Logger {
	return l.WithLevel(slog.Fatal)
}

// Panic creates a new filtered logger on level slog.Panic.
func (l *LogEntry) Panic() slog.Logger {
	return l.WithLevel(slog.Panic)
}

// WithLevel creates a new filtered logger on the given level.
func (l *LogEntry) WithLevel(level slog.LogLevel) slog.Logger {
	err := l.checkWithLevel(1, level)
	switch err {
	case nil:
		// Proceed with level change.
		return doWithLevel(l.config, &l.loglet, level)
	case errSkip:
		// Skip - return self.
		return l
	default:
		// Error - panic immediately.
		panic(err)
	}
}

// checkWithLevel determines if LogEntry.WithLevel should proceed.
// Returns nil to proceed, errSkip to skip, or an error to panic.
func (l *LogEntry) checkWithLevel(skip int, level slog.LogLevel) error {
	err := validateLogLevel(skip+1, level)
	switch {
	case err != nil:
		// invalid level
		return err
	case !l.check():
		// invalid instance
		return core.NewPanicErrorf(skip+1, "invalid logger entry state")
	case l.Level() == level:
		// Same level.
		return errSkip
	default:
		// Create a new entry with the requested level, potentially disabled
		// by threshold.
		return nil
	}
}

// WithStack would, if conditions are met, attach a call stack to the log entry.
func (l *LogEntry) WithStack(skip int) slog.Logger {
	if !l.shouldCollectFields() {
		return l
	}

	if skip < 0 {
		skip = 0
	}

	return doWithStack(l.config, &l.loglet, skip+1)
}

// WithField would, if conditions are met, attach a field to the log entry. This
// field could be altered if a FieldFilter is used.
func (l *LogEntry) WithField(label string, value any) slog.Logger {
	if label == "" || !l.shouldCollectFields() {
		return l
	}

	if out := doWithField(l.config, &l.loglet, label, value); out != nil {
		return out
	}

	// Field was filtered out, return unchanged.
	return l
}

// WithFields would, if conditions are met, attach fields to the log entry.
// These fields could be altered if a FieldFilter is used.
func (l *LogEntry) WithFields(fields map[string]any) slog.Logger {
	if !internal.HasFields(fields) || !l.shouldCollectFields() {
		return l
	}

	if out := doWithFields(l.config, &l.loglet, fields); out != nil {
		return out
	}

	// Fields were filtered out, return unchanged.
	return l
}

// shouldCollectFields checks if fields should be collected.
func (l *LogEntry) shouldCollectFields() bool {
	if !l.check() {
		return false
	}

	// Parentless never collects fields (nowhere to send them).
	if l.config.Parent == nil {
		return false
	}

	level := l.loglet.Level()

	// No level set - collect fields speculatively.
	if level == slog.UndefinedLevel {
		return true
	}

	// Level is set - only collect if within threshold.
	return level <= l.config.Threshold
}
