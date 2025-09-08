package filter_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
	"darvaza.org/slog/handlers/mock"
	slogtest "darvaza.org/slog/internal/testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = fieldsFilterPrimaryTestCase{}
var _ core.TestCase = fieldsFilterReturnFalseTestCase{}
var _ core.TestCase = fieldsFilterComplexHierarchyTestCase{}

// fieldsFilterPrimaryTestCase tests FieldsFilter as primary transformer
type fieldsFilterPrimaryTestCase struct {
	inputFields    map[string]any
	expectedFields map[string]any
	name           string
	description    string
}

func (tc fieldsFilterPrimaryTestCase) Name() string {
	return tc.name
}

func (tc fieldsFilterPrimaryTestCase) Test(t *testing.T) {
	t.Helper()

	base := mock.NewLogger()
	transformCalls := 0

	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		// Only FieldsFilter, no FieldFilter
		FieldsFilter: func(fields slog.Fields) (slog.Fields, bool) {
			transformCalls++
			result := make(map[string]any)
			for k, v := range fields {
				// Transform: prefix with "transformed_"
				result["transformed_"+k] = v
			}
			return result, true
		},
	}

	entry := logger.Info().WithFields(tc.inputFields)
	entry.Print("test message")

	msgs := base.GetMessages()
	slogtest.AssertMustMessageCount(t, msgs, 1)
	core.AssertEqual(t, 1, transformCalls, "FieldsFilter calls")

	actualFields := msgs[0].Fields
	for expectedKey, expectedValue := range tc.expectedFields {
		core.AssertEqual(t, expectedValue, actualFields[expectedKey],
			"field %s", expectedKey)
	}

	core.AssertEqual(t, len(tc.expectedFields), len(actualFields),
		"total field count")
}

func newFieldsFilterPrimaryTestCase(name, description string,
	inputFields, expectedFields map[string]any) fieldsFilterPrimaryTestCase {
	return fieldsFilterPrimaryTestCase{
		name:           name,
		description:    description,
		inputFields:    inputFields,
		expectedFields: expectedFields,
	}
}

func fieldsFilterPrimaryTestCases() []fieldsFilterPrimaryTestCase {
	return []fieldsFilterPrimaryTestCase{
		newFieldsFilterPrimaryTestCase(
			"Single field",
			"FieldsFilter should transform single field",
			map[string]any{"key": "value"},
			map[string]any{"transformed_key": "value"},
		),
		newFieldsFilterPrimaryTestCase(
			"Multiple fields",
			"FieldsFilter should transform all fields",
			map[string]any{"key1": "value1", "key2": "value2", "key3": 123},
			map[string]any{"transformed_key1": "value1", "transformed_key2": "value2", "transformed_key3": 123},
		),
		// Note: FieldsFilter is not called for empty field maps
		// This is an optimization - no point filtering nothing
	}
}

func TestFieldsFilterAsPrimary(t *testing.T) {
	core.RunTestCases(t, fieldsFilterPrimaryTestCases())
}

// fieldsFilterReturnFalseTestCase tests FieldsFilter returning false (drop all fields)
type fieldsFilterReturnFalseTestCase struct {
	inputFields map[string]any
	name        string
	shouldDrop  bool
}

func (tc fieldsFilterReturnFalseTestCase) Name() string {
	return tc.name
}

func (tc fieldsFilterReturnFalseTestCase) Test(t *testing.T) {
	t.Helper()

	base := mock.NewLogger()
	filterCalls := 0

	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		FieldsFilter: func(fields slog.Fields) (slog.Fields, bool) {
			filterCalls++
			// Return false to drop all fields when shouldDrop is true
			if tc.shouldDrop {
				return nil, false
			}
			// Otherwise pass through
			return fields, true
		},
	}

	entry := logger.Info().WithFields(tc.inputFields)
	entry.Print("test message")

	msgs := base.GetMessages()
	slogtest.AssertMustMessageCount(t, msgs, 1)
	core.AssertEqual(t, 1, filterCalls, "FieldsFilter calls")

	actualFields := msgs[0].Fields
	if tc.shouldDrop {
		core.AssertEqual(t, 0, len(actualFields), "all fields should be dropped")
	} else {
		core.AssertEqual(t, len(tc.inputFields), len(actualFields), "fields should pass through")
	}
}

func newFieldsFilterReturnFalseTestCase(name string, inputFields map[string]any,
	shouldDrop bool) fieldsFilterReturnFalseTestCase {
	return fieldsFilterReturnFalseTestCase{
		name:        name,
		inputFields: inputFields,
		shouldDrop:  shouldDrop,
	}
}

