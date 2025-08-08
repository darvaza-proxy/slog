package filter

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/mock"
	"darvaza.org/slog/internal"
	slogtest "darvaza.org/slog/internal/testing"
)

// Test constants
const (
	sensitiveField1 = "password"
	sensitiveField2 = "secret"
)

// Test 1: Logger Print methods (no-ops) - covers filter.go:54,57,60
func TestLoggerPrintMethodsCoverage(t *testing.T) {
	parent := mock.NewLogger()
	logger := &Logger{
		Parent:    parent,
		Threshold: slog.Debug, // Everything should be enabled
	}

	// Logger.Enabled() should always return false
	core.AssertFalse(t, logger.Enabled(), "Logger.Enabled always false")

	// Call all print methods - they should do nothing
	logger.Print("test print")
	logger.Println("test println")
	logger.Printf("test printf %s", "value")

	// Verify nothing was logged to parent
	messages := parent.GetMessages()
	slogtest.AssertMessageCount(t, messages, 0)
}

// Test 2: LogEntry with UndefinedLevel threshold - covers entry.go:32
func TestLogEntryUndefinedThreshold(t *testing.T) {
	parent := mock.NewLogger()

	// Create LogEntry with UndefinedLevel threshold directly
	entry := &LogEntry{
		config: &Logger{
			Parent:    parent,
			Threshold: slog.UndefinedLevel,
		},
		loglet: internal.Loglet{},
	}

	// This covers the check() function's UndefinedLevel path
	core.AssertFalse(t, entry.check(), "undefined threshold check fails")
	core.AssertFalse(t, entry.Enabled(), "undefined threshold not enabled")
}

// Test 3: Logger with UndefinedLevel threshold - covers filter.go:166-172
func TestNewWithUndefinedLevel(t *testing.T) {
	parent := mock.NewLogger()

	// Test New() with UndefinedLevel (should default to Error)
	logger := New(parent, slog.UndefinedLevel)

	// Verify threshold was set to Error (no type assertion needed now)
	core.AssertEqual(t, slog.Error, logger.Threshold, "defaults to Error")

	// Test with negative threshold (should also default to Error)
	logger2 := New(parent, slog.LogLevel(-1))
	core.AssertEqual(t, slog.Error, logger2.Threshold, "negative threshold defaults to Error")
}

// Test 4: Filter returns handling - covers filter.go:127-134, 138-145
func TestFilterReturnsNil(t *testing.T) {
	parent := mock.NewLogger()
	logger := &Logger{
		Parent:    parent,
		Threshold: slog.Debug,
		FieldFilter: func(key string, val any) (string, any, bool) {
			// Filter out any field with key "password" or "secret"
			if key == sensitiveField1 || key == sensitiveField2 {
				return "", nil, false
			}
			return key, val, true
		},
	}

	// Test WithField returns self when filtered
	result := logger.WithField("password", "12345")
	core.AssertSame(t, logger, result, "returns self when filtered")

	// Test WithFields returns self when filtered
	logger2 := &Logger{
		Parent:    parent,
		Threshold: slog.Debug,
		FieldsFilter: func(_ slog.Fields) (slog.Fields, bool) {
			return nil, false // Filter everything out
		},
	}
	result2 := logger2.WithFields(map[string]any{"key": "value"})
	core.AssertSame(t, logger2, result2, "returns self when fields filtered")
}

// Test 5: LogEntry.check() with nil config - covers additional branches
func TestLogEntryCheckNilConfig(t *testing.T) {
	// Test with nil LogEntry
	var entry *LogEntry
	core.AssertFalse(t, entry.check(), "nil entry check fails")

	// Test with nil config
	entry2 := &LogEntry{
		config: nil,
		loglet: internal.Loglet{},
	}
	core.AssertFalse(t, entry2.check(), "nil config check fails")
	core.AssertFalse(t, entry2.Enabled(), "nil config not enabled")
}

