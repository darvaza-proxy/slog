package filter_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
	"darvaza.org/slog/handlers/mock"
	slogtest "darvaza.org/slog/internal/testing"
)

const testDropFieldKey = "drop"

func TestLevel(t *testing.T) {
	// Test nil receiver for LogEntry.Level()
	var nilEntry *filter.LogEntry
	core.AssertEqual(t, slog.UndefinedLevel, nilEntry.Level(), "nil log entry level")

	// Test normal functionality
	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Info)

	infoEntry := logger.Info()
	filterEntry := core.AssertMustTypeIs[*filter.LogEntry](t, infoEntry, "info entry type")
	core.AssertEqual(t, slog.Info, filterEntry.Level(), "info level")

	errorEntry := logger.Error()
	filterError := core.AssertMustTypeIs[*filter.LogEntry](t, errorEntry, "error entry type")
	core.AssertEqual(t, slog.Error, filterError.Level(), "error level")
}

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = filterLogletTestCase{}
var _ core.TestCase = filterFieldsTestCase{}
var _ core.TestCase = filterStackTestCase{}
var _ core.TestCase = filterChainingTestCase{}
var _ core.TestCase = filterFieldTransformationTestCase{}
var _ core.TestCase = filterMessageFilterTestCase{}
var _ core.TestCase = filterParentlessTestCase{}
var _ core.TestCase = filterHierarchyTestCase{}

type filterLogletTestCase struct {
	method  func() slog.Logger
	name    string
	level   slog.LogLevel
	enabled bool
}

func (tc filterLogletTestCase) Name() string {
	return tc.name
}

func (tc filterLogletTestCase) Test(t *testing.T) {
	t.Helper()

	l := tc.method()
	core.AssertMustNotNil(t, l, "logger method returned nil")

	core.AssertEqual(t, tc.enabled, l.Enabled(), "Enabled() for level %s", tc.name)
}

func newFilterLogletTestCase(name string,
	method func() slog.Logger,
	level slog.LogLevel, enabled bool) filterLogletTestCase {
	return filterLogletTestCase{
		name:    name,
		method:  method,
		level:   level,
		enabled: enabled,
	}
}

func newTestLogger() slog.Logger {
	return mock.NewLogger()
}

func filterLogletTestCases() []filterLogletTestCase {
	base := newTestLogger()
	logger := filter.New(base, slog.Info)
	return []filterLogletTestCase{
		newFilterLogletTestCase("Debug", logger.Debug, slog.Debug, false),
		newFilterLogletTestCase("Info", logger.Info, slog.Info, true),
		newFilterLogletTestCase("Warn", logger.Warn, slog.Warn, true),
		newFilterLogletTestCase("Error", logger.Error, slog.Error, true),
		newFilterLogletTestCase("Fatal", logger.Fatal, slog.Fatal, true),
		newFilterLogletTestCase("Panic", logger.Panic, slog.Panic, true),
	}
}

func TestFilterLoglet(t *testing.T) {
	// Test basic level methods using internal/testing utilities
	t.Run("LevelMethods", testFilterLevelMethods)
	t.Run("ThresholdFiltering", testFilterThresholdFiltering)
	t.Run("LegacyTestCases", testFilterLogletLegacyCases)
}

func testFilterLogletLegacyCases(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, filterLogletTestCases())
}

func testFilterLevelMethods(t *testing.T) {
	t.Helper()
	slogtest.TestLevelMethods(t, func() slog.Logger {
		base := newTestLogger()
		return filter.New(base, slog.Debug) // Use Debug to allow all levels
	})
}

func testFilterThresholdFiltering(t *testing.T) {
	base := newTestLogger()
	logger := filter.New(base, slog.Info)

	// Test level transitions with threshold
	testLevels := []struct {
		method  func() slog.Logger
		name    string
		level   slog.LogLevel
		enabled bool
	}{
		{logger.Debug, "Debug", slog.Debug, false},
		{logger.Info, "Info", slog.Info, true},
		{logger.Warn, "Warn", slog.Warn, true},
		{logger.Error, "Error", slog.Error, true},
		{logger.Fatal, "Fatal", slog.Fatal, true},
		{logger.Panic, "Panic", slog.Panic, true},
	}

	for _, tt := range testLevels {
		slogtest.RunWithLogger(t, tt.name, logger, func(t core.T, _ slog.Logger) {
			l := tt.method()
			core.AssertMustNotNil(t, l, "logger method returned nil")
			core.AssertEqual(t, tt.enabled, l.Enabled(), "Expected %s enabled=%t", tt.name, tt.enabled)
		})
	}
}

