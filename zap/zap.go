// Package zap provides a slog.Logger adaptor using a go.uber.org/zap Logger as backend
package zap

import (
	"fmt"

	// we aren't using zap yet we need to handle zap.Logger and zap.SugaredLogger
	_ "go.uber.org/zap"

	"github.com/darvaza-proxy/slog"
)

var (
	_ slog.Logger = (*Logger)(nil)
)

// LoggerEntry is either a zap.Logger or a zap.SugaredLogger
type LoggerEntry interface{}

// Logger is an adaptor using go.uber.org/zap as slog.Logger
type Logger struct {
	entry LoggerEntry
}

// Print adds a log entry with arguments handled in the manner of fmt.Print
func (zl *Logger) Print(args ...any) {
	zl.print(fmt.Sprint(args...))
}

// Println adds a log entry with arguments handled in the manner of fmt.Println
func (zl *Logger) Println(args ...any) {
	zl.print(fmt.Sprintln(args...))
}

// Printf adds a log entry with arguments handled in the manner of fmt.Printf
func (zl *Logger) Printf(format string, args ...any) {
	zl.print(fmt.Sprintf(format, args...))
}

func (zl *Logger) print(string) {}

// Debug returns a new logger set to add entries as level Debug
func (zl *Logger) Debug() slog.Logger {
	return zl.WithLevel(slog.Debug)
}

// Info returns a new logger set to add entries as level Info
func (zl *Logger) Info() slog.Logger {
	return zl.WithLevel(slog.Info)
}

// Warn returns a new logger set to add entries as level Warn
func (zl *Logger) Warn() slog.Logger {
	return zl.WithLevel(slog.Warn)
}

// Error returns a new logger set to add entries as level Error
func (zl *Logger) Error() slog.Logger {
	return zl.WithLevel(slog.Error)
}

// Fatal returns a new logger set to add entries as level Fatal
func (zl *Logger) Fatal() slog.Logger {
	return zl.WithLevel(slog.Fatal)
}

// WithLevel returns a new logger set to add entries to the specified level
func (zl *Logger) WithLevel(level slog.LogLevel) slog.Logger {
	return zl
}

// WithStack attaches a call stack to a new logger
func (zl *Logger) WithStack(skip int) slog.Logger {
	return zl
}

// WithField returns a new logger with a field attached
func (zl *Logger) WithField(label string, value any) slog.Logger {
	return zl
}

// WithFields returns a new logger with a set of fields attached
func (zl *Logger) WithFields(fields map[string]any) slog.Logger {
	return zl
}

// NewWrapper creates a slog.Logger adaptor using a zap logger as backend
func NewWrapper(entry LoggerEntry) slog.Logger {
	return &Logger{
		entry: entry,
	}
}
