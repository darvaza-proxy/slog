package testing

import (
	"errors"
	"fmt"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// TestBidirectional tests that a bidirectional adapter correctly preserves
// log messages, fields, and levels when round-tripping through the adapter.
// The fn parameter should return a logger that uses the given logger as backend.
func TestBidirectional(t *testing.T, name string, fn func(slog.Logger) slog.Logger) {
	TestBidirectionalWithOptions(t, name, fn, nil)
}

// TestBidirectionalWithOptions tests a bidirectional adapter with custom options.
// This is useful for adapters that have known limitations, such as logr which
// doesn't have a native Warn level.
func TestBidirectionalWithOptions(t *testing.T, name string,
	fn func(slog.Logger) slog.Logger, opts *BidirectionalTestOptions) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		// Run subtests for different scenarios
		t.Run("BasicLogging", func(t *testing.T) {
			testBidirectionalBasic(t, fn)
		})

		t.Run("WithFields", func(t *testing.T) {
			testBidirectionalFields(t, fn, opts)
		})

		t.Run("AllLevels", func(t *testing.T) {
			testBidirectionalLevels(t, fn, opts)
		})

		t.Run("FieldChaining", func(t *testing.T) {
			testBidirectionalChaining(t, fn, opts)
		})

		t.Run("ComplexFields", func(t *testing.T) {
			testBidirectionalComplexFields(t, fn, opts)
		})
	})
}

// testBidirectionalBasic tests basic message logging
func testBidirectionalBasic(t *testing.T, fn func(slog.Logger) slog.Logger) {
	t.Helper()

	// Create recorder and adapter
	recorder := NewLogger()
	adapter := fn(recorder)

	// Log a simple message
	adapter.Info().Print("test message")

	// Verify the message was recorded correctly
	messages := recorder.GetMessages()
	AssertMustMessageCount(t, messages, 1)
	AssertMustMessage(t, messages[0], slog.Info, "test message")
}

// assertIntField checks that an int field has the expected value
// It handles the case where the value might be stored as int, int64, or float64
func assertIntField(t core.T, fields map[string]any, key string, expected int) {
	t.Helper()

	value, exists := fields[key]
	if !core.AssertTrue(t, exists, "field %q exists", key) {
		return
	}

	isExpectedValue := false

	switch v := value.(type) {
	case int:
		isExpectedValue = (v == expected)
	case int64:
		isExpectedValue = (v == int64(expected))
	case float64:
		isExpectedValue = (v == float64(expected))
	default:
		// Type not recognized for int conversion
		core.AssertTrue(t, false, "field %q type %T convertible to int", key, value)
		return
	}

	core.AssertTrue(t, isExpectedValue, "field %q value %v equals %d", key, value, expected)
}

// testBidirectionalFields tests field preservation
func testBidirectionalFields(t *testing.T, fn func(slog.Logger) slog.Logger, _ *BidirectionalTestOptions) {
	t.Helper()

	recorder := NewLogger()
	adapter := fn(recorder)

	// Log with various field types
	// Use Info level instead of Debug to ensure message is not filtered
	adapter.Info().
		WithField("string", "value").
		WithField("int", 42).
		WithField("bool", true).
		WithField("float", 3.14).
		Print("fields test")

	messages := recorder.GetMessages()
	AssertMustMessageCount(t, messages, 1)

	msg := messages[0]
	AssertMustMessage(t, msg, slog.Info, "fields test")
	AssertField(t, msg, "string", "value")
	assertIntField(t, msg.Fields, "int", 42)
	AssertField(t, msg, "bool", true)
	AssertField(t, msg, "float", 3.14)
}

// testBidirectionalLevels tests all log levels
func testBidirectionalLevels(t *testing.T, fn func(slog.Logger) slog.Logger, opts *BidirectionalTestOptions) {
	t.Helper()

	testCases := buildLevelTestCases()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSingleLevel(t, fn, opts, tc)
		})
	}
}

// levelTestCase represents a single level test case
type levelTestCase struct {
	name    string
	logFunc func(slog.Logger, string)
	level   slog.LogLevel
}

// buildLevelTestCases creates test cases for all log levels
func buildLevelTestCases() []levelTestCase {
	return []levelTestCase{
		{
			name: "Debug",
			logFunc: func(l slog.Logger, msg string) {
				l.Debug().Print(msg)
			},
			level: slog.Debug,
		},
		{
			name: "Info",
			logFunc: func(l slog.Logger, msg string) {
				l.Info().Print(msg)
			},
			level: slog.Info,
		},
		{
			name: "Warn",
			logFunc: func(l slog.Logger, msg string) {
				l.Warn().Print(msg)
			},
			level: slog.Warn,
		},
		{
			name: "Error",
			logFunc: func(l slog.Logger, msg string) {
				l.Error().Print(msg)
			},
			level: slog.Error,
		},
	}
}

