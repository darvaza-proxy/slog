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

// GetEvent returns the current zerolog.Event
func (zl *Logger) GetEvent() *zerolog.Event {
	ev := zl.event
	if ev == nil {
		ev = zl.logger.Log()
	}
	return ev
}

// IsDisabled tells if the underlying logger is disabled or not.
func (zl *Logger) IsDisabled() bool {
	if zl == nil || zl.logger == nil || zl.logger.GetLevel() == zerolog.Disabled {
		return true
	}
	return false
}

// Print adds a log entry with arguments handled in the manner of fmt.Print.
func (zl *Logger) Print(args ...any) {
	if !zl.IsDisabled() {
		zl.print(fmt.Sprint(args...))
	}
}

// Println adds a log entry with arguments handled in the manner of fmt.Println.
func (zl *Logger) Println(args ...any) {
	if !zl.IsDisabled() {
		zl.print(fmt.Sprintln(args...))
	}
}

// Printf adds a log entry with arguments handled in the manner of fmt.Printf.
func (zl *Logger) Printf(format string, args ...any) {
	if !zl.IsDisabled() {
		zl.print(fmt.Sprintf(format, args...))
	}
}

func (zl *Logger) print(msg string) {
	zl.GetEvent().
		Msg(strings.TrimSpace(msg))
}

// Debug returns a new logger set to add entries as level Debug.
func (zl *Logger) Debug() slog.Logger {
	return zl.WithLevel(slog.Debug)
}

// Info returns a new logger set to add entries as level Info.
func (zl *Logger) Info() slog.Logger {
	return zl.WithLevel(slog.Info)
}

// Warn returns a new logger set to add entries as level Warn.
func (zl *Logger) Warn() slog.Logger {
	return zl.WithLevel(slog.Warn)
}

// Error returns a new logger set to add entries as level Error.
func (zl *Logger) Error() slog.Logger {
	return zl.WithLevel(slog.Error)
}

// Fatal returns a new logger set to add entries as level Fatal.
func (zl *Logger) Fatal() slog.Logger {
	return zl.WithLevel(slog.Fatal)
}

// WithLevel returns a new logger set to add entries to the specified level.
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

		// unreachable
		return zl
	} else if zl.IsDisabled() {
		// don't bother, it's disabled
		return zl
	} else {
		zlevel := levels[level]

		// new event
		return &Logger{
			logger: zl.logger,
			event:  zl.logger.WithLevel(zlevel),
		}
	}
}

// WithStack attaches a call stack to a new logger.
func (zl *Logger) WithStack(skip int) slog.Logger {
	if zl.IsDisabled() {
		return zl // NOP
	}

	return zl.NewWithCallback(func(ev *zerolog.Event) *zerolog.Event {
		return ev.CallerSkipFrame(skip + 2).Stack()
	})
}

// WithField returns a new logger with a field attached.
func (zl *Logger) WithField(label string, value any) slog.Logger {
	if zl.IsDisabled() {
		return zl // NOP
	}

	return zl.NewWithCallback(func(ev *zerolog.Event) *zerolog.Event {
		return ev.Interface(label, value)
	})
}

// WithFields returns a new logger with a set of fields attached.
func (zl *Logger) WithFields(fields map[string]any) slog.Logger {
	if zl.IsDisabled() {
		return zl // NOP
	}

	return zl.NewWithCallback(func(ev *zerolog.Event) *zerolog.Event {
		return ev.Fields(fields)
	})
}

// New creates a slog.Logger adaptor using a zerolog as backend, if
// one was passed.
func New(logger *zerolog.Logger) slog.Logger {
	var zl *Logger

	if logger != nil {
		zl = &Logger{
			logger: logger,
		}
	}

	return zl
}

// NewWithCallback creates a new zerolog.Event using a callback to modify it.
func (zl *Logger) NewWithCallback(fn func(ev *zerolog.Event) *zerolog.Event) *Logger {

	ev := zl.GetEvent()
	if fn != nil {
		if ev1 := fn(ev); ev1 != nil {
			ev = ev1
		}
	}

	return &Logger{
		logger: zl.logger,
		event:  ev,
	}
}
