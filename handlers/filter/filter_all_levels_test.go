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
var _ core.TestCase = allLevelsTestCase{}
var _ core.TestCase = levelMethodTestCase{}

// allLevelsTestCase tests all log level methods on LogEntry
type allLevelsTestCase struct {
	method      func(slog.Logger) slog.Logger
	methodName  string
	name        string
	methodLevel slog.LogLevel
	threshold   slog.LogLevel
	shouldLog   bool
}

func (tc allLevelsTestCase) Name() string {
	return tc.name
}

func (tc allLevelsTestCase) Test(t *testing.T) {
	t.Helper()

	parent := mock.NewLogger()
	logger := filter.New(parent, tc.threshold)

	// Start with an entry at the highest enabled level to ensure it's enabled
	// This way level methods will actually create new entries
	var entry slog.Logger
	if tc.threshold >= slog.Debug {
		entry = logger.Debug()
	} else if tc.threshold >= slog.Info {
		entry = logger.Info()
	} else if tc.threshold >= slog.Warn {
		entry = logger.Warn()
	} else if tc.threshold >= slog.Error {
		entry = logger.Error()
	} else if tc.threshold >= slog.Fatal {
		entry = logger.Fatal()
	} else {
		entry = logger.Panic()
	}

	// Call the level method using the function pointer
	levelEntry := tc.method(entry)

	// Verify the level was set correctly
	// We need to cast to *filter.LogEntry to access Level() method for testing
	logEntry := core.AssertMustTypeIs[*filter.LogEntry](t, levelEntry, "entry type")

	// Get the starting entry's level for comparison
	startEntry := core.AssertMustTypeIs[*filter.LogEntry](t, entry, "start entry type")
	startLevel := startEntry.Level()

	// Level methods should ALWAYS create an entry at the requested level
	// This maintains semantic correctness: .Debug() creates a Debug entry
	core.AssertEqual(t, tc.methodLevel, logEntry.Level(), "entry level")

	// Verify immutability: level methods should create new instances when changing level
	// but reuse the same instance when the level doesn't change (optimization)
	if startLevel == tc.methodLevel {
		core.AssertSame(t, entry, levelEntry, "same level returns same instance")
	} else {
		core.AssertMustNotSame(t, entry, levelEntry, "different level creates new instance")
	}

	// Test logging
	levelEntry.Print("test message")

	messages := parent.GetMessages()
	if tc.shouldLog {
		slogtest.AssertMessageCount(t, messages, 1)
		if len(messages) > 0 {
			// The logged level should match what was actually set
			slogtest.AssertMessage(t, messages[0], logEntry.Level(), "test message")
		}
	} else {
		slogtest.AssertMessageCount(t, messages, 0)
	}
}

// Factory function for allLevelsTestCase
func newAllLevelsTestCase(name, methodName string,
	method func(slog.Logger) slog.Logger,
	methodLevel, threshold slog.LogLevel,
	shouldLog bool) allLevelsTestCase {
	return allLevelsTestCase{
		name:        name,
		methodName:  methodName,
		method:      method,
		methodLevel: methodLevel,
		threshold:   threshold,
		shouldLog:   shouldLog,
	}
}

// levelMethodTestCase tests level methods with various configurations
type levelMethodTestCase struct {
	method    func(slog.Logger) slog.Logger
	name      string
	level     slog.LogLevel
	threshold slog.LogLevel
	hasParent bool
	expectLog bool
}

func (tc levelMethodTestCase) Name() string {
	return tc.name
}