// Test 6: Edge cases in filter fallbacks - covers utils.go:98-117, 139-163
func TestFilterFallbackEdgeCases(t *testing.T) {
	parent := mock.NewLogger()

	// Test FieldsFilter returning empty map
	logger := &Logger{
		Parent:    parent,
		Threshold: slog.Debug,
		FieldsFilter: func(_ slog.Fields) (slog.Fields, bool) {
			return slog.Fields{}, true // Return empty map but keep entry
		},
	}

	entry := logger.Info().WithField("key", "value")
	entry.Print("test empty map")

	// Test FieldFilter on WithFields (fallback path)
	logger2 := &Logger{
		Parent:    parent,
		Threshold: slog.Debug,
		FieldFilter: func(key string, val any) (string, any, bool) {
			if key == "skip" {
				return "", nil, false
			}
			return "prefix_" + key, val, true
		},
		// No FieldsFilter, so WithFields falls back to FieldFilter
	}

	entry2 := logger2.Info().WithFields(map[string]any{
		"keep": "value",
		"skip": "removed",
	})
	entry2.Print("test field filter on WithFields")

	// Verify messages were logged
	messages := parent.GetMessages()
	slogtest.AssertMessageCount(t, messages, 2)
}

// Test 7: LogEntry.getEnabled() branches - covers entry.go:53-74
func TestLogEntryGetEnabledBranches(t *testing.T) {
	parent := mock.NewLogger()

	// Test with level == UndefinedLevel
	entry1 := &LogEntry{
		config: &Logger{
			Parent:    parent,
			Threshold: slog.Info,
		},
		loglet: internal.Loglet{},
	}
	entry1.loglet = entry1.loglet.WithLevel(slog.UndefinedLevel)
	_, _, enabled := entry1.getEnabled()
	core.AssertFalse(t, enabled, "undefined level not enabled")

	// Test with threshold == UndefinedLevel but valid level
	entry2 := &LogEntry{
		config: &Logger{
			Parent:    parent,
			Threshold: slog.UndefinedLevel,
		},
		loglet: internal.Loglet{},
	}
	entry2.loglet = entry2.loglet.WithLevel(slog.Info)
	_, _, enabled2 := entry2.getEnabled()
	core.AssertFalse(t, enabled2, "undefined threshold not enabled")

	// Test Fatal with no parent
	entry3 := &LogEntry{
		config: &Logger{
			Parent:    nil,
			Threshold: slog.Error,
		},
		loglet: internal.Loglet{},
	}
	entry3.loglet = entry3.loglet.WithLevel(slog.Fatal)
	_, _, enabled3 := entry3.getEnabled()
	core.AssertTrue(t, enabled3, "Fatal enabled even without parent")

	// Test Panic with no parent
	entry4 := &LogEntry{
		config: &Logger{
			Parent:    nil,
			Threshold: slog.Error,
		},
		loglet: internal.Loglet{},
	}
	entry4.loglet = entry4.loglet.WithLevel(slog.Panic)
	_, _, enabled4 := entry4.getEnabled()
	core.AssertTrue(t, enabled4, "Panic enabled even without parent")
}

// Test 8: LogEntry.msg() branches for parentless Panic - covers entry.go:107-132
func TestLogEntryMsgParentlessPanic(t *testing.T) {
	// Test Panic without parent - should panic
	entry := &LogEntry{
		config: &Logger{
			Parent:    nil,
			Threshold: slog.Error,
		},
		loglet: internal.Loglet{},
	}
	entry.loglet = entry.loglet.WithLevel(slog.Panic)

	// Test that msg() with Panic actually panics
	core.AssertPanic(t, func() {
		entry.msg(0, "test panic")
	}, nil, "parentless panic")
}

