// Package stdslog provides bidirectional adapters between slog.Logger
// and the standard library log/slog: a slog.Logger backed by a stdlib
// Handler, and a stdlib slog.Handler backed by a slog.Logger.
package stdslog

import (
	"context"
	"fmt"
	stdslog "log/slog"
	"os"
	"strings"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/internal"
)

var (
	_ slog.Logger = (*Logger)(nil)
)

// Logger is an adaptor using a standard library log/slog Handler as
// backend.
type Logger struct {
	handler stdslog.Handler
	loglet  internal.Loglet
}

// Unwrap returns the underlying stdlib handler
func (l *Logger) Unwrap() stdslog.Handler {
	return l.handler
}

// Level returns the current log level. Exposed for testing only.
func (l *Logger) Level() slog.LogLevel {
	if l == nil {
		return slog.UndefinedLevel
	}
	return l.loglet.Level()
}

// Enabled tells if this logger is enabled. A logger with no level set
// reports disabled; set a level first. Fatal and Panic always report
// enabled: terminal entries are delivered and then exit or panic
// regardless of the backend's filter.
func (l *Logger) Enabled() bool {
	switch {
	case l == nil, l.handler == nil:
		return false
	case l.loglet.Level() == slog.UndefinedLevel:
		return false
	}

	level, ok := MapToSLogLevel(l.loglet.Level())
	if !ok {
		// fix your code
		l.Panic().WithStack(1).Printf("slog: invalid log level %v",
			l.loglet.Level())
	}

	if ll := l.loglet.Level(); ll >= slog.Panic && ll <= slog.Fatal {
		return true
	}
	return l.handler.Enabled(context.Background(), level)
}

// WithEnabled passes the logger and if it's enabled
func (l *Logger) WithEnabled() (slog.Logger, bool) {
	return l, l.Enabled() // skipcq: GO-W4006
}

// Print adds a log entry with arguments handled in the manner of fmt.Print
func (l *Logger) Print(args ...any) {
	if l.Enabled() {
		l.logMessage(fmt.Sprint(args...))
	}
}

// Println adds a log entry with arguments handled in the manner of fmt.Println
func (l *Logger) Println(args ...any) {
	if l.Enabled() {
		l.logMessage(fmt.Sprintln(args...))
	}
}

// Printf adds a log entry with arguments handled in the manner of fmt.Printf
func (l *Logger) Printf(format string, args ...any) {
	if l.Enabled() {
		l.logMessage(fmt.Sprintf(format, args...))
	}
}

func (l *Logger) logMessage(msg string) {
	// unreachable: the Print methods gate on Enabled(), which
	// panics or returns false on invalid levels.
	level := core.MustOK(MapToSLogLevel(l.loglet.Level()))

	// Normalise once so the delivered record and the terminal-level
	// payload (the Panic value) cannot diverge on the message.
	msg = strings.TrimSpace(msg)

	// The record's PC stays zero: slog has no caller-attribution
	// concept, and the Handler contract tells backends to ignore it.
	record := stdslog.NewRecord(time.Now(), level, msg, 0)
	record.AddAttrs(l.attrs()...)

	// slog.Logger has no error channel; delivery errors are dropped.
	_ = l.handler.Handle(context.Background(), record)

	l.handleTerminalLevels(msg)
}

// attrs converts the loglet's fields and call stack to stdlib
// attributes, fields first in sorted key order for consistent output.
func (l *Logger) attrs() []stdslog.Attr {
	fields := l.loglet.FieldsMap()
	attrs := make([]stdslog.Attr, 0, len(fields)+1)
	for _, k := range core.SortedKeys(fields) {
		attrs = append(attrs, stdslog.Any(k, fields[k]))
	}

	if stack := l.loglet.CallStack(); len(stack) > 0 {
		attrs = append(attrs, stdslog.String("stack",
			fmt.Sprintf("%+v", stack)))
	}
	return attrs
}