func (tc levelMethodTestCase) Test(t *testing.T) {
	t.Helper()

	var logger slog.Logger
	var parent *mock.Logger

	if tc.hasParent {
		parent = mock.NewLogger()
		logger = filter.New(parent, tc.threshold)
	} else {
		logger = filter.NewNoop()
	}

	// Special handling for Panic without parent
	if tc.level == slog.Panic && !tc.hasParent && tc.expectLog {
		// Panic with no parent should panic
		core.AssertPanic(t, func() {
			tc.method(logger).Print("test panic")
		}, nil, "panic without parent")
		return
	}

	// Test the level method using the function pointer
	entry := tc.method(logger)

	// Verify level is set
	// All level methods should create a LogEntry with the appropriate level
	logEntry := core.AssertMustTypeIs[*filter.LogEntry](t, entry, "entry type")
	core.AssertEqual(t, tc.level, logEntry.Level(), "entry level")

	// Test enabled state
	core.AssertEqual(t, tc.expectLog, entry.Enabled(), "enabled state")

	// Test logging (except for Fatal which would exit)
	if tc.level != slog.Fatal {
		entry.Print("test message")

		if tc.hasParent {
			messages := parent.GetMessages()
			if tc.expectLog {
				slogtest.AssertMessageCount(t, messages, 1)
				if len(messages) > 0 {
					slogtest.AssertMessage(t, messages[0], tc.level, "test message")
				}
			} else {
				slogtest.AssertMessageCount(t, messages, 0)
			}
		}
	}
}

// Factory function for levelMethodTestCase
func newLevelMethodTestCase(name string,
	method func(slog.Logger) slog.Logger,
	level, threshold slog.LogLevel,
	hasParent, expectLog bool) levelMethodTestCase {
	return levelMethodTestCase{
		name:      name,
		method:    method,
		level:     level,
		threshold: threshold,
		hasParent: hasParent,
		expectLog: expectLog,
	}
}

// TestAllLogEntryLevelMethods tests Debug, Warn, Fatal, Panic methods on LogEntry
func TestAllLogEntryLevelMethods(t *testing.T) {
	testCases := []allLevelsTestCase{
		// Debug method
		newAllLevelsTestCase("Debug from Info, Debug threshold",
			"Debug", (slog.Logger).Debug, slog.Debug, slog.Debug, true),
		newAllLevelsTestCase("Debug from Info, Info threshold",
			"Debug", (slog.Logger).Debug, slog.Debug, slog.Info, false),
		newAllLevelsTestCase("Debug from Error, Debug threshold",
			"Debug", (slog.Logger).Debug, slog.Debug, slog.Debug, true),

		// Warn method
		newAllLevelsTestCase("Warn from Info, Debug threshold",
			"Warn", (slog.Logger).Warn, slog.Warn, slog.Debug, true),
		newAllLevelsTestCase("Warn from Info, Warn threshold",
			"Warn", (slog.Logger).Warn, slog.Warn, slog.Warn, true),
		newAllLevelsTestCase("Warn from Info, Error threshold",
			"Warn", (slog.Logger).Warn, slog.Warn, slog.Error, false),
		newAllLevelsTestCase("Warn from Debug, Warn threshold",
			"Warn", (slog.Logger).Warn, slog.Warn, slog.Warn, true),

		// Fatal method
		newAllLevelsTestCase("Fatal from Info, Debug threshold",
			"Fatal", (slog.Logger).Fatal, slog.Fatal, slog.Debug, true),
		newAllLevelsTestCase("Fatal from Info, Fatal threshold",
			"Fatal", (slog.Logger).Fatal, slog.Fatal, slog.Fatal, true),
		newAllLevelsTestCase("Fatal from Info, Panic threshold",
			"Fatal", (slog.Logger).Fatal, slog.Fatal, slog.Panic, false),
		newAllLevelsTestCase("Fatal from Debug, Fatal threshold",
			"Fatal", (slog.Logger).Fatal, slog.Fatal, slog.Fatal, true),

		// Panic method
		newAllLevelsTestCase("Panic from Info, Debug threshold",
			"Panic", (slog.Logger).Panic, slog.Panic, slog.Debug, true),
		newAllLevelsTestCase("Panic from Info, Panic threshold",
			"Panic", (slog.Logger).Panic, slog.Panic, slog.Panic, true),
		newAllLevelsTestCase("Panic from Debug, Panic threshold",
			"Panic", (slog.Logger).Panic, slog.Panic, slog.Panic, true),
		newAllLevelsTestCase("Panic from Error, Debug threshold",
			"Panic", (slog.Logger).Panic, slog.Panic, slog.Debug, true),
	}

	core.RunTestCases(t, testCases)
}