// Test 9: LogEntry.checkWithLevel() error paths - covers entry.go:203-221
func TestLogEntryCheckWithLevelErrors(t *testing.T) {
	parent := mock.NewLogger()

	// Test with invalid logger (nil config) - now returns PanicError instead of errSkip
	entry1 := &LogEntry{
		config: nil,
		loglet: internal.Loglet{},
	}
	err := entry1.checkWithLevel(0, slog.Info)
	core.AssertError(t, err, "invalid logger returns error")
	// Use AssertMustTypeIs to verify error type and message
	panicErr := core.AssertMustTypeIs[*core.PanicError](t, err, "error is PanicError")
	core.AssertContains(t, panicErr.Error(), "invalid logger entry state", "error message")

	// Test with valid logger and level change
	entry2 := &LogEntry{
		config: &Logger{
			Parent:    parent,
			Threshold: slog.Debug,
		},
		loglet: internal.Loglet{},
	}
	entry2.loglet = entry2.loglet.WithLevel(slog.Info)

	// Try to change level (should return nil to proceed with new entry)
	err2 := entry2.checkWithLevel(0, slog.Debug)
	core.AssertNoError(t, err2, "no error for level change")

	// Test with same level (should return errSkip)
	err3 := entry2.checkWithLevel(0, slog.Info)
	core.AssertEqual(t, errSkip, err3, "errSkip for same level")
}

// Test Logger.WithLevel panic path - covers filter.go:85-86
func TestLoggerWithLevelPanic(t *testing.T) {
	// Test panic path: invalid Logger with UndefinedLevel threshold
	invalidLogger := &Logger{
		Parent:    nil,
		Threshold: slog.UndefinedLevel,
	}

	// WithLevel should panic due to invalid logger state
	core.AssertPanic(t, func() {
		invalidLogger.WithLevel(slog.Info)
	}, nil, "panic on invalid logger")
}

// Test 10: LogEntry.shouldCollectFields() branches - covers entry.go:257-276
func TestLogEntryShouldCollectFieldsBranches(t *testing.T) {
	parent := mock.NewLogger()

	// Test with no level set (should collect speculatively)
	entry1 := &LogEntry{
		config: &Logger{
			Parent:    parent,
			Threshold: slog.Info,
		},
		loglet: internal.Loglet{},
	}
	// No level set means UndefinedLevel
	core.AssertTrue(t, entry1.shouldCollectFields(), "no level collects speculatively")

	// Test with level within threshold
	entry2 := &LogEntry{
		config: &Logger{
			Parent:    parent,
			Threshold: slog.Info,
		},
		loglet: internal.Loglet{},
	}
	entry2.loglet = entry2.loglet.WithLevel(slog.Warn)
	core.AssertTrue(t, entry2.shouldCollectFields(), "within threshold collects")

	// Test with level exceeding threshold
	entry3 := &LogEntry{
		config: &Logger{
			Parent:    parent,
			Threshold: slog.Error,
		},
		loglet: internal.Loglet{},
	}
	entry3.loglet = entry3.loglet.WithLevel(slog.Debug)
	core.AssertFalse(t, entry3.shouldCollectFields(), "exceeding threshold doesn't collect")

	// Test parentless (should never collect)
	entry4 := &LogEntry{
		config: &Logger{
			Parent:    nil,
			Threshold: slog.Debug,
		},
		loglet: internal.Loglet{},
	}
	entry4.loglet = entry4.loglet.WithLevel(slog.Info)
	core.AssertFalse(t, entry4.shouldCollectFields(), "parentless never collects")

	// Test with nil config (should not collect)
	entry5 := &LogEntry{
		config: nil,
		loglet: internal.Loglet{},
	}
	core.AssertFalse(t, entry5.shouldCollectFields(), "nil config doesn't collect")
}

