// Package logrus provides a slog.Logger using
// github.com/sirupsen/logrus Logger as backend
package logrus

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"darvaza.org/slog"
	"darvaza.org/slog/internal"
)

var (
	_ slog.Logger = (*Logger)(nil)
)

const (
	// CallerFieldName is the field name to be used by WithStack()
	// attempting to mimic the effect of logrus' own SetReportCaller()
	CallerFieldName = "method"

	// StackFieldName is the field name used to store the formatted callstack
	StackFieldName = "call-stack"
)

// Logger is an adaptor for using github.com/sirupsen/logrus as slog.Logger
type Logger struct {
	loglet internal.Loglet

	logger *logrus.Logger
	entry  *logrus.Entry
}

// Level returns the current log level. Exposed for testing only.
func (rl *Logger) Level() slog.LogLevel {
	if rl == nil {
		return slog.UndefinedLevel
	}
	return rl.loglet.Level()
}

// Enabled tells if the logger is enabled
func (rl *Logger) Enabled() bool {
	if rl == nil || rl.logger == nil || rl.entry == nil {
		// invalid
		return false
	}
	level := mapToLogrusLevel(rl.loglet.Level())
	return rl.logger.IsLevelEnabled(level)
}

// WithEnabled tells if the logger would log or not
func (rl *Logger) WithEnabled() (slog.Logger, bool) {
	return rl, rl.Enabled()
}

// Print adds a log entry with arguments handled in the manner of fmt.Print
func (rl *Logger) Print(args ...any) {
	if rl.Enabled() {
		rl.msg(fmt.Sprint(args...))
	}
}

// Println adds a log entry with arguments handled in the manner of fmt.Println
func (rl *Logger) Println(args ...any) {
	if rl.Enabled() {
		rl.msg(fmt.Sprintln(args...))
	}
}

// Printf adds a log entry with arguments handled in the manner of fmt.Printf
func (rl *Logger) Printf(format string, args ...any) {
	if rl.Enabled() {
		rl.msg(fmt.Sprintf(format, args...))
	}
}

func (rl *Logger) msg(msg string) {
	level := mapToLogrusLevel(rl.loglet.Level())

	// Build entry with fields from Loglet chain
	entry := rl.entry
	if fields := rl.loglet.FieldsMap(); fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}

	// Add stack trace if present
	if stack := rl.loglet.CallStack(); len(stack) > 0 {
		caller := stack[0]
		entry = entry.WithFields(logrus.Fields{
			CallerFieldName: fmt.Sprintf("%+n", caller),
			StackFieldName:  fmt.Sprintf("%+n", stack),
		})
	}

	entry.Log(level, strings.TrimSpace(msg))
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
	if level <= slog.UndefinedLevel {
		// fix your code
		rl.Panic().WithStack(1).Printf("slog: invalid log level %v", level)
	} else if level == rl.loglet.Level() {
		return rl
	}

	return &Logger{
		loglet: rl.loglet.WithLevel(level),
		logger: rl.logger,
		entry:  rl.entry,
	}
}

// WithStack attaches a call stack to the log entry
func (rl *Logger) WithStack(skip int) slog.Logger {
	return &Logger{
		loglet: rl.loglet.WithStack(skip + 1),
		logger: rl.logger,
		entry:  rl.entry,
	}
}

// WithField adds a field to the log entry
func (rl *Logger) WithField(label string, value any) slog.Logger {
	if label != "" {
		return &Logger{
			loglet: rl.loglet.WithField(label, value),
			logger: rl.logger,
			entry:  rl.entry,
		}
	}
	return rl
}

// WithFields adds fields to the log entry
func (rl *Logger) WithFields(fields map[string]any) slog.Logger {
	if internal.HasFields(fields) {
		return &Logger{
			loglet: rl.loglet.WithFields(fields),
			logger: rl.logger,
			entry:  rl.entry,
		}
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
	}
}

const invalidLogrusLevel = logrus.TraceLevel + 1

// mapToLogrusLevel maps slog levels to logrus levels
func mapToLogrusLevel(level slog.LogLevel) logrus.Level {
	switch level {
	case slog.Panic:
		return logrus.PanicLevel
	case slog.Fatal:
		return logrus.FatalLevel
	case slog.Error:
		return logrus.ErrorLevel
	case slog.Warn:
		return logrus.WarnLevel
	case slog.Info:
		return logrus.InfoLevel
	case slog.Debug:
		return logrus.DebugLevel
	default:
		// Return an invalid logrus level for undefined slog levels
		return invalidLogrusLevel
	}
}
