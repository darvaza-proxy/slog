// Package logrus provides a slog.Logger adaptor using a github.com/sirupsen/logrus Logger as backend
package logrus

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/darvaza-proxy/slog"
)

var (
	_ slog.Logger = (*Logger)(nil)
)

// Logger is an adaptor for using github.com/sirupsen/logrus as slog.Logger
type Logger struct {
	entry *logrus.Logger
}

// Print adds a log entry with arguments handled in the manner of fmt.Print
func (rl *Logger) Print(args ...any) {
	rl.print(fmt.Sprint(args...))
}

// Println adds a log entry with arguments handled in the manner of fmt.Println
func (rl *Logger) Println(args ...any) {
	rl.print(fmt.Sprintln(args...))
}

// Printf adds a log entry with arguments handled in the manner of fmt.Printf
func (rl *Logger) Printf(format string, args ...any) {
	rl.print(fmt.Sprintf(format, args...))
}

func (rl *Logger) print(string) {}

// Debug returns a new logger set to add entries as level Debug
func (rl *Logger) Debug() slog.Logger {
	return rl.WithLevel(slog.Debug)
}

// Info returns a new logger set to add entries as level Info
func (rl *Logger) Info() slog.Logger {
	return rl.WithLevel(slog.Info)
}

// Warn returns a new logger set to add entries as level Warn
func (rl *Logger) Warn() slog.Logger {
	return rl.WithLevel(slog.Warn)
}

// Error returns a new logger set to add entries as level Error
func (rl *Logger) Error() slog.Logger {
	return rl.WithLevel(slog.Error)
}

// Fatal returns a new logger set to add entries as level Fatal
func (rl *Logger) Fatal() slog.Logger {
	return rl.WithLevel(slog.Fatal)
}

// WithLevel returns a new logger set to add entries to the specified level
func (rl *Logger) WithLevel(level slog.LogLevel) slog.Logger {
	return rl
}

// WithStack attaches a call stack to a new logger
func (rl *Logger) WithStack(skip int) slog.Logger {
	return rl
}

// WithField returns a new logger with a field attached
func (rl *Logger) WithField(label string, value any) slog.Logger {
	return rl
}

// WithFields returns a new logger with a set of fields attached
func (rl *Logger) WithFields(fields map[string]any) slog.Logger {
	return rl
}

// NewWrapper creates a slog.Logger adaptor using a logrus as backend
func NewWrapper(entry *logrus.Logger) slog.Logger {
	return &Logger{
		entry: entry,
	}
}
