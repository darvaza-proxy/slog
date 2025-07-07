package filter_test

import (
	"strings"
	"testing"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
)

const (
	testValue = "value"
)

func TestLogEntryWithFieldsExtended(t *testing.T) {
	t.Run("EmptyFieldKey", func(t *testing.T) {
		base := newTestLogger()
		logger := filter.New(base, slog.Info)
		entry := logger.Info()

		// Should return same entry
		newEntry := entry.WithField("", "value")
		if newEntry != entry {
			t.Error("WithField with empty key should return same entry")
		}
	})

	t.Run("DisabledEntryFields", func(t *testing.T) {
		base := newTestLogger()
		logger := filter.New(base, slog.Error)
		// Info is below Error threshold, so disabled
		entry := logger.Info().WithField("key", "value")

		// Fields should still be stored in Loglet even if disabled
		entry.Print("should not appear")

		// Verify no message was sent
		msgs := base.GetMessages()
		if len(msgs) != 0 {
			t.Error("Disabled entry should not send messages")
		}
	})

	t.Run("FieldFilterWithFields", func(t *testing.T) {
		base := newTestLogger()
		filterCalls := 0
		logger := &filter.Logger{
			Parent:    base,
			Threshold: slog.Info,
			FieldFilter: func(key string, val any) (string, any, bool) {
				filterCalls++
				if key == "remove" {
					return "", nil, false
				}
				return "prefix_" + key, val, true
			},
		}

		entry := logger.Info()
		entry.WithFields(map[string]any{
			"key1":   "value1",
			"key2":   "value2",
			"remove": "this",
		}).Print("test")

		msgs := base.GetMessages()
		if len(msgs) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(msgs))
		}

		msg := msgs[0]
		if filterCalls != 3 {
			t.Errorf("Expected 3 filter calls, got %d", filterCalls)
		}

		// Check filtered fields
		if msg.Fields["prefix_key1"] != "value1" {
			t.Error("key1 not properly filtered")
		}
		if msg.Fields["prefix_key2"] != "value2" {
			t.Error("key2 not properly filtered")
		}
		if _, exists := msg.Fields["remove"]; exists {
			t.Error("remove field should have been filtered out")
		}
	})

	t.Run("EmptyFieldsMap", func(t *testing.T) {
		base := newTestLogger()
		logger := filter.New(base, slog.Info)
		entry := logger.Info()

		// Should return same entry
		newEntry := entry.WithFields(map[string]any{})
		if newEntry != entry {
			t.Error("WithFields with empty map should return same entry")
		}

		// nil map should also return same entry
		newEntry2 := entry.WithFields(nil)
		if newEntry2 != entry {
			t.Error("WithFields with nil map should return same entry")
		}
	})
}

func TestLogEntryOverridesExtended(t *testing.T) {
	t.Run("FieldOverridePreventsFurtherProcessing", func(t *testing.T) {
		base := newTestLogger()
		fieldFilterCalled := false
		fieldOverrideCalled := false

		logger := &filter.Logger{
			Parent:    base,
			Threshold: slog.Info,
			FieldOverride: func(_ slog.Logger, _ string, _ any) {
				fieldOverrideCalled = true
				// Custom handling - don't add field to entry
			},
			FieldFilter: func(key string, val any) (string, any, bool) {
				fieldFilterCalled = true
				return key, val, true
			},
		}

		logger.Info().WithField("test", "value").Print("message")

		if !fieldOverrideCalled {
			t.Error("FieldOverride should have been called")
		}
		if fieldFilterCalled {
			t.Error("FieldFilter should not be called when FieldOverride is set")
		}

		msgs := base.GetMessages()
		if len(msgs) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(msgs))
		}
		if len(msgs[0].Fields) > 0 {
			t.Error("No fields should be added when FieldOverride doesn't add them")
		}
	})

	t.Run("FieldsOverrideWithFieldOverride", func(t *testing.T) {
		base := newTestLogger()
		fieldOverrideCalls := 0

		logger := &filter.Logger{
			Parent:    base,
			Threshold: slog.Info,
			FieldOverride: func(_ slog.Logger, _ string, _ any) {
				fieldOverrideCalls++
			},
		}

		// When FieldsOverride is not set but FieldOverride is,
		// WithFields should call FieldOverride for each field
		logger.Info().WithFields(map[string]any{
			"key1": "value1",
			"key2": "value2",
		}).Print("message")

		if fieldOverrideCalls != 2 {
			t.Errorf("Expected exactly 2 FieldOverride calls, got %d", fieldOverrideCalls)
		}
	})

	t.Run("FieldsOverridePreventsFurtherProcessing", func(t *testing.T) {
		base := newTestLogger()
		fieldsOverrideCalled := false
		fieldFilterCalled := false

		logger := &filter.Logger{
			Parent:    base,
			Threshold: slog.Info,
			FieldsOverride: func(_ slog.Logger, _ map[string]any) {
				fieldsOverrideCalled = true
				// Custom handling
			},
			FieldFilter: func(key string, val any) (string, any, bool) {
				fieldFilterCalled = true
				return key, val, true
			},
		}

		logger.Info().WithFields(map[string]any{
			"key1": "value1",
		}).Print("message")

		if !fieldsOverrideCalled {
			t.Error("FieldsOverride should have been called")
		}
		if fieldFilterCalled {
			t.Error("FieldFilter should not be called when FieldsOverride is set")
		}
	})
}

