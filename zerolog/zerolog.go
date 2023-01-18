// Package zerolog provides a slog.Logger adaptor using a github.com/rs/zerolog Logger as backend.
package zerolog

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"

	"github.com/darvaza-proxy/slog"
)

var (
	_ slog.Logger = (*Logger)(nil)
)

// Logger is an adaptor for using github.com/rs/zerolog as slog.Logger.
type Logger struct {
	logger *zerolog.Logger
	event  *zerolog.Event
}

// Enabled tells if the underlying logger is enabled or not.
func (zl *Logger) Enabled() bool {
	if zl == nil || zl.logger == nil || zl.logger.GetLevel() == zerolog.Disabled {
		// logger disabled
		return false
	}

	return zl.event.Enabled()
}

// Print adds a log entry with arguments handled in the manner of fmt.Print.
func (zl *Logger) Print(args ...any) {
	if zl.Enabled() {
		zl.print(fmt.Sprint(args...))
	}
}

// Println adds a log entry with arguments handled in the manner of fmt.Println.
func (zl *Logger) Println(args ...any) {
	if zl.Enabled() {
		zl.print(fmt.Sprintln(args...))
	}
}

// Printf adds a log entry with arguments handled in the manner of fmt.Printf.
func (zl *Logger) Printf(format string, args ...any) {
	if zl.Enabled() {
		zl.print(fmt.Sprintf(format, args...))
	}
}

func (zl *Logger) print(msg string) {
	zl.event.Msg(strings.TrimSpace(msg))
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

// WithLevel returns a new Event Context set to add entries to the specified level.
func (zl *Logger) WithLevel(level slog.LogLevel) slog.Logger {
	var levels = []zerolog.Level{
		slog.UndefinedLevel: zerolog.NoLevel,
		slog.Fatal:          zerolog.FatalLevel,
		slog.Error:          zerolog.ErrorLevel,
		slog.Warn:           zerolog.WarnLevel,
		slog.Info:           zerolog.InfoLevel,
		slog.Debug:          zerolog.DebugLevel,
	}

	if level < slog.UndefinedLevel || int(level) >= len(levels) {
		// fix your code
		err := fmt.Errorf("slog: invalid log level %v", level)
		zl.logger.Panic().Stack().Err(err).Send()

	} else if zl.Enabled() {
		zlevel := levels[level]

		// new event
		ev := zl.logger.WithLevel(zlevel)
		return newLogger(zl.logger, ev)
	}

	// NOP
	return zl
}

// WithStack attaches a call stack to the Event Context
func (zl *Logger) WithStack(skip int) slog.Logger {
	if zl.Enabled() {
		zl.event.CallerSkipFrame(skip + 1)
		zl.event.Stack()
	}
	return zl
}

// WithField adds a field to the Event Context
func (zl *Logger) WithField(label string, value any) slog.Logger {
	if zl.Enabled() {
		zl.event.Interface(label, value)
	}
	return zl
}

// WithFields adds fields to the Event Context
func (zl *Logger) WithFields(fields map[string]any) slog.Logger {
	if zl.Enabled() {
		zl.event.Fields(fields)
	}
	return zl
}

// New creates a slog.Logger adaptor using a zerolog as backend, if
// one was passed.
func New(logger *zerolog.Logger) slog.Logger {

	if logger == nil {
		return nil
	}

	return newLogger(logger, nil)
}

// NewWithCallback creates a new zerolog.Event using a callback to modify it.
func (zl *Logger) NewWithCallback(fn func(ev *zerolog.Event)) *Logger {

	// new event
	ev := zl.logger.Log()
	if fn != nil {
		fn(ev)
	}

	return newLogger(zl.logger, ev)
}

func newLogger(logger *zerolog.Logger, ev *zerolog.Event) *Logger {
	if ev == nil {
		ev = logger.Log()
	}

	return &Logger{
		logger: logger,
		event:  ev,
	}
}
