package filter_test

import (
	"fmt"
	"sort"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
	"darvaza.org/slog/handlers/mock"
)

// Test field collection and iteration timing
func TestFieldCollectionTiming(t *testing.T) {
	base := mock.NewLogger()

	// Track when field filter is called
	filterCallOrder := []string{}

	filterLogger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		FieldFilter: func(key string, val any) (string, any, bool) {
			filterCallOrder = append(filterCallOrder, fmt.Sprintf("filter_%s", key))
			return "filtered_" + key, val, true
		},
	}

	// Build up fields in stages
	logger := filterLogger.Info()
	filterCallOrder = append(filterCallOrder, "stage1_complete")

	logger = logger.WithField("field1", "value1")
	// Field filter is called immediately
	filterCallOrder = append(filterCallOrder, "stage2_complete")

	logger = logger.WithField("field2", "value2")
	// Field filter is called immediately
	filterCallOrder = append(filterCallOrder, "stage3_complete")

	// Fields are filtered eagerly when WithField is called
	core.AssertEqual(t, 5, len(filterCallOrder), "filter calls during field attachment")
	core.AssertEqual(t, "stage3_complete", filterCallOrder[4], "stages recorded")

	// Now trigger the actual logging
	logger.Print("test message")

	// Verify fields were filtered at print time (lazy evaluation)
	core.AssertTrue(t, len(filterCallOrder) > 3, "filters called at print time")

	msgs := base.GetMessages()
	core.AssertEqual(t, 1, len(msgs), "message logged")

	// Check filtered fields
	msg := msgs[0]
	core.AssertEqual(t, "value1", msg.Fields["filtered_field1"], "field1 filtered")
	core.AssertEqual(t, "value2", msg.Fields["filtered_field2"], "field2 filtered")
}

// Test field iteration order and completeness
func TestFieldIterationCompleteness(t *testing.T) {
	base := mock.NewLogger()

	// Create a logger that tracks all fields seen
	seenFields := make(map[string]any)

	filterLogger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		FieldsFilter: func(fields slog.Fields) (slog.Fields, bool) {
			// Record all fields we see
			for k, v := range fields {
				seenFields[k] = v
			}
			return fields, true
		},
	}

	// Add fields at different levels
	rootLogger := filterLogger.WithField("root", "value")
	entry := rootLogger.Info().
		WithField("entry1", "value1").
		WithFields(map[string]any{
			"bulk1": "b1",
			"bulk2": "b2",
		}).
		WithField("entry2", "value2")

	entry.Print("test")

	// Verify all fields were seen
	expectedFields := map[string]any{
		"root":   "value",
		"entry1": "value1",
		"bulk1":  "b1",
		"bulk2":  "b2",
		"entry2": "value2",
	}

	for key, expectedVal := range expectedFields {
		actualVal, exists := seenFields[key]
		core.AssertTrue(t, exists, "field %s should be present", key)
		core.AssertEqual(t, expectedVal, actualVal, "field %s value", key)
	}

	// Verify message has all fields
	msgs := base.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "message count")

	msg := msgs[0]
	core.AssertEqual(t, len(expectedFields), len(msg.Fields), "all fields present")
}

// Test field override behaviour in iteration
func TestFieldOverrideInIteration(t *testing.T) {
	base := mock.NewLogger()
	logger := filter.New(base, slog.Debug)

	// Add same field multiple times with different values
	entry := logger.Info().
		WithField("key", "value1").
		WithField("key", "value2"). // Override
		WithFields(map[string]any{
			"key":   "value3", // Override again
			"other": "data",
		}).
		WithField("key", "final") // Final override

	entry.Print("test overrides")

	msgs := base.GetMessages()
	core.AssertEqual(t, 1, len(msgs), "message count")

	// Last value should win
	msg := msgs[0]
	core.AssertEqual(t, "final", msg.Fields["key"], "last value wins")
	core.AssertEqual(t, "data", msg.Fields["other"], "other field preserved")
}