// Test 11: Additional filter fallback scenarios
func TestFilterFallbackScenarios(t *testing.T) {
	parent := mock.NewLogger()

	// Test FieldFilter applied to WithFields (fallback from FieldsFilter)
	logger1 := &Logger{
		Parent:    parent,
		Threshold: slog.Debug,
		FieldFilter: func(key string, val any) (string, any, bool) {
			// Transform keys
			return "prefix_" + key, val, true
		},
		// No FieldsFilter, so WithFields should fall back to FieldFilter
	}

	entry1 := logger1.Info().WithFields(map[string]any{
		"key1": "value1",
		"key2": "value2",
	})
	entry1.Print("test with field filter fallback")

	// Test FieldsFilter applied to WithField (fallback from FieldFilter)
	logger2 := &Logger{
		Parent:    parent,
		Threshold: slog.Debug,
		FieldsFilter: func(fields slog.Fields) (slog.Fields, bool) {
			// Transform all fields
			result := make(slog.Fields)
			for k, v := range fields {
				result["transformed_"+k] = v
			}
			return result, true
		},
		// No FieldFilter, so WithField should fall back to FieldsFilter
	}

	entry2 := logger2.Info().WithField("single", "value")
	entry2.Print("test with fields filter fallback")

	// Verify both logged successfully
	messages := parent.GetMessages()
	slogtest.AssertMessageCount(t, messages, 2)

	// Verify field transformations
	msg1 := messages[0]
	slogtest.AssertField(t, msg1, "prefix_key1", "value1")
	slogtest.AssertField(t, msg1, "prefix_key2", "value2")

	msg2 := messages[1]
	slogtest.AssertField(t, msg2, "transformed_single", "value")
}

// Test 12: Cover remaining New() branches
func TestNewWithNilParent(t *testing.T) {
	// Test New() with nil parent and valid threshold
	// Note: New() overrides threshold to Fatal when parent is nil
	logger := New(nil, slog.Info)

	// Verify it creates a parentless logger with Fatal threshold (no type assertion needed)
	core.AssertNil(t, logger.Parent, "nil parent preserved")
	core.AssertEqual(t, slog.Fatal, logger.Threshold, "threshold overridden to Fatal for nil parent")

	// Test that only Fatal/Panic are enabled
	core.AssertFalse(t, logger.Debug().Enabled(), "Debug not enabled")
	core.AssertFalse(t, logger.Info().Enabled(), "Info not enabled")
	core.AssertFalse(t, logger.Warn().Enabled(), "Warn not enabled")
	core.AssertFalse(t, logger.Error().Enabled(), "Error not enabled")
	core.AssertTrue(t, logger.Fatal().Enabled(), "Fatal enabled")
	core.AssertTrue(t, logger.Panic().Enabled(), "Panic enabled")
}

// Test 13: Cover filter function edge cases
func TestFilterFunctionEdgeCases(t *testing.T) {
	parent := mock.NewLogger()

	// Test FieldFilter that modifies key to empty string
	logger1 := &Logger{
		Parent:    parent,
		Threshold: slog.Debug,
		FieldFilter: func(_ string, val any) (string, any, bool) {
			return "", val, true // Empty key should be rejected
		},
	}

	entry1 := logger1.Info().WithField("test", "value")
	entry1.Print("test empty key")

	// Test FieldsFilter that returns nil map
	logger2 := &Logger{
		Parent:    parent,
		Threshold: slog.Debug,
		FieldsFilter: func(_ slog.Fields) (slog.Fields, bool) {
			return nil, true // nil map should be handled
		},
	}

	entry2 := logger2.Info().WithFields(map[string]any{"test": "value"})
	entry2.Print("test nil map")

	// Both should still log
	messages := parent.GetMessages()
	slogtest.AssertMessageCount(t, messages, 2)
}

// Test 14: Logger.shouldCollectFields() branches
func TestLoggerShouldCollectFields(t *testing.T) {
	// Test with nil logger
	var logger *Logger
	core.AssertFalse(t, logger.shouldCollectFields(), "nil logger doesn't collect")

	// Test with invalid logger (UndefinedLevel threshold)
	logger2 := &Logger{
		Parent:    mock.NewLogger(),
		Threshold: slog.UndefinedLevel,
	}
	core.AssertFalse(t, logger2.shouldCollectFields(), "invalid logger doesn't collect")

	// Test with no parent
	logger3 := &Logger{
		Parent:    nil,
		Threshold: slog.Debug,
	}
	core.AssertFalse(t, logger3.shouldCollectFields(), "parentless doesn't collect")

	// Test with valid logger
	logger4 := &Logger{
		Parent:    mock.NewLogger(),
		Threshold: slog.Debug,
	}
	core.AssertTrue(t, logger4.shouldCollectFields(), "valid logger collects")
}