type filterFieldsTestCase struct {
	name string
}

func (tc filterFieldsTestCase) Name() string {
	return tc.name
}

func (tc filterFieldsTestCase) Test(t *testing.T) {
	t.Helper()

	base := newTestLogger()
	logger := filter.New(base, slog.Info)

	l1 := logger.WithField("root", "value")
	core.AssertMustNotNil(t, l1, "WithField returned nil")

	l2 := logger.Info().WithField("key1", "value1")
	core.AssertMustNotNil(t, l2, "WithField on enabled logger returned nil")

	fields := map[string]any{
		"key2": "value2",
		"key3": 123,
	}
	l3 := logger.Info().WithFields(fields)
	core.AssertNotNil(t, l3, "WithFields returned nil")
}

func newFilterFieldsTestCase(name string) filterFieldsTestCase {
	return filterFieldsTestCase{
		name: name,
	}
}

func TestFilterWithFields(t *testing.T) {
	// Use TestFieldMethods which tests both WithField and WithFields
	t.Run("FieldMethods", testFilterFieldMethods)
	t.Run("LegacyTestCases", testFilterFieldsLegacyCases)
}

func testFilterFieldMethods(t *testing.T) {
	t.Helper()
	slogtest.TestFieldMethods(t, func() slog.Logger {
		base := newTestLogger()
		return filter.New(base, slog.Info)
	})
}

func testFilterFieldsLegacyCases(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, []filterFieldsTestCase{
		newFilterFieldsTestCase("WithFields"),
	})
}

type filterStackTestCase struct {
	name string
}

func (tc filterStackTestCase) Name() string {
	return tc.name
}

func (tc filterStackTestCase) Test(t *testing.T) {
	t.Helper()

	base := newTestLogger()
	logger := filter.New(base, slog.Info)

	l1 := logger.WithStack(1)
	core.AssertMustNotNil(t, l1, "WithStack on root returned nil")

	l2 := logger.Info().WithStack(1)
	core.AssertNotNil(t, l2, "WithStack on enabled logger returned nil")
}

func newFilterStackTestCase(name string) filterStackTestCase {
	return filterStackTestCase{
		name: name,
	}
}

func TestFilterWithStack(t *testing.T) {
	// Use slogtest utility for comprehensive testing
	t.Run("StackMethods", testFilterStackMethods)
	t.Run("LegacyTestCases", testFilterStackLegacyCases)
}

func testFilterStackMethods(t *testing.T) {
	t.Helper()
	base := newTestLogger()
	logger := filter.New(base, slog.Info)
	slogtest.TestWithStack(t, logger)
}

func testFilterStackLegacyCases(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, []filterStackTestCase{
		newFilterStackTestCase("WithStack"),
	})
}

type filterChainingTestCase struct {
	name string
}

func (tc filterChainingTestCase) Name() string {
	return tc.name
}

func (tc filterChainingTestCase) Test(t *testing.T) {
	t.Helper()

	base := newTestLogger()
	logger := filter.New(base, slog.Info)

	l := logger.
		WithField("key1", "value1").
		WithField("key2", "value2").
		Info().
		WithField("key3", "value3")

	core.AssertMustNotNil(t, l, "Chained logger is nil")

	core.AssertTrue(t, l.Enabled(), "Info logger enabled")
}

func newFilterChainingTestCase(name string) filterChainingTestCase {
	return filterChainingTestCase{
		name: name,
	}
}

func TestFilterChaining(t *testing.T) {
	core.RunTestCases(t, []filterChainingTestCase{
		newFilterChainingTestCase("Chaining"),
	})
}

type filterFieldTransformationTestCase struct {
	name string
}

func (tc filterFieldTransformationTestCase) Name() string {
	return tc.name
}

func (tc filterFieldTransformationTestCase) Test(t *testing.T) {
	t.Helper()

	base := newTestLogger()

	transformed := false
	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Info,
		FieldFilter: func(key string, val any) (string, any, bool) {
			transformed = true
			if key == sensitiveKey1 {
				return key, redactedValue, true
			}
			return key, val, true
		},
	}

	l := logger.Info().WithField("password", "secret123")
	l.Print("test")

	core.AssertTrue(t, transformed, "FieldFilter was not called")
}

func newFilterFieldTransformationTestCase(name string) filterFieldTransformationTestCase {
	return filterFieldTransformationTestCase{
		name: name,
	}
}

func TestFilterFieldTransformation(t *testing.T) {
	core.RunTestCases(t, []filterFieldTransformationTestCase{
		newFilterFieldTransformationTestCase("FieldTransformation"),
	})
}