func TestLogEntryMessageHandlingExtended(t *testing.T) {
	t.Run("MessageFilterDropsMessage", func(t *testing.T) {
		base := newTestLogger()
		logger := &filter.Logger{
			Parent:    base,
			Threshold: slog.Info,
			MessageFilter: func(msg string) (string, bool) {
				if strings.Contains(msg, "drop") {
					return "", false
				}
				return msg, true
			},
		}

		// This should be dropped
		logger.Info().Print("drop this message")

		// This should go through
		logger.Info().Print("keep this message")

		msgs := base.GetMessages()
		if len(msgs) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(msgs))
		}
		if msgs[0].Message != "keep this message" {
			t.Errorf("Wrong message received: %q", msgs[0].Message)
		}
	})

	t.Run("PrintMethods", func(t *testing.T) {
		base := newTestLogger()
		logger := filter.New(base, slog.Info)
		entry := logger.Info()

		// Test Print
		entry.Print("hello", " ", "world")
		msgs := base.GetMessages()
		if len(msgs) != 1 || msgs[0].Message != "hello world" {
			t.Errorf("Print: expected 'hello world', got %q", msgs[0].Message)
		}

		// Test Println
		entry.Println("hello", "world")
		msgs = base.GetMessages()
		if len(msgs) != 2 {
			t.Fatalf("Println: expected 2 messages, got %d", len(msgs))
		}
		// fmt.Sprintln adds spaces between args and a trailing newline
		expected := "hello world\n"
		if msgs[1].Message != expected {
			t.Errorf("Println: expected %q, got %q", expected, msgs[1].Message)
		}

		// Test Printf
		entry.Printf("hello %s %d", "world", 42)
		msgs = base.GetMessages()
		if len(msgs) != 3 || msgs[2].Message != "hello world 42" {
			t.Errorf("Printf: expected 'hello world 42', got %d messages", len(msgs))
			if len(msgs) > 2 {
				t.Errorf("Printf: message was %q", msgs[2].Message)
			}
		}
	})
}

func TestLogEntryComplexScenariosExtended(t *testing.T) {
	t.Run("ChainedFiltersAndTransforms", func(t *testing.T) {
		base := newTestLogger()
		// Create a filter with all transforms
		logger := &filter.Logger{
			Parent:    base,
			Threshold: slog.Info,
			FieldFilter: func(key string, val any) (string, any, bool) {
				// Remove private fields
				if strings.HasPrefix(key, "_") {
					return "", nil, false
				}
				// Uppercase keys
				return strings.ToUpper(key), val, true
			},
			MessageFilter: func(msg string) (string, bool) {
				// Add timestamp prefix
				return "[TIMESTAMP] " + msg, true
			},
		}

		entry := logger.Info().
			WithField("public", "value").
			WithField("_private", "secret").
			WithStack(0)

		entry.Printf("User %s logged in", "john")

		msgs := base.GetMessages()
		if len(msgs) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(msgs))
		}

		msg := msgs[0]

		// Check message transform
		if msg.Message != "[TIMESTAMP] User john logged in" {
			t.Errorf("Message not properly transformed: %q", msg.Message)
		}

		// Check field transform
		if msg.Fields["PUBLIC"] != testValue {
			t.Error("Public field not transformed to uppercase")
		}
		if _, exists := msg.Fields["_private"]; exists {
			t.Error("Private field should have been filtered")
		}

		// Check stack was preserved
		if !msg.Stack {
			t.Error("Stack should have been preserved")
		}
	})

	t.Run("DisabledToEnabledTransition", func(t *testing.T) {
		base := newTestLogger()
		logger := filter.New(base, slog.Error)

		// Start with disabled level
		entry := logger.Debug().WithField("debug", "value")
		if entry.Enabled() {
			t.Error("Debug entry should be disabled")
		}

		// Transition to enabled level
		// Note: Fields added to disabled entries are stored in Loglet but not passed to parent
		errorEntry := entry.Error().WithField("error", "value")
		if !errorEntry.Enabled() {
			t.Error("Error entry should be enabled")
		}

		errorEntry.Print("test")
		msgs := base.GetMessages()
		if len(msgs) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(msgs))
		}

		msg := msgs[0]
		// Only the error field should be present (debug was on disabled entry)
		if _, hasDebug := msg.Fields["debug"]; hasDebug {
			t.Error("Debug field from disabled entry should not be passed to parent")
		}
		if msg.Fields["error"] != "value" {
			t.Error("Error field not added")
		}
	})
}
