package internal

import (
	"reflect"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
)

func TestLogletLevel(t *testing.T) {
	t.Run("DefaultLevel", testLogletLevelDefault)
	t.Run("SetLevel", testLogletLevelSet)
	t.Run("WithLevel", testLogletLevelWith)
}

func testLogletLevelDefault(t *testing.T) {
	var loglet Loglet

	// Default level should be UndefinedLevel (zero value)
	if loglet.Level() != slog.UndefinedLevel {
		t.Errorf("Default level should be UndefinedLevel, got %v", loglet.Level())
	}
}

func testLogletLevelSet(t *testing.T) {
	var loglet Loglet

	// Test setting different levels
	levels := []slog.LogLevel{
		slog.Panic,
		slog.Fatal,
		slog.Error,
		slog.Warn,
		slog.Info,
		slog.Debug,
	}

	for _, level := range levels {
		newLoglet := loglet.WithLevel(level)
		if newLoglet.Level() != level {
			t.Errorf("WithLevel(%v) should set level to %v, got %v", level, level, newLoglet.Level())
		}
	}
}

func testLogletLevelWith(t *testing.T) {
	var loglet Loglet

	// Test that WithLevel with same level returns same loglet
	loglet1 := loglet.WithLevel(slog.Info)
	loglet2 := loglet1.WithLevel(slog.Info)

	// Check they have same level and field count (can't compare structs directly)
	if loglet1.Level() != loglet2.Level() || loglet1.FieldsCount() != loglet2.FieldsCount() {
		t.Error("WithLevel with same level should return equivalent loglet")
	}

	// Test that WithLevel with different level returns new loglet
	loglet3 := loglet1.WithLevel(slog.Error)
	if loglet3.Level() == loglet1.Level() {
		t.Error("WithLevel with different level should change level")
	}
	if loglet3.Level() != slog.Error {
		t.Errorf("New loglet should have Error level, got %v", loglet3.Level())
	}
}

func TestLogletStack(t *testing.T) {
	t.Run("DefaultStack", testLogletStackDefault)
	t.Run("WithStack", testLogletStackWith)
	t.Run("StackChaining", testLogletStackChaining)
}

func testLogletStackDefault(t *testing.T) {
	var loglet Loglet

	// Default stack should be nil/empty
	stack := loglet.CallStack()
	if stack != nil {
		t.Error("Default stack should be nil")
	}
}

func testLogletStackWith(t *testing.T) {
	var loglet Loglet

	// WithStack should create a new loglet with stack
	newLoglet := loglet.WithStack(1)

	// Original should have nil stack, new should have non-nil stack
	if loglet.CallStack() != nil {
		t.Error("Original loglet should have nil stack")
	}

	// Stack should not be nil
	if newLoglet.CallStack() == nil {
		t.Error("WithStack should create non-nil stack")
	}
}

func testLogletStackChaining(t *testing.T) {
	var loglet Loglet

	// Chain multiple WithStack calls
	loglet1 := loglet.WithStack(1)
	loglet2 := loglet1.WithStack(2)

	// Both should have non-nil stacks
	if loglet1.CallStack() == nil {
		t.Error("First WithStack should create non-nil stack")
	}
	if loglet2.CallStack() == nil {
		t.Error("Second WithStack should create non-nil stack")
	}
}

func TestLogletFields(t *testing.T) {
	t.Run("WithField", testLogletWithField)
	t.Run("WithFields", testLogletWithFields)
	t.Run("FieldsCount", testLogletFieldsCount)
	t.Run("EmptyKey", testLogletEmptyKey)
	t.Run("FieldsMap", testLogletFieldsMap)
}

func testLogletWithField(t *testing.T) {
	var loglet Loglet

	// Add single field
	newLoglet := loglet.WithField("key1", "value1")

	// Count should be 1
	if newLoglet.FieldsCount() != 1 {
		t.Errorf("FieldsCount should be 1, got %d", newLoglet.FieldsCount())
	}

	// Add another field
	newLoglet2 := newLoglet.WithField("key2", "value2")

	// Count should be 2
	if newLoglet2.FieldsCount() != 2 {
		t.Errorf("FieldsCount should be 2, got %d", newLoglet2.FieldsCount())
	}

	// Original should still be 1
	if newLoglet.FieldsCount() != 1 {
		t.Errorf("Original FieldsCount should still be 1, got %d", newLoglet.FieldsCount())
	}
}

