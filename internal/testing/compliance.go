// Package testing provides shared test utilities for slog handler testing.
package testing

import (
	"testing"

	"darvaza.org/slog"
)

// ComplianceTest runs a comprehensive test suite to verify that a logger
// implementation correctly implements the slog.Logger interface.
type ComplianceTest struct {
	AdapterOptions
	FactoryOptions

	// SkipEnabledTests skips tests that require checking the Enabled state.
	SkipEnabledTests bool

	// SkipPanicTests skips tests that would cause panics.
	SkipPanicTests bool
}

// Run executes the full compliance test suite.
func (ct ComplianceTest) Run(t *testing.T) {
	t.Run("Interface", func(t *testing.T) {
		logger := ct.NewLogger()

		// Verify it's not nil
		if logger == nil {
			t.Fatal("NewLogger returned nil")
		}

		// Verify it implements slog.Logger
		var _ slog.Logger = logger
	})

	t.Run("LevelMethods", ct.testLevelMethods)

	t.Run("FieldMethods", ct.testFieldMethods)

	t.Run("PrintMethods", ct.testPrintMethods)

	if !ct.SkipEnabledTests {
		t.Run("EnabledMethod", ct.testEnabledMethod)
	}

	t.Run("WithStack", ct.testWithStack)

	t.Run("Immutability", ct.testImmutability)

	t.Run("Concurrency", ct.testConcurrency)
}

func (ct ComplianceTest) testLevelMethods(t *testing.T) {
	t.Helper()
	levels := logLevels()

	for _, tc := range levels {
		t.Run(tc.name, func(t *testing.T) {
			ct.testSingleLevelMethod(t, tc.method, tc.level)
		})
	}
}

// testSingleLevelMethod tests a single level method in compliance test.
func (ct ComplianceTest) testSingleLevelMethod(t *testing.T,
	method func(slog.Logger) slog.Logger, level slog.LogLevel) {
	t.Helper()

	if ct.SkipPanicTests && (level == slog.Panic || level == slog.Fatal) {
		t.Skip("Skipping panic/fatal test")
	}

	logger := ct.NewLogger()
	levelLogger := method(logger)

	// Test that we get a logger back
	if levelLogger == nil {
		t.Fatal("level method returned nil")
	}

	// Test method chaining
	chained := levelLogger.WithField("test", "value")
	if chained == nil {
		t.Fatal("chained level method returned nil")
	}
}

func (ct ComplianceTest) testFieldMethods(t *testing.T) {
	t.Helper()
	RunWithLogger(t, "WithField", ct.NewLogger(), TestWithField)

	RunWithLogger(t, "WithFields", ct.NewLogger(), TestWithFields)
}

func (ct ComplianceTest) testPrintMethods(t *testing.T) {
	t.Helper()
	tests := []struct {
		name   string
		method func(slog.Logger, ...any)
	}{
		{
			name: "Print",
			method: func(l slog.Logger, args ...any) {
				l.Print(args...)
			},
		},
		{
			name: "Println",
			method: func(l slog.Logger, args ...any) {
				l.Println(args...)
			},
		},
		{
			name: "Printf",
			method: func(l slog.Logger, args ...any) {
				if len(args) > 0 {
					l.Printf("%v", args[0])
				} else {
					l.Printf("test")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ct.testPrintMethod(t, tc.method)
		})
	}
}

// testPrintMethod tests a single print method with various arguments
func (ct ComplianceTest) testPrintMethod(t *testing.T, method func(slog.Logger, ...any)) {
	t.Helper()
	logger := ct.NewLogger()

	// Test with no args
	method(logger.Info())

	// Test with single arg
	method(logger.Info(), "test")

	// Test with multiple args
	method(logger.Info(), "test", 123, true)

	// Test with nil arg
	method(logger.Info(), nil)
}

func (ct ComplianceTest) testEnabledMethod(t *testing.T) {
	t.Helper()
	t.Run("Enabled", func(_ *testing.T) {
		logger := ct.NewLogger()

		// Test Enabled method exists and returns bool
		_ = logger.Enabled()
	})

	t.Run("WithEnabled", func(t *testing.T) {
		logger := ct.NewLogger()

		// Test WithEnabled returns logger and bool
		l, enabled := logger.WithEnabled()
		if l == nil {
			t.Error("WithEnabled returned nil logger")
		}

		// If logger is enabled, the returned logger should be the same
		if enabled && l != logger {
			t.Error("WithEnabled should return same logger when enabled")
		}
	})
}

func (ct ComplianceTest) testWithStack(t *testing.T) {
	t.Helper()
	TestWithStack(t, ct.NewLogger())
}

func (ct ComplianceTest) testImmutability(t *testing.T) {
	t.Helper()
	base := ct.NewLogger()

	// Add fields to a logger
	l1 := base.WithField("field1", "value1")
	l2 := base.WithField("field2", "value2")

	// l1 and l2 should be independent
	if l1 == l2 {
		t.Error("WithField should return new logger instances")
	}

	// Original should be unchanged
	if base == l1 || base == l2 {
		t.Error("WithField should not modify original logger")
	}

	// Test with levels
	l3 := base.Info()
	l4 := base.Debug()

	if l3 == l4 {
		t.Error("level methods should return new logger instances")
	}
}

func (ct ComplianceTest) testConcurrency(t *testing.T) {
	t.Helper()

	if ct.NewLoggerWithRecorder != nil {
		// Use factory pattern for adapters
		opts := &ConcurrencyTestOptions{
			AdapterOptions: ct.AdapterOptions,
			FactoryOptions: ct.FactoryOptions,
		}
		RunConcurrentTestWithOptions(t, nil, DefaultConcurrencyTest(), opts)
	} else {
		// Use direct logger for simple implementations
		logger := ct.NewLogger()
		RunConcurrentTest(t, logger, DefaultConcurrencyTest())
	}

	// Test concurrent field operations
	TestConcurrentFields(t, ct.NewLogger)
}
