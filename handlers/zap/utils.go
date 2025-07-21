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

// getConfigLevel extracts the current log level from a zap config
func getConfigLevel(cfg *zap.Config) slog.LogLevel {
	if cfg == nil || cfg.Level == (zap.AtomicLevel{}) {
		return DefaultLogLevel
	}

	// Get the current level from the config
	zapLevel := cfg.Level.Level()
	return mapFromZapLevel(zapLevel)
}