type filterMessageFilterTestCase struct {
	name string
}

func (tc filterMessageFilterTestCase) Name() string {
	return tc.name
}

func (tc filterMessageFilterTestCase) Test(t *testing.T) {
	t.Helper()

	base := newTestLogger()

	filtered := false
	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Info,
		MessageFilter: func(msg string) (string, bool) {
			filtered = true
			return "[FILTERED] " + msg, true
		},
	}

	l := logger.Info()
	l.Print("test message")

	core.AssertTrue(t, filtered, "MessageFilter was not called")
}

func newFilterMessageFilterTestCase(name string) filterMessageFilterTestCase {
	return filterMessageFilterTestCase{
		name: name,
	}
}

func TestFilterMessageFilter(t *testing.T) {
	core.RunTestCases(t, []filterMessageFilterTestCase{
		newFilterMessageFilterTestCase("MessageFilter"),
	})
}

type filterParentlessTestCase struct {
	name string
}

func (tc filterParentlessTestCase) Name() string {
	return tc.name
}

func (tc filterParentlessTestCase) Test(t *testing.T) {
	t.Helper()

	logger := filter.NewNoop()

	logger.Debug().Print("test")
	logger.Info().Print("test")
	logger.Error().Print("test")

	core.AssertTrue(t, logger.Fatal().Enabled(), "Fatal parentless enabled")
	core.AssertTrue(t, logger.Panic().Enabled(), "Panic parentless enabled")
}

func newFilterParentlessTestCase(name string) filterParentlessTestCase {
	return filterParentlessTestCase{
		name: name,
	}
}

func TestFilterParentless(t *testing.T) {
	core.RunTestCases(t, []filterParentlessTestCase{
		newFilterParentlessTestCase("Parentless"),
	})
}

// filterHierarchyTestCase tests the filter hierarchy for WithField/WithFields
type filterHierarchyTestCase struct {
	operation      func(entry slog.Logger) slog.Logger
	expectedFields map[string]any
	name           string
	description    string
	expectedCalls  hierarchyCallTracker
}

func (tc filterHierarchyTestCase) Name() string {
	return tc.name
}

func (tc filterHierarchyTestCase) Test(t *testing.T) {
	t.Helper()

	base := mock.NewLogger()
	tracker := newHierarchyCallTracker()

	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		FieldFilter: func(key string, val any) (string, any, bool) {
			tracker.fieldFilterCalls++
			tracker.fieldFilterKeys = append(tracker.fieldFilterKeys, key)
			if key == testDropFieldKey {
				return "", nil, false
			}
			return "field_" + key, val, true
		},
		FieldsFilter: func(fields slog.Fields) (slog.Fields, bool) {
			tracker.fieldsFilterCalls++
			result := make(slog.Fields)
			for k, v := range fields {
				if k != testDropFieldKey {
					result["fields_"+k] = v
				}
			}
			return result, len(result) > 0
		},
	}

	// Execute the operation
	entry := logger.Info()
	result := tc.operation(entry)
	result.Print("test")

	// Verify calls match expectations
	core.AssertEqual(t, tc.expectedCalls.fieldFilterCalls, tracker.fieldFilterCalls,
		"FieldFilter calls")
	core.AssertEqual(t, tc.expectedCalls.fieldsFilterCalls, tracker.fieldsFilterCalls,
		"FieldsFilter calls")

	// Verify resulting fields
	msgs := base.GetMessages()
	slogtest.AssertMustMessageCount(t, msgs, 1)

	actualFields := msgs[0].Fields
	for expectedKey, expectedValue := range tc.expectedFields {
		core.AssertEqual(t, expectedValue, actualFields[expectedKey],
			"field %s", expectedKey)
	}

	// Verify no unexpected fields
	core.AssertEqual(t, len(tc.expectedFields), len(actualFields),
		"total field count")
}

// hierarchyCallTracker tracks filter function calls
type hierarchyCallTracker struct {
	fieldFilterKeys   []string
	fieldFilterCalls  int
	fieldsFilterCalls int
}

func newHierarchyCallTracker() *hierarchyCallTracker {
	return &hierarchyCallTracker{
		fieldFilterKeys: make([]string, 0),
	}
}

// Factory functions for filter hierarchy test cases

func newFilterHierarchyTestCase(name, description string,
	operation func(entry slog.Logger) slog.Logger,
	expectedFields map[string]any,
	expectedCalls hierarchyCallTracker) filterHierarchyTestCase {
	return filterHierarchyTestCase{
		name:           name,
		description:    description,
		operation:      operation,
		expectedFields: expectedFields,
		expectedCalls:  expectedCalls,
	}
}

