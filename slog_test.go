package slog_test

import (
	"fmt"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = logLevelConstantTestCase{}
var _ core.TestCase = logLevelOrderingTestCase{}

func TestLogLevel(t *testing.T) {
	t.Run("Constants", testLogLevelConstants)
	t.Run("Ordering", testLogLevelOrdering)
}

// logLevelConstantTestCase represents a test case for log level constants.
type logLevelConstantTestCase struct {
	name  string
	level slog.LogLevel
	value slog.LogLevel
}

func (tc logLevelConstantTestCase) Name() string {
	return tc.name
}

func (tc logLevelConstantTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.value, tc.level, "log level constant %s", tc.name)
}

func newLogLevelConstantTestCase(name string, level, value slog.LogLevel) logLevelConstantTestCase {
	return logLevelConstantTestCase{
		name:  name,
		level: level,
		value: value,
	}
}

func logLevelConstantTestCases() []logLevelConstantTestCase {
	return []logLevelConstantTestCase{
		newLogLevelConstantTestCase("UndefinedLevel", slog.UndefinedLevel, 0),
		newLogLevelConstantTestCase("Panic", slog.Panic, 1),
		newLogLevelConstantTestCase("Fatal", slog.Fatal, 2),
		newLogLevelConstantTestCase("Error", slog.Error, 3),
		newLogLevelConstantTestCase("Warn", slog.Warn, 4),
		newLogLevelConstantTestCase("Info", slog.Info, 5),
		newLogLevelConstantTestCase("Debug", slog.Debug, 6),
	}
}

func testLogLevelConstants(t *testing.T) {
	core.RunTestCases(t, logLevelConstantTestCases())
}

// logLevelOrderingTestCase represents a test case for log level ordering.
type logLevelOrderingTestCase struct {
	name      string
	prevLevel slog.LogLevel
	currLevel slog.LogLevel
}

func (tc logLevelOrderingTestCase) Name() string {
	return tc.name
}

func (tc logLevelOrderingTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertTrue(t, tc.prevLevel < tc.currLevel, "level ordering %s", tc.name)
}

func newLogLevelOrderingTestCase(name string, prevLevel, currLevel slog.LogLevel) logLevelOrderingTestCase {
	return logLevelOrderingTestCase{
		name:      name,
		prevLevel: prevLevel,
		currLevel: currLevel,
	}
}

func logLevelOrderingTestCases() []logLevelOrderingTestCase {
	levels := []slog.LogLevel{
		slog.UndefinedLevel,
		slog.Panic,
		slog.Fatal,
		slog.Error,
		slog.Warn,
		slog.Info,
		slog.Debug,
	}

	var cases []logLevelOrderingTestCase
	for i := 1; i < len(levels); i++ {
		cases = append(cases, newLogLevelOrderingTestCase(
			fmt.Sprintf("%v < %v", levels[i-1], levels[i]),
			levels[i-1], levels[i]))
	}
	return cases
}

func testLogLevelOrdering(t *testing.T) {
	core.RunTestCases(t, logLevelOrderingTestCases())
}

func TestFields(t *testing.T) {
	t.Run("TypeAlias", testFieldsTypeAlias)
	t.Run("MapCompatibility", testFieldsMapCompatibility)
}

func testFieldsTypeAlias(t *testing.T) {
	t.Helper()
	// Test that Fields is a proper map alias
	fields := slog.Fields{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	// Should behave like a map
	core.AssertEqual(t, 3, len(fields), "field count")
	core.AssertEqual(t, "value1", fields["key1"], "field access")
}

func testFieldsMapCompatibility(t *testing.T) {
	t.Helper()
	// Test that Fields can be used where map[string]any is expected
	fields := slog.Fields{
		"test": "value",
	}

	// Should be assignable to map[string]any
	var m map[string]any = fields
	core.AssertEqual(t, "value", m["test"], "map compatibility")

	// Should be convertible from map[string]any
	original := map[string]any{
		"convert": "test",
	}
	converted := slog.Fields(original)
	core.AssertEqual(t, "test", converted["convert"], "map conversion")
}

func TestErrorFieldName(t *testing.T) {
	// Test the error field name constant
	core.AssertEqual(t, "error", slog.ErrorFieldName, "error field name")
}