// Test field collection with disabled entries
func TestFieldCollectionDisabled(t *testing.T) {
	base := mock.NewLogger()

	filterCallCount := 0
	filterLogger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Error, // High threshold
		FieldFilter: func(key string, val any) (string, any, bool) {
			filterCallCount++
			return key, val, true
		},
	}

	// Debug entry is disabled
	debugEntry := filterLogger.Debug().
		WithField("debug_field", "value1").
		WithField("debug_field2", "value2")

	core.AssertFalse(t, debugEntry.Enabled(), "debug entry disabled")

	// Try to print (should not trigger field collection)
	debugEntry.Print("debug message")

	core.AssertEqual(t, 0, filterCallCount, "no filter calls for disabled entry")

	msgs := base.GetMessages()
	core.AssertEqual(t, 0, len(msgs), "no messages logged")

	// Now test with enabled entry
	errorEntry := filterLogger.Error().
		WithField("error_field", "value")

	errorEntry.Print("error message")

	core.AssertTrue(t, filterCallCount > 0, "filter called for enabled entry")

	msgs = base.GetMessages()
	core.AssertEqual(t, 1, len(msgs), "error message logged")
}

// Test field collection with filter chains
func TestFieldCollectionInChain(t *testing.T) {
	base := mock.NewLogger()

	// First filter adds prefix
	filter1 := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		FieldFilter: func(key string, val any) (string, any, bool) {
			return "f1_" + key, val, true
		},
	}

	// Second filter adds another prefix
	filter2 := &filter.Logger{
		Parent:    filter1,
		Threshold: slog.Debug,
		FieldFilter: func(key string, val any) (string, any, bool) {
			return "f2_" + key, val, true
		},
	}

	// Add fields at different levels
	logger := filter2.Info().WithField("root", "rootVal")
	entry := logger.WithField("entry", "entryVal")

	entry.Print("chained message")

	msgs := base.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "message count")

	msg := msgs[0]

	// Debug: show what fields we actually got
	fieldKeys := make([]string, 0, len(msg.Fields))
	for k := range msg.Fields {
		fieldKeys = append(fieldKeys, k)
	}
	sort.Strings(fieldKeys)
	t.Logf("Actual fields: %v", fieldKeys)
	for _, k := range fieldKeys {
		t.Logf("  %s = %v", k, msg.Fields[k])
	}

	// Fields should be double-transformed (both filters apply)
	core.AssertEqual(t, "rootVal", msg.Fields["f1_f2_root"], "root field transformed by both")
	core.AssertEqual(t, "entryVal", msg.Fields["f1_f2_entry"], "entry field transformed by both")
}

// Test large field collection performance
func TestLargeFieldCollection(t *testing.T) {
	base := mock.NewLogger()
	logger := filter.New(base, slog.Debug)

	// Start with a base logger
	entry := logger.Info()

	// Add many fields incrementally
	const fieldCount = 100
	for i := 0; i < fieldCount; i++ {
		fieldKey := fmt.Sprintf("field_%03d", i)
		fieldValue := fmt.Sprintf("value_%03d", i)
		entry = entry.WithField(fieldKey, fieldValue)
	}

	// Add a batch of fields
	batchFields := make(map[string]any)
	for i := 0; i < 50; i++ {
		batchKey := fmt.Sprintf("batch_%02d", i)
		batchValue := fmt.Sprintf("batch_value_%02d", i)
		batchFields[batchKey] = batchValue
	}
	entry = entry.WithFields(batchFields)

	// Log the message
	entry.Print("large field collection")

	msgs := base.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "message count")

	msg := msgs[0]
	expectedTotal := fieldCount + len(batchFields)
	core.AssertEqual(t, expectedTotal, len(msg.Fields), "all fields collected")

	// Spot check some fields
	core.AssertEqual(t, "value_000", msg.Fields["field_000"], "first field")
	core.AssertEqual(t, "value_099", msg.Fields["field_099"], "last incremental field")
	core.AssertEqual(t, "batch_value_00", msg.Fields["batch_00"], "first batch field")
}