func newWithFieldTestCase(name, description, key string, value any,
	expectedFields map[string]any, expectedCalls hierarchyCallTracker) filterHierarchyTestCase {
	return newFilterHierarchyTestCase(name, description,
		func(entry slog.Logger) slog.Logger {
			return entry.WithField(key, value)
		},
		expectedFields, expectedCalls)
}

func newWithFieldsTestCase(name, description string, fields map[string]any,
	expectedFields map[string]any, expectedCalls hierarchyCallTracker) filterHierarchyTestCase {
	return newFilterHierarchyTestCase(name, description,
		func(entry slog.Logger) slog.Logger {
			return entry.WithFields(fields)
		},
		expectedFields, expectedCalls)
}

// Test case list factory
func filterHierarchyTestCases() []filterHierarchyTestCase {
	return []filterHierarchyTestCase{
		// WithField() hierarchy tests
		newWithFieldTestCase(
			"WithField-FieldFilter",
			"WithField should try FieldFilter first (most specific)",
			"test", "value",
			map[string]any{"field_test": "value"},
			hierarchyCallTracker{fieldFilterCalls: 1, fieldsFilterCalls: 0},
		),
		newWithFieldTestCase(
			"WithField-Dropped",
			"WithField should drop field when FieldFilter returns false",
			testDropFieldKey, "value",
			map[string]any{},
			hierarchyCallTracker{fieldFilterCalls: 1, fieldsFilterCalls: 0},
		),

		// WithFields() hierarchy tests
		newWithFieldsTestCase(
			"WithFields-FieldsFilter",
			"WithFields should try FieldsFilter first (most specific)",
			map[string]any{"test": "value"},
			map[string]any{"fields_test": "value"},
			hierarchyCallTracker{fieldFilterCalls: 0, fieldsFilterCalls: 1},
		),
		newWithFieldsTestCase(
			"WithFields-Multiple",
			"WithFields should process all fields through FieldsFilter",
			map[string]any{"key1": "value1", "key2": "value2"},
			map[string]any{"fields_key1": "value1", "fields_key2": "value2"},
			hierarchyCallTracker{fieldFilterCalls: 0, fieldsFilterCalls: 1},
		),
		newWithFieldsTestCase(
			"WithFields-PartialDrop",
			"WithFields should drop some fields when filtered",
			map[string]any{"keep": "value", testDropFieldKey: "remove"},
			map[string]any{"fields_keep": "value"},
			hierarchyCallTracker{fieldFilterCalls: 0, fieldsFilterCalls: 1},
		),
	}
}

func TestFilterHierarchy(t *testing.T) {
	core.RunTestCases(t, filterHierarchyTestCases())
}

// Additional test for fallback behaviour - when FieldsFilter is not defined
func filterHierarchyFallbackTestCases() []filterHierarchyTestCase {
	return []filterHierarchyTestCase{
		// WithField() with no FieldsFilter - should only try FieldFilter
		newWithFieldTestCase(
			"WithField-NoFieldsFilter",
			"WithField with only FieldFilter should work normally",
			"test", "value",
			map[string]any{"field_test": "value"},
			hierarchyCallTracker{fieldFilterCalls: 1, fieldsFilterCalls: 0},
		),
	}
}

func runTestFilterHierarchyFallbackCase(t *testing.T, tc filterHierarchyTestCase) {
	t.Helper()

	base := mock.NewLogger()
	tracker := newHierarchyCallTracker()

	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		FieldFilter: func(key string, val any) (string, any, bool) {
			tracker.fieldFilterCalls++
			return "field_" + key, val, true
		},
		// FieldsFilter intentionally nil
	}

	entry := logger.Info()
	result := tc.operation(entry)
	result.Print("test")

	core.AssertEqual(t, tc.expectedCalls.fieldFilterCalls, tracker.fieldFilterCalls,
		"FieldFilter calls")
}

func runTestFilterHierarchyFallbackOnlyFieldFilter(t *testing.T) {
	t.Helper()

	// Test cases for logger with only FieldFilter defined
	for _, tc := range filterHierarchyFallbackTestCases() {
		t.Run(tc.Name(), func(t *testing.T) {
			runTestFilterHierarchyFallbackCase(t, tc)
		})
	}
}

func TestFilterHierarchyFallback(t *testing.T) {
	t.Run("OnlyFieldFilter", runTestFilterHierarchyFallbackOnlyFieldFilter)
}
