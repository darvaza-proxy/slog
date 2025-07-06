// Package cblog provides a channel based logger
package cblog

import (
	"fmt"
	"runtime"
	"strings"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/internal"
)

const (
	// DefaultOutputBufferSize is the default size of the channel buffer used for logging.
	DefaultOutputBufferSize = 1024
)

var (
	_ slog.Logger = (*Logger)(nil)
)

// LogMsg represents one structured log entry
type LogMsg struct {
	Message string
	Level   slog.LogLevel
	Fields  map[string]any
	Stack   core.Stack
}

// Logger is a slog.Logger using a channel as backend
type Logger struct {
	internal.Loglet

	l *cblog
}

type cblog struct {
	ch chan LogMsg

	Logger
}

// Enabled tells this logger is enabled
func (*Logger) Enabled() bool {
	return true
}

// WithEnabled passes the logger and if it's enabled
func (l *Logger) WithEnabled() (slog.Logger, bool) {
	return l, true
}

// Print adds a log entry with arguments handled in the manner of fmt.Print
func (l *Logger) Print(args ...any) {
	l.sendMsg(fmt.Sprint(args...))
}

// Println adds a log entry with arguments handled in the manner of fmt.Println
func (l *Logger) Println(args ...any) {
	l.sendMsg(fmt.Sprintln(args...))
}

// Printf adds a log entry with arguments handled in the manner of fmt.Printf
func (l *Logger) Printf(format string, args ...any) {
	l.sendMsg(fmt.Sprintf(format, args...))
}

func (l *Logger) sendMsg(msg string) {
	var m map[string]any

	if n := l.FieldsCount(); n > 0 {
		iter := l.Fields()

		m = make(map[string]any, n)

		for iter.Next() {
			k, v := iter.Field()

			m[k] = v
		}
	}

	l.l.ch <- LogMsg{
		Message: strings.TrimSpace(msg),
		Level:   l.Level(),
		Fields:  m,
		Stack:   l.CallStack(),
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
	} else if level == l.Level() {
		return l
	}

	out := &Logger{
		Loglet: l.Loglet.WithLevel(level),
		l:      l.l,
	}
	return out
}

// WithStack attaches a call stack to a new logger
func (l *Logger) WithStack(skip int) slog.Logger {
	out := &Logger{
		Loglet: l.Loglet.WithStack(skip + 1),
		l:      l.l,
	}
	return out
}

// WithField returns a new logger with a field attached
func (l *Logger) WithField(label string, value any) slog.Logger {
	if label != "" {
		out := &Logger{
			Loglet: l.Loglet.WithField(label, value),
			l:      l.l,
		}
		return out
	}
	return l
}

// WithFields returns a new logger with a set of fields attached
func (l *Logger) WithFields(fields map[string]any) slog.Logger {
	if internal.HasFields(fields) {
		out := &Logger{
			Loglet: l.Loglet.WithFields(fields),
			l:      l.l,
		}
		return out
	}
	return l
}

// New creates a new Channel Based Logger
func New(ch chan LogMsg) (*Logger, <-chan LogMsg) {
	var createdChannel bool
	if ch == nil {
		ch = make(chan LogMsg, DefaultOutputBufferSize)
		createdChannel = true
	}

	l := newLogger(ch)

	// Set finaliser to close the channel if we created it
	if createdChannel {
		runtime.SetFinalizer(l, func(l *cblog) {
			close(l.ch)
		})
	}

	return &l.Logger, ch
}

func newLogger(ch chan LogMsg) *cblog {
	l := &cblog{
		ch: ch,
	}
	l.Logger.l = l
	return l
}
