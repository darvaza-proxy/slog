package filter_test

import (
	"fmt"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
	"darvaza.org/slog/handlers/mock"
	slogtest "darvaza.org/slog/internal/testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = parentlessBehaviourTestCase{}
var _ core.TestCase = disabledLoggerTestCase{}

// parentlessBehaviourTestCase tests parentless logger behaviour
type parentlessBehaviourTestCase struct {
	name     string
	level    slog.LogLevel
	expected bool
}

func (tc parentlessBehaviourTestCase) Name() string {
	return tc.name
}

func (tc parentlessBehaviourTestCase) Test(t *testing.T) {
	t.Helper()

	logger := filter.NewNoop()
	entry := logger.WithLevel(tc.level)

	core.AssertEqual(t, tc.expected, entry.Enabled(), "enabled state")

	if tc.level == slog.Panic && tc.expected {
		// Test that panic actually panics
		core.AssertPanic(t, func() {
			entry.Print("test panic")
		}, nil, "panic")

		// Test panic with fields
		core.AssertPanic(t, func() {
			logger.Panic().
				WithField("key", "value").
				Print("panic with fields")
		}, nil, "panic with fields")
	}
	// Note: Fatal behaviour cannot be tested as it calls os.Exit
}

func newParentlessBehaviourTestCase(name string, level slog.LogLevel, expected bool) parentlessBehaviourTestCase {
	return parentlessBehaviourTestCase{
		name:     name,
		level:    level,
		expected: expected,
	}
}

func parentlessBehaviourTestCases() []parentlessBehaviourTestCase {
	return []parentlessBehaviourTestCase{
		newParentlessBehaviourTestCase("Debug disabled", slog.Debug, false),
		newParentlessBehaviourTestCase("Info disabled", slog.Info, false),
		newParentlessBehaviourTestCase("Warn disabled", slog.Warn, false),
		newParentlessBehaviourTestCase("Error disabled", slog.Error, false),
		newParentlessBehaviourTestCase("Fatal enabled", slog.Fatal, true),
		newParentlessBehaviourTestCase("Panic enabled", slog.Panic, true),
	}
}

// disabledLoggerTestCase tests disabled logger behaviour
type disabledLoggerTestCase struct {
	setupFunc     func() slog.Logger
	operationFunc func(slog.Logger) slog.Logger
	name          string
	shouldEnable  bool
}

func (tc disabledLoggerTestCase) Name() string {
	return tc.name
}

func (tc disabledLoggerTestCase) Test(t *testing.T) {
	t.Helper()

	logger := tc.setupFunc()
	result := tc.operationFunc(logger)

	if tc.shouldEnable {
		core.AssertTrue(t, result.Enabled(), "should be enabled")
	} else {
		core.AssertFalse(t, result.Enabled(), "should be disabled")
	}
}

func newDisabledLoggerTestCase(name string, setupFunc func() slog.Logger,
	operationFunc func(slog.Logger) slog.Logger, shouldEnable bool) disabledLoggerTestCase {
	return disabledLoggerTestCase{
		name:          name,
		setupFunc:     setupFunc,
		operationFunc: operationFunc,
		shouldEnable:  shouldEnable,
	}
}

func disabledLoggerTestCases() []disabledLoggerTestCase {
	return []disabledLoggerTestCase{
		newDisabledLoggerTestCase(
			"WithLevel on disabled stays disabled",
			func() slog.Logger {
				base := mock.NewLogger()
				return filter.New(base, slog.Error).Debug() // Debug is disabled
			},
			func(l slog.Logger) slog.Logger {
				return l.WithField("test", "value").WithLevel(slog.Debug) // Still disabled
			},
			false,
		),
		newDisabledLoggerTestCase(
			"WithField on disabled stays disabled",
			func() slog.Logger {
				base := mock.NewLogger()
				return filter.New(base, slog.Error).Info() // Info is disabled
			},
			func(l slog.Logger) slog.Logger {
				return l.WithField("key", "value")
			},
			false,
		),
	}
}

func runTestEmptyFieldKeyIgnored(t *testing.T) {
	t.Helper()

	base := mock.NewLogger()
	logger := filter.New(base, slog.Info)
	entry := logger.Info()

	// Should return same entry
	newEntry := entry.WithField("", "value")
	core.AssertSame(t, entry, newEntry, "empty key returns same entry")

	// Valid field should work
	newEntry = entry.WithField("valid", "included")
	newEntry.Print("test")

	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, 1)
	slogtest.AssertField(t, msgs[0], "valid", "included")
	slogtest.AssertNoField(t, msgs[0], "")
}

