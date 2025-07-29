package slog_test

import (
	"testing"

	"darvaza.org/slog"
)

func TestLogLevel(t *testing.T) {
	t.Run("Constants", testLogLevelConstants)
	t.Run("Ordering", testLogLevelOrdering)
}

func testLogLevelConstants(t *testing.T) {
	// Test that log level constants have expected values
	levels := []struct {
		name  string
		level slog.LogLevel
		value slog.LogLevel
	}{
		{"UndefinedLevel", slog.UndefinedLevel, 0},
		{"Panic", slog.Panic, 1},
		{"Fatal", slog.Fatal, 2},
		{"Error", slog.Error, 3},
		{"Warn", slog.Warn, 4},
		{"Info", slog.Info, 5},
		{"Debug", slog.Debug, 6},
	}

	for _, tc := range levels {
		t.Run(tc.name, func(t *testing.T) {
			if tc.level != tc.value {
				t.Errorf("%s should be %d, got %d", tc.name, tc.value, tc.level)
			}
		})
	}
}

func testLogLevelOrdering(t *testing.T) {
	// Test that log levels are in correct order (lower values = higher priority)
	levels := []slog.LogLevel{
		slog.UndefinedLevel,
		slog.Panic,
		slog.Fatal,
		slog.Error,
		slog.Warn,
		slog.Info,
		slog.Debug,
	}

	for i := 1; i < len(levels); i++ {
		if levels[i-1] >= levels[i] {
			t.Errorf("Level %v should be less than %v", levels[i-1], levels[i])
		}
	}
}

func TestFields(t *testing.T) {
	t.Run("TypeAlias", testFieldsTypeAlias)
	t.Run("MapCompatibility", testFieldsMapCompatibility)
}

func testFieldsTypeAlias(t *testing.T) {
	// Test that Fields is a proper map alias
	fields := slog.Fields{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	// Should behave like a map
	if len(fields) != 3 {
		t.Errorf("Fields should have 3 elements, got %d", len(fields))
	}

	if fields["key1"] != "value1" {
		t.Errorf("Fields access should work like map")
	}
}

func testFieldsMapCompatibility(t *testing.T) {
	// Test that Fields can be used where map[string]any is expected
	fields := slog.Fields{
		"test": "value",
	}

	// Should be assignable to map[string]any
	var m map[string]any = fields
	if m["test"] != "value" {
		t.Error("Fields should be compatible with map[string]any")
	}

	// Should be convertible from map[string]any
	original := map[string]any{
		"convert": "test",
	}
	converted := slog.Fields(original)
	if converted["convert"] != "test" {
		t.Error("Should be able to convert map[string]any to Fields")
	}
}

func TestErrorFieldName(t *testing.T) {
	// Test the error field name constant
	if slog.ErrorFieldName != "error" {
		t.Errorf("ErrorFieldName should be 'error', got %q", slog.ErrorFieldName)
	}
}