// testSingleLevel tests a single log level
func testSingleLevel(t *testing.T, fn func(slog.Logger) slog.Logger, opts *BidirectionalTestOptions, tc levelTestCase) {
	t.Helper()

	recorder := NewLogger()
	adapter := fn(recorder)

	msg := fmt.Sprintf("%s level test", tc.name)
	tc.logFunc(adapter, msg)

	messages := recorder.GetMessages()
	expectedLevel := opts.ExpectedLevel(tc.level)

	verifyLevelTestResult(t, messages, expectedLevel, msg)
}

// verifyLevelTestResult verifies the result of a level test
func verifyLevelTestResult(t *testing.T, messages []Message, expectedLevel slog.LogLevel, msg string) {
	t.Helper()

	if expectedLevel == slog.UndefinedLevel {
		AssertMustMessageCount(t, messages, 0)
		return
	}

	AssertMustMessageCount(t, messages, 1)
	AssertMustMessage(t, messages[0], expectedLevel, msg)
}

// testBidirectionalChaining tests field chaining behaviour
func testBidirectionalChaining(t *testing.T, fn func(slog.Logger) slog.Logger, _ *BidirectionalTestOptions) {
	t.Helper()

	recorder := NewLogger()
	adapter := fn(recorder)

	// Create a base logger with fields
	base := adapter.WithField("app", "test").WithField("version", "1.0")

	// Branch off with different fields
	userLogger := base.WithField("component", "user")
	adminLogger := base.WithField("component", "admin")

	// Log from each branch
	userLogger.Info().Print("user action")
	adminLogger.Info().Print("admin action")

	// Verify both messages have the base fields and their specific fields
	messages := recorder.GetMessages()
	AssertMustMessageCount(t, messages, 2)

	// First message - user
	msg1 := messages[0]
	AssertMustMessage(t, msg1, slog.Info, "user action")
	AssertField(t, msg1, "app", "test")
	AssertField(t, msg1, "version", "1.0")
	AssertField(t, msg1, "component", "user")

	// Second message - admin
	msg2 := messages[1]
	AssertMustMessage(t, msg2, slog.Info, "admin action")
	AssertField(t, msg2, "app", "test")
	AssertField(t, msg2, "version", "1.0")
	AssertField(t, msg2, "component", "admin")
}

// testBidirectionalComplexFields tests complex field types
func testBidirectionalComplexFields(t *testing.T, fn func(slog.Logger) slog.Logger, _ *BidirectionalTestOptions) {
	t.Helper()

	recorder := NewLogger()
	adapter := fn(recorder)

	// Test with various complex types
	type custom struct {
		Name  string
		Value int
	}

	err := errors.New("test error")
	customVal := custom{Name: "test", Value: 123}
	slice := []string{"a", "b", "c"}
	mapVal := map[string]int{"one": 1, "two": 2}

	adapter.Error().
		WithField("error", err).
		WithField("custom", customVal).
		WithField("slice", slice).
		WithField("map", mapVal).
		WithField("nil", nil).
		Print("complex fields")

	messages := recorder.GetMessages()
	AssertMustMessageCount(t, messages, 1)

	msg := messages[0]
	AssertMustMessage(t, msg, slog.Error, "complex fields")

	// Verify fields exist (exact value checking depends on adapter implementation)
	_, exists := msg.Fields["error"]
	core.AssertTrue(t, exists, "field %q exists", "error")
	_, exists = msg.Fields["custom"]
	core.AssertTrue(t, exists, "field %q exists", "custom")
	_, exists = msg.Fields["slice"]
	core.AssertTrue(t, exists, "field %q exists", "slice")
	_, exists = msg.Fields["map"]
	core.AssertTrue(t, exists, "field %q exists", "map")
	_, exists = msg.Fields["nil"]
	core.AssertTrue(t, exists, "field %q exists", "nil")
}

// TestBidirectionalWithAdapter is a convenience function that creates an adapter
// and tests it for bidirectional compatibility. It's useful for handlers that
// have a simple New() function that returns the adapter.
func TestBidirectionalWithAdapter(t *testing.T, name string, newAdapter func() slog.Logger) {
	t.Helper()

	TestBidirectional(t, name, func(_ slog.Logger) slog.Logger {
		// This assumes the adapter has some way to configure its backend
		// The actual implementation depends on the specific adapter
		return newAdapter()
	})
}
