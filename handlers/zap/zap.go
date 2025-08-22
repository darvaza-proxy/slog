// Package zap provides a slog.Logger adaptor using a go.uber.org/zap Logger as backend
package zap

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"darvaza.org/slog"
	"darvaza.org/slog/internal"
)

const (
	// DefaultLogLevel is the default log level for new loggers
	DefaultLogLevel = slog.Info
)

var (
	_ slog.Logger = (*Logger)(nil)
)

// Logger is an adaptor using go.uber.org/zap as slog.Logger
type Logger struct {
	loglet internal.Loglet

	logger *zap.Logger
	config *zap.Config
}

// Unwrap returns the underlying zap logger
func (zpl *Logger) Unwrap() (*zap.Logger, *zap.Config) {
	return zpl.logger, zpl.config
}

// Level returns the current log level. Exposed for testing only.
func (zpl *Logger) Level() slog.LogLevel {
	if zpl == nil {
		return slog.UndefinedLevel
	}
	return zpl.loglet.Level()
}

// Enabled tells this logger is enabled
func (zpl *Logger) Enabled() bool {
	if zpl == nil || zpl.logger == nil {
		return false
	}

	level := mapToZapLevel(zpl.loglet.Level())
	if level == zapcore.InvalidLevel {
		zpl.Panic().WithStack(1).Printf("slog: invalid log level %v", zpl.loglet.Level())
	}

	return zpl.logger.Core().Enabled(level)
}

// WithEnabled passes the logger and if it's enabled
func (zpl *Logger) WithEnabled() (slog.Logger, bool) {
	return zpl, zpl.Enabled()
}

// Print adds a log entry with arguments handled in the manner of fmt.Print
func (zpl *Logger) Print(args ...any) {
	if zpl.Enabled() {
		zpl.logMessage(fmt.Sprint(args...))
	}
}

// Println adds a log entry with arguments handled in the manner of fmt.Println
func (zpl *Logger) Println(args ...any) {
	if zpl.Enabled() {
		zpl.logMessage(fmt.Sprintln(args...))
	}
}

// Printf adds a log entry with arguments handled in the manner of fmt.Printf
func (zpl *Logger) Printf(format string, args ...any) {
	if zpl.Enabled() {
		zpl.logMessage(fmt.Sprintf(format, args...))
	}
}

func (zpl *Logger) logMessage(msg string) {
	msg = strings.TrimSpace(msg)
	level := mapToZapLevel(zpl.loglet.Level())
	if level == zapcore.InvalidLevel {
		zpl.Panic().WithStack(1).Printf("slog: invalid log level %v", zpl.loglet.Level())
	}

	// Check if we can log at this level
	if ce := zpl.logger.Check(level, msg); ce != nil {
		// Add fields from Loglet chain
		if fieldsMap := zpl.loglet.FieldsMap(); len(fieldsMap) > 0 {
			fields := make([]zap.Field, 0, len(fieldsMap))
			for k, v := range fieldsMap {
				fields = append(fields, zap.Any(k, v))
			}
			ce.Write(fields...)
		} else {
			ce.Write()
		}
	}
}

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
	if level <= slog.UndefinedLevel {
		// fix your code
		zpl.Panic().WithStack(1).Printf("slog: invalid log level %v", level)
	} else if level == zpl.loglet.Level() {
		return zpl
	}

	return &Logger{
		loglet: zpl.loglet.WithLevel(level),
		logger: zpl.logger,
		config: zpl.config,
	}
}

// WithStack attaches a call stack to a new logger
func (zpl *Logger) WithStack(skip int) slog.Logger {
	return &Logger{
		loglet: zpl.loglet.WithStack(skip + 1),
		logger: zpl.logger,
		config: zpl.config,
	}
}

// WithField returns a new logger with a field attached
func (zpl *Logger) WithField(label string, value any) slog.Logger {
	if label != "" {
		return &Logger{
			loglet: zpl.loglet.WithField(label, value),
			logger: zpl.logger,
			config: zpl.config,
		}
	}
	return zpl
}

// WithFields returns a new logger with a set of fields attached
func (zpl *Logger) WithFields(fields map[string]any) slog.Logger {
	if internal.HasFields(fields) {
		return &Logger{
			loglet: zpl.loglet.WithFields(fields),
			logger: zpl.logger,
			config: zpl.config,
		}
	}
	return zpl
}

// New creates a slog.Logger adaptor using a zap as backend. If
// none was passed it will create an opinionated new one.
func New(cfg *zap.Config, zapOptions ...zap.Option) (slog.Logger, error) {
	return newLogger(cfg, zapOptions...)
}

// NewWithCallback creates a new zap logger using a callback to modify it.
func (zpl *Logger) NewWithCallback(fn func(lv zapcore.Entry) error) *Logger {
	if fn != nil && zpl != nil {
		return &Logger{
			loglet: zpl.loglet.Copy(),
			logger: zpl.logger.WithOptions(zap.Hooks(fn)),
			config: zpl.config,
		}
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

func newLogger(cfg *zap.Config, zapOptions ...zap.Option) (*Logger, error) {
	if cfg == nil {
		cfg = NewDefaultConfig()
	}

	lg, err := cfg.Build(zapOptions...)
	if err != nil {
		return nil, err
	}

	loglet := &internal.Loglet{}
	return &Logger{
		loglet: loglet.WithLevel(getConfigLevel(cfg)),
		logger: lg,
		config: cfg,
	}, nil
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
