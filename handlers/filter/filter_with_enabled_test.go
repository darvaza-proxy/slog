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
var _ core.TestCase = withEnabledLoggerTestCase{}
var _ core.TestCase = withEnabledEntryTestCase{}

// withEnabledLoggerTestCase tests Logger.WithEnabled() behaviour
type withEnabledLoggerTestCase struct {
	name      string
	threshold slog.LogLevel
}

func (tc withEnabledLoggerTestCase) Name() string {
	return tc.name
}

func (tc withEnabledLoggerTestCase) Test(t *testing.T) {
	t.Helper()

	parent := mock.NewLogger()
	logger := filter.New(parent, tc.threshold)

	gotLogger, gotEnabled := logger.WithEnabled()

	// Logger.WithEnabled() always returns (self, false)
	core.AssertSame(t, logger, gotLogger, "logger instance")
	core.AssertFalse(t, gotEnabled, "enabled state always false")
}

// Factory function for withEnabledLoggerTestCase
func newWithEnabledLoggerTestCase(name string,
	threshold slog.LogLevel) withEnabledLoggerTestCase {
	return withEnabledLoggerTestCase{
		name:      name,
		threshold: threshold,
	}
}

// withEnabledEntryTestCase tests LogEntry.WithEnabled() behaviour
type withEnabledEntryTestCase struct {
	entryLevel  slog.LogLevel
	threshold   slog.LogLevel
	name        string
	hasParent   bool
	expectState bool
}

func (tc withEnabledEntryTestCase) Name() string {
	return tc.name
}

func (tc withEnabledEntryTestCase) Test(t *testing.T) {
	t.Helper()

	var logger slog.Logger
	if tc.hasParent {
		parent := mock.NewLogger()
		logger = filter.New(parent, tc.threshold)
	} else {
		logger = filter.NewNoop()
	}

	entry := logger.WithLevel(tc.entryLevel)
	gotEntry, gotEnabled := entry.WithEnabled()

	// LogEntry.WithEnabled() always returns same instance
	core.AssertSame(t, entry, gotEntry, "entry instance")
	core.AssertEqual(t, tc.expectState, gotEnabled, "enabled state")
}

// Factory function for withEnabledEntryTestCase
func newWithEnabledEntryTestCase(name string, entryLevel, threshold slog.LogLevel,
	hasParent bool, expectState bool) withEnabledEntryTestCase {
	return withEnabledEntryTestCase{
		name:        name,
		entryLevel:  entryLevel,
		threshold:   threshold,
		hasParent:   hasParent,
		expectState: expectState,
	}
}

// TestLoggerWithEnabled tests Logger.WithEnabled() method
func TestLoggerWithEnabled(t *testing.T) {
	testCases := []withEnabledLoggerTestCase{
		// Logger.WithEnabled() always returns (self, false)
		newWithEnabledLoggerTestCase("Debug threshold", slog.Debug),
		newWithEnabledLoggerTestCase("Info threshold", slog.Info),
		newWithEnabledLoggerTestCase("Warn threshold", slog.Warn),
		newWithEnabledLoggerTestCase("Error threshold", slog.Error),
		newWithEnabledLoggerTestCase("Fatal threshold", slog.Fatal),
		newWithEnabledLoggerTestCase("Panic threshold", slog.Panic),
	}

	core.RunTestCases(t, testCases)
}

// TestLoggerWithEnabledNoop tests Logger.WithEnabled() with no parent
func TestLoggerWithEnabledNoop(t *testing.T) {
	logger := filter.NewNoop()
	gotLogger, gotEnabled := logger.WithEnabled()

	// Noop logger also returns (self, false)
	core.AssertSame(t, logger, gotLogger, "noop logger instance")
	core.AssertFalse(t, gotEnabled, "noop enabled state")
}

