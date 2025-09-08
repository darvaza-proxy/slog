package zap

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = mapToZapLevelTestCase{}
var _ core.TestCase = mapFromZapLevelTestCase{}
var _ core.TestCase = getConfigLevelTestCase{}

type mapToZapLevelTestCase struct {
	name      string
	slogLevel slog.LogLevel
	zapLevel  zapcore.Level
}

func (tc mapToZapLevelTestCase) Name() string {
	return tc.name
}

func (tc mapToZapLevelTestCase) Test(t *testing.T) {
	t.Helper()
	result := mapToZapLevel(tc.slogLevel)
	core.AssertEqual(t, tc.zapLevel, result, "mapToZapLevel(%v)", tc.slogLevel)
}

func newMapToZapLevelTestCase(name string, slogLevel slog.LogLevel, zapLevel zapcore.Level) mapToZapLevelTestCase {
	return mapToZapLevelTestCase{
		name:      name,
		slogLevel: slogLevel,
		zapLevel:  zapLevel,
	}
}

func mapToZapLevelTestCases() []mapToZapLevelTestCase {
	return []mapToZapLevelTestCase{
		newMapToZapLevelTestCase("Debug", slog.Debug, zapcore.DebugLevel),
		newMapToZapLevelTestCase("Info", slog.Info, zapcore.InfoLevel),
		newMapToZapLevelTestCase("Warn", slog.Warn, zapcore.WarnLevel),
		newMapToZapLevelTestCase("Error", slog.Error, zapcore.ErrorLevel),
		newMapToZapLevelTestCase("Fatal", slog.Fatal, zapcore.FatalLevel),
		newMapToZapLevelTestCase("Panic", slog.Panic, zapcore.PanicLevel),
		newMapToZapLevelTestCase("Invalid", slog.LogLevel(99), zapcore.InvalidLevel),
	}
}

func TestMapToZapLevel(t *testing.T) {
	core.RunTestCases(t, mapToZapLevelTestCases())
}

type mapFromZapLevelTestCase struct {
	name      string
	zapLevel  zapcore.Level
	slogLevel slog.LogLevel
}

func (tc mapFromZapLevelTestCase) Name() string {
	return tc.name
}

func (tc mapFromZapLevelTestCase) Test(t *testing.T) {
	t.Helper()
	result := mapFromZapLevel(tc.zapLevel)
	core.AssertEqual(t, tc.slogLevel, result, "mapFromZapLevel(%v)", tc.zapLevel)
}

func newMapFromZapLevelTestCase(name string, zapLevel zapcore.Level, slogLevel slog.LogLevel) mapFromZapLevelTestCase {
	return mapFromZapLevelTestCase{
		name:      name,
		zapLevel:  zapLevel,
		slogLevel: slogLevel,
	}
}

func mapFromZapLevelTestCases() []mapFromZapLevelTestCase {
	return []mapFromZapLevelTestCase{
		newMapFromZapLevelTestCase("Debug", zapcore.DebugLevel, slog.Debug),
		newMapFromZapLevelTestCase("Info", zapcore.InfoLevel, slog.Info),
		newMapFromZapLevelTestCase("Warn", zapcore.WarnLevel, slog.Warn),
		newMapFromZapLevelTestCase("Error", zapcore.ErrorLevel, slog.Error),
		newMapFromZapLevelTestCase("Fatal", zapcore.FatalLevel, slog.Fatal),
		newMapFromZapLevelTestCase("Panic", zapcore.PanicLevel, slog.Panic),
		newMapFromZapLevelTestCase("DPanic", zapcore.DPanicLevel, slog.Panic),
		newMapFromZapLevelTestCase("Invalid", zapcore.InvalidLevel, slog.Info),
		newMapFromZapLevelTestCase("Unknown", zapcore.Level(99), slog.Info),
	}
}

func TestMapFromZapLevel(t *testing.T) {
	core.RunTestCases(t, mapFromZapLevelTestCases())
}

type getConfigLevelTestCase struct {
	config   *zap.Config
	name     string
	expected slog.LogLevel
}

func (tc getConfigLevelTestCase) Name() string {
	return tc.name
}

func (tc getConfigLevelTestCase) Test(t *testing.T) {
	t.Helper()
	result := getConfigLevel(tc.config)
	core.AssertEqual(t, tc.expected, result, "getConfigLevel()")
}

func newGetConfigLevelTestCase(name string, config *zap.Config, expected slog.LogLevel) getConfigLevelTestCase {
	return getConfigLevelTestCase{
		name:     name,
		config:   config,
		expected: expected,
	}
}

func getConfigLevelTestCases() []getConfigLevelTestCase {
	return []getConfigLevelTestCase{
		newGetConfigLevelTestCase("NilConfig", nil, DefaultLogLevel),
		newGetConfigLevelTestCase("DebugLevel", func() *zap.Config {
			cfg := zap.NewDevelopmentConfig()
			cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
			return &cfg
		}(), slog.Debug),
		newGetConfigLevelTestCase("InfoLevel", func() *zap.Config {
			cfg := zap.NewProductionConfig()
			cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
			return &cfg
		}(), slog.Info),
		newGetConfigLevelTestCase("WarnLevel", func() *zap.Config {
			cfg := zap.NewProductionConfig()
			cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
			return &cfg
		}(), slog.Warn),
		newGetConfigLevelTestCase("ErrorLevel", func() *zap.Config {
			cfg := zap.NewProductionConfig()
			cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
			return &cfg
		}(), slog.Error),
	}
}

func TestGetConfigLevel(t *testing.T) {
	core.RunTestCases(t, getConfigLevelTestCases())
}
