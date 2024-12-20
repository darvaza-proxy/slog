// Package zap provides a slog.Logger adaptor using a go.uber.org/zap Logger as backend
package zap

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"darvaza.org/core"
	"darvaza.org/slog"
)

var (
	_ slog.Logger = (*Logger)(nil)
)

// Logger is an adaptor using go.uber.org/zap as slog.Logger
type Logger struct {
	logger *zap.Logger
	config *zap.Config
}

// Unwrap returns the underlying zap logger
func (zpl *Logger) Unwrap() (*zap.Logger, *zap.Config) {
	return zpl.logger, zpl.config
}

// Enabled tells this logger is enabled
func (zpl *Logger) Enabled() bool {
	if zpl == nil || zpl.logger == nil || zpl.logger.Level() == zapcore.InvalidLevel {
		return false
	}

	return zpl.logger.Check(zpl.logger.Level(), "") != nil
}

// WithEnabled passes the logger and if it's enabled
func (zpl *Logger) WithEnabled() (slog.Logger, bool) {
	return zpl, zpl.Enabled()
}

// Print adds a log entry with arguments handled in the manner of fmt.Print
func (zpl *Logger) Print(args ...any) {
	if zpl.Enabled() {
		zpl.print(fmt.Sprint(args...))
	}
}

// Println adds a log entry with arguments handled in the manner of fmt.Println
func (zpl *Logger) Println(args ...any) {
	if zpl.Enabled() {
		zpl.print(fmt.Sprintln(args...))
	}
}

// Printf adds a log entry with arguments handled in the manner of fmt.Printf
func (zpl *Logger) Printf(format string, args ...any) {
	if zpl.Enabled() {
		zpl.print(fmt.Sprintf(format, args...))
	}
}

// revive:disable:confusing-naming
func (zpl *Logger) print(msg string) {
	msg = strings.TrimSpace(msg)
	if ce := zpl.logger.Check(zpl.logger.Level(), msg); ce != nil {
		ce.Write()
	}
}

// revive:enable:confusing-naming

// Debug returns a new logger set to add entries as level Debug
func (zpl *Logger) Debug() slog.Logger {
	return zpl.WithLevel(slog.Debug)
}

// Info returns a new logger set to add entries as level Info
func (zpl *Logger) Info() slog.Logger {
	return zpl.WithLevel(slog.Info)
}

// Warn returns a new logger set to add entries as level Warn
func (zpl *Logger) Warn() slog.Logger {
	return zpl.WithLevel(slog.Warn)
}

// Error returns a new logger set to add entries as level Error
func (zpl *Logger) Error() slog.Logger {
	return zpl.WithLevel(slog.Error)
}

// Fatal returns a new logger set to add entries as level Fatal
func (zpl *Logger) Fatal() slog.Logger {
	return zpl.WithLevel(slog.Fatal)
}

// Panic returns a new logger set to add entries as level Panic
func (zpl *Logger) Panic() slog.Logger {
	return zpl.WithLevel(slog.Panic)
}

// WithLevel returns a new logger set to add entries to the specified level
func (zpl *Logger) WithLevel(level slog.LogLevel) slog.Logger {
	var levels = []zapcore.Level{
		slog.UndefinedLevel: zapcore.InvalidLevel,
		slog.Panic:          zapcore.PanicLevel,
		slog.Fatal:          zapcore.FatalLevel,
		slog.Error:          zapcore.ErrorLevel,
		slog.Warn:           zapcore.WarnLevel,
		slog.Info:           zapcore.InfoLevel,
		slog.Debug:          zapcore.DebugLevel,
	}

	if level <= slog.UndefinedLevel || int(level) >= len(levels) {
		// fix your code
		zpl.Panic().WithStack(1).Printf("slog: invalid log level %v", level)
	} else if zpl.logger.Core().Enabled(levels[level]) {
		zpl.config.Level.SetLevel(levels[level])
	}

	return zpl
}

// WithStack attaches a call stack to a new logger
func (zpl *Logger) WithStack(skip int) slog.Logger {
	zpl.logger = zpl.logger.WithOptions(
		zap.AddStacktrace(zpl.logger.Level()),
		zap.AddCallerSkip(skip+1))
	return zpl
}

// WithField returns a new logger with a field attached
func (zpl *Logger) WithField(label string, value any) slog.Logger {
	if zpl.Enabled() && label != "" {
		zpl.logger = zpl.logger.With(zap.Any(label, value))
	}
	return zpl
}

// WithFields returns a new logger with a set of fields attached
func (zpl *Logger) WithFields(fields map[string]any) slog.Logger {
	if zpl.Enabled() {
		zs := make([]zap.Field, len(fields))
		for _, k := range core.SortedKeys(fields) {
			zs = append(zs, zap.Any(k, fields[k]))
		}
		zpl.logger = zpl.logger.With(zs...)
	}
	return zpl
}

// New creates a slog.Logger adaptor using a zap as backend. If
// none was passed it will create an opiniated new one.
func New(cfg *zap.Config) slog.Logger {
	return newLogger(cfg)
}

// NewWithCallback creates a new zap logger using a callback to modify it.
func (zpl *Logger) NewWithCallback(fn func(lv zapcore.Entry) error) *Logger {
	if fn != nil && zpl != nil {
		zpl.logger = zpl.logger.WithOptions(zap.Hooks(fn))
	}
	return zpl
}

// NewNoop returns a no-op Logger. It never writes out logs or internal errors
func NewNoop() *Logger {
	nop := zap.NewNop()

	return &Logger{
		logger: nop,
		config: nil,
	}
}

func newLogger(cfg *zap.Config) *Logger {
	if cfg == nil {
		cfg = NewDefaultConfig()
	}

	lg, err := cfg.Build()
	if err != nil {
		return nil
	}

	return &Logger{
		logger: lg,
		config: cfg,
	}
}

// NewDefaultConfig creates a new [zap.Config] logging to the
// console.
func NewDefaultConfig() *zap.Config {
	cfg := zap.NewProductionConfig()
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	cfg.Encoding = "console"
	cfg.EncoderConfig = encoderConfig
	cfg.Sampling = nil
	cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	cfg.DisableStacktrace = true
	cfg.DisableCaller = true
	return &cfg
}
