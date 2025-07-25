// Package zerolog provides a slog.Logger adaptor using a github.com/rs/zerolog Logger as backend.
package zerolog

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/internal"
)

var (
	_ slog.Logger = (*Logger)(nil)
)

// Logger is an adaptor for using github.com/rs/zerolog as slog.Logger.
type Logger struct {
	internal.Loglet

	logger *zerolog.Logger
}

// Enabled tells if the underlying logger is enabled or not.
func (zl *Logger) Enabled() bool {
	if zl == nil || zl.logger == nil || zl.logger.GetLevel() == zerolog.Disabled {
		// logger disabled
		return false
	}

	level := mapToZerologLevel(zl.Level())
	return zl.logger.GetLevel() <= level
}

// WithEnabled tells if the logger would log or not
func (zl *Logger) WithEnabled() (slog.Logger, bool) {
	return zl, zl.Enabled()
}

// Print adds a log entry with arguments handled in the manner of fmt.Print.
func (zl *Logger) Print(args ...any) {
	if zl.Enabled() {
		zl.msg(fmt.Sprint(args...))
	}
}

// Println adds a log entry with arguments handled in the manner of fmt.Println.
func (zl *Logger) Println(args ...any) {
	if zl.Enabled() {
		zl.msg(fmt.Sprintln(args...))
	}
}

// Printf adds a log entry with arguments handled in the manner of fmt.Printf.
func (zl *Logger) Printf(format string, args ...any) {
	if zl.Enabled() {
		zl.msg(fmt.Sprintf(format, args...))
	}
}

func (zl *Logger) msg(msg string) {
	level := mapToZerologLevel(zl.Level())
	event := zl.logger.WithLevel(level)

	// Add fields from Loglet chain
	zl.addFields(event)

	// Add stack trace if present
	if stack := zl.CallStack(); len(stack) > 0 {
		event.CallerSkipFrame(1).Stack()
	}

	event.Msg(strings.TrimSpace(msg))

	// Handle Fatal/Panic
	zl.handleTerminalLevels(msg)
}

func (zl *Logger) addFields(event *zerolog.Event) {
	if zl.FieldsCount() == 0 {
		return
	}

	iter := zl.Fields()
	for iter.Next() {
		k, v := iter.Field()
		zl.addField(event, k, v)
	}
}

func (*Logger) addField(event *zerolog.Event, k string, v any) {
	if k == slog.ErrorFieldName {
		if err, ok := v.(error); ok {
			event.Err(err)
			return
		}
	}
	event.Interface(k, v)
}

func (zl *Logger) handleTerminalLevels(msg string) {
	switch zl.Level() {
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
	}
}

// Debug returns a new Event Context set to add entries as level Debug.
func (zl *Logger) Debug() slog.Logger {
	return zl.WithLevel(slog.Debug)
}

// Info returns a new Event Context set to add entries as level Info.
func (zl *Logger) Info() slog.Logger {
	return zl.WithLevel(slog.Info)
}

// Warn returns a new Event Context set to add entries as level Warn.
func (zl *Logger) Warn() slog.Logger {
	return zl.WithLevel(slog.Warn)
}

// Error returns a new Event Context set to add entries as level Error.
func (zl *Logger) Error() slog.Logger {
	return zl.WithLevel(slog.Error)
}

// Fatal returns a new Event Context set to add entries as level Fatal.
func (zl *Logger) Fatal() slog.Logger {
	return zl.WithLevel(slog.Fatal)
}

// Panic returns a new Event Context set to add entries as level Panic.
func (zl *Logger) Panic() slog.Logger {
	return zl.WithLevel(slog.Panic)
}

// WithLevel returns a new Event Context set to add entries to the specified level.
func (zl *Logger) WithLevel(level slog.LogLevel) slog.Logger {
	if level <= slog.UndefinedLevel {
		// fix your code
		zl.Panic().WithStack(1).Printf("slog: invalid log level %v", level)
	} else if level == zl.Level() {
		return zl
	}

	return &Logger{
		Loglet: zl.Loglet.WithLevel(level),
		logger: zl.logger,
	}
}

// WithStack attaches a call stack to the Event Context
func (zl *Logger) WithStack(skip int) slog.Logger {
	return &Logger{
		Loglet: zl.Loglet.WithStack(skip + 1),
		logger: zl.logger,
	}
}

// WithField adds a field to the Event Context
func (zl *Logger) WithField(label string, value any) slog.Logger {
	if label != "" {
		return &Logger{
			Loglet: zl.Loglet.WithField(label, value),
			logger: zl.logger,
		}
	}
	return zl
}

// WithFields adds fields to the Event Context
func (zl *Logger) WithFields(fields map[string]any) slog.Logger {
	if internal.HasFields(fields) {
		return &Logger{
			Loglet: zl.Loglet.WithFields(fields),
			logger: zl.logger,
		}
	}
	return zl
}

// New creates a slog.Logger adaptor using a zerolog as backend, if
// one was passed.
func New(logger *zerolog.Logger) slog.Logger {
	if logger == nil {
		return nil
	}

	return &Logger{
		logger: logger,
	}
}

// mapToZerologLevel maps slog levels to zerolog levels
func mapToZerologLevel(level slog.LogLevel) zerolog.Level {
	switch level {
	case slog.Panic:
		return zerolog.PanicLevel
	case slog.Fatal:
		return zerolog.FatalLevel
	case slog.Error:
		return zerolog.ErrorLevel
	case slog.Warn:
		return zerolog.WarnLevel
	case slog.Info:
		return zerolog.InfoLevel
	case slog.Debug:
		return zerolog.DebugLevel
	default:
		return zerolog.NoLevel
	}
}