func fieldsFilterReturnFalseTestCases() []fieldsFilterReturnFalseTestCase {
	return []fieldsFilterReturnFalseTestCase{
		newFieldsFilterReturnFalseTestCase(
			"Drop all fields",
			map[string]any{"key1": "value1", "key2": "value2"},
			true,
		),
		newFieldsFilterReturnFalseTestCase(
			"Pass through fields",
			map[string]any{"key1": "value1", "key2": "value2"},
			false,
		),
		// Note: FieldsFilter is not called for empty/nil maps
		// as an optimization
	}
}

func TestFieldsFilterReturnFalse(t *testing.T) {
	core.RunTestCases(t, fieldsFilterReturnFalseTestCases())
}

// fieldsFilterComplexHierarchyTestCase tests complex filter hierarchy scenarios
type fieldsFilterComplexHierarchyTestCase struct {
	setupLogger    func() *filter.Logger
	inputFields    map[string]any
	expectedFields map[string]any
	name           string
	description    string
}

func (tc fieldsFilterComplexHierarchyTestCase) Name() string {
	return tc.name
}

func (tc fieldsFilterComplexHierarchyTestCase) Test(t *testing.T) {
	t.Helper()

	logger := tc.setupLogger()
	base := logger.Parent.(*mock.Logger)

	entry := logger.Info()
	// Test both WithField and WithFields in combination
	entry = entry.WithField("single", "value")
	entry = entry.WithFields(tc.inputFields)
	entry.Print("test message")

	msgs := base.GetMessages()
	slogtest.AssertMustMessageCount(t, msgs, 1)

	actualFields := msgs[0].Fields
	for expectedKey, expectedValue := range tc.expectedFields {
		core.AssertEqual(t, expectedValue, actualFields[expectedKey],
			"field %s", expectedKey)
	}

	// Allow for extra fields from the single WithField
	core.AssertTrue(t, len(actualFields) >= len(tc.expectedFields),
		"should have at least expected fields")
}

func newFieldsFilterComplexHierarchyTestCase(name, description string,
	setupLogger func() *filter.Logger,
	inputFields, expectedFields map[string]any) fieldsFilterComplexHierarchyTestCase {
	return fieldsFilterComplexHierarchyTestCase{
		name:           name,
		description:    description,
		setupLogger:    setupLogger,
		inputFields:    inputFields,
		expectedFields: expectedFields,
	}
}

func fieldsFilterComplexHierarchyTestCases() []fieldsFilterComplexHierarchyTestCase {
	return []fieldsFilterComplexHierarchyTestCase{
		newFieldsFilterComplexHierarchyTestCase(
			"Both filters present",
			"Both FieldFilter and FieldsFilter should work in hierarchy",
			func() *filter.Logger {
				base := mock.NewLogger()
				return &filter.Logger{
					Parent:    base,
					Threshold: slog.Debug,
					FieldFilter: func(key string, val any) (string, any, bool) {
						// For single fields, prefix with "field_"
						return "field_" + key, val, true
					},
					FieldsFilter: func(fields slog.Fields) (slog.Fields, bool) {
						// For multiple fields, prefix with "fields_"
						result := make(map[string]any)
						for k, v := range fields {
							result["fields_"+k] = v
						}
						return result, true
					},
				}
			},
			map[string]any{"multi1": "v1", "multi2": "v2"},
			map[string]any{
				"field_single":  "value", // From WithField
				"fields_multi1": "v1",    // From WithFields
				"fields_multi2": "v2",    // From WithFields
			},
		),
		newFieldsFilterComplexHierarchyTestCase(
			"FieldsFilter with fallback",
			"FieldsFilter falls back to FieldFilter for WithFields when nil",
			func() *filter.Logger {
				base := mock.NewLogger()
				return &filter.Logger{
					Parent:    base,
					Threshold: slog.Debug,
					FieldFilter: func(key string, val any) (string, any, bool) {
						// Should be used for both WithField and WithFields
						return "fallback_" + key, val, true
					},
					// No FieldsFilter
				}
			},
			map[string]any{"multi1": "v1", "multi2": "v2"},
			map[string]any{
				"fallback_single": "value", // From WithField
				"fallback_multi1": "v1",    // From WithFields (fallback)
				"fallback_multi2": "v2",    // From WithFields (fallback)
			},
		),
		newFieldsFilterComplexHierarchyTestCase(
			"Selective field dropping",
			"Filters should selectively drop fields based on content",
			func() *filter.Logger {
				base := mock.NewLogger()
				return &filter.Logger{
					Parent:    base,
					Threshold: slog.Debug,
					FieldFilter: func(key string, val any) (string, any, bool) {
						// Drop fields starting with underscore
						if key[0] == '_' {
							return "", nil, false
						}
						return "single_" + key, val, true
					},
					FieldsFilter: func(fields slog.Fields) (slog.Fields, bool) {
						result := make(map[string]any)
						for k, v := range fields {
							// Drop private fields
							if k[0] != '_' {
								result["multi_"+k] = v
							}
						}
						return result, len(result) > 0
					},
				}
			},
			map[string]any{"public": "visible", "_private": "hidden"},
			map[string]any{
				"single_single": "value",   // From WithField
				"multi_public":  "visible", // From WithFields
				// "_private" should be dropped
			},
		),
	}
}

