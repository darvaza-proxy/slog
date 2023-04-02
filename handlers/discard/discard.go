// Package discard is a Logger that doesn't really log anything
package discard

import (
	"fmt"
	"log"
	"os"
	"strings"

	"darvaza.org/slog"
)

var (
	_ slog.Logger = (*Logger)(nil)
)

// Logger implements slog.Logger but doesn't log anything
type Logger struct {
	level slog.LogLevel
}

// Enabled tells that we only handle Fatal
func (nl *Logger) Enabled() bool {
	if nl == nil || nl.level > slog.Fatal {
		return false
	}
	return true
}

// WithEnabled passes the logger, but also indicates if it's enabled or not.
// This logger is only enabled for Fatal entries
func (nl *Logger) WithEnabled() (slog.Logger, bool) {
	return nl, nl.Enabled()
}

// Print pretends to add a log entry with arguments handled in the manner of fmt.Print
func (nl *Logger) Print(args ...any) {
	if nl.Enabled() {
		nl.print(fmt.Sprint(args...))
	}
}

// Println pretends to add a log entry with arguments handled in the manner of fmt.Println
func (nl *Logger) Println(args ...any) {
	if nl.Enabled() {
		nl.print(fmt.Sprintln(args...))
	}
}

// Printf pretends to add a log entry with arguments handled in the manner of fmt.Printf
func (nl *Logger) Printf(format string, args ...any) {
	if nl.Enabled() {
		nl.print(fmt.Sprintf(format, args...))
	}
}

// revive:disable:confusing-naming
func (nl *Logger) print(msg string) {
	msg = strings.TrimSpace(msg)
	_ = log.Output(3, msg)

	if nl.level != slog.Fatal {
		panic(msg)
	}
	// revive:disable:deep-exit
	os.Exit(1)
}

// revive:enable:confusing-naming

// Debug pretends to return a new NOOP logger
func (nl *Logger) Debug() slog.Logger { return nl }

// Info pretends to return a new NOOP logger
func (nl *Logger) Info() slog.Logger { return nl }

// Warn pretends to return a new NOOP logger
func (nl *Logger) Warn() slog.Logger { return nl }

// Error pretends to return a new NOOP logger
func (nl *Logger) Error() slog.Logger { return nl }

// Fatal return a new Fatal logger
func (nl *Logger) Fatal() slog.Logger {
	return nl.WithLevel(slog.Fatal)
}

// Panic return a new Panic logger
func (nl *Logger) Panic() slog.Logger {
	return nl.WithLevel(slog.Panic)
}

// WithLevel pretends to return a new logger set to add entries to the
// level.
func (nl *Logger) WithLevel(level slog.LogLevel) slog.Logger {
	if level <= slog.UndefinedLevel {
		// fix your code
		nl.Panic().Printf("slog: invalid log level %v", level)
	}
	return &Logger{level}
}

// WithStack pretends to attach a call stack to the logger
func (nl *Logger) WithStack(int) slog.Logger { return nl }

// WithField pretends to add a fields to the Logger
func (nl *Logger) WithField(string, any) slog.Logger { return nl }

// WithFields pretends to add fields to the Logger
func (nl *Logger) WithFields(map[string]any) slog.Logger { return nl }

// New creates a slog.Logger that doesn't really log anything
func New() slog.Logger { return &Logger{} }
