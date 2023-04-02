// Package zerolog provides a slog.Logger adaptor using a github.com/rs/zerolog Logger as backend.
package zerolog

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"

	"darvaza.org/core"
	"github.com/darvaza-proxy/slog"
	"github.com/darvaza-proxy/slog/internal"
)

var (
	_ slog.Logger = (*Logger)(nil)
)

// Logger is an adaptor for using github.com/rs/zerolog as slog.Logger.
type Logger struct {
	logger *zerolog.Logger
	event  *zerolog.Event
	action func(string, error)
	err    error
}

// Enabled tells if the underlying logger is enabled or not.
func (zl *Logger) Enabled() bool {
	if zl == nil || zl.logger == nil || zl.logger.GetLevel() == zerolog.Disabled {
		// logger disabled
		return false
	}

	return zl.event.Enabled()
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
	zl.event.Msg(strings.TrimSpace(msg))
	if fn := zl.action; fn != nil {
		fn(msg, zl.err)
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
	var levels = []zerolog.Level{
		slog.UndefinedLevel: zerolog.NoLevel,
		slog.Panic:          zerolog.PanicLevel,
		slog.Fatal:          zerolog.FatalLevel,
		slog.Error:          zerolog.ErrorLevel,
		slog.Warn:           zerolog.WarnLevel,
		slog.Info:           zerolog.InfoLevel,
		slog.Debug:          zerolog.DebugLevel,
	}

	if level <= slog.UndefinedLevel || int(level) >= len(levels) {
		// fix your code
		zl.Panic().WithStack(1).Printf("slog: invalid log level %v", level)
	} else if zl.Enabled() {
		var fn func(string, error)

		zlevel := levels[level]

		switch level {
		case slog.Fatal:
			fn = zl.triggerExit
		case slog.Panic:
			fn = zl.triggerPanic
		}

		// new event
		ev := zl.logger.WithLevel(zlevel)
		return newLogger(zl.logger, ev, fn)
	}

	// NOP
	return zl
}

func (*Logger) triggerExit(string, error) {
	// revive:disable:deep-exit
	os.Exit(1)
	// revive:enable:deep-exit
}

func (*Logger) triggerPanic(msg string, err error) {
	const skip = 2 // whoever called Print
	var perr error
	if msg == "" {
		perr = core.NewPanicError(skip, err)
	} else if err == nil {
		perr = core.NewPanicError(skip, msg)
	} else {
		perr = core.NewPanicWrap(skip, err, msg)
	}
	panic(perr)
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
	if zl.Enabled() && label != "" {
		zl.addField(label, value)
	}
	return zl
}

// WithFields adds fields to the Event Context
func (zl *Logger) WithFields(fields map[string]any) slog.Logger {
	if zl.Enabled() && len(fields) > 0 {
		// append in order
		for _, key := range internal.SortedKeys(fields) {
			zl.addField(key, fields[key])
		}
	}
	return zl
}

func (zl *Logger) addField(label string, value any) {
	if label == slog.ErrorFieldName {
		if err, ok := value.(error); ok {
			zl.event.Err(err)
			zl.err = err
			return
		}
	}
	zl.event.Interface(label, value)
}

// New creates a slog.Logger adaptor using a zerolog as backend, if
// one was passed.
func New(logger *zerolog.Logger) slog.Logger {
	if logger == nil {
		return nil
	}

	return newLogger(logger, nil, nil)
}

// NewWithCallback creates a new zerolog.Event using a callback to modify it.
func (zl *Logger) NewWithCallback(fn func(ev *zerolog.Event)) *Logger {
	// new event
	ev := zl.logger.Log()
	if fn != nil {
		fn(ev)
	}

	return newLogger(zl.logger, ev, nil)
}

func newLogger(logger *zerolog.Logger, ev *zerolog.Event, fn func(string, error)) *Logger {
	if ev == nil {
		ev = logger.Log()
	}

	return &Logger{
		logger: logger,
		event:  ev,
		action: fn,
	}
}