func TestFieldsFilterComplexHierarchy(t *testing.T) {
	core.RunTestCases(t, fieldsFilterComplexHierarchyTestCases())
}

func runTestFieldsFilterChaining(t *testing.T) {
	t.Helper()
	base := mock.NewLogger()

	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		FieldsFilter: func(fields slog.Fields) (slog.Fields, bool) {
			// Add a prefix to all fields
			result := make(map[string]any)
			for k, v := range fields {
				result["chained_"+k] = v
			}
			return result, true
		},
	}

	// Chain multiple field additions
	entry := logger.Info().
		WithFields(map[string]any{"first": "1"}).
		WithFields(map[string]any{"second": "2"}).
		WithFields(map[string]any{"third": "3"})

	entry.Print("chained test")

	msgs := base.GetMessages()
	slogtest.AssertMustMessageCount(t, msgs, 1)

	actualFields := msgs[0].Fields
	expectedFields := map[string]any{
		"chained_first":  "1",
		"chained_second": "2",
		"chained_third":  "3",
	}

	for expectedKey, expectedValue := range expectedFields {
		core.AssertEqual(t, expectedValue, actualFields[expectedKey],
			"field %s", expectedKey)
	}
}

// Test FieldsFilter with chained loggers
func TestFieldsFilterChaining(t *testing.T) {
	t.Run("chained fields", runTestFieldsFilterChaining)
}

func runTestFieldsFilterWithStack(t *testing.T) {
	t.Helper()
	base := mock.NewLogger()

	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		FieldsFilter: func(fields slog.Fields) (slog.Fields, bool) {
			// Pass through but track that we were called
			result := make(map[string]any)
			for k, v := range fields {
				result[k] = v
			}
			result["filter_called"] = true
			return result, true
		},
	}

	entry := logger.Info().
		WithStack(0).
		WithFields(map[string]any{"key": "value"})

	entry.Print("with stack")

	msgs := base.GetMessages()
	slogtest.AssertMustMessageCount(t, msgs, 1)

	msg := msgs[0]
	core.AssertTrue(t, msg.Stack, "stack should be present")
	slogtest.AssertField(t, msg, "key", "value")
	slogtest.AssertField(t, msg, "filter_called", true)
}

// Test FieldsFilter interaction with stack traces
func TestFieldsFilterWithStack(t *testing.T) {
	t.Run("fields with stack", runTestFieldsFilterWithStack)
}

func runTestFieldsFilterEmptyMapOptimization(t *testing.T) {
	t.Helper()
	base := mock.NewLogger()
	filterCalled := false

	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		FieldsFilter: func(fields slog.Fields) (slog.Fields, bool) {
			filterCalled = true
			return fields, true
		},
	}

	// Test with empty map
	entry := logger.Info().WithFields(map[string]any{})
	entry.Print("empty map test")

	core.AssertFalse(t, filterCalled, "FieldsFilter should NOT be called for empty map")

	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, 1)
	core.AssertEqual(t, 0, len(msgs[0].Fields), "no fields in message")

	// Test with nil map
	filterCalled = false
	entry2 := logger.Info().WithFields(nil)
	entry2.Print("nil map test")

	core.AssertFalse(t, filterCalled, "FieldsFilter should NOT be called for nil map")

	msgs = base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, 2)
	core.AssertEqual(t, 0, len(msgs[1].Fields), "no fields in second message")

	// Verify it IS called for non-empty maps
	filterCalled = false
	entry3 := logger.Info().WithFields(map[string]any{"key": "value"})
	entry3.Print("non-empty map test")

	core.AssertTrue(t, filterCalled, "FieldsFilter SHOULD be called for non-empty map")
}

// Test that FieldsFilter is NOT called for empty maps (optimization)
func TestFieldsFilterEmptyMapOptimization(t *testing.T) {
	t.Run("empty map optimization", runTestFieldsFilterEmptyMapOptimization)
}
