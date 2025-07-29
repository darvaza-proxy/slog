package zap

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"darvaza.org/slog"
)

func TestMapToZapLevel(t *testing.T) {
	tests := []struct {
		name      string
		slogLevel slog.LogLevel
		zapLevel  zapcore.Level
	}{
		{"Debug", slog.Debug, zapcore.DebugLevel},
		{"Info", slog.Info, zapcore.InfoLevel},
		{"Warn", slog.Warn, zapcore.WarnLevel},
		{"Error", slog.Error, zapcore.ErrorLevel},
		{"Fatal", slog.Fatal, zapcore.FatalLevel},
		{"Panic", slog.Panic, zapcore.PanicLevel},
		{"Invalid", slog.LogLevel(99), zapcore.InvalidLevel},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := mapToZapLevel(tc.slogLevel)
			if result != tc.zapLevel {
				t.Errorf("mapToZapLevel(%v) = %v, want %v", tc.slogLevel, result, tc.zapLevel)
			}
		})
	}
}

func TestMapFromZapLevel(t *testing.T) {
	tests := []struct {
		name      string
		zapLevel  zapcore.Level
		slogLevel slog.LogLevel
	}{
		{"Debug", zapcore.DebugLevel, slog.Debug},
		{"Info", zapcore.InfoLevel, slog.Info},
		{"Warn", zapcore.WarnLevel, slog.Warn},
		{"Error", zapcore.ErrorLevel, slog.Error},
		{"Fatal", zapcore.FatalLevel, slog.Fatal},
		{"Panic", zapcore.PanicLevel, slog.Panic},
		{"DPanic", zapcore.DPanicLevel, slog.Panic},
		{"Invalid", zapcore.InvalidLevel, slog.Info},
		{"Unknown", zapcore.Level(99), slog.Info},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := mapFromZapLevel(tc.zapLevel)
			if result != tc.slogLevel {
				t.Errorf("mapFromZapLevel(%v) = %v, want %v", tc.zapLevel, result, tc.slogLevel)
			}
		})
	}
}

func TestGetConfigLevel(t *testing.T) {
	tests := []struct {
		name     string
		config   *zap.Config
		expected slog.LogLevel
	}{
		{
			name:     "NilConfig",
			config:   nil,
			expected: DefaultLogLevel,
		},
		{
			name: "DebugLevel",
			config: func() *zap.Config {
				cfg := zap.NewDevelopmentConfig()
				cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
				return &cfg
			}(),
			expected: slog.Debug,
		},
		{
			name: "InfoLevel",
			config: func() *zap.Config {
				cfg := zap.NewProductionConfig()
				cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
				return &cfg
			}(),
			expected: slog.Info,
		},
		{
			name: "WarnLevel",
			config: func() *zap.Config {
				cfg := zap.NewProductionConfig()
				cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
				return &cfg
			}(),
			expected: slog.Warn,
		},
		{
			name: "ErrorLevel",
			config: func() *zap.Config {
				cfg := zap.NewProductionConfig()
				cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
				return &cfg
			}(),
			expected: slog.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := getConfigLevel(tc.config)
			if result != tc.expected {
				t.Errorf("getConfigLevel() = %v, want %v", result, tc.expected)
			}
		})
	}
}