func runTestNilMapHandledGracefully(t *testing.T) {
	t.Helper()

	base := mock.NewLogger()
	logger := filter.New(base, slog.Info)
	entry := logger.Info()

	// Should return same entry for nil
	newEntry := entry.WithFields(nil)
	core.AssertSame(t, entry, newEntry, "nil map returns same entry")

	// Empty map also returns same
	newEntry = entry.WithFields(map[string]any{})
	core.AssertSame(t, entry, newEntry, "empty map returns same entry")
}

func runTestNilParentConfiguration(t *testing.T) {
	t.Helper()

	// Create filter with nil parent
	logger := filter.New(nil, slog.Info)

	// Info should be disabled with nil parent
	core.AssertFalse(t, logger.Info().Enabled(), "info disabled with nil parent")
	core.AssertFalse(t, logger.Debug().Enabled(), "debug disabled")

	// Fatal/Panic always enabled for termination
	core.AssertTrue(t, logger.Fatal().Enabled(), "fatal always enabled")
	core.AssertTrue(t, logger.Panic().Enabled(), "panic always enabled")
}

func runTestStackWithVariousSkipValues(t *testing.T) {
	t.Helper()

	base := mock.NewLogger()
	logger := filter.New(base, slog.Info)

	// Negative skip should work (treated as 0)
	entry := logger.Info().WithStack(-1)
	entry.Print("negative skip")

	// Small skip should work
	entry = logger.Info().WithStack(2)
	entry.Print("small skip")

	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, 2)
	core.AssertTrue(t, msgs[0].Stack, "stack present with negative skip")
	core.AssertTrue(t, msgs[1].Stack, "stack present with small skip")
}

func runTestExtremeFieldCounts(t *testing.T) {
	t.Helper()

	base := mock.NewLogger()
	logger := filter.New(base, slog.Debug)
	entry := logger.Info()

	// Add many fields
	const maxFields = 1000
	for i := range maxFields {
		key := fmt.Sprintf("field_%d", i)
		value := fmt.Sprintf("value_%d", i)
		entry = entry.WithField(key, value)
	}

	entry.Print("extreme field count")

	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, 1)
	core.AssertEqual(t, 1000, len(msgs[0].Fields), "field count")
}

func runTestRecursiveFilterScenarios(t *testing.T) {
	t.Helper()

	base := mock.NewLogger()

	// Create nested filter loggers
	filter1 := filter.New(base, slog.Debug)
	filter2 := filter.New(filter1, slog.Info)
	filter3 := filter.New(filter2, slog.Warn)

	// Test that thresholds cascade correctly
	core.AssertFalse(t, filter3.Debug().Enabled(), "debug blocked by all")
	core.AssertFalse(t, filter3.Info().Enabled(), "info blocked by filter3")
	core.AssertTrue(t, filter3.Warn().Enabled(), "warn passes all")
	core.AssertTrue(t, filter3.Error().Enabled(), "error passes all")

	// Log through the chain
	filter3.Error().WithField("nested", "value").Print("nested message")

	// Verify message reached the base
	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, 1)
	slogtest.AssertMessage(t, msgs[0], slog.Error, "nested message")
	slogtest.AssertField(t, msgs[0], "nested", "value")
}

func TestParentlessBehaviour(t *testing.T) {
	core.RunTestCases(t, parentlessBehaviourTestCases())
}

func TestDisabledLogger(t *testing.T) {
	core.RunTestCases(t, disabledLoggerTestCases())
}

func TestEdgeCases(t *testing.T) {
	t.Run("empty field key ignored", runTestEmptyFieldKeyIgnored)
	t.Run("nil map handled gracefully", runTestNilMapHandledGracefully)
	t.Run("nil parent configuration", runTestNilParentConfiguration)
	t.Run("stack with various skip values", runTestStackWithVariousSkipValues)
	t.Run("extreme field counts", runTestExtremeFieldCounts)
	t.Run("recursive filter scenarios", runTestRecursiveFilterScenarios)
}