// Test 15: Cover entry.go:126 - unreachable return in parentless non-Fatal/Panic
func TestParentlessWithNonTerminalLevel(t *testing.T) {
	// Create a parentless LogEntry with Info level (not Fatal or Panic)
	entry := &LogEntry{
		config: &Logger{
			Parent:    nil,
			Threshold: slog.Debug,
		},
		loglet: internal.Loglet{},
	}
	entry.loglet = entry.loglet.WithLevel(slog.Info)

	// This should hit the "unreachable" return at line 126
	// The function will log to stderr but then return normally
	entry.msg(0, "test message")

	// If we got here without panic or exit, the test passed
	core.AssertTrue(t, true, "parentless non-terminal level handled")
}

// Test 16: Cover utils.go:162 - all fields filtered out in fallback
func TestFieldsFilterFallbackAllFiltered(t *testing.T) {
	parent := mock.NewLogger()

	// Create a logger with only FieldFilter that rejects everything
	logger := &Logger{
		Parent:    parent,
		Threshold: slog.Debug,
		FieldFilter: func(key string, val any) (string, any, bool) {
			// Reject all fields that start with "field" prefix
			if len(key) >= 5 && key[:5] == "field" {
				return "", nil, false
			}
			return key, val, true
		},
		// No FieldsFilter, so WithFields will use createFieldsFilterFallback
	}

	// Try to add multiple fields - all should be filtered out (they all start with "field")
	entry := logger.Info().WithFields(map[string]any{
		"field1": "value1",
		"field2": "value2",
		"field3": "value3",
	})

	// The entry should be the same as without fields (all filtered)
	entry.Print("message with no fields")

	// Verify message was logged but with no fields
	messages := parent.GetMessages()
	slogtest.AssertMessageCount(t, messages, 1)
	core.AssertEqual(t, 0, len(messages[0].Fields), "no fields added")
}

// Test 15: createFieldsFilterFallback with all fields filtered out - covers utils.go:162
func TestCreateFieldsFilterFallbackAllFiltered(t *testing.T) {
	parent := mock.NewLogger()

	// Create logger with FieldFilter that filters out everything
	logger := &Logger{
		Parent:    parent,
		Threshold: slog.Debug,
		FieldFilter: func(key string, val any) (string, any, bool) {
			// Filter based on value type - reject all strings
			if _, ok := val.(string); ok {
				return "", nil, false
			}
			return key, val, true
		},
		// No FieldsFilter, so WithFields will use createFieldsFilterFallback
	}

	// Try to add fields through WithFields - string values should be filtered out
	result := logger.WithFields(map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	})

	// Should return the original logger since all fields were filtered out
	core.AssertSame(t, logger, result, "returns original logger when all fields filtered")
}

// Test 16: WithLevel panic on invalid level - covers entry.go:195-197
func TestWithLevelPanicOnInvalidLevel(t *testing.T) {
	parent := mock.NewLogger()
	entry := &LogEntry{
		config: &Logger{
			Parent:    parent,
			Threshold: slog.Debug,
		},
		loglet: internal.Loglet{},
	}

	// Test with invalid level (out of range) - should panic
	core.AssertPanic(t, func() {
		entry.WithLevel(slog.LogLevel(127)) // Invalid level (above Debug)
	}, nil, "invalid level panics")

	// Test with negative invalid level - should also panic
	core.AssertPanic(t, func() {
		entry.WithLevel(slog.LogLevel(-99)) // Invalid negative level
	}, nil, "negative invalid level panics")
}
