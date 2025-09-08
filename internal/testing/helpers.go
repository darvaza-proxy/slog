package testing

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// logLevelInfo holds information about a log level for testing purposes.
type logLevelInfo struct {
	method func(slog.Logger) slog.Logger
	name   string
	level  slog.LogLevel
}

// newLogLevelInfo creates a new log level info instance.
func newLogLevelInfo(name string, method func(slog.Logger) slog.Logger, level slog.LogLevel) logLevelInfo {
	return logLevelInfo{
		name:   name,
		method: method,
		level:  level,
	}
}

// Run executes a test function with the provided name, adapting to different test interfaces.
// It automatically handles both *testing.T and core.MockT implementations.
func Run(t core.T, name string, fn func(core.T)) {
	t.Helper()

	switch tt := t.(type) {
	case interface {
		Run(name string, fn func(*testing.T)) bool
	}:
		tt.Run(name, func(t *testing.T) { fn(t) })
	case interface {
		Run(name string, fn func(core.T)) bool
	}:
		tt.Run(name, fn)
	default:
		fn(t)
	}
}

// AssertMessage verifies that a message matches expected properties.
// Returns true if all assertions pass, false otherwise.
func AssertMessage(t core.T, msg Message, level slog.LogLevel, text string) bool {
	t.Helper()
	ok := core.AssertEqual(t, level, msg.Level, "message level")
	if !core.AssertEqual(t, text, msg.Message, "message text") {
		ok = false
	}
	return ok
}

// AssertMustMessage verifies that a message matches expected properties.
// If the assertion fails, the test is terminated immediately with t.FailNow().
func AssertMustMessage(t core.T, msg Message, level slog.LogLevel, text string) {
	t.Helper()
	if !AssertMessage(t, msg, level, text) {
		t.FailNow()
	}
}

// AssertField verifies that a message contains a field with the expected value.
// Returns true if the field exists and has the expected value, false otherwise.
func AssertField(t core.T, msg Message, key string, value any) bool {
	t.Helper()
	got, exists := msg.Fields[key]
	if !core.AssertTrue(t, exists, "field %q exists", key) {
		return false
	}
	return core.AssertEqual(t, value, got, "field %q value", key)
}

// AssertMustField verifies that a message contains a field with the expected value.
// If the assertion fails, the test is terminated immediately with t.FailNow().
func AssertMustField(t core.T, msg Message, key string, value any) {
	t.Helper()
	if !AssertField(t, msg, key, value) {
		t.FailNow()
	}
}

// AssertNoField verifies that a message does not contain a specific field.
// Returns true if the field does not exist, false if it exists.
func AssertNoField(t core.T, msg Message, key string) bool {
	t.Helper()
	_, exists := msg.Fields[key]
	return core.AssertFalse(t, exists, "field %q should not exist", key)
}

// AssertMustNoField verifies that a message does not contain a specific field.
// If the assertion fails, the test is terminated immediately with t.FailNow().
func AssertMustNoField(t core.T, msg Message, key string) {
	t.Helper()
	if !AssertNoField(t, msg, key) {
		t.FailNow()
	}
}

// AssertMessageCount verifies the expected number of messages were recorded.
// Returns true if the count matches, false otherwise. On failure, logs all messages for debugging.
func AssertMessageCount(t core.T, messages []Message, expected int) bool {
	t.Helper()
	ok := core.AssertEqual(t, expected, len(messages), "message count")
	if !ok {
		for i, msg := range messages {
			t.Logf("  [%d] level=%v, message=%q", i, msg.Level, msg.Message)
		}
	}
	return ok
}

// AssertMustMessageCount verifies the expected number of messages were recorded.
// If the assertion fails, the test is terminated immediately with t.FailNow().
func AssertMustMessageCount(t core.T, messages []Message, expected int) {
	t.Helper()
	if !AssertMessageCount(t, messages, expected) {
		t.FailNow()
	}
}

// RunWithLogger is a helper that runs a test function with a given logger instance.
func RunWithLogger(t core.T, name string, logger slog.Logger, fn func(core.T, slog.Logger)) {
	Run(t, name, func(subT core.T) { fn(subT, logger) })
}

