package filter_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
	"darvaza.org/slog/handlers/mock"
	slogtest "darvaza.org/slog/internal/testing"
)

// TestFilterCompliance runs the comprehensive compliance test suite for the filter handler.
func TestFilterCompliance(t *testing.T) {
	t.Run("StandardFilter", runTestStandardFilterCompliance)
	t.Run("FilterWithThresholds", runTestFilterWithThresholds)
	t.Run("FilterWithCustomFilters", runTestFilterWithCustomFilters)
	t.Run("NoopFilter", runTestNoopFilterCompliance)
}

func runTestStandardFilterCompliance(t *testing.T) {
	t.Helper()

	compliance := slogtest.ComplianceTest{
		FactoryOptions: slogtest.FactoryOptions{
			NewLogger: func() slog.Logger {
				// Create filter with mock backend at Debug level (all messages pass)
				backend := mock.NewLogger()
				return filter.New(backend, slog.Debug)
			},
			NewLoggerWithRecorder: func(recorder slog.Logger) slog.Logger {
				// Create filter that writes to the given recorder
				return filter.New(recorder, slog.Debug)
			},
		},
		// Filter with parent should support all features
		SkipPanicTests:   false,
		SkipEnabledTests: false,
	}

	compliance.Run(t)
}

func runTestFilterWithThresholds(t *testing.T) {
	t.Helper()

	thresholds := []struct {
		name      string
		threshold slog.LogLevel
	}{
		{"Info", slog.Info},
		{"Warn", slog.Warn},
		{"Error", slog.Error},
	}

	for _, tc := range thresholds {
		t.Run(tc.name, func(t *testing.T) {
			runTestFilterThresholdCompliance(t, tc.threshold)
		})
	}
}

func runTestFilterWithCustomFilters(t *testing.T) {
	t.Helper()

	compliance := slogtest.ComplianceTest{
		FactoryOptions: slogtest.FactoryOptions{
			NewLogger: func() slog.Logger {
				backend := mock.NewLogger()
				logger := &filter.Logger{
					Parent:    backend,
					Threshold: slog.Debug,
					// Filter out fields starting with underscore
					FieldFilter: func(key string, val any) (string, any, bool) {
						if len(key) > 0 && key[0] == '_' {
							return "", nil, false
						}
						// Transform sensitive fields
						if key == sensitiveKey1 || key == sensitiveKey2 {
							return key, redactedValue, true
						}
						return key, val, true
					},
					// Add prefix to all messages
					MessageFilter: func(msg string) (string, bool) {
						if msg == "" {
							return msg, false // Filter out empty messages
						}
						return "[FILTERED] " + msg, true
					},
				}
				return logger
			},
			NewLoggerWithRecorder: func(recorder slog.Logger) slog.Logger {
				logger := &filter.Logger{
					Parent:    recorder,
					Threshold: slog.Debug,
					FieldFilter: func(key string, val any) (string, any, bool) {
						if len(key) > 0 && key[0] == '_' {
							return "", nil, false
						}
						if key == sensitiveKey1 || key == sensitiveKey2 {
							return key, redactedValue, true
						}
						return key, val, true
					},
					MessageFilter: func(msg string) (string, bool) {
						if msg == "" {
							return msg, false
						}
						return "[FILTERED] " + msg, true
					},
				}
				return logger
			},
		},
		SkipPanicTests:   false,
		SkipEnabledTests: false,
	}

	compliance.Run(t)
}

func runTestNoopFilterCompliance(t *testing.T) {
	t.Helper()

	// Noop filter behaves differently - it returns itself for many operations
	// since it has no parent to write to and doesn't collect fields

	t.Run("BasicFunctionality", runTestNoopBasicFunctionality)
	t.Run("PrintMethodsDoNothing", runTestNoopPrintMethodsDoNothing)
	t.Run("PanicBehaviour", runTestNoopPanicBehaviour)
	t.Run("WithEnabled", runTestNoopWithEnabled)
}

