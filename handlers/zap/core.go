package zap

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"darvaza.org/slog"
)

var (
	_ zapcore.Core = (*SlogCore)(nil)
)

// SlogCore implements zapcore.Core using slog.Logger as backend
type SlogCore struct {
	logger slog.Logger
	level  zapcore.LevelEnabler
	fields []zapcore.Field
}

// NewCore creates a zapcore.Core that writes to the provided slog.Logger
func NewCore(logger slog.Logger, level zapcore.LevelEnabler) zapcore.Core {
	if logger == nil {
		panic("nil slog.Logger")
	}
	if level == nil {
		level = zap.InfoLevel
	}
	return &SlogCore{
		logger: logger,
		level:  level,
	}
}

// NewZapLogger creates a *zap.Logger that writes to the provided slog.Logger
func NewZapLogger(logger slog.Logger, opts ...zap.Option) *zap.Logger {
	core := NewCore(logger, zap.InfoLevel)
	return zap.New(core, opts...)
}

// Enabled returns whether the given level is enabled
func (c *SlogCore) Enabled(level zapcore.Level) bool {
	return c.level.Enabled(level)
}

// With returns a new Core with additional fields
func (c *SlogCore) With(fields []zapcore.Field) zapcore.Core {
	if len(fields) == 0 {
		return c // No need to clone if no new fields
	}

	newFields := make([]zapcore.Field, len(c.fields)+len(fields))
	copy(newFields, c.fields)
	copy(newFields[len(c.fields):], fields)

	return &SlogCore{
		logger: c.logger,
		level:  c.level,
		fields: newFields,
	}
}

// Check determines whether the logger should log at the given level
func (c *SlogCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checked.AddCore(entry, c)
	}
	return checked
}

// Write serializes the Entry and any Fields to the slog backend
func (c *SlogCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Convert zap level to slog level
	slogLevel := mapZapToSlogLevel(entry.Level)

	// Start with the appropriate log level
	logger := c.logger.WithLevel(slogLevel)

	// Add stack trace if available
	if entry.Stack != "" {
		logger = logger.WithField("stacktrace", entry.Stack)
	}

	// Add caller information if available
	if entry.Caller.Defined {
		logger = logger.WithField("caller", entry.Caller.String())
	}

	// Convert and add accumulated fields from With() calls
	if len(c.fields) > 0 {
		logger = logger.WithFields(convertFields(c.fields))
	}

	// Convert and add fields from this Write call
	if len(fields) > 0 {
		logger = logger.WithFields(convertFields(fields))
	}

	// Log the message
	logger.Print(entry.Message)

	// Handle Fatal and Panic as zap expects
	switch entry.Level {
	case zapcore.FatalLevel:
		// zap expects Fatal to exit
		c.logger.Fatal().Print("zap fatal exit")
	case zapcore.PanicLevel, zapcore.DPanicLevel:
		// zap expects Panic to panic
		panic(fmt.Sprintf("zap panic: %s", entry.Message))
	}

	return nil
}

// Sync flushes any buffered log entries
func (*SlogCore) Sync() error {
	// slog doesn't have a Sync concept, so this is a no-op
	return nil
}

// convertFields converts zap fields to a map for slog
func convertFields(zapFields []zapcore.Field) map[string]any {
	if len(zapFields) == 0 {
		return nil
	}

	// Use zap's MapObjectEncoder to extract field values
	enc := zapcore.NewMapObjectEncoder()
	for _, field := range zapFields {
		field.AddTo(enc)
	}

	return enc.Fields
}

// mapZapToSlogLevel maps zap levels to slog levels
func mapZapToSlogLevel(level zapcore.Level) slog.LogLevel {
	switch level {
	case zapcore.DebugLevel:
		return slog.Debug
	case zapcore.InfoLevel:
		return slog.Info
	case zapcore.WarnLevel:
		return slog.Warn
	case zapcore.ErrorLevel:
		return slog.Error
	case zapcore.DPanicLevel, zapcore.PanicLevel:
		return slog.Panic
	case zapcore.FatalLevel:
		return slog.Fatal
	default:
		// Unknown levels default to Info
		return slog.Info
	}
}
