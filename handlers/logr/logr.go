// Package logr provides adapters between slog.Logger and go-logr/logr
package logr

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/logr"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/internal"
)

var (
	_ slog.Logger = (*Logger)(nil)
)

// Logger is an adaptor using go-logr/logr as slog.Logger
type Logger struct {
	loglet internal.Loglet

	logger logr.Logger
}

// Unwrap returns the underlying logr logger
func (ll *Logger) Unwrap() logr.Logger {
	return ll.logger
}

// Level returns the current log level. Exposed for testing only.
func (ll *Logger) Level() slog.LogLevel {
	if ll == nil {
		return slog.UndefinedLevel
	}
	return ll.loglet.Level()
}

// Enabled tells if this logger is enabled
func (ll *Logger) Enabled() bool {
	if ll == nil || ll.logger.GetSink() == nil {
		return false
	}

	level := mapToLogrLevel(ll.loglet.Level())
	if level < 0 {
		// Error levels are always enabled in logr
		return true
	}
	return ll.logger.V(level).Enabled()
}

// WithEnabled passes the logger and if it's enabled
func (ll *Logger) WithEnabled() (slog.Logger, bool) {
	return ll, ll.Enabled()
}

// Print adds a log entry with arguments handled in the manner of fmt.Print
func (ll *Logger) Print(args ...any) {
	if ll.Enabled() {
		ll.doPrint(fmt.Sprint(args...))
	}
}

// Println adds a log entry with arguments handled in the manner of fmt.Println
func (ll *Logger) Println(args ...any) {
	if ll.Enabled() {
		ll.doPrint(fmt.Sprintln(args...))
	}
}

// Printf adds a log entry with arguments handled in the manner of fmt.Printf
func (ll *Logger) Printf(format string, args ...any) {
	if ll.Enabled() {
		ll.doPrint(fmt.Sprintf(format, args...))
	}
}

// collectKeysAndValues creates the key-value slice for logging from Loglet fields and stack
func (ll *Logger) collectKeysAndValues() []any {
	var keysAndValues []any

	// Collect fields from Loglet chain
	if n := ll.loglet.FieldsCount(); n > 0 {
		// Collect fields into a map first for sorting
		fields := make(map[string]any, n)
		iter := ll.loglet.Fields()
		for iter.Next() {
			k, v := iter.Field()
			fields[k] = v
		}

		// Add fields in sorted order for consistent output
		keysAndValues = make([]any, 0, n*2)
		sortedKeys := core.SortedKeys(fields)
		for _, k := range sortedKeys {
			keysAndValues = append(keysAndValues, k, fields[k])
		}
	}

	// Add stack trace if present
	if stack := ll.loglet.CallStack(); len(stack) > 0 {
		keysAndValues = append(keysAndValues, "stack", fmt.Sprintf("%+v", stack))
	}

	return keysAndValues
}

func (ll *Logger) doPrint(msg string) {
	msg = strings.TrimSpace(msg)
	level := ll.loglet.Level()
	keysAndValues := ll.collectKeysAndValues()

	// Log based on level
	switch level {
	case slog.Panic:
		ll.logger.Error(nil, msg, keysAndValues...)
		core.Panic(msg)
	case slog.Fatal:
		ll.logger.Error(nil, msg, keysAndValues...)
		// revive:disable:deep-exit
		os.Exit(1)
		// revive:enable:deep-exit
	case slog.Error:
		ll.logger.Error(nil, msg, keysAndValues...)
	default:
		// Info, Warn, Debug all use Info with V-levels
		ll.logger.V(mapToLogrLevel(level)).Info(msg, keysAndValues...)
	}
}

// Debug returns a new logger set to add entries as level Debug
func (ll *Logger) Debug() slog.Logger {
	return ll.WithLevel(slog.Debug)
}

// Info returns a new logger set to add entries as level Info
func (ll *Logger) Info() slog.Logger {
	return ll.WithLevel(slog.Info)
}

// Warn returns a new logger set to add entries as level Warn
func (ll *Logger) Warn() slog.Logger {
	return ll.WithLevel(slog.Warn)
}

// Error returns a new logger set to add entries as level Error
func (ll *Logger) Error() slog.Logger {
	return ll.WithLevel(slog.Error)
}

// Fatal returns a new logger set to add entries as level Fatal
func (ll *Logger) Fatal() slog.Logger {
	return ll.WithLevel(slog.Fatal)
}

// Panic returns a new logger set to add entries as level Panic
func (ll *Logger) Panic() slog.Logger {
	return ll.WithLevel(slog.Panic)
}

// WithLevel returns a new logger set to add entries to the specified level
func (ll *Logger) WithLevel(level slog.LogLevel) slog.Logger {
	if level <= slog.UndefinedLevel {
		// fix your code
		ll.Panic().WithStack(1).Printf("slog: invalid log level %v", level)
	} else if level == ll.loglet.Level() {
		return ll
	}

	return &Logger{
		loglet: ll.loglet.WithLevel(level),
		logger: ll.logger,
	}
}

// WithStack attaches a call stack to a new logger
func (ll *Logger) WithStack(skip int) slog.Logger {
	return &Logger{
		loglet: ll.loglet.WithStack(skip + 1),
		logger: ll.logger,
	}
}

// WithField returns a new logger with a field attached
func (ll *Logger) WithField(label string, value any) slog.Logger {
	if label != "" {
		return &Logger{
			loglet: ll.loglet.WithField(label, value),
			logger: ll.logger,
		}
	}
	return ll
}

// WithFields returns a new logger with a set of fields attached
func (ll *Logger) WithFields(fields map[string]any) slog.Logger {
	if internal.HasFields(fields) {
		return &Logger{
			loglet: ll.loglet.WithFields(fields),
			logger: ll.logger,
		}
	}
	return ll
}

// New creates a slog.Logger adaptor using a logr.Logger as backend
func New(logger logr.Logger) slog.Logger {
	return &Logger{
		logger: logger,
	}
}

// mapToLogrLevel maps slog levels to logr V-levels
// logr uses verbosity levels where 0 is most important
// We map: Error/Fatal/Panic -> -1 (use Error() method instead)
// Warn/Info -> V(0), Debug -> V(1)
// Note: logr doesn't have a warn level, so we map it to info
func mapToLogrLevel(level slog.LogLevel) int {
	switch level {
	case slog.Warn, slog.Info:
		return 0
	case slog.Debug:
		return 1
	default:
		// Error, Fatal, Panic don't use V-levels
		// Return -1 to indicate these should use Error() method
		return -1
	}
}