func testLogletWithFields(t *testing.T) {
	var loglet Loglet

	fields := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	newLoglet := loglet.WithFields(fields)

	// Count should match map size
	if newLoglet.FieldsCount() != len(fields) {
		t.Errorf("FieldsCount should be %d, got %d", len(fields), newLoglet.FieldsCount())
	}
}

func testLogletFieldsCount(t *testing.T) {
	var loglet Loglet

	// Start with 0
	if loglet.FieldsCount() != 0 {
		t.Errorf("Initial FieldsCount should be 0, got %d", loglet.FieldsCount())
	}

	// Add fields progressively using proper chaining (not reassignment)
	loglet1 := loglet.WithField("a", 1)
	if loglet1.FieldsCount() != 1 {
		t.Errorf("FieldsCount should be 1, got %d", loglet1.FieldsCount())
	}

	loglet2 := loglet1.WithFields(map[string]any{"b": 2, "c": 3})
	if loglet2.FieldsCount() != 3 {
		t.Errorf("FieldsCount should be 3, got %d", loglet2.FieldsCount())
	}

	loglet3 := loglet2.WithField("d", 4)
	if loglet3.FieldsCount() != 4 {
		t.Errorf("FieldsCount should be 4, got %d", loglet3.FieldsCount())
	}
}

func testLogletEmptyKey(t *testing.T) {
	var loglet Loglet

	// Empty key should not be added
	newLoglet := loglet.WithField("", "value")

	// Should have same properties (can't compare structs directly)
	if newLoglet.FieldsCount() != loglet.FieldsCount() {
		t.Error("WithField with empty key should not change field count")
	}

	if newLoglet.FieldsCount() != 0 {
		t.Errorf("FieldsCount should remain 0, got %d", newLoglet.FieldsCount())
	}
}

func TestLogletFieldsIterator(t *testing.T) {
	t.Run("EmptyIterator", testLogletIteratorEmpty)
	t.Run("SingleField", testLogletIteratorSingle)
	t.Run("MultipleFields", testLogletIteratorMultiple)
	t.Run("ChainedFields", testLogletIteratorChained)
}

func testLogletIteratorEmpty(t *testing.T) {
	var loglet Loglet

	iter := loglet.Fields()

	// Should not have any fields
	if iter.Next() {
		t.Error("Empty loglet iterator should return false for Next()")
	}
}

func testLogletIteratorSingle(t *testing.T) {
	var loglet Loglet
	loglet1 := loglet.WithField("key", "value")

	iter := loglet1.Fields()

	// Should have one field
	if !iter.Next() {
		t.Fatal("Iterator should have one field")
	}

	key := iter.Key()
	value := iter.Value()

	if key != "key" {
		t.Errorf("Expected key 'key', got %q", key)
	}
	if value != "value" {
		t.Errorf("Expected value 'value', got %v", value)
	}

	// Should not have more fields
	if iter.Next() {
		t.Error("Iterator should not have more fields")
	}
}

func testLogletIteratorMultiple(t *testing.T) {
	var loglet Loglet

	fields := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	loglet1 := loglet.WithFields(fields)
	iter := loglet1.Fields()

	found := make(map[string]any)

	for iter.Next() {
		key, value := iter.Field()
		found[key] = value
	}

	// Should have found all fields
	if len(found) != len(fields) {
		t.Errorf("Expected %d fields, found %d", len(fields), len(found))
	}

	for k, v := range fields {
		if found[k] != v {
			t.Errorf("Field %q: expected %v, got %v", k, v, found[k])
		}
	}
}

func testLogletIteratorChained(t *testing.T) {
	var loglet Loglet

	// Create chain of loglets using proper chaining
	loglet1 := loglet.WithField("root", "rootValue")
	loglet2 := loglet1.WithField("child", "childValue")
	loglet3 := loglet2.WithFields(map[string]any{
		"grand": "grandValue",
		"leaf":  "leafValue",
	})

	iter := loglet3.Fields()
	found := make(map[string]any)

	for iter.Next() {
		key, value := iter.Field()
		found[key] = value
	}

	expected := map[string]any{
		"root":  "rootValue",
		"child": "childValue",
		"grand": "grandValue",
		"leaf":  "leafValue",
	}

	if len(found) != len(expected) {
		t.Errorf("Expected %d fields, found %d", len(expected), len(found))
	}

	for k, v := range expected {
		if found[k] != v {
			t.Errorf("Field %q: expected %v, got %v", k, v, found[k])
		}
	}
}