// Test field iteration with transformation errors
func TestFieldIterationWithErrors(t *testing.T) {
	base := mock.NewLogger()

	processedFields := []string{}
	filterLogger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		FieldFilter: func(key string, val any) (string, any, bool) {
			processedFields = append(processedFields, key)

			// Drop fields with "drop" prefix
			if len(key) > 4 && key[:4] == "drop" {
				return "", nil, false
			}

			// Transform others
			return "ok_" + key, val, true
		},
	}

	entry := filterLogger.Info().
		WithField("keep1", "v1").
		WithField("drop1", "gone").
		WithField("keep2", "v2").
		WithField("drop2", "gone").
		WithField("keep3", "v3")

	entry.Print("test selective drop")

	// All fields should have been processed
	sort.Strings(processedFields)
	expectedProcessed := []string{"drop1", "drop2", "keep1", "keep2", "keep3"}
	sort.Strings(expectedProcessed)
	core.AssertSliceEqual(t, expectedProcessed, processedFields, "all fields processed")

	// Only non-dropped fields should appear in output
	msgs := base.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "message count")

	msg := msgs[0]
	core.AssertEqual(t, 3, len(msg.Fields), "only kept fields in output")
	core.AssertEqual(t, "v1", msg.Fields["ok_keep1"], "keep1 present")
	core.AssertEqual(t, "v2", msg.Fields["ok_keep2"], "keep2 present")
	core.AssertEqual(t, "v3", msg.Fields["ok_keep3"], "keep3 present")
	core.AssertNil(t, msg.Fields["drop1"], "drop1 absent")
	core.AssertNil(t, msg.Fields["drop2"], "drop2 absent")
}

// Test field collection with WithStack
func TestFieldCollectionWithStack(t *testing.T) {
	base := mock.NewLogger()

	// Track field processing order relative to stack
	operations := []string{}

	filterLogger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		FieldFilter: func(key string, val any) (string, any, bool) {
			operations = append(operations, fmt.Sprintf("field:%s", key))
			return key, val, true
		},
	}

	// Mix field and stack operations
	entry := filterLogger.Info().
		WithField("before_stack", "v1").
		WithStack(0).
		WithField("after_stack", "v2")

	entry.Print("mixed operations")

	msgs := base.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "message count")

	msg := msgs[0]
	core.AssertTrue(t, msg.Stack, "stack attached")
	core.AssertEqual(t, "v1", msg.Fields["before_stack"], "field before stack")
	core.AssertEqual(t, "v2", msg.Fields["after_stack"], "field after stack")

	// Verify both fields were processed
	core.AssertContains(t, fmt.Sprint(operations), "field:before_stack", "before_stack processed")
	core.AssertContains(t, fmt.Sprint(operations), "field:after_stack", "after_stack processed")
}

// Test iterator pattern implementation
func TestFieldIteratorPattern(t *testing.T) {
	base := mock.NewLogger()

	// Use FieldsFilter to observe the iterator pattern
	var capturedFields []string

	filterLogger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		FieldsFilter: func(fields slog.Fields) (slog.Fields, bool) {
			// Capture field keys in order seen
			for k := range fields {
				capturedFields = append(capturedFields, k)
			}
			return fields, true
		},
	}

	// Create a complex field chain
	logger := filterLogger.
		WithField("a", 1).
		WithField("b", 2)

	entry := logger.Info().
		WithField("c", 3).
		WithFields(map[string]any{
			"d": 4,
			"e": 5,
		}).
		WithField("f", 6)

	entry.Print("iterator test")

	// Verify all fields were iterated
	expectedFields := map[string]bool{
		"a": true, "b": true, "c": true,
		"d": true, "e": true, "f": true,
	}

	for _, field := range capturedFields {
		delete(expectedFields, field)
	}

	core.AssertEqual(t, 0, len(expectedFields), "all fields iterated")

	msgs := base.GetMessages()
	msg := msgs[0]
	core.AssertEqual(t, 6, len(msg.Fields), "all fields in message")
}
