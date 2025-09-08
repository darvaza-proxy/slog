package testing

import (
	"errors"
	"fmt"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// TestBidirectional tests that a bidirectional adapter correctly preserves
// log messages, fields, and levels when round-tripping through the adapter.
// The fn parameter should return a logger that uses the given logger as backend.
func TestBidirectional(t core.T, name string, fn func(slog.Logger) slog.Logger) {
	TestBidirectionalWithOptions(t, name, fn, nil)
}

// TestBidirectionalWithOptions tests a bidirectional adapter with custom options.
// This is useful for adapters that have known limitations, such as logr which
// doesn't have a native Warn level.
func TestBidirectionalWithOptions(t core.T, name string,
	fn func(slog.Logger) slog.Logger, opts *BidirectionalTestOptions) {
	t.Helper()

	Run(t, name, func(runner core.T) {
		Run(runner, "BasicLogging", func(subT core.T) {
			testBidirectionalBasic(subT, fn)
		})
		Run(runner, "WithFields", func(subT core.T) {
			testBidirectionalFields(subT, fn, opts)
		})
		Run(runner, "AllLevels", func(subT core.T) {
			testBidirectionalLevels(subT, fn, opts)
		})
		Run(runner, "FieldChaining", func(subT core.T) {
			testBidirectionalChaining(subT, fn, opts)
		})
		Run(runner, "ComplexFields", func(subT core.T) {
			testBidirectionalComplexFields(subT, fn, opts)
		})
	})
}

// testBidirectionalBasic tests basic message logging
func testBidirectionalBasic(t core.T, fn func(slog.Logger) slog.Logger) {
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

// testBidirectionalFields tests field preservation
func testBidirectionalFields(t core.T, fn func(slog.Logger) slog.Logger, _ *BidirectionalTestOptions) {
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
	AssertField(t, msg, "int", 42)
	AssertField(t, msg, "bool", true)
	AssertField(t, msg, "float", 3.14)
}

// testBidirectionalLevels tests all log levels
func testBidirectionalLevels(t core.T, fn func(slog.Logger) slog.Logger, opts *BidirectionalTestOptions) {
	t.Helper()

	testCases := buildLevelTestCases()

	for _, tc := range testCases {
		Run(t, tc.name, func(subT core.T) {
			testSingleLevel(subT, fn, opts, tc)
		})
	}
}

// levelTestCase represents a single level test case
type levelTestCase struct {
	logFunc func(slog.Logger, string)
	name    string
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
func testSingleLevel(t core.T, fn func(slog.Logger) slog.Logger, opts *BidirectionalTestOptions, tc levelTestCase) {
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
func verifyLevelTestResult(t core.T, messages []Message, expectedLevel slog.LogLevel, msg string) {
	t.Helper()

	if expectedLevel == slog.UndefinedLevel {
		AssertMustMessageCount(t, messages, 0)
		return
	}

	AssertMustMessageCount(t, messages, 1)
	AssertMustMessage(t, messages[0], expectedLevel, msg)
}

// testBidirectionalChaining tests field chaining behaviour
func testBidirectionalChaining(t core.T, fn func(slog.Logger) slog.Logger, _ *BidirectionalTestOptions) {
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
func testBidirectionalComplexFields(t core.T, fn func(slog.Logger) slog.Logger, _ *BidirectionalTestOptions) {
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
func TestBidirectionalWithAdapter(t core.T, name string, newAdapter func() slog.Logger) {
	t.Helper()

	TestBidirectional(t, name, func(_ slog.Logger) slog.Logger {
		// This assumes the adapter has some way to configure its backend
		// The actual implementation depends on the specific adapter
		return newAdapter()
	})
}