func TestLogletIntegration(t *testing.T) {
	t.Run("ComplexChaining", testLogletComplexChaining)
	t.Run("LevelAndFields", testLogletLevelAndFields)
	t.Run("StackAndFields", testLogletStackAndFields)
}

func testLogletComplexChaining(t *testing.T) {
	var loglet Loglet

	// Complex chain with level, stack, and fields using proper chaining
	loglet1 := loglet.WithLevel(slog.Info)
	loglet2 := loglet1.WithStack(1)
	loglet3 := loglet2.WithField("service", "api")
	loglet4 := loglet3.WithFields(map[string]any{
		"version": "1.0",
		"env":     "test",
	})
	loglet5 := loglet4.WithField("user", "john")

	// Verify level preserved
	if loglet5.Level() != slog.Info {
		t.Errorf("Level should be Info, got %v", loglet5.Level())
	}

	// Verify stack preserved
	if loglet5.CallStack() == nil {
		t.Error("Stack should be preserved")
	}

	// Verify field count
	if loglet5.FieldsCount() != 4 {
		t.Errorf("Should have 4 fields, got %d", loglet5.FieldsCount())
	}

	// Verify all fields present
	iter := loglet5.Fields()
	found := make(map[string]any)

	for iter.Next() {
		key, value := iter.Field()
		found[key] = value
	}

	expected := map[string]any{
		"service": "api",
		"version": "1.0",
		"env":     "test",
		"user":    "john",
	}

	for k, v := range expected {
		if found[k] != v {
			t.Errorf("Field %q: expected %v, got %v", k, v, found[k])
		}
	}
}

func testLogletLevelAndFields(t *testing.T) {
	var loglet Loglet

	// Add fields, then change level using proper chaining
	loglet1 := loglet.WithField("before", "value")
	loglet2 := loglet1.WithLevel(slog.Error)
	loglet3 := loglet2.WithField("after", "value")

	// Both fields should be present
	if loglet3.FieldsCount() != 2 {
		t.Errorf("Should have 2 fields, got %d", loglet3.FieldsCount())
	}

	// Level should be Error
	if loglet3.Level() != slog.Error {
		t.Errorf("Level should be Error, got %v", loglet3.Level())
	}
}

func testLogletStackAndFields(t *testing.T) {
	var loglet Loglet

	// Add fields, then stack using proper chaining
	loglet1 := loglet.WithField("before", "value")
	loglet2 := loglet1.WithStack(1)
	loglet3 := loglet2.WithField("after", "value")

	// Both fields should be present
	if loglet3.FieldsCount() != 2 {
		t.Errorf("Should have 2 fields, got %d", loglet3.FieldsCount())
	}

	// Stack should be present
	if loglet3.CallStack() == nil {
		t.Error("Stack should be present")
	}

	// Should implement CallStacker interface
	var _ core.CallStacker = &loglet3
}

// TestLogletIsZero tests the IsZero method for all branches
func TestLogletIsZero(t *testing.T) {
	t.Run("NilLoglet", testLogletIsZeroNil)
	t.Run("ZeroLoglet", testLogletIsZeroTrue)
	t.Run("NonZeroLoglet", testLogletIsZeroFalse)
}

func testLogletIsZeroNil(t *testing.T) {
	var loglet *Loglet
	if !loglet.IsZero() {
		t.Error("nil loglet should be zero")
	}
}

func testLogletIsZeroTrue(t *testing.T) {
	var loglet Loglet
	if !loglet.IsZero() {
		t.Error("empty loglet should be zero")
	}
}

func testLogletIsZeroFalse(t *testing.T) {
	var loglet Loglet

	// Test with field
	loglet1 := loglet.WithField("key", "value")
	if loglet1.IsZero() {
		t.Error("loglet with field should not be zero")
	}

	// Test with level
	loglet2 := loglet.WithLevel(slog.Info)
	if loglet2.IsZero() {
		t.Error("loglet with level should not be zero")
	}

	// Test with stack
	loglet3 := loglet.WithStack(1)
	if loglet3.IsZero() {
		t.Error("loglet with stack should not be zero")
	}
}

