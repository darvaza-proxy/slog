package filter_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
	"darvaza.org/slog/handlers/mock"
	slogtest "darvaza.org/slog/internal/testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = invalidLevelTestCase{}
var _ core.TestCase = withStackDisabledTestCase{}
var _ core.TestCase = withLevelDisabledTestCase{}
var _ core.TestCase = filterNewEdgeTestCase{}

// invalidLevelTestCase tests panic on invalid log level
type invalidLevelTestCase struct {
	invalidLevel slog.LogLevel
	name         string
}

func (tc invalidLevelTestCase) Name() string {
	return tc.name
}

func (tc invalidLevelTestCase) Test(t *testing.T) {
	t.Helper()

	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Info)

	// Invalid level should cause panic
	core.AssertPanic(t, func() {
		logger.WithLevel(tc.invalidLevel)
	}, nil, "invalid level panic")

	// Also test with LogEntry.WithLevel
	entry := logger.Info()
	core.AssertPanic(t, func() {
		entry.WithLevel(tc.invalidLevel)
	}, nil, "entry invalid level panic")
}

// Factory function for invalidLevelTestCase
func newInvalidLevelTestCase(name string, invalidLevel slog.LogLevel) invalidLevelTestCase {
	return invalidLevelTestCase{
		name:         name,
		invalidLevel: invalidLevel,
	}
}

// withStackDisabledTestCase tests WithStack on disabled entries
type withStackDisabledTestCase struct {
	entryLevel slog.LogLevel
	threshold  slog.LogLevel
	name       string
}

func (tc withStackDisabledTestCase) Name() string {
	return tc.name
}

func (tc withStackDisabledTestCase) Test(t *testing.T) {
	t.Helper()

	parent := mock.NewLogger()
	logger := filter.New(parent, tc.threshold)

	// Create disabled entry
	entry := logger.WithLevel(tc.entryLevel)
	core.AssertFalse(t, entry.Enabled(), "entry should be disabled")

	// WithStack on disabled entry should return same instance
	withStack := entry.WithStack(1)
	core.AssertSame(t, entry, withStack, "disabled WithStack returns same")

	// Verify no logging occurs
	withStack.Print("should not log")
	messages := parent.GetMessages()
	slogtest.AssertMessageCount(t, messages, 0)
}

// Factory function for withStackDisabledTestCase
func newWithStackDisabledTestCase(name string, entryLevel, threshold slog.LogLevel) withStackDisabledTestCase {
	return withStackDisabledTestCase{
		name:       name,
		entryLevel: entryLevel,
		threshold:  threshold,
	}
}

// withLevelDisabledTestCase tests WithLevel on disabled logger
type withLevelDisabledTestCase struct {
	level     slog.LogLevel
	threshold slog.LogLevel
	name      string
}

func (tc withLevelDisabledTestCase) Name() string {
	return tc.name
}

func (tc withLevelDisabledTestCase) Test(t *testing.T) {
	t.Helper()

	parent := mock.NewLogger()
	logger := NewTestFilter(parent, tc.threshold)

	// If threshold is UndefinedLevel, WithLevel should panic due to invalid logger
	if tc.threshold == slog.UndefinedLevel {
		core.AssertPanic(t, func() {
			logger.WithLevel(tc.level)
		}, nil, "invalid logger panics")
		return
	}

	// WithLevel on logger with disabled level
	withLevel := logger.WithLevel(tc.level)

	// In slog: Debug=6, Info=5, Warn=4, Error=3, Fatal=2, Panic=1
	// A level is enabled if level <= threshold
	if tc.level <= tc.threshold {
		// Level is enabled, should get new LogEntry
		core.AssertNotSame(t, logger, withLevel, "enabled WithLevel returns new")
		core.AssertTrue(t, withLevel.Enabled(), "enabled level")
	} else {
		// Level is disabled
		core.AssertFalse(t, withLevel.Enabled(), "disabled level")
	}

	// WithStack on logger should also handle disabled case
	withStack := logger.WithStack(1)
	if tc.threshold >= slog.Panic {
		// Some level is enabled, should get new instance
		core.AssertNotSame(t, logger, withStack, "WithStack returns new")
	} else {
		// All levels disabled (threshold < Panic), should return same
		core.AssertSame(t, logger, withStack, "disabled WithStack returns same")
	}
}

// Factory function for withLevelDisabledTestCase
func newWithLevelDisabledTestCase(name string, level, threshold slog.LogLevel) withLevelDisabledTestCase {
	return withLevelDisabledTestCase{
		name:      name,
		level:     level,
		threshold: threshold,
	}
}

