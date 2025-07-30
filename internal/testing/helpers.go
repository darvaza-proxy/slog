package testing

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// AssertMessage verifies that a message matches expected properties.
func AssertMessage(t *testing.T, msg Message, level slog.LogLevel, text string) {
	t.Helper()
	if msg.Level != level {
		t.Errorf("expected level %v, got %v", level, msg.Level)
	}
	if msg.Message != text {
		t.Errorf("expected message %q, got %q", text, msg.Message)
	}
}

// AssertField verifies that a message contains a field with the expected value.
func AssertField(t *testing.T, msg Message, key string, value any) {
	t.Helper()
	got, exists := msg.Fields[key]
	if !exists {
		t.Errorf("expected field %q not found", key)
		return
	}
	if got != value {
		t.Errorf("field %q: expected %v, got %v", key, value, got)
	}
}

// AssertNoField verifies that a message does not contain a specific field.
func AssertNoField(t *testing.T, msg Message, key string) {
	t.Helper()
	if value, exists := msg.Fields[key]; exists {
		t.Errorf("unexpected field %q with value %v", key, value)
	}
}

// AssertMessageCount verifies the expected number of messages were recorded.
func AssertMessageCount(t *testing.T, messages []Message, expected int) {
	t.Helper()
	if len(messages) != expected {
		t.Errorf("expected %d messages, got %d", expected, len(messages))
		for i, msg := range messages {
			t.Logf("  [%d] level=%v, message=%q", i, msg.Level, msg.Message)
		}
	}
}

// RunWithLogger is a helper that runs a test function with a given logger instance.
func RunWithLogger(t *testing.T, name string, logger slog.Logger, fn func(*testing.T, slog.Logger)) {
	t.Run(name, func(t *testing.T) {
		t.Helper()
		fn(t, logger)
	})
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
func RunWithLoggerFactory(t *testing.T, name string, newLogger func() slog.Logger, fn func(*testing.T, slog.Logger)) {
	t.Run(name, func(t *testing.T) {
		t.Helper()
		fn(t, newLogger())
	})
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
func TestLevelMethods(t *testing.T, newLogger func() slog.Logger) {
	levels := logLevels()

	for _, tc := range levels {
		t.Run(tc.name, func(t *testing.T) {
			testLevelMethod(t, newLogger, tc.method, tc.level)
		})
	}
}

// testLevelMethod tests a single level method.
func testLevelMethod(t *testing.T, newLogger func() slog.Logger,
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
func TestFieldMethods(t *testing.T, newLogger func() slog.Logger) {
	RunWithLogger(t, "WithField", newLogger(), TestWithField)
	RunWithLogger(t, "WithFields", newLogger(), TestWithFields)
}

// TestWithField tests the WithField method.
func TestWithField(t *testing.T, logger slog.Logger) {
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
func TestWithFields(t *testing.T, logger slog.Logger) {
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
func TestWithStack(t *testing.T, logger slog.Logger) {
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
