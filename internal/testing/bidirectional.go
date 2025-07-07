package testing

import (
	"errors"
	"fmt"
	"testing"

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
			testBidirectionalFields(t, fn)
		})

		t.Run("AllLevels", func(t *testing.T) {
			testBidirectionalLevels(t, fn, opts)
		})

		t.Run("FieldChaining", func(t *testing.T) {
			testBidirectionalChaining(t, fn)
		})

		t.Run("ComplexFields", func(t *testing.T) {
			testBidirectionalComplexFields(t, fn)
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
	AssertMessageCount(t, messages, 1)
	AssertMessage(t, messages[0], slog.Info, "test message")
}

// assertIntField checks that an int field has the expected value
// It handles the case where the value might be stored as int, int64, or float64
func assertIntField(t *testing.T, fields map[string]any, key string, expected int) {
	t.Helper()

	value, exists := fields[key]
	if !exists {
		t.Errorf("Expected field '%s' not found", key)
		return
	}

	isExpectedType := true
	isExpectedValue := false

	switch v := value.(type) {
	case int:
		isExpectedValue = (v == expected)
	case int64:
		isExpectedValue = (v == int64(expected))
	case float64:
		isExpectedValue = (v == float64(expected))
	default:
		isExpectedType = false
	}

	if !isExpectedType {
		t.Errorf("Unexpected type for %s field: %T", key, value)
	} else if !isExpectedValue {
		t.Errorf("Expected %s field to be %d, got %v", key, expected, value)
	}
}

// testBidirectionalFields tests field preservation
func testBidirectionalFields(t *testing.T, fn func(slog.Logger) slog.Logger) {
	t.Helper()

	recorder := NewLogger()
	adapter := fn(recorder)

	// Log with various field types
	adapter.Debug().
		WithField("string", "value").
		WithField("int", 42).
		WithField("bool", true).
		WithField("float", 3.14).
		Print("fields test")

	messages := recorder.GetMessages()
	AssertMessageCount(t, messages, 1)

	msg := messages[0]
	AssertMessage(t, msg, slog.Debug, "fields test")
	AssertField(t, msg, "string", "value")
	assertIntField(t, msg.Fields, "int", 42)
	AssertField(t, msg, "bool", true)
	AssertField(t, msg, "float", 3.14)
}

// testBidirectionalLevels tests all log levels
func testBidirectionalLevels(t *testing.T, fn func(slog.Logger) slog.Logger, opts *BidirectionalTestOptions) {
	t.Helper()

	testCases := []struct {
		name    string
		logFunc func(slog.Logger, string)
		level   slog.LogLevel
	}{
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := NewLogger()
			adapter := fn(recorder)

			msg := fmt.Sprintf("%s level test", tc.name)
			tc.logFunc(adapter, msg)

			messages := recorder.GetMessages()
			AssertMessageCount(t, messages, 1)

			// Use options to get expected level
			expectedLevel := opts.ExpectedLevel(tc.level)
			AssertMessage(t, messages[0], expectedLevel, msg)
		})
	}
}

// testBidirectionalChaining tests field chaining behavior
func testBidirectionalChaining(t *testing.T, fn func(slog.Logger) slog.Logger) {
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
	AssertMessageCount(t, messages, 2)

	// First message - user
	msg1 := messages[0]
	AssertMessage(t, msg1, slog.Info, "user action")
	AssertField(t, msg1, "app", "test")
	AssertField(t, msg1, "version", "1.0")
	AssertField(t, msg1, "component", "user")

	// Second message - admin
	msg2 := messages[1]
	AssertMessage(t, msg2, slog.Info, "admin action")
	AssertField(t, msg2, "app", "test")
	AssertField(t, msg2, "version", "1.0")
	AssertField(t, msg2, "component", "admin")
}

// testBidirectionalComplexFields tests complex field types
func testBidirectionalComplexFields(t *testing.T, fn func(slog.Logger) slog.Logger) {
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
	AssertMessageCount(t, messages, 1)

	msg := messages[0]
	AssertMessage(t, msg, slog.Error, "complex fields")

	// Verify fields exist (exact value checking depends on adapter implementation)
	if _, exists := msg.Fields["error"]; !exists {
		t.Error("Expected error field to exist")
	}
	if _, exists := msg.Fields["custom"]; !exists {
		t.Error("Expected custom field to exist")
	}
	if _, exists := msg.Fields["slice"]; !exists {
		t.Error("Expected slice field to exist")
	}
	if _, exists := msg.Fields["map"]; !exists {
		t.Error("Expected map field to exist")
	}
	if _, exists := msg.Fields["nil"]; !exists {
		t.Error("Expected nil field to exist")
	}
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
