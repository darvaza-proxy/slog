package testing

import (
	"testing"

	"darvaza.org/slog"
)

func TestCompareMessages(t *testing.T) {
	// Create test messages
	msg1 := Message{Level: slog.Info, Message: "test1", Fields: map[string]any{"a": 1}}
	msg2 := Message{Level: slog.Info, Message: "test2", Fields: map[string]any{"b": 2}}
	msg3 := Message{Level: slog.Debug, Message: "test3", Fields: map[string]any{"c": 3}}
	msgDup := Message{Level: slog.Info, Message: "test1", Fields: map[string]any{"a": 1}} // Same as msg1

	// Messages with multiple fields to test field ordering
	msgMulti1 := Message{
		Level:   slog.Info,
		Message: "multi",
		Fields:  map[string]any{"z": 26, "a": 1, "m": 13}, // Intentionally unsorted
	}
	msgMulti2 := Message{
		Level:   slog.Info,
		Message: "multi",
		Fields:  map[string]any{"a": 1, "m": 13, "z": 26}, // Same fields, different order
	}

	t.Run("empty sets", func(t *testing.T) {
		testCompareMessagesCase(t, compareTestCase{
			first: []Message{}, second: []Message{},
			wantOnlyFirst: 0, wantOnlySecond: 0, wantBoth: 0,
		})
	})

	t.Run("disjoint sets", func(t *testing.T) {
		testCompareMessagesCase(t, compareTestCase{
			first: []Message{msg1}, second: []Message{msg2},
			wantOnlyFirst: 1, wantOnlySecond: 1, wantBoth: 0,
		})
	})

	t.Run("identical sets", func(t *testing.T) {
		testCompareMessagesCase(t, compareTestCase{
			first: []Message{msg1, msg2}, second: []Message{msg1, msg2},
			wantOnlyFirst: 0, wantOnlySecond: 0, wantBoth: 2,
		})
	})

	t.Run("overlapping sets", func(t *testing.T) {
		testCompareMessagesCase(t, compareTestCase{
			first: []Message{msg1, msg2}, second: []Message{msg2, msg3},
			wantOnlyFirst: 1, wantOnlySecond: 1, wantBoth: 1,
		})
	})

	t.Run("duplicates in same set", func(t *testing.T) {
		testCompareMessagesCase(t, compareTestCase{
			first: []Message{msg1, msgDup}, second: []Message{msg2},
			wantOnlyFirst: 1, wantOnlySecond: 1, wantBoth: 0,
		})
	})

	t.Run("messages with multiple fields - same content different order", func(t *testing.T) {
		testCompareMessagesCase(t, compareTestCase{
			first: []Message{msgMulti1}, second: []Message{msgMulti2},
			wantOnlyFirst: 0, wantOnlySecond: 0, wantBoth: 1,
		})
	})

	t.Run("multi-field vs single-field messages", func(t *testing.T) {
		testCompareMessagesCase(t, compareTestCase{
			first: []Message{msgMulti1, msg1}, second: []Message{msg1, msg2},
			wantOnlyFirst: 1, wantOnlySecond: 1, wantBoth: 1,
		})
	})
}

// compareTestCase holds expected values for comparison tests
type compareTestCase struct {
	first          []Message
	second         []Message
	wantOnlyFirst  int
	wantOnlySecond int
	wantBoth       int
}

// testCompareMessagesCase tests a single CompareMessages case
func testCompareMessagesCase(t *testing.T, tc compareTestCase) {
	t.Helper()

	onlyFirst, onlySecond, both := CompareMessages(tc.first, tc.second)

	if len(onlyFirst) != tc.wantOnlyFirst {
		t.Errorf("onlyFirst: got %d messages, want %d", len(onlyFirst), tc.wantOnlyFirst)
		logMessages(t, "only in first", onlyFirst)
	}

	if len(onlySecond) != tc.wantOnlySecond {
		t.Errorf("onlySecond: got %d messages, want %d", len(onlySecond), tc.wantOnlySecond)
		logMessages(t, "only in second", onlySecond)
	}

	if len(both) != tc.wantBoth {
		t.Errorf("both: got %d messages, want %d", len(both), tc.wantBoth)
		logMessages(t, "in both", both)
	}
}

// logMessages logs a slice of messages with a prefix
func logMessages(t *testing.T, prefix string, messages []Message) {
	t.Helper()
	for _, msg := range messages {
		t.Logf("  %s: %s", prefix, msg.String())
	}
}

func TestTransformMessages(t *testing.T) {
	messages := []Message{
		{Level: slog.Debug, Message: "debug"},
		{Level: slog.Info, Message: "info"},
		{Level: slog.Warn, Message: "warn"},
		{Level: slog.Error, Message: "error"},
	}

	t.Run("no options", func(t *testing.T) {
		testTransformMessagesNoOptions(t, messages)
	})

	t.Run("with level exceptions", func(t *testing.T) {
		testTransformMessagesWithExceptions(t, messages)
	})

	t.Run("with undefined level mapping", func(t *testing.T) {
		testTransformMessagesWithUndefinedLevel(t, messages)
	})

	t.Run("comparison with transformation", func(t *testing.T) {
		testTransformMessagesComparison(t)
	})
}