// TestLogletGetParent tests the GetParent method for all branches
func TestLogletGetParent(t *testing.T) {
	t.Run("NilLoglet", testLogletGetParentNil)
	t.Run("SelfReference", testLogletGetParentSelf)
	t.Run("NormalParent", testLogletGetParentNormal)
}

func testLogletGetParentNil(t *testing.T) {
	// Test GetParent method indirectly
	// We'll test indirectly through FieldsCount which uses GetParent
	var nilLoglet *Loglet
	count := nilLoglet.FieldsCount()
	if count != 0 {
		t.Errorf("nil loglet should have 0 fields, got %d", count)
	}
}

func testLogletGetParentSelf(t *testing.T) {
	t.Run("CircularReference", testCircularReference)
	t.Run("ReassignmentPattern", testReassignmentPattern)
	t.Run("ProperChaining", testProperChaining)
	t.Run("DirectSelfReference", testDirectSelfReference)
}

func testCircularReference(t *testing.T) {
	// Test the problematic reassignment pattern
	var loglet Loglet
	loglet = loglet.WithField("key1", "value1")
	loglet = loglet.WithField("key2", "value2")

	// Should work without infinite loop due to GetParent protection
	count := loglet.FieldsCount()
	if count != 1 {
		t.Errorf("circular reference should result in 1 field, got %d", count)
	}

	// Test field iteration
	iter := loglet.Fields()
	fieldCount := 0
	for iter.Next() {
		fieldCount++
		_ = iter.Key()
		_ = iter.Value()
	}

	if fieldCount != 1 {
		t.Errorf("iterator should find 1 field, got %d", fieldCount)
	}
}

func testReassignmentPattern(t *testing.T) {
	// Test reassignment with non-zero base
	var base Loglet
	base = base.WithLevel(slog.Info)
	base = base.WithField("service", "test")
	base = base.WithField("version", "1.0")

	if base.FieldsCount() != 1 {
		t.Errorf("reassignment should have 1 field, got %d", base.FieldsCount())
	}
}

func testProperChaining(t *testing.T) {
	// Test proper chaining creates field chains
	var chain Loglet
	chain1 := chain.WithField("key1", "value1")
	chain2 := chain1.WithField("key2", "value2")
	chain3 := chain2.WithField("key3", "value3")

	if chain3.FieldsCount() != 3 {
		t.Errorf("proper chaining should have 3 fields, got %d", chain3.FieldsCount())
	}

	// Test field iteration
	iter := chain3.Fields()
	fieldCount := 0
	for iter.Next() {
		fieldCount++
	}

	if fieldCount != 3 {
		t.Errorf("proper chain iterator should find 3 fields, got %d", fieldCount)
	}
}

func testDirectSelfReference(t *testing.T) {
	// Manually create self-reference
	selfRef := Loglet{parent: nil}
	selfRef.parent = &selfRef

	// Should be caught by GetParent() protection
	if selfRef.GetParent() != nil {
		t.Error("GetParent should return nil for self-reference")
	}

	// Should not cause infinite loop in FieldsCount
	if selfRef.FieldsCount() != 0 {
		t.Errorf("self-referencing loglet should have 0 fields, got %d", selfRef.FieldsCount())
	}
}

func testLogletGetParentNormal(t *testing.T) {
	var loglet Loglet
	loglet1 := loglet.WithField("key", "value")

	// Test that parent relationship works correctly
	if loglet1.FieldsCount() != 1 {
		t.Errorf("should have 1 field, got %d", loglet1.FieldsCount())
	}
}

// TestFilterFields tests the filterFields function through WithFields
func TestFilterFields(t *testing.T) {
	t.Run("EmptyMap", testFilterFieldsEmpty)
	t.Run("EmptyKeys", testFilterFieldsEmptyKeys)
	t.Run("MixedKeys", testFilterFieldsMixed)
}

func testFilterFieldsEmpty(t *testing.T) {
	var loglet Loglet

	// Test with empty map
	emptyFields := map[string]any{}
	newLoglet := loglet.WithFields(emptyFields)

	if newLoglet.FieldsCount() != 0 {
		t.Errorf("empty fields should result in 0 fields, got %d", newLoglet.FieldsCount())
	}
}

func testFilterFieldsEmptyKeys(t *testing.T) {
	var loglet Loglet

	// Test with map containing only empty keys
	fieldsWithEmptyKeys := map[string]any{
		"":  "value1",
		" ": "value2", // Non-empty key
	}

	newLoglet := loglet.WithFields(fieldsWithEmptyKeys)

	// Should only include the non-empty key
	if newLoglet.FieldsCount() != 1 {
		t.Errorf("should filter out empty keys, got %d fields", newLoglet.FieldsCount())
	}
}