// TransformMessages applies transformations to a slice of messages based on options.
// Returns a new slice with transformed messages.
// Messages that transform to slog.UndefinedLevel are omitted.
func TransformMessages(messages []Message, opts *AdapterOptions) []Message {
	if opts == nil || len(opts.LevelExceptions) == 0 {
		// No transformations needed, just copy the slice
		return core.SliceCopy(messages)
	}

	return core.SliceAsFn(func(msg Message) (Message, bool) {
		expected := opts.ExpectedLevel(msg.Level)
		if expected == slog.UndefinedLevel {
			return msg, false
		}

		out := msg
		out.Level = expected
		return out, true
	}, messages)
}

// CompareMessages compares two message arrays as sets.
// Returns three slices:
// - onlyInFirst: messages that appear only in the first array
// - onlyInSecond: messages that appear only in the second array
// - inBoth: messages that appear in both arrays
func CompareMessages(first, second []Message) (onlyInFirst, onlyInSecond, inBoth []Message) {
	// Use custom equality function based on String() representation
	eq := func(a, b Message) bool {
		return a.String() == b.String()
	}

	// Get unique messages from each set
	firstUnique := core.SliceUniqueFn(first, eq)
	secondUnique := core.SliceUniqueFn(second, eq)

	// Get differences using core utilities
	onlyInFirst = core.SliceMinusFn(firstUnique, secondUnique, eq)
	onlyInSecond = core.SliceMinusFn(secondUnique, firstUnique, eq)

	// Get intersection - need to implement this ourselves
	inBoth = sliceIntersectFn(firstUnique, secondUnique, eq)

	return onlyInFirst, onlyInSecond, inBoth
}

// sliceIntersectFn returns elements that appear in both slices
func sliceIntersectFn[T any](a, b []T, eq func(T, T) bool) []T {
	var result []T
	for _, va := range a {
		if core.SliceContainsFn(b, va, eq) {
			result = append(result, va)
		}
	}
	return result
}

// RunWithLoggerFactory is a helper that runs a test function with a fresh logger instance.
func RunWithLoggerFactory(t core.T, name string, newLogger func() slog.Logger, fn func(core.T, slog.Logger)) {
	Run(t, name, func(subT core.T) { fn(subT, newLogger()) })
}

// logLevels returns all log levels for testing.
func logLevels() []logLevelInfo {
	return []logLevelInfo{
		newLogLevelInfo("Debug", (slog.Logger).Debug, slog.Debug),
		newLogLevelInfo("Info", (slog.Logger).Info, slog.Info),
		newLogLevelInfo("Warn", (slog.Logger).Warn, slog.Warn),
		newLogLevelInfo("Error", (slog.Logger).Error, slog.Error),
		newLogLevelInfo("Fatal", (slog.Logger).Fatal, slog.Fatal),
		newLogLevelInfo("Panic", (slog.Logger).Panic, slog.Panic),
	}
}

// TestLevelMethods runs standard tests for all log level methods.
func TestLevelMethods(t core.T, newLogger func() slog.Logger) {
	levels := logLevels()

	for _, tc := range levels {
		Run(t, tc.name, func(t core.T) {
			testLevelMethod(t, newLogger, tc.method, tc.level)
		})
	}
}

// testLevelMethod tests a single level method.
func testLevelMethod(t core.T, newLogger func() slog.Logger,
	method func(slog.Logger) slog.Logger, expectedLevel slog.LogLevel) {
	t.Helper()
	logger := newLogger()
	levelLogger := method(logger)

	// Test that we get a logger back
	core.AssertMustNotNil(t, levelLogger, "level method")

	// For testable loggers, verify the level
	if tl, ok := logger.(*Logger); ok {
		tl.Clear()
		levelLogger.Print("test")
		msgs := tl.GetMessages()
		if len(msgs) == 1 {
			AssertMessage(t, msgs[0], expectedLevel, "test")
		}
	}
}