// testTransformMessagesNoOptions tests transformation without options
func testTransformMessagesNoOptions(t *testing.T, messages []Message) {
	t.Helper()

	result := TransformMessages(messages, nil)
	if len(result) != len(messages) {
		t.Errorf("expected %d messages, got %d", len(messages), len(result))
	}

	for i, msg := range result {
		if msg.Level != messages[i].Level {
			t.Errorf("message %d: level should not change", i)
		}
	}
}

// testTransformMessagesWithExceptions tests transformation with level exceptions
func testTransformMessagesWithExceptions(t *testing.T, messages []Message) {
	t.Helper()

	opts := AdapterOptions{
		LevelExceptions: map[slog.LogLevel]slog.LogLevel{
			slog.Warn: slog.Info, // logr style mapping
		},
	}

	result := TransformMessages(messages, &opts)
	verifyTransformations(t, messages, result, &opts)
}

// testTransformMessagesWithUndefinedLevel tests transformation with UndefinedLevel mapping
func testTransformMessagesWithUndefinedLevel(t *testing.T, messages []Message) {
	t.Helper()

	opts := AdapterOptions{
		LevelExceptions: map[slog.LogLevel]slog.LogLevel{
			slog.Warn:  slog.UndefinedLevel, // Skip Warn messages
			slog.Debug: slog.UndefinedLevel, // Skip Debug messages
		},
	}

	result := TransformMessages(messages, &opts)

	// We should only have Info and Error messages left
	if len(result) != 2 {
		t.Errorf("expected 2 messages after filtering, got %d", len(result))
		for i, msg := range result {
			t.Logf("  [%d] level=%v, message=%q", i, msg.Level, msg.Message)
		}
	}

	// Verify only Info and Error messages remain
	expectedMessages := map[string]bool{
		"info":  false,
		"error": false,
	}

	for _, msg := range result {
		expectedMessages[msg.Message] = true
	}

	if !expectedMessages["info"] {
		t.Error("expected Info message to be present")
	}
	if !expectedMessages["error"] {
		t.Error("expected Error message to be present")
	}
}

// verifyTransformations verifies that transformations were applied correctly
func verifyTransformations(t *testing.T, original, transformed []Message, opts *AdapterOptions) {
	t.Helper()

	for i, msg := range transformed {
		expected := opts.ExpectedLevel(original[i].Level)
		if msg.Level != expected {
			t.Errorf("message %d: expected level %v, got %v", i, expected, msg.Level)
		}
	}
}

// testTransformMessagesComparison tests comparison after transformation
func testTransformMessagesComparison(t *testing.T) {
	t.Helper()

	expected := []Message{
		{Level: slog.Info, Message: "test1"},
		{Level: slog.Warn, Message: "test2"}, // Will be transformed to Info
	}

	actual := []Message{
		{Level: slog.Info, Message: "test1"},
		{Level: slog.Info, Message: "test2"}, // Already Info (as adapter would return)
	}

	opts := AdapterOptions{
		LevelExceptions: map[slog.LogLevel]slog.LogLevel{
			slog.Warn: slog.Info,
		},
	}

	expectedTransformed := TransformMessages(expected, &opts)
	verifyComparisonResult(t, expectedTransformed, actual)
}

// verifyComparisonResult verifies the comparison result
func verifyComparisonResult(t *testing.T, expected, actual []Message) {
	t.Helper()

	onlyExpected, onlyActual, both := CompareMessages(expected, actual)

	if len(onlyExpected) != 0 {
		t.Errorf("expected no messages only in expected, got %d", len(onlyExpected))
	}
	if len(onlyActual) != 0 {
		t.Errorf("expected no messages only in actual, got %d", len(onlyActual))
	}
	if len(both) != 2 {
		t.Errorf("expected 2 messages in both, got %d", len(both))
	}
}

func TestMessageString(t *testing.T) {
	// Note: This test is currently expected to fail because LogLevel
	// doesn't have a String() method, so it prints as a number.
	// This documents the current behaviour.
	tests := []struct {
		name string
		msg  Message
		want string
	}{
		{
			name: "basic message",
			msg:  Message{Level: slog.Info, Message: "hello"},
			want: `[5] "hello"`, // Info = 5
		},
		{
			name: "message with fields",
			msg: Message{
				Level:   slog.Debug,
				Message: "test",
				Fields:  map[string]any{"b": 2, "a": 1}, // Intentionally unsorted
			},
			want: `[6] "test" a=1 b=2`, // Debug = 6, fields sorted
		},
		{
			name: "message with stack",
			msg: Message{
				Level:   slog.Error,
				Message: "error",
				Stack:   true,
			},
			want: `[3] "error" [stack]`, // Error = 3
		},
		{
			name: "message with everything",
			msg: Message{
				Level:   slog.Warn,
				Message: "warning",
				Fields:  map[string]any{"code": 500, "msg": "internal"},
				Stack:   true,
			},
			want: `[4] "warning" code=500 msg=internal [stack]`, // Warn = 4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.msg.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