func testFilterFieldsMixed(t *testing.T) {
	var loglet Loglet

	// Test with mix of valid and invalid keys
	mixedFields := map[string]any{
		"":      "filtered_out",
		"valid": "kept",
		"also":  "kept",
	}

	newLoglet := loglet.WithFields(mixedFields)

	// Should only include valid keys
	if newLoglet.FieldsCount() != 2 {
		t.Errorf("should have 2 valid fields, got %d", newLoglet.FieldsCount())
	}
}

// TestWithFieldsEdgeCases tests additional edge cases for WithFields
func TestWithFieldsEdgeCases(t *testing.T) {
	t.Run("ZeroLogletParent", testWithFieldsZeroLogletParent)
	t.Run("NonZeroLogletParent", testWithFieldsNonZeroLogletParent)
}

func testWithFieldsZeroLogletParent(t *testing.T) {
	var loglet Loglet // Zero loglet

	fields := map[string]any{"key": "value"}
	newLoglet := loglet.WithFields(fields)

	// Should not set parent for zero loglet
	if newLoglet.FieldsCount() != 1 {
		t.Errorf("should have 1 field, got %d", newLoglet.FieldsCount())
	}
}

func testWithFieldsNonZeroLogletParent(t *testing.T) {
	var loglet Loglet
	loglet1 := loglet.WithLevel(slog.Info) // Make it non-zero

	fields := map[string]any{"key": "value"}
	newLoglet := loglet1.WithFields(fields)

	// Should set parent for non-zero loglet
	if newLoglet.FieldsCount() != 1 {
		t.Errorf("should have 1 field, got %d", newLoglet.FieldsCount())
	}
}

func testLogletFieldsMap(t *testing.T) {
	t.Run("NilLoglet", testFieldsMapNil)
	t.Run("EmptyLoglet", testFieldsMapEmpty)
	t.Run("SingleField", testFieldsMapSingle)
	t.Run("MultipleFields", testFieldsMapMultiple)
	t.Run("ChainedFields", testFieldsMapChained)
	t.Run("FieldOverride", testFieldsMapOverride)
	t.Run("Caching", testFieldsMapCaching)
	t.Run("ImmutableCache", testFieldsMapImmutableCache)
	t.Run("SharedMapWarning", testFieldsMapSharedWarning)
}

func testFieldsMapNil(t *testing.T) {
	var loglet *Loglet
	fields := loglet.FieldsMap()
	if fields != nil {
		t.Error("nil loglet should return nil FieldsMap")
	}
}

func testFieldsMapEmpty(t *testing.T) {
	var loglet Loglet
	fields := loglet.FieldsMap()
	if fields == nil || len(fields) != 0 {
		t.Errorf("empty loglet should return empty map, got %v", fields)
	}
}

func testFieldsMapSingle(t *testing.T) {
	var loglet Loglet
	loglet1 := loglet.WithField("key", "value")

	fields := loglet1.FieldsMap()
	if fields == nil {
		t.Fatal("single field loglet should return non-nil FieldsMap")
	}

	if len(fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(fields))
	}

	if fields["key"] != "value" {
		t.Errorf("expected value 'value', got %v", fields["key"])
	}
}