// TestEntryWithEnabled tests LogEntry.WithEnabled() method
func TestEntryWithEnabled(t *testing.T) {
	testCases := []withEnabledEntryTestCase{
		// With parent - enabled based on threshold
		newWithEnabledEntryTestCase("Debug entry, Debug threshold",
			slog.Debug, slog.Debug, true, true),
		newWithEnabledEntryTestCase("Debug entry, Info threshold",
			slog.Debug, slog.Info, true, false),
		newWithEnabledEntryTestCase("Info entry, Debug threshold",
			slog.Info, slog.Debug, true, true),
		newWithEnabledEntryTestCase("Info entry, Info threshold",
			slog.Info, slog.Info, true, true),
		newWithEnabledEntryTestCase("Warn entry, Error threshold",
			slog.Warn, slog.Error, true, false),
		newWithEnabledEntryTestCase("Error entry, Error threshold",
			slog.Error, slog.Error, true, true),
		newWithEnabledEntryTestCase("Fatal entry, Panic threshold",
			slog.Fatal, slog.Panic, true, false),
		newWithEnabledEntryTestCase("Panic entry, Panic threshold",
			slog.Panic, slog.Panic, true, true),

		// Without parent (noop) - only Fatal and Panic are enabled
		newWithEnabledEntryTestCase("Debug entry, no parent",
			slog.Debug, slog.Debug, false, false),
		newWithEnabledEntryTestCase("Info entry, no parent",
			slog.Info, slog.Info, false, false),
		newWithEnabledEntryTestCase("Warn entry, no parent",
			slog.Warn, slog.Warn, false, false),
		newWithEnabledEntryTestCase("Error entry, no parent",
			slog.Error, slog.Error, false, false),
		newWithEnabledEntryTestCase("Fatal entry, no parent",
			slog.Fatal, slog.Fatal, false, true),
		newWithEnabledEntryTestCase("Panic entry, no parent",
			slog.Panic, slog.Panic, false, true),
	}

	core.RunTestCases(t, testCases)
}

// TestEntryWithEnabledChaining tests that WithEnabled preserves field chains
func TestEntryWithEnabledChaining(t *testing.T) {
	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Info)

	// Create entry with fields
	entry := logger.Info().
		WithField("key1", "value1").
		WithField("key2", "value2")

	// WithEnabled should preserve the entry and its fields
	gotEntry, gotEnabled := entry.WithEnabled()

	core.AssertSame(t, entry, gotEntry, "chained entry instance")
	core.AssertTrue(t, gotEnabled, "chained entry enabled")

	// Verify fields are preserved by logging
	entry.Print("test message")

	messages := parent.GetMessages()
	slogtest.AssertMessageCount(t, messages, 1)
	if len(messages) > 0 {
		msg := messages[0]
		slogtest.AssertField(t, msg, "key1", "value1")
		slogtest.AssertField(t, msg, "key2", "value2")
	}
}

// TestEntryWithEnabledDisabledLevel tests WithEnabled with disabled level
func TestEntryWithEnabledDisabledLevel(t *testing.T) {
	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Error) // Only Error and above

	// Create Debug entry (disabled)
	entry := logger.Debug()
	gotEntry, gotEnabled := entry.WithEnabled()

	core.AssertSame(t, entry, gotEntry, "disabled entry instance")
	core.AssertFalse(t, gotEnabled, "disabled entry state")

	// Verify no message is logged
	entry.Print("should not appear")
	messages := parent.GetMessages()
	slogtest.AssertMessageCount(t, messages, 0)
}

// TestEntryWithEnabledWithStack tests WithEnabled interaction with WithStack
func TestEntryWithEnabledWithStack(t *testing.T) {
	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Info)

	// Create entry with stack
	entry := logger.Info().WithStack(1)
	gotEntry, gotEnabled := entry.WithEnabled()

	core.AssertSame(t, entry, gotEntry, "entry with stack instance")
	core.AssertTrue(t, gotEnabled, "entry with stack enabled")

	// Test disabled entry with stack attempt
	disabledEntry := logger.Debug().WithStack(1)
	gotDisabled, gotDisabledState := disabledEntry.WithEnabled()

	core.AssertSame(t, disabledEntry, gotDisabled, "disabled with stack instance")
	core.AssertFalse(t, gotDisabledState, "disabled with stack state")
}