func (l *Logger) handleTerminalLevels(msg string) {
	switch l.loglet.Level() {
	case slog.Fatal:
		// revive:disable:deep-exit
		os.Exit(1)
		// revive:enable:deep-exit
	case slog.Panic:
		const skip = 2 // whoever called Print
		var perr error
		if msg == "" {
			perr = core.NewPanicError(skip, nil)
		} else {
			perr = core.NewPanicError(skip, msg)
		}
		panic(perr)
	default:
		// Non-terminal levels just return — the record has already
		// been handled.
	}
}

// Debug returns a new logger set to add entries as level Debug
func (l *Logger) Debug() slog.Logger {
	return l.WithLevel(slog.Debug)
}

// Info returns a new logger set to add entries as level Info
func (l *Logger) Info() slog.Logger {
	return l.WithLevel(slog.Info)
}

// Warn returns a new logger set to add entries as level Warn
func (l *Logger) Warn() slog.Logger {
	return l.WithLevel(slog.Warn)
}

// Error returns a new logger set to add entries as level Error
func (l *Logger) Error() slog.Logger {
	return l.WithLevel(slog.Error)
}

// Fatal returns a new logger set to add entries as level Fatal
func (l *Logger) Fatal() slog.Logger {
	return l.WithLevel(slog.Fatal)
}

// Panic returns a new logger set to add entries as level Panic
func (l *Logger) Panic() slog.Logger {
	return l.WithLevel(slog.Panic)
}

// WithLevel returns a new logger set to add entries to the specified level
func (l *Logger) WithLevel(level slog.LogLevel) slog.Logger {
	if level <= slog.UndefinedLevel {
		// fix your code
		l.Panic().WithStack(1).Printf("slog: invalid log level %v", level)
	} else if level == l.loglet.Level() {
		return l
	}

	return &Logger{
		loglet:  l.loglet.WithLevel(level),
		handler: l.handler,
	}
}

// shouldCollectFields reports whether attaching fields or a call stack
// is worth the work. Unleveled entries collect speculatively — the
// level is set later. Once a level is set, every action is bound to
// Enabled(): a disabled leveled entry discards everything on Print, so
// the collection is wasted. Fatal and Panic are always enabled, so
// their fields survive to the terminal record.
func (l *Logger) shouldCollectFields() bool {
	if l == nil {
		return false
	}
	if l.loglet.Level() == slog.UndefinedLevel {
		return true
	}
	return l.Enabled()
}

// WithStack attaches a call stack to a new logger
func (l *Logger) WithStack(skip int) slog.Logger {
	if !l.shouldCollectFields() {
		return l
	}
	return &Logger{
		loglet:  l.loglet.WithStack(skip + 1),
		handler: l.handler,
	}
}

// WithField returns a new logger with a field attached
func (l *Logger) WithField(label string, value any) slog.Logger {
	if label == "" || !l.shouldCollectFields() {
		return l
	}
	return &Logger{
		loglet:  l.loglet.WithField(label, value),
		handler: l.handler,
	}
}

// WithFields returns a new logger with a set of fields attached
func (l *Logger) WithFields(fields map[string]any) slog.Logger {
	if !internal.HasFields(fields) || !l.shouldCollectFields() {
		return l
	}
	return &Logger{
		loglet:  l.loglet.WithFields(fields),
		handler: l.handler,
	}
}

// New creates a slog.Logger adaptor using a standard library
// *slog.Logger as backend. When logger is nil the stdlib default
// logger is snapshot at construction; a later slog.SetDefault does
// not reach the adaptor.
func New(logger *stdslog.Logger) slog.Logger {
	if logger == nil {
		logger = stdslog.Default()
	}
	return NewWithHandler(logger.Handler())
}

// NewWithHandler creates a slog.Logger adaptor using a standard
// library slog.Handler as backend, if one was passed.
func NewWithHandler(handler stdslog.Handler) slog.Logger {
	if handler == nil {
		return nil
	}
	return &Logger{
		handler: handler,
	}
}
