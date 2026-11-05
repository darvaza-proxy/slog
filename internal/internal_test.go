package internal_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/internal"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ core.TestCase = hasFieldsTestCase{}
	_ core.TestCase = isTerminalLevelTestCase{}
)

func TestHasFields(t *testing.T) {
	t.Run("EmptyMap", testHasFieldsEmpty)
	t.Run("NilMap", testHasFieldsNil)
	t.Run("ValidFields", testHasFieldsValid)
	t.Run("EmptyKeys", testHasFieldsEmptyKeys)
	t.Run("MixedKeys", testHasFieldsMixed)
}

func testHasFieldsEmpty(t *testing.T) {
	t.Helper()
	fields := map[string]any{}

	core.AssertFalse(t, internal.HasFields(fields), "HasFields with empty map")
}

func testHasFieldsNil(t *testing.T) {
	t.Helper()
	var fields map[string]any

	core.AssertFalse(t, internal.HasFields(fields), "HasFields with nil map")
}

// hasFieldsTestCase represents a test case for HasFields function.
type hasFieldsTestCase struct {
	fields map[string]any
	name   string
	want   bool
}

func (tc hasFieldsTestCase) Name() string {
	return tc.name
}

func (tc hasFieldsTestCase) Test(t *testing.T) {
	t.Helper()
	got := internal.HasFields(tc.fields)
	core.AssertEqual(t, tc.want, got, "HasFields result")
}

func newHasFieldsTestCase(name string, fields map[string]any, want bool) hasFieldsTestCase {
	return hasFieldsTestCase{
		name:   name,
		fields: fields,
		want:   want,
	}
}

func hasFieldsValidTestCases() []hasFieldsTestCase {
	return []hasFieldsTestCase{
		newHasFieldsTestCase("SingleField", map[string]any{"key": "value"}, true),
		newHasFieldsTestCase("MultipleFields", map[string]any{"key1": "value1", "key2": "value2"}, true),
		newHasFieldsTestCase("NilValue", map[string]any{"key": nil}, true),
		newHasFieldsTestCase("ZeroValue", map[string]any{"key": 0}, true),
		newHasFieldsTestCase("EmptyStringValue", map[string]any{"key": ""}, true),
	}
}

func testHasFieldsValid(t *testing.T) {
	core.RunTestCases(t, hasFieldsValidTestCases())
}

func hasFieldsEmptyKeysTestCases() []hasFieldsTestCase {
	return []hasFieldsTestCase{
		newHasFieldsTestCase("OnlyEmptyKey", map[string]any{"": "value"}, false),
		newHasFieldsTestCase("OnlyEmptyKeySingle", map[string]any{"": "value1"}, false),
	}
}

func testHasFieldsEmptyKeys(t *testing.T) {
	core.RunTestCases(t, hasFieldsEmptyKeysTestCases())
}

func hasFieldsMixedTestCases() []hasFieldsTestCase {
	return []hasFieldsTestCase{
		newHasFieldsTestCase("ValidAndEmptyKeys",
			map[string]any{"": "empty", "valid": "value"}, true), // Should return true because "valid" is non-empty
		newHasFieldsTestCase("OnlyEmptyKeysMultiple", map[string]any{"": "value"}, false),
	}
}

func testHasFieldsMixed(t *testing.T) {
	core.RunTestCases(t, hasFieldsMixedTestCases())
}

func TestHasFieldsEdgeCases(t *testing.T) {
	t.Run("WhitespaceKeys", testHasFieldsWhitespaceKeys)
	t.Run("LargeMap", testHasFieldsLargeMap)
	t.Run("UnicodeKeys", testHasFieldsUnicodeKeys)
}

func testHasFieldsWhitespaceKeys(t *testing.T) {
	fields := map[string]any{
		" ":  "space",
		"\t": "tab",
		"\n": "newline",
		"  ": "spaces",
	}

	// Whitespace keys are not empty strings, so should return true
	core.AssertTrue(t, internal.HasFields(fields), "HasFields with whitespace keys")
}

func testHasFieldsLargeMap(t *testing.T) {
	fields := make(map[string]any, 1000)

	// Fill with empty keys
	for i := range 999 {
		fields[""] = i // All will overwrite to same empty key
	}

	// Add one valid key
	fields["valid"] = "value"

	core.AssertTrue(t, internal.HasFields(fields), "HasFields with large map containing one valid key")
}

func testHasFieldsUnicodeKeys(t *testing.T) {
	fields := map[string]any{
		"测试":   "chinese",
		"🔑":    "emoji",
		"café": "accented",
	}

	core.AssertTrue(t, internal.HasFields(fields), "HasFields with unicode keys")
}

// isTerminalLevelTestCase tests IsTerminalLevel on one level.
type isTerminalLevelTestCase struct {
	name  string
	level slog.LogLevel
	want  bool
}

func (tc isTerminalLevelTestCase) Name() string {
	return tc.name
}

func (tc isTerminalLevelTestCase) Test(t *testing.T) {
	core.AssertEqual(t, tc.want, internal.IsTerminalLevel(tc.level), "terminal")
}

func newIsTerminalLevelTestCase(name string, level slog.LogLevel, want bool) isTerminalLevelTestCase {
	return isTerminalLevelTestCase{
		name:  name,
		level: level,
		want:  want,
	}
}

func newTerminalLevelTestCase(name string, level slog.LogLevel) isTerminalLevelTestCase {
	return newIsTerminalLevelTestCase(name, level, true)
}

func newNonTerminalLevelTestCase(name string, level slog.LogLevel) isTerminalLevelTestCase {
	return newIsTerminalLevelTestCase(name, level, false)
}

func TestIsTerminalLevel(t *testing.T) {
	testCases := []core.TestCase{
		newTerminalLevelTestCase("Panic", slog.Panic),
		newTerminalLevelTestCase("Fatal", slog.Fatal),
		newNonTerminalLevelTestCase("UndefinedLevel", slog.UndefinedLevel),
		newNonTerminalLevelTestCase("Error", slog.Error),
		newNonTerminalLevelTestCase("Warn", slog.Warn),
		newNonTerminalLevelTestCase("Info", slog.Info),
		newNonTerminalLevelTestCase("Debug", slog.Debug),
		newNonTerminalLevelTestCase("BelowUndefinedLevel", -1),
		newNonTerminalLevelTestCase("AboveDebug", 7),
	}

	core.RunTestCases(t, testCases)
}