// filterNewEdgeTestCase tests edge cases in filter.New
type filterNewEdgeTestCase struct {
	name       string
	parent     slog.Logger
	threshold  slog.LogLevel
	expectNoop bool
}

func (tc filterNewEdgeTestCase) Name() string {
	return tc.name
}

func (tc filterNewEdgeTestCase) Test(t *testing.T) {
	t.Helper()

	logger := filter.New(tc.parent, tc.threshold)
	core.AssertNotNil(t, logger, "logger created")

	// Test behaviour based on parent
	if tc.expectNoop {
		// Should behave like noop
		entry := logger.Info()
		core.AssertFalse(t, entry.Enabled(), "noop info disabled")

		fatal := logger.Fatal()
		core.AssertTrue(t, fatal.Enabled(), "noop fatal enabled")
	} else {
		// Normal behaviour: level is enabled if level <= threshold
		entry := logger.Info()
		enabled := slog.Info <= tc.threshold
		core.AssertEqual(t, enabled, entry.Enabled(), "normal enabled state")
	}
}

// Factory function for filterNewEdgeTestCase
func newFilterNewEdgeTestCase(name string, parent slog.Logger,
	threshold slog.LogLevel, expectNoop bool) filterNewEdgeTestCase {
	return filterNewEdgeTestCase{
		name:       name,
		parent:     parent,
		threshold:  threshold,
		expectNoop: expectNoop,
	}
}

// TestInvalidLogLevel tests panic on invalid log levels
func TestInvalidLogLevel(t *testing.T) {
	testCases := []invalidLevelTestCase{
		newInvalidLevelTestCase("UndefinedLevel", slog.UndefinedLevel),
		newInvalidLevelTestCase("Negative level", slog.LogLevel(-1)),
		newInvalidLevelTestCase("Very negative level", slog.LogLevel(-100)),
	}

	core.RunTestCases(t, testCases)
}

// TestWithStackDisabled tests WithStack behaviour on disabled entries
func TestWithStackDisabled(t *testing.T) {
	testCases := []withStackDisabledTestCase{
		// Debug=6 > Info=5, so Debug is disabled when threshold is Info
		newWithStackDisabledTestCase("Debug entry, Info threshold",
			slog.Debug, slog.Info),
		// Info=5 > Warn=4, so Info is disabled when threshold is Warn
		newWithStackDisabledTestCase("Info entry, Warn threshold",
			slog.Info, slog.Warn),
		// Warn=4 > Error=3, so Warn is disabled when threshold is Error
		newWithStackDisabledTestCase("Warn entry, Error threshold",
			slog.Warn, slog.Error),
		// Error=3 > Fatal=2, so Error is disabled when threshold is Fatal
		newWithStackDisabledTestCase("Error entry, Fatal threshold",
			slog.Error, slog.Fatal),
		// Fatal=2 > Panic=1, so Fatal is disabled when threshold is Panic
		newWithStackDisabledTestCase("Fatal entry, Panic threshold",
			slog.Fatal, slog.Panic),
	}

	core.RunTestCases(t, testCases)
}

// TestWithLevelDisabled tests WithLevel behaviour with disabled levels
func TestWithLevelDisabled(t *testing.T) {
	testCases := []withLevelDisabledTestCase{
		// Debug=6 > Info=5, so Debug is disabled
		newWithLevelDisabledTestCase("Debug level, Info threshold",
			slog.Debug, slog.Info),
		// Info=5 > Warn=4, so Info is disabled
		newWithLevelDisabledTestCase("Info level, Warn threshold",
			slog.Info, slog.Warn),
		// Warn=4 > Error=3, so Warn is disabled
		newWithLevelDisabledTestCase("Warn level, Error threshold",
			slog.Warn, slog.Error),
		// Error=3 > Fatal=2, so Error is disabled
		newWithLevelDisabledTestCase("Error level, Fatal threshold",
			slog.Error, slog.Fatal),
		// Fatal=2 > Panic=1, so Fatal is disabled
		newWithLevelDisabledTestCase("Fatal level, Panic threshold",
			slog.Fatal, slog.Panic),
		// Test with threshold below all valid levels
		newWithLevelDisabledTestCase("Debug level, UndefinedLevel threshold",
			slog.Debug, slog.UndefinedLevel),
	}

	core.RunTestCases(t, testCases)
}

