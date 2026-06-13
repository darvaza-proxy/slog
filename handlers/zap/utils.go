package zap

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"darvaza.org/slog"
)

// mapToZapLevel maps slog levels to zap levels
func mapToZapLevel(level slog.LogLevel) zapcore.Level {
	switch level {
	case slog.Panic:
		return zapcore.PanicLevel
	case slog.Fatal:
		return zapcore.FatalLevel
	case slog.Error:
		return zapcore.ErrorLevel
	case slog.Warn:
		return zapcore.WarnLevel
	case slog.Info:
		return zapcore.InfoLevel
	case slog.Debug:
		return zapcore.DebugLevel
	default:
		return zapcore.InvalidLevel
	}
}

// mapFromZapLevel maps zap levels to slog levels
func mapFromZapLevel(level zapcore.Level) slog.LogLevel {
	switch level {
	case zapcore.DebugLevel:
		return slog.Debug
	case zapcore.WarnLevel:
		return slog.Warn
	case zapcore.ErrorLevel, zapcore.DPanicLevel:
		// DPanic only panics in development mode, which the
		// zap.Logger layer handles; slog has no conditional
		// equivalent, so use the strongest non-terminal level.
		return slog.Error
	case zapcore.PanicLevel:
		return slog.Panic
	case zapcore.FatalLevel:
		return slog.Fatal
	default:
		// InfoLevel is the explicit mapping; unknown levels reuse it
		// as the safest fallback.
		return slog.Info
	}
}

// toZapLevel maps slog levels to zap levels, rejecting values outside
// the range slog defines.
func toZapLevel(level slog.LogLevel) (zapcore.Level, bool) {
	zl := mapToZapLevel(level)
	return zl, zl != zapcore.InvalidLevel
}

// fromZapLevel maps zap levels to slog levels, rejecting values
// outside the range zap defines.
func fromZapLevel(level zapcore.Level) (slog.LogLevel, bool) {
	if level < zapcore.DebugLevel || level > zapcore.FatalLevel {
		return slog.UndefinedLevel, false
	}
	return mapFromZapLevel(level), true
}

// zapFields converts a slog fields map to zap fields.
func zapFields(m map[string]any) []zap.Field {
	if len(m) == 0 {
		return nil
	}

	fields := make([]zap.Field, 0, len(m))
	for k, v := range m {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}

// getConfigLevel extracts the current log level from a zap config
func getConfigLevel(cfg *zap.Config) slog.LogLevel {
	if cfg == nil || cfg.Level == (zap.AtomicLevel{}) {
		return DefaultLogLevel
	}

	// Get the current level from the config
	zapLevel := cfg.Level.Level()
	return mapFromZapLevel(zapLevel)
}
