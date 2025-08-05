package testing

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
)

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
	ok := true
	if !core.AssertEqual(t, level, msg.Level, "message level") {
		ok = false
	}
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
func logLevels() []struct {
	name   string
	method func(slog.Logger) slog.Logger
	level  slog.LogLevel
} {
	return []struct {
		name   string
		method func(slog.Logger) slog.Logger
		level  slog.LogLevel
	}{
		{"Debug", func(l slog.Logger) slog.Logger { return l.Debug() }, slog.Debug},
		{"Info", func(l slog.Logger) slog.Logger { return l.Info() }, slog.Info},
		{"Warn", func(l slog.Logger) slog.Logger { return l.Warn() }, slog.Warn},
		{"Error", func(l slog.Logger) slog.Logger { return l.Error() }, slog.Error},
		{"Fatal", func(l slog.Logger) slog.Logger { return l.Fatal() }, slog.Fatal},
		{"Panic", func(l slog.Logger) slog.Logger { return l.Panic() }, slog.Panic},
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
	if levelLogger == nil {
		t.Fatal("level method returned nil")
	}

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
	if l1 == nil {
		t.Fatal("WithField returned nil")
	}

	// Test empty key (should return same logger)
	l2 := logger.WithField("", "value")
	if l2 != logger {
		t.Error("WithField with empty key should return same logger")
	}

	// Test nil value (should work)
	l3 := logger.WithField("nil", nil)
	if l3 == nil {
		t.Fatal("WithField with nil value returned nil")
	}

	// Test chaining
	l4 := logger.WithField("a", 1).WithField("b", 2).WithField("c", 3)
	if l4 == nil {
		t.Fatal("chained WithField returned nil")
	}
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
	if l1 == nil {
		t.Fatal("WithFields returned nil")
	}

	// Test empty map (should return same logger)
	l2 := logger.WithFields(nil)
	if l2 != logger {
		t.Error("WithFields with nil should return same logger")
	}

	l3 := logger.WithFields(map[string]any{})
	if l3 != logger {
		t.Error("WithFields with empty map should return same logger")
	}

	// Test empty keys are filtered
	l4 := logger.WithFields(map[string]any{
		"":     "ignored",
		"kept": "value",
	})
	if l4 == nil {
		t.Fatal("WithFields with mixed keys returned nil")
	}
}

// TestWithStack tests the WithStack method with various skip values.
func TestWithStack(t core.T, logger slog.Logger) {
	t.Helper()

	skipValues := []int{0, 1, 5, -1}

	for _, skip := range skipValues {
		l := logger.WithStack(skip)
		if l == nil {
			t.Fatalf("WithStack(%d) returned nil", skip)
		}

		// Test chaining
		l2 := l.WithField("test", "value")
		if l2 == nil {
			t.Fatal("chaining after WithStack returned nil")
		}
	}
}