// TestFilterNewEdgeCases tests edge cases in filter.New
func TestFilterNewEdgeCases(t *testing.T) {
	testCases := []filterNewEdgeTestCase{
		newFilterNewEdgeTestCase("nil parent, Debug threshold",
			nil, slog.Debug, true),
		newFilterNewEdgeTestCase("nil parent, Info threshold",
			nil, slog.Info, true),
		newFilterNewEdgeTestCase("nil parent, Fatal threshold",
			nil, slog.Fatal, true),
		newFilterNewEdgeTestCase("valid parent, Debug threshold",
			mock.NewLogger(), slog.Debug, false),
		newFilterNewEdgeTestCase("valid parent, Panic threshold",
			mock.NewLogger(), slog.Panic, false),
	}

	core.RunTestCases(t, testCases)
}

// TestEntryMsgPanicPath tests the panic path in entry.msg()
func TestEntryMsgPanicPath(t *testing.T) {
	// This test covers the panic path when parentless Panic level is logged
	logger := filter.NewNoop()
	entry := logger.Panic()

	core.AssertPanic(t, func() {
		// This triggers the panic path in entry.msg()
		entry.Print("panic message")
	}, nil, "parentless panic")

	// Also test with Printf and Println
	core.AssertPanic(t, func() {
		logger.Panic().Printf("panic %s", "formatted")
	}, nil, "parentless panic printf")

	core.AssertPanic(t, func() {
		logger.Panic().Println("panic", "line")
	}, nil, "parentless panic println")
}

// TestEntryEnabledEdgeCases tests edge cases in entry.Enabled()
func TestEntryEnabledEdgeCases(t *testing.T) {
	t.Run("entry.Enabled with various thresholds", runTestEntryEnabledWithVariousThresholds)

	t.Run("entry.isEnabled edge cases", runTestEntryIsEnabledEdgeCases)
}

// TestWithStackEnabled tests WithStack behaviour on enabled entries
func TestWithStackEnabled(t *testing.T) {
	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Info)

	// Create enabled entry
	entry := logger.Info()
	core.AssertTrue(t, entry.Enabled(), "entry should be enabled")

	// WithStack on enabled entry should return new instance
	withStack := entry.WithStack(1)
	core.AssertNotSame(t, entry, withStack, "enabled WithStack returns new")

	// Log and verify stack is included
	withStack.Print("message with stack")
	messages := parent.GetMessages()
	slogtest.AssertMessageCount(t, messages, 1)
	// Note: Stack trace verification would depend on internal implementation
}

// TestLoggerWithStackDisabled tests Logger.WithStack with all levels disabled
func TestLoggerWithStackDisabled(t *testing.T) {
	parent := mock.NewLogger()
	// Set threshold to UndefinedLevel (0) so nothing is enabled
	logger := NewTestFilter(parent, slog.UndefinedLevel)

	// WithStack should return same instance when nothing is enabled
	withStack := logger.WithStack(1)
	core.AssertSame(t, logger, withStack, "all disabled WithStack returns same")
}

// TestLoggerWithLevelEnabled tests Logger.WithLevel with enabled levels
func TestLoggerWithLevelEnabled(t *testing.T) {
	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Debug)

	// WithLevel with enabled level should return new LogEntry
	entry := logger.WithLevel(slog.Info)
	core.AssertNotSame(t, logger, entry, "enabled WithLevel returns new")
	core.AssertTrue(t, entry.Enabled(), "entry is enabled")

	// Log to verify it works
	entry.Print("test message")
	messages := parent.GetMessages()
	slogtest.AssertMessageCount(t, messages, 1)
}

func runTestEntryEnabledWithVariousThresholds(t *testing.T) {
	t.Helper()
	parent := mock.NewLogger()

	// Test with each threshold level
	for threshold := slog.Panic; threshold <= slog.Debug; threshold++ {
		logger := filter.New(parent, threshold)

		// Test each entry level
		for level := slog.Panic; level <= slog.Debug; level++ {
			entry := logger.WithLevel(level)
			expected := level <= threshold
			core.AssertEqual(t, expected, entry.Enabled(),
				"level %v vs threshold %v", level, threshold)
		}
	}
}

func runTestEntryIsEnabledEdgeCases(t *testing.T) {
	t.Helper()
	// Test parentless Fatal and Panic
	noop := filter.NewNoop()

	fatal := noop.Fatal()
	core.AssertTrue(t, fatal.Enabled(), "parentless fatal enabled")

	panicEntry := noop.Panic()
	core.AssertTrue(t, panicEntry.Enabled(), "parentless panic enabled")

	// Other levels should be disabled
	debug := noop.Debug()
	core.AssertFalse(t, debug.Enabled(), "parentless debug disabled")

	info := noop.Info()
	core.AssertFalse(t, info.Enabled(), "parentless info disabled")
}