// TestFieldMethods runs standard tests for field handling.
func TestFieldMethods(t core.T, newLogger func() slog.Logger) {
	RunWithLogger(t, "WithField", newLogger(), TestWithField)
	RunWithLogger(t, "WithFields", newLogger(), TestWithFields)
}

// TestWithField tests the WithField method.
func TestWithField(t core.T, logger slog.Logger) {
	t.Helper()

	// Test single field
	l1 := logger.WithField("key1", "value1")
	core.AssertMustNotNil(t, l1, "WithField")

	// Test empty key (should return same logger)
	l2 := logger.WithField("", "value")
	core.AssertEqual(t, logger, l2, "WithField empty key")

	// Test nil value (should work)
	l3 := logger.WithField("nil", nil)
	core.AssertMustNotNil(t, l3, "WithField nil value")

	// Test chaining
	l4 := logger.WithField("a", 1).WithField("b", 2).WithField("c", 3)
	core.AssertMustNotNil(t, l4, "WithField chaining")
}

// TestWithFields tests the WithFields method.
func TestWithFields(t core.T, logger slog.Logger) {
	t.Helper()

	// Test multiple fields
	fields := map[string]any{
		"key1": "value1",
		"key2": 42,
		"key3": true,
		"key4": nil,
	}
	l1 := logger.WithFields(fields)
	core.AssertMustNotNil(t, l1, "WithFields")

	// Test empty map (should return same logger)
	l2 := logger.WithFields(nil)
	core.AssertEqual(t, logger, l2, "WithFields nil")

	l3 := logger.WithFields(map[string]any{})
	core.AssertEqual(t, logger, l3, "WithFields empty")

	// Test empty keys are filtered
	l4 := logger.WithFields(map[string]any{
		"":     "ignored",
		"kept": "value",
	})
	core.AssertMustNotNil(t, l4, "WithFields mixed keys")
}

// TestWithStack tests the WithStack method with various skip values.
func TestWithStack(t core.T, logger slog.Logger) {
	t.Helper()

	skipValues := []int{0, 1, 5, -1}

	for _, skip := range skipValues {
		l := logger.WithStack(skip)
		core.AssertMustNotNil(t, l, "WithStack %d", skip)

		// Test chaining
		l2 := l.WithField("test", "value")
		core.AssertMustNotNil(t, l2, "WithStack chaining")
	}
}

// AssertSame verifies that two values are the same instance using reflection.
// This checks pointer equality for reference types and value equality for value types.
//
// Deprecated: Use core.AssertSame instead. This function will be removed in a future version.
func AssertSame(t core.T, expected, actual any, name string, args ...any) bool {
	t.Helper()
	return core.AssertSame(t, expected, actual, name, args...)
}

// AssertNotSame verifies that two values are not the same instance using reflection.
// This checks that values are not pointer-equal for reference types and not value-equal for value types.
//
// Deprecated: Use core.AssertNotSame instead. This function will be removed in a future version.
func AssertNotSame(t core.T, expected, actual any, name string, args ...any) bool {
	t.Helper()
	return core.AssertNotSame(t, expected, actual, name, args...)
}

// AssertMustNotSame verifies that two values are not the same instance using reflection.
// If the assertion fails, the test is terminated immediately with t.FailNow().
//
// Deprecated: Use core.AssertMustNotSame instead. This function will be removed in a future version.
func AssertMustNotSame(t core.T, expected, actual any, name string, args ...any) {
	t.Helper()
	core.AssertMustNotSame(t, expected, actual, name, args...)
}

// AssertMustSame verifies that two values are the same instance using reflection.
// If the assertion fails, the test is terminated immediately with t.FailNow().
//
// Deprecated: Use core.AssertMustSame instead. This function will be removed in a future version.
func AssertMustSame(t core.T, expected, actual any, name string, args ...any) {
	t.Helper()
	core.AssertMustSame(t, expected, actual, name, args...)
}

// IsSame checks if two values are the same instance using reflection.
// Returns true if the values are the same instance for reference types,
// or equal for value types.
//
// Deprecated: Use core.IsSame instead. This function will be removed in a future version.
func IsSame(expected, actual any) bool {
	return core.IsSame(expected, actual)
}
