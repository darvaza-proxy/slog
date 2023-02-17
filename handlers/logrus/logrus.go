// Package logrus provides a slog.Logger using
// github.com/sirupsen/logrus Logger as backend
package logrus

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/darvaza-proxy/slog"
	"github.com/darvaza-proxy/slog/internal"
)

var (
	_ slog.Logger = (*Logger)(nil)
)

const (
	// CallerFieldName is the field name to be used by WithStack()
	// attempting to mimick the effect of logrus' own SetReportCaller()
	CallerFieldName = "method"

	// StackFieldName is the field name used to store the formatted callstack
	StackFieldName = "call-stack"
)

// Logger is an adaptor for using github.com/sirupsen/logrus as slog.Logger
type Logger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
	level  logrus.Level
}

// Enabled tells if the logger is enabled
func (rl *Logger) Enabled() bool {
	if rl == nil || rl.logger == nil || rl.entry == nil {
		// invalid
		return false
	}
	return rl.logger.IsLevelEnabled(rl.level)
}

// WithEnabled tells if the logger would log or not
func (rl *Logger) WithEnabled() (slog.Logger, bool) {
	return rl, rl.Enabled()
}

// Print adds a log entry with arguments handled in the manner of fmt.Print
func (rl *Logger) Print(args ...any) {
	if rl.Enabled() {
		rl.print(fmt.Sprint(args...))
	}
}

// Println adds a log entry with arguments handled in the manner of fmt.Println
func (rl *Logger) Println(args ...any) {
	if rl.Enabled() {
		rl.print(fmt.Sprintln(args...))
	}
}

// Printf adds a log entry with arguments handled in the manner of fmt.Printf
func (rl *Logger) Printf(format string, args ...any) {
	if rl.Enabled() {
		rl.print(fmt.Sprintf(format, args...))
	}
}

func (rl *Logger) print(msg string) {
	rl.entry.Log(rl.level, strings.TrimSpace(msg))
}

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

// Panic returns a new logger set to add entries as level Panic
func (rl *Logger) Panic() slog.Logger {
	return rl.WithLevel(slog.Panic)
}

// WithLevel returns a new logger set to add entries to the specified level
func (rl *Logger) WithLevel(level slog.LogLevel) slog.Logger {
	var levels = []logrus.Level{
		slog.UndefinedLevel: logrus.TraceLevel + 1,
		slog.Panic:          logrus.PanicLevel,
		slog.Fatal:          logrus.FatalLevel,
		slog.Error:          logrus.ErrorLevel,
		slog.Warn:           logrus.WarnLevel,
		slog.Info:           logrus.InfoLevel,
		slog.Debug:          logrus.DebugLevel,
	}

	if level <= slog.UndefinedLevel || int(level) >= len(levels) {
		// fix your code
		rl.Panic().WithStack(1).Printf("slog: invalid log level %v", level)
	}

	out := rl.dup(nil)
	out.level = levels[level]
	return out
}

// WithStack attaches a call stack to the log entry
func (rl *Logger) WithStack(skip int) slog.Logger {
	if rl.Enabled() {
		frames := internal.StackTrace(skip + 1)
		if len(frames) > 0 {
			caller := frames[0]

			entry := rl.entry.WithFields(logrus.Fields{
				CallerFieldName: fmt.Sprintf("%+n", caller),
				StackFieldName:  fmt.Sprintf("%+n", frames),
			})

			return rl.dup(entry)
		}
	}
	return rl
}

// WithField adds a field to the log entry
func (rl *Logger) WithField(label string, value any) slog.Logger {
	if rl.Enabled() {
		entry := rl.entry.WithFields(logrus.Fields{
			label: value,
		})
		return rl.dup(entry)
	}
	return rl
}

// WithFields adds fields to the log entry
func (rl *Logger) WithFields(fields map[string]any) slog.Logger {
	if rl.Enabled() {
		entry := rl.entry.WithFields(fields)
		return rl.dup(entry)
	}
	return rl
}

// New creates a slog.Logger adaptor using a logrus as backend
func New(logger *logrus.Logger) slog.Logger {
	if logger == nil {
		return nil
	}

	return &Logger{
		logger: logger,
		entry:  logrus.NewEntry(logger),
		level:  logger.GetLevel(),
	}
}

// dup duplicates the entry to be modified
func (rl *Logger) dup(entry *logrus.Entry) *Logger {
	if entry == nil {
		// unless one is given, duplicate the current
		entry = rl.entry.Dup()
	}

	return &Logger{
		logger: rl.logger,
		entry:  entry,
		level:  rl.level,
	}
}