// TestLevelMethodsDirectly tests level methods called directly on logger
func TestLevelMethodsDirectly(t *testing.T) {
	testCases := []levelMethodTestCase{
		// With parent - Debug
		newLevelMethodTestCase("Debug with parent, Debug threshold",
			(slog.Logger).Debug, slog.Debug, slog.Debug, true, true),
		newLevelMethodTestCase("Debug with parent, Info threshold",
			(slog.Logger).Debug, slog.Debug, slog.Info, true, false),

		// With parent - Warn
		newLevelMethodTestCase("Warn with parent, Debug threshold",
			(slog.Logger).Warn, slog.Warn, slog.Debug, true, true),
		newLevelMethodTestCase("Warn with parent, Warn threshold",
			(slog.Logger).Warn, slog.Warn, slog.Warn, true, true),
		newLevelMethodTestCase("Warn with parent, Error threshold",
			(slog.Logger).Warn, slog.Warn, slog.Error, true, false),

		// With parent - Fatal
		newLevelMethodTestCase("Fatal with parent, Fatal threshold",
			(slog.Logger).Fatal, slog.Fatal, slog.Fatal, true, true),
		newLevelMethodTestCase("Fatal with parent, Panic threshold",
			(slog.Logger).Fatal, slog.Fatal, slog.Panic, true, false),

		// With parent - Panic
		newLevelMethodTestCase("Panic with parent, Panic threshold",
			(slog.Logger).Panic, slog.Panic, slog.Panic, true, true),

		// Without parent (noop) - only Fatal and Panic enabled
		newLevelMethodTestCase("Debug without parent",
			(slog.Logger).Debug, slog.Debug, slog.Debug, false, false),
		newLevelMethodTestCase("Warn without parent",
			(slog.Logger).Warn, slog.Warn, slog.Warn, false, false),
		newLevelMethodTestCase("Fatal without parent",
			(slog.Logger).Fatal, slog.Fatal, slog.Fatal, false, true),
		newLevelMethodTestCase("Panic without parent",
			(slog.Logger).Panic, slog.Panic, slog.Panic, false, true),
	}

	core.RunTestCases(t, testCases)
}

// TestLevelMethodChaining tests that level methods preserve field chains
func TestLevelMethodChaining(t *testing.T) {
	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Debug)

	// Start with Info and add fields
	entry := logger.Info().
		WithField("key1", "value1").
		WithField("key2", "value2")

	// Change to Debug level
	debugEntry := entry.Debug()
	e := core.AssertMustTypeIs[*filter.LogEntry](t, debugEntry, "debug entry type")
	core.AssertEqual(t, slog.Debug, e.Level(), "debug level")
	core.AssertMustNotSame(t, entry, debugEntry, "debug creates new instance")

	// Log and verify fields are preserved
	debugEntry.Print("debug message")

	messages := parent.GetMessages()
	slogtest.AssertMessageCount(t, messages, 1)
	if len(messages) > 0 {
		msg := messages[0]
		slogtest.AssertMessage(t, msg, slog.Debug, "debug message")
		slogtest.AssertField(t, msg, "key1", "value1")
		slogtest.AssertField(t, msg, "key2", "value2")
	}

	parent.Clear()

	// Change to Warn level
	warnEntry := entry.Warn()
	e2 := core.AssertMustTypeIs[*filter.LogEntry](t, warnEntry, "warn entry type")
	core.AssertEqual(t, slog.Warn, e2.Level(), "warn level")
	core.AssertMustNotSame(t, entry, warnEntry, "warn creates new instance")
	core.AssertMustNotSame(t, debugEntry, warnEntry, "warn is different from debug")

	warnEntry.Print("warn message")
	messages = parent.GetMessages()
	slogtest.AssertMessageCount(t, messages, 1)
	if len(messages) > 0 {
		msg := messages[0]
		slogtest.AssertMessage(t, msg, slog.Warn, "warn message")
		slogtest.AssertField(t, msg, "key1", "value1")
		slogtest.AssertField(t, msg, "key2", "value2")
	}
}

