package testing

import (
	"reflect"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = messageStringTestCase{}

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

	if !core.AssertEqual(t, tc.wantOnlyFirst, len(onlyFirst), "onlyFirst count") {
		logMessages(t, "only in first", onlyFirst)
	}

	if !core.AssertEqual(t, tc.wantOnlySecond, len(onlySecond), "onlySecond count") {
		logMessages(t, "only in second", onlySecond)
	}

	if !core.AssertEqual(t, tc.wantBoth, len(both), "both count") {
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
	core.AssertMustEqual(t, len(messages), len(result), "message count")

	for i, msg := range result {
		core.AssertEqual(t, messages[i].Level, msg.Level, "message %d level", i)
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
	if !core.AssertEqual(t, 2, len(result), "filtered count") {
		for i, msg := range result {
			t.Logf("  [%d] level=%v, message=%q", i, msg.Level, msg.Message)
		}
		return
	}

	// Verify only Info and Error messages remain
	expectedMessages := map[string]bool{
		"info":  false,
		"error": false,
	}

	for _, msg := range result {
		expectedMessages[msg.Message] = true
	}

	core.AssertTrue(t, expectedMessages["info"], "info present")
	core.AssertTrue(t, expectedMessages["error"], "error present")
}

// verifyTransformations verifies that transformations were applied correctly
func verifyTransformations(t *testing.T, original, transformed []Message, opts *AdapterOptions) {
	t.Helper()

	for i, msg := range transformed {
		expected := opts.ExpectedLevel(original[i].Level)
		core.AssertEqual(t, expected, msg.Level, "message %d level", i)
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

	core.AssertEqual(t, 0, len(onlyExpected), "expected only")
	core.AssertEqual(t, 0, len(onlyActual), "actual only")
	core.AssertEqual(t, 2, len(both), "both")
}

// messageStringTestCase represents a test case for Message String method.
type messageStringTestCase struct {
	msg  Message
	want string
	name string
}

func (tc messageStringTestCase) Name() string {
	return tc.name
}

func (tc messageStringTestCase) Test(t *testing.T) {
	t.Helper()
	got := tc.msg.String()
	core.AssertEqual(t, tc.want, got, "string representation")
}

func newMessageStringTestCase(name string, msg Message, want string) messageStringTestCase {
	return messageStringTestCase{
		name: name,
		msg:  msg,
		want: want,
	}
}

func messageStringTestCases() []messageStringTestCase {
	// Note: This test is currently expected to fail because LogLevel
	// doesn't have a String() method, so it prints as a number.
	// This documents the current behaviour.
	return []messageStringTestCase{
		newMessageStringTestCase("basic message",
			Message{Level: slog.Info, Message: "hello"},
			`[5] "hello"`), // Info = 5
		newMessageStringTestCase("message with fields",
			Message{
				Level:   slog.Debug,
				Message: "test",
				Fields:  map[string]any{"b": 2, "a": 1}, // Intentionally unsorted
			},
			`[6] "test" a=1 b=2`), // Debug = 6, fields sorted
		newMessageStringTestCase("message with stack",
			Message{
				Level:   slog.Error,
				Message: "error",
				Stack:   true,
			},
			`[3] "error" [stack]`), // Error = 3
		newMessageStringTestCase("message with everything",
			Message{
				Level:   slog.Warn,
				Message: "warning",
				Fields:  map[string]any{"code": 500, "msg": "internal"},
				Stack:   true,
			},
			`[4] "warning" code=500 msg=internal [stack]`), // Warn = 4
	}
}

func TestIsSame(t *testing.T) {
	t.Run("nil values", func(t *testing.T) {
		testIsSameNilValues(t)
	})

	t.Run("pointer types", func(t *testing.T) {
		testIsSamePointerTypes(t)
	})

	t.Run("interface types", func(t *testing.T) {
		testIsSameInterfaceTypes(t)
	})

	t.Run("value types", func(t *testing.T) {
		testIsSameValueTypes(t)
	})

	t.Run("different types", func(t *testing.T) {
		testIsSameDifferentTypes(t)
	})
}

// testIsSameNilValues tests IsSame with nil values
func testIsSameNilValues(t *testing.T) {
	t.Helper()

	// Both nil
	core.AssertTrue(t, IsSame(nil, nil), "IsSame(nil, nil)")

	// One nil, one non-nil
	value := 42
	core.AssertFalse(t, IsSame(nil, value), "IsSame(nil, non-nil)")
	core.AssertFalse(t, IsSame(value, nil), "IsSame(non-nil, nil)")

	// Nil pointers
	var ptr1, ptr2 *int
	core.AssertTrue(t, IsSame(ptr1, ptr2), "IsSame with nil pointers")

	// Nil interfaces
	var interface1, interface2 any
	core.AssertTrue(t, IsSame(interface1, interface2), "IsSame with nil interfaces")
}

// testIsSamePointerTypes tests IsSame with pointer types
func testIsSamePointerTypes(t *testing.T) {
	t.Helper()

	value1 := 42
	value2 := 42
	ptr1 := &value1
	ptr2 := &value2
	ptr3 := ptr1

	// Same pointer
	if !IsSame(ptr1, ptr3) {
		t.Error("IsSame with same pointer should return true")
	}

	// Different pointers to same value
	if IsSame(ptr1, ptr2) {
		t.Error("IsSame with different pointers should return false")
	}

	// Nil pointer vs non-nil pointer
	var nilPtr *int
	if IsSame(ptr1, nilPtr) {
		t.Error("IsSame with nil and non-nil pointer should return false")
	}
}

// testIsSameInterfaceTypes tests IsSame with interface types
func testIsSameInterfaceTypes(t *testing.T) {
	t.Helper()

	value := 42
	ptr := &value

	value2 := 42 // Different variable with same value

	var interface1, interface2, interface3 any
	interface1 = ptr
	interface2 = ptr     // Same pointer wrapped in interface
	interface3 = &value2 // Different pointer to same value

	// Same underlying pointer
	if !IsSame(interface1, interface2) {
		t.Error("IsSame with same underlying pointer should return true")
	}

	// Different underlying pointers
	if IsSame(interface1, interface3) {
		t.Error("IsSame with different underlying pointers should return false")
	}

	// One nil interface
	var nilInterface any
	if IsSame(interface1, nilInterface) {
		t.Error("IsSame with nil and non-nil interface should return false")
	}
}

// testIsSameValueTypes tests IsSame with value types
func testIsSameValueTypes(t *testing.T) {
	t.Helper()

	// Value types should return false (not same instance)
	value1 := 42
	value2 := 42

	if IsSame(value1, value2) {
		t.Error("IsSame with value types should return false")
	}

	// String values
	str1 := "hello"
	str2 := "hello"

	if IsSame(str1, str2) {
		t.Error("IsSame with string values should return false")
	}

	// Struct values
	type testStruct struct{ x int }
	struct1 := testStruct{x: 1}
	struct2 := testStruct{x: 1}

	if IsSame(struct1, struct2) {
		t.Error("IsSame with struct values should return false")
	}
}

// testIsSameDifferentTypes tests IsSame with different types
func testIsSameDifferentTypes(t *testing.T) {
	t.Helper()

	// Different types should return false
	if IsSame(42, "42") {
		t.Error("IsSame with different types should return false")
	}

	if IsSame(42, 42.0) {
		t.Error("IsSame with int and float should return false")
	}

	// Pointer vs value
	value := 42
	ptr := &value

	if IsSame(value, ptr) {
		t.Error("IsSame with value and pointer should return false")
	}
}

func TestAssertSame(t *testing.T) {
	t.Run("same instances pass", func(t *testing.T) {
		testAssertSamePassing(t)
	})

	t.Run("different instances fail", func(t *testing.T) {
		testAssertSameFailing(t)
	})
}

func TestAssertNotSame(t *testing.T) {
	t.Run("different instances pass", func(t *testing.T) {
		testAssertNotSamePassing(t)
	})

	t.Run("same instances fail", func(t *testing.T) {
		testAssertNotSameFailing(t)
	})
}

// testAssertSamePassing tests AssertSame with cases that should pass
func testAssertSamePassing(t *testing.T) {
	t.Helper()

	// Use a mock test to capture assertions
	mock := &core.MockT{}

	// Test with same pointers
	value := 42
	ptr1 := &value
	ptr2 := ptr1

	if !AssertSame(mock, ptr1, ptr2, "same pointer test") {
		t.Error("AssertSame should return true for same pointers")
	}

	// Mock should not have any failures
	if mock.Failed() {
		t.Error("AssertSame should not fail for same instances")
	}

	// Test with nil values
	mock = &core.MockT{}
	if !AssertSame(mock, nil, nil, "nil test") {
		t.Error("AssertSame should return true for both nil")
	}

	if mock.Failed() {
		t.Error("AssertSame should not fail for both nil")
	}
}

// testAssertSameFailing tests AssertSame with cases that should fail
func testAssertSameFailing(t *testing.T) {
	t.Helper()

	// Use a mock test to capture assertions
	mock := &core.MockT{}

	// Test with different pointers
	value1 := 42
	value2 := 42
	ptr1 := &value1
	ptr2 := &value2

	if AssertSame(mock, ptr1, ptr2, "different pointer test") {
		t.Error("AssertSame should return false for different pointers")
	}

	// Mock should have a failure (from core.AssertEqual fallback)
	if !mock.Failed() {
		t.Error("AssertSame should fail and call core.AssertEqual for different instances")
	}

	// Test with value types
	mock = &core.MockT{}
	val1 := 42
	val2 := 42

	if AssertSame(mock, val1, val2, "value test") {
		t.Error("AssertSame should return false for value types")
	}

	// Mock should have a failure
	if !mock.Failed() {
		t.Error("AssertSame should fail for value types")
	}
}

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = assertNoFieldTestCase{}

type assertNoFieldTestCase struct {
	msg        Message
	key        string
	name       string
	expectPass bool
}

func (tc assertNoFieldTestCase) Name() string {
	return tc.name
}

func (tc assertNoFieldTestCase) Test(t *testing.T) {
	t.Helper()

	// Use MockT to test assertion function without failing the build
	mock := &core.MockT{}
	result := AssertNoField(mock, tc.msg, tc.key)

	core.AssertEqual(t, tc.expectPass, result, "result")
}

func newAssertNoFieldTestCase(name, key string, msg Message, expectPass bool) assertNoFieldTestCase {
	return assertNoFieldTestCase{
		name:       name,
		key:        key,
		msg:        msg,
		expectPass: expectPass,
	}
}

func assertNoFieldTestCases() []assertNoFieldTestCase {
	return []assertNoFieldTestCase{
		newAssertNoFieldTestCase("field does not exist", "non-existent",
			Message{Level: slog.Info, Message: "test", Fields: map[string]any{"existing": "value"}},
			true),
		newAssertNoFieldTestCase("field exists", "existing",
			Message{Level: slog.Info, Message: "test", Fields: map[string]any{"existing": "value"}},
			false),
		newAssertNoFieldTestCase("empty fields map", "someKey",
			Message{Level: slog.Info, Message: "test", Fields: map[string]any{}},
			true),
		newAssertNoFieldTestCase("nil fields map", "someKey",
			Message{Level: slog.Info, Message: "test", Fields: nil},
			true),
		newAssertNoFieldTestCase("field with nil value", "nilField",
			Message{Level: slog.Info, Message: "test", Fields: map[string]any{"nilField": nil}},
			false),
	}
}

func TestAssertNoField(t *testing.T) {
	core.RunTestCases(t, assertNoFieldTestCases())
}

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = runWithLoggerFactoryTestCase{}

type runWithLoggerFactoryTestCase struct {
	factory       func() slog.Logger
	name          string
	expectedCalls int
	expectNil     bool
}

func (tc runWithLoggerFactoryTestCase) Name() string {
	return tc.name
}

func (tc runWithLoggerFactoryTestCase) Test(t *testing.T) {
	t.Helper()

	callCount := 0
	actualFactory := func() slog.Logger {
		callCount++
		return tc.factory()
	}

	RunWithLoggerFactory(t, "subtest", actualFactory, func(subT core.T, logger slog.Logger) {
		if tc.expectNil {
			core.AssertNil(subT, logger, "logger")
		} else {
			core.AssertNotNil(subT, logger, "logger")
		}
	})

	core.AssertEqual(t, tc.expectedCalls, callCount, "factory call count")
}

func newRunWithLoggerFactoryTestCase(name string, expectedCalls int,
	factory func() slog.Logger, expectNil bool) runWithLoggerFactoryTestCase {
	return runWithLoggerFactoryTestCase{
		name:          name,
		expectedCalls: expectedCalls,
		factory:       factory,
		expectNil:     expectNil,
	}
}

func runWithLoggerFactoryTestCases() []runWithLoggerFactoryTestCase {
	return []runWithLoggerFactoryTestCase{
		newRunWithLoggerFactoryTestCase("factory called once", 1,
			func() slog.Logger { return NewLogger() },
			false),
		newRunWithLoggerFactoryTestCase("test receives logger", 1,
			func() slog.Logger { return NewLogger() },
			false),
		newRunWithLoggerFactoryTestCase("nil logger factory", 1,
			func() slog.Logger { return nil },
			true),
	}
}

func TestRunWithLoggerFactory(t *testing.T) {
	core.RunTestCases(t, runWithLoggerFactoryTestCases())
}

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = isSameInterfaceTestCase{}

type isSameInterfaceTestCase struct {
	valueA   any
	valueB   any
	expected bool
	name     string
}

func (tc isSameInterfaceTestCase) Name() string {
	return tc.name
}

func (tc isSameInterfaceTestCase) Test(t *testing.T) {
	t.Helper()

	va := reflect.ValueOf(&tc.valueA).Elem()
	vb := reflect.ValueOf(&tc.valueB).Elem()

	result := isSameInterface(va, vb)
	core.AssertEqual(t, tc.expected, result, "interface sameness")
}

func newIsSameInterfaceTestCase(name string, valueA, valueB any, expected bool) isSameInterfaceTestCase {
	return isSameInterfaceTestCase{
		name:     name,
		valueA:   valueA,
		valueB:   valueB,
		expected: expected,
	}
}

func isSameInterfaceTestCases() []isSameInterfaceTestCase {
	value := 42
	ptr := &value
	value1 := 42
	value2 := 42
	ptr1 := &value1
	ptr2 := &value2

	// testAssertNotSamePassing tests AssertNotSame with cases that should pass
	return []isSameInterfaceTestCase{
		newIsSameInterfaceTestCase("both nil interfaces", nil, nil, true),
		newIsSameInterfaceTestCase("nil and non-nil interface", nil, 42, false),
		newIsSameInterfaceTestCase("non-nil and nil interface", 42, nil, false),
		newIsSameInterfaceTestCase("same underlying pointer", ptr, ptr, true),
		newIsSameInterfaceTestCase("different underlying pointers", ptr1, ptr2, false),
		newIsSameInterfaceTestCase("same value types", value1, value2, false),
		newIsSameInterfaceTestCase("different underlying types", 42, "42", false),
	}
}

func TestIsSameInterface(t *testing.T) {
	core.RunTestCases(t, isSameInterfaceTestCases())
}

// testAssertNotSamePassing tests AssertNotSame with cases that should pass
func testAssertNotSamePassing(t *testing.T) {
	t.Helper()

	// Use a mock test to capture assertions
	mock := &core.MockT{}

	// Test with different pointers
	value1 := 42
	value2 := 42
	ptr1 := &value1
	ptr2 := &value2

	if !AssertNotSame(mock, ptr1, ptr2, "different pointer test") {
		t.Error("AssertNotSame should return true for different pointers")
	}

	// Mock should not have any failures
	if mock.Failed() {
		t.Error("AssertNotSame should not fail for different instances")
	}

	// Test with value types
	mock = &core.MockT{}
	val1 := 42
	val2 := 42

	if !AssertNotSame(mock, val1, val2, "value test") {
		t.Error("AssertNotSame should return true for value types")
	}

	if mock.Failed() {
		t.Error("AssertNotSame should not fail for value types")
	}
}

// testAssertNotSameFailing tests AssertNotSame with cases that should fail
func testAssertNotSameFailing(t *testing.T) {
	t.Helper()

	// Use a mock test to capture assertions
	mock := &core.MockT{}

	// Test with same pointers
	value := 42
	ptr1 := &value
	ptr2 := ptr1

	if AssertNotSame(mock, ptr1, ptr2, "same pointer test") {
		t.Error("AssertNotSame should return false for same pointers")
	}

	// Mock should have a failure
	if !mock.Failed() {
		t.Error("AssertNotSame should fail for same instances")
	}

	// Test with nil values
	mock = &core.MockT{}
	if AssertNotSame(mock, nil, nil, "nil test") {
		t.Error("AssertNotSame should return false for both nil")
	}

	if !mock.Failed() {
		t.Error("AssertNotSame should fail for both nil")
	}
}

func TestMessageString(t *testing.T) {
	core.RunTestCases(t, messageStringTestCases())
}