func testFieldsMapMultiple(t *testing.T) {
	var loglet Loglet
	inputFields := map[string]any{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	loglet1 := loglet.WithFields(inputFields)
	fields := loglet1.FieldsMap()

	if fields == nil {
		t.Fatal("multi-field loglet should return non-nil FieldsMap")
	}

	if len(fields) != len(inputFields) {
		t.Errorf("expected %d fields, got %d", len(inputFields), len(fields))
	}

	for k, v := range inputFields {
		if fields[k] != v {
			t.Errorf("field %q: expected %v, got %v", k, v, fields[k])
		}
	}
}

func testFieldsMapChained(t *testing.T) {
	var loglet Loglet

	// Create chain of loglets using proper chaining
	loglet1 := loglet.WithField("root", "rootValue")
	loglet2 := loglet1.WithField("child", "childValue")
	loglet3 := loglet2.WithFields(map[string]any{
		"grand": "grandValue",
		"leaf":  "leafValue",
	})

	fields := loglet3.FieldsMap()
	if fields == nil {
		t.Fatal("chained loglet should return non-nil FieldsMap")
	}

	expected := map[string]any{
		"root":  "rootValue",
		"child": "childValue",
		"grand": "grandValue",
		"leaf":  "leafValue",
	}

	if len(fields) != len(expected) {
		t.Errorf("expected %d fields, got %d", len(expected), len(fields))
	}

	for k, v := range expected {
		if fields[k] != v {
			t.Errorf("field %q: expected %v, got %v", k, v, fields[k])
		}
	}
}

func testFieldsMapOverride(t *testing.T) {
	var loglet Loglet

	// Create chain where child overrides parent field
	loglet1 := loglet.WithField("key", "parentValue")
	loglet2 := loglet1.WithField("key", "childValue") // Same key

	fields := loglet2.FieldsMap()
	if fields == nil {
		t.Fatal("override loglet should return non-nil FieldsMap")
	}

	// Child value should override parent value
	if fields["key"] != "childValue" {
		t.Errorf("expected child value 'childValue', got %v", fields["key"])
	}

	if len(fields) != 1 {
		t.Errorf("expected 1 field after override, got %d", len(fields))
	}
}

func testFieldsMapCaching(t *testing.T) {
	var loglet Loglet
	loglet1 := loglet.WithField("key", "value")

	// First call should build the cache
	fields1 := loglet1.FieldsMap()
	if fields1 == nil {
		t.Fatal("first call should return non-nil FieldsMap")
	}

	// Second call should return same cached map
	fields2 := loglet1.FieldsMap()
	if fields2 == nil {
		t.Fatal("second call should return non-nil FieldsMap")
	}

	// Should be the same map instance (cached)
	// Compare map pointers using reflection
	ptr1 := reflect.ValueOf(fields1).Pointer()
	ptr2 := reflect.ValueOf(fields2).Pointer()
	if ptr1 != ptr2 {
		t.Error("cached calls should return same map instance")
	}
}

func testFieldsMapImmutableCache(t *testing.T) {
	var loglet Loglet
	loglet1 := loglet.WithField("original", "value")

	// Cache the fields map
	fields1 := loglet1.FieldsMap()
	if fields1 == nil {
		t.Fatal("should return non-nil FieldsMap")
	}

	// Create new loglet with additional field - should not have cache
	loglet2 := loglet1.WithField("additional", "value2")
	fields2 := loglet2.FieldsMap()

	if fields2 == nil {
		t.Fatal("new loglet should return non-nil FieldsMap")
	}

	// New loglet should have both fields
	if len(fields2) != 2 {
		t.Errorf("new loglet should have 2 fields, got %d", len(fields2))
	}

	// Original loglet should still have cached single field
	if len(fields1) != 1 {
		t.Errorf("original loglet should still have 1 field, got %d", len(fields1))
	}

	// Different map instances
	ptr1 := reflect.ValueOf(fields1).Pointer()
	ptr2 := reflect.ValueOf(fields2).Pointer()
	if ptr1 == ptr2 {
		t.Error("different loglets should have different map instances")
	}
}

func testFieldsMapSharedWarning(t *testing.T) {
	// Test that demonstrates why the returned map should not be modified
	var loglet Loglet
	loglet1 := loglet.WithField("key", "original")

	// Get the cached map
	fields1 := loglet1.FieldsMap()
	if fields1 == nil {
		t.Fatal("should return non-nil FieldsMap")
	}

	if fields1["key"] != "original" {
		t.Errorf("expected 'original', got %v", fields1["key"])
	}

	// Get the map again (should be same cached instance)
	fields2 := loglet1.FieldsMap()
	ptr1 := reflect.ValueOf(fields1).Pointer()
	ptr2 := reflect.ValueOf(fields2).Pointer()
	if ptr1 != ptr2 {
		t.Error("should return same cached map instance")
	}

	// This demonstrates why modifying the map would be dangerous:
	// If we modified fields1["key"] = "modified", it would affect
	// all future calls to FieldsMap() on this loglet instance,
	// breaking the immutable contract.
	//
	// Instead, handlers that need to modify fields should build their own copy:
	// modifiableFields := make(map[string]any)
	// iter := loglet.Fields()
	// for iter.Next() {
	//     k, v := iter.Field()
	//     modifiableFields[k] = transformValue(v)  // safe to modify
	// }
}