// TestFatalAndPanicWithNoop tests Fatal and Panic behaviour with no parent
func TestFatalAndPanicWithNoop(t *testing.T) {
	t.Run("Fatal with noop", runTestFatalWithNoop)

	t.Run("Panic with noop", runTestPanicWithNoop)

	t.Run("Panic with fields and noop", runTestPanicWithFieldsAndNoop)
}

// TestLevelTransitions tests transitions between different levels
func TestLevelTransitions(t *testing.T) {
	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Debug)

	// Create entry and transition through levels
	entry1 := logger.Info()
	e1 := core.AssertMustTypeIs[*filter.LogEntry](t, entry1, "info entry")
	core.AssertEqual(t, slog.Info, e1.Level(), "initial info")

	entry2 := entry1.Debug()
	e2 := core.AssertMustTypeIs[*filter.LogEntry](t, entry2, "debug entry")
	core.AssertEqual(t, slog.Debug, e2.Level(), "transition to debug")
	core.AssertMustNotSame(t, entry1, entry2, "debug creates new instance")

	entry3 := entry2.Warn()
	e3 := core.AssertMustTypeIs[*filter.LogEntry](t, entry3, "warn entry")
	core.AssertEqual(t, slog.Warn, e3.Level(), "transition to warn")
	core.AssertMustNotSame(t, entry2, entry3, "warn creates new instance")

	entry4 := entry3.Error()
	e4 := core.AssertMustTypeIs[*filter.LogEntry](t, entry4, "error entry")
	core.AssertEqual(t, slog.Error, e4.Level(), "transition to error")
	core.AssertMustNotSame(t, entry3, entry4, "error creates new instance")

	entry5 := entry4.Fatal()
	e5 := core.AssertMustTypeIs[*filter.LogEntry](t, entry5, "fatal entry")
	core.AssertEqual(t, slog.Fatal, e5.Level(), "transition to fatal")
	core.AssertMustNotSame(t, entry4, entry5, "fatal creates new instance")

	entry6 := entry5.Panic()
	e6 := core.AssertMustTypeIs[*filter.LogEntry](t, entry6, "panic entry")
	core.AssertEqual(t, slog.Panic, e6.Level(), "transition to panic")
	core.AssertMustNotSame(t, entry5, entry6, "panic creates new instance")

	// Back to Info
	entry7 := entry6.Info()
	e7 := core.AssertMustTypeIs[*filter.LogEntry](t, entry7, "back to info entry")
	core.AssertEqual(t, slog.Info, e7.Level(), "back to info")
	core.AssertMustNotSame(t, entry6, entry7, "info creates new instance")
}

func runTestFatalWithNoop(t *testing.T) {
	t.Helper()
	logger := filter.NewNoop()
	entry := logger.Fatal()

	e := core.AssertMustTypeIs[*filter.LogEntry](t, entry, "fatal entry type")
	core.AssertEqual(t, slog.Fatal, e.Level(), "fatal level")
	core.AssertTrue(t, entry.Enabled(), "fatal enabled for termination")

	// Note: We can't test actual Fatal behaviour as it calls os.Exit
}

func runTestPanicWithNoop(t *testing.T) {
	t.Helper()
	logger := filter.NewNoop()
	entry := logger.Panic()

	e := core.AssertMustTypeIs[*filter.LogEntry](t, entry, "panic entry type")
	core.AssertEqual(t, slog.Panic, e.Level(), "panic level")
	core.AssertTrue(t, entry.Enabled(), "panic enabled for termination")

	// Test that it actually panics
	core.AssertPanic(t, func() {
		entry.Print("panic message")
	}, nil, "panic behaviour")
}

func runTestPanicWithFieldsAndNoop(t *testing.T) {
	t.Helper()
	logger := filter.NewNoop()

	core.AssertPanic(t, func() {
		logger.Panic().
			WithField("error", "critical").
			WithField("code", 500).
			Print("system panic")
	}, nil, "panic with fields")
}