func runTestNoopBasicFunctionality(t *testing.T) {
	t.Helper()
	logger := filter.NewNoop()

	// Test that we get a logger
	core.AssertNotNil(t, logger, "NewNoop")

	// Test level methods return something
	core.AssertNotNil(t, logger.Debug(), "Debug()")
	core.AssertNotNil(t, logger.Info(), "Info()")
	core.AssertNotNil(t, logger.Warn(), "Warn()")
	core.AssertNotNil(t, logger.Error(), "Error()")
	core.AssertNotNil(t, logger.Fatal(), "Fatal()")
	core.AssertNotNil(t, logger.Panic(), "Panic()")

	// Test WithField returns self (no fields collected)
	l1 := logger.WithField("key", "value")
	core.AssertSame(t, logger, l1, "WithField returns self")

	// Test WithFields returns self
	l2 := logger.WithFields(map[string]any{"k": "v"})
	core.AssertSame(t, logger, l2, "WithFields returns self")

	// Test WithStack returns self (no stack collected)
	l3 := logger.WithStack(0)
	core.AssertSame(t, logger, l3, "WithStack returns self")

	// Test Enabled returns false for non-terminal levels
	core.AssertFalse(t, logger.Enabled(), "Logger not enabled")
	core.AssertFalse(t, logger.Debug().Enabled(), "Debug not enabled")
	core.AssertFalse(t, logger.Info().Enabled(), "Info not enabled")
	core.AssertFalse(t, logger.Warn().Enabled(), "Warn not enabled")
	core.AssertFalse(t, logger.Error().Enabled(), "Error not enabled")

	// Fatal and Panic should be enabled (for termination)
	core.AssertTrue(t, logger.Fatal().Enabled(), "Fatal enabled")
	core.AssertTrue(t, logger.Panic().Enabled(), "Panic enabled")
}

func runTestNoopPanicBehaviour(t *testing.T) {
	t.Helper()
	logger := filter.NewNoop()

	// Test that Panic actually panics
	core.AssertPanic(t, func() {
		logger.Panic().Print("test panic")
	}, nil, "Panic() should panic")
}

func runTestNoopWithEnabled(t *testing.T) {
	t.Helper()
	logger := filter.NewNoop()

	// WithEnabled should return self and false for noop
	l, enabled := logger.WithEnabled()
	core.AssertSame(t, logger, l, "WithEnabled returns self")
	core.AssertFalse(t, enabled, "WithEnabled returns false")

	// WithEnabled on Fatal should return entry and true
	fatalEntry := logger.Fatal()
	l2, enabled2 := fatalEntry.WithEnabled()
	core.AssertSame(t, fatalEntry, l2, "Fatal WithEnabled returns self")
	core.AssertTrue(t, enabled2, "Fatal WithEnabled returns true")
}

func runTestFilterThresholdCompliance(t *testing.T, threshold slog.LogLevel) {
	t.Helper()

	compliance := slogtest.ComplianceTest{
		FactoryOptions: slogtest.FactoryOptions{
			NewLogger: func() slog.Logger {
				backend := mock.NewLogger()
				return filter.New(backend, threshold)
			},
			NewLoggerWithRecorder: func(recorder slog.Logger) slog.Logger {
				return filter.New(recorder, threshold)
			},
		},
		// The filter still supports all operations, just filters some messages
		SkipPanicTests:   false,
		SkipEnabledTests: false,
	}

	// Run the full compliance suite - the filter should handle all cases
	compliance.Run(t)
}

func runTestNoopPrintMethodsDoNothing(t *testing.T) {
	t.Helper()
	logger := filter.NewNoop()
	core.AssertNotNil(t, logger, "logger is not nil")
	core.AssertNil(t, logger.Parent, "logger.Parent is not nil")
	core.AssertEqual(t, slog.Fatal, logger.Threshold, "Threshold level is Fatal")

	core.AssertNoPanic(t, func() {
		// These should not panic, just do nothing
		logger.Print("test")
		logger.Println("test")
		logger.Printf("test %s", "value")

		// Level methods print should also not panic (except Fatal/Panic)
		logger.Debug().Print("debug")
		logger.Info().Print("info")
		logger.Warn().Print("warn")
		logger.Error().Print("error")
	}, "Noop does nothing")
}
