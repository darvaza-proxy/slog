package internal

import (
	"fmt"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
)

const testValue = "value"

func TestLogletLevel(t *testing.T) {
	t.Run("DefaultLevel", testLogletLevelDefault)
	t.Run("SetLevel", testLogletLevelSet)
	t.Run("WithLevel", testLogletLevelWith)
}

func testLogletLevelDefault(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// Default level should be UndefinedLevel (zero value)
	core.AssertEqual(t, slog.UndefinedLevel, loglet.Level(), "default level")
}

func testLogletLevelSet(t *testing.T) {
	t.Helper()
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
		core.AssertEqual(t, level, newLoglet.Level(), "level for %v", level)
	}
}

func testLogletLevelWith(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// Test that WithLevel with same level returns same loglet
	loglet1 := loglet.WithLevel(slog.Info)
	loglet2 := loglet1.WithLevel(slog.Info)

	// Check they have same level and field count (can't compare structs directly)
	core.AssertEqual(t, loglet1.Level(), loglet2.Level(), "same level after WithLevel")
	core.AssertEqual(t, loglet1.FieldsCount(), loglet2.FieldsCount(), "same field count after WithLevel")

	// Test that WithLevel with different level returns new loglet
	loglet3 := loglet1.WithLevel(slog.Error)
	core.AssertNotEqual(t, loglet1.Level(), loglet3.Level(), "different level after WithLevel")
	core.AssertEqual(t, slog.Error, loglet3.Level(), "new level")
}

func TestLogletStack(t *testing.T) {
	t.Run("DefaultStack", testLogletStackDefault)
	t.Run("WithStack", testLogletStackWith)
	t.Run("StackChaining", testLogletStackChaining)
}

func testLogletStackDefault(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// Default stack should be nil/empty
	stack := loglet.CallStack()
	core.AssertNil(t, stack, "default stack")
}

func testLogletStackWith(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// WithStack should create a new loglet with stack
	newLoglet := loglet.WithStack(1)

	// Original should have nil stack, new should have non-nil stack
	core.AssertNil(t, loglet.CallStack(), "original loglet stack")

	// Stack should not be nil
	core.AssertNotNil(t, newLoglet.CallStack(), "new loglet stack")
}

func testLogletStackChaining(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// Chain multiple WithStack calls
	loglet1 := loglet.WithStack(1)
	loglet2 := loglet1.WithStack(2)

	// Both should have non-nil stacks
	core.AssertNotNil(t, loglet1.CallStack(), "first WithStack result")
	core.AssertNotNil(t, loglet2.CallStack(), "second WithStack result")
}

func TestLogletFields(t *testing.T) {
	t.Run("WithField", testLogletWithField)
	t.Run("WithFields", testLogletWithFields)
	t.Run("FieldsCount", testLogletFieldsCount)
	t.Run("EmptyKey", testLogletEmptyKey)
	t.Run("FieldsMap", testLogletFieldsMap)
}

func testLogletWithField(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// Add single field
	newLoglet := loglet.WithField("key1", "value1")

	// Count should be 1
	core.AssertEqual(t, 1, newLoglet.FieldsCount(), "field count after single field")

	// Add another field
	newLoglet2 := newLoglet.WithField("key2", "value2")

	// Count should be 2
	core.AssertEqual(t, 2, newLoglet2.FieldsCount(), "field count after second field")

	// Original should still be 1
	core.AssertEqual(t, 1, newLoglet.FieldsCount(), "original field count unchanged")
}

func testLogletWithFields(t *testing.T) {
	t.Helper()
	var loglet Loglet

	fields := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	newLoglet := loglet.WithFields(fields)

	// Count should match map size
	core.AssertEqual(t, len(fields), newLoglet.FieldsCount(), "field count after WithFields")
}

func testLogletFieldsCount(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// Start with 0
	core.AssertEqual(t, 0, loglet.FieldsCount(), "initial field count")

	// Add fields progressively using proper chaining (not reassignment)
	loglet1 := loglet.WithField("a", 1)
	core.AssertEqual(t, 1, loglet1.FieldsCount(), "field count after first field")

	loglet2 := loglet1.WithFields(map[string]any{"b": 2, "c": 3})
	core.AssertEqual(t, 3, loglet2.FieldsCount(), "field count after WithFields")

	loglet3 := loglet2.WithField("d", 4)
	core.AssertEqual(t, 4, loglet3.FieldsCount(), "field count after final field")
}

func testLogletEmptyKey(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// Empty key should not be added
	newLoglet := loglet.WithField("", testValue)

	// Should have same properties (can't compare structs directly)
	core.AssertEqual(t, loglet.FieldsCount(), newLoglet.FieldsCount(), "field count unchanged with empty key")
	core.AssertEqual(t, 0, newLoglet.FieldsCount(), "field count remains zero")
}

func TestLogletFieldsIterator(t *testing.T) {
	t.Run("EmptyIterator", testLogletIteratorEmpty)
	t.Run("SingleField", testLogletIteratorSingle)
	t.Run("MultipleFields", testLogletIteratorMultiple)
	t.Run("ChainedFields", testLogletIteratorChained)
}

func testLogletIteratorEmpty(t *testing.T) {
	t.Helper()
	var loglet Loglet

	iter := loglet.Fields()

	// Should not have any fields
	core.AssertFalse(t, iter.Next(), "empty iterator Next()")
}

func testLogletIteratorSingle(t *testing.T) {
	t.Helper()
	var loglet Loglet
	loglet1 := loglet.WithField("key", testValue)

	iter := loglet1.Fields()

	// Should have one field
	core.AssertMustTrue(t, iter.Next(), "iterator has one field")

	key := iter.Key()
	value := iter.Value()

	core.AssertEqual(t, "key", key, "field key")
	core.AssertEqual(t, testValue, value, "field value")

	// Should not have more fields
	core.AssertFalse(t, iter.Next(), "iterator should not have more fields")
}

func testLogletIteratorMultiple(t *testing.T) {
	t.Helper()
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
	core.AssertEqual(t, len(fields), len(found), "found field count")

	for k, v := range fields {
		core.AssertEqual(t, v, found[k], "field %q", k)
	}
}

func testLogletIteratorChained(t *testing.T) {
	t.Helper()
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

	core.AssertEqual(t, len(expected), len(found), "expected field count")

	for k, v := range expected {
		core.AssertEqual(t, v, found[k], "complex chaining field %q", k)
	}
}

func TestLogletIntegration(t *testing.T) {
	t.Run("ComplexChaining", testLogletComplexChaining)
	t.Run("LevelAndFields", testLogletLevelAndFields)
	t.Run("StackAndFields", testLogletStackAndFields)
}

func testLogletComplexChaining(t *testing.T) {
	t.Helper()
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
	core.AssertEqual(t, slog.Info, loglet5.Level(), "preserved level")

	// Verify stack preserved
	core.AssertNotNil(t, loglet5.CallStack(), "preserved stack")

	// Verify field count
	core.AssertEqual(t, 4, loglet5.FieldsCount(), "complex chaining field count")

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
		core.AssertEqual(t, v, found[k], "field %q", k)
	}
}

func testLogletLevelAndFields(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// Add fields, then change level using proper chaining
	loglet1 := loglet.WithField("before", testValue)
	loglet2 := loglet1.WithLevel(slog.Error)
	loglet3 := loglet2.WithField("after", testValue)

	// Both fields should be present
	core.AssertEqual(t, 2, loglet3.FieldsCount(), "field count")

	// Level should be Error
	core.AssertEqual(t, slog.Error, loglet3.Level(), "level")
}

func testLogletStackAndFields(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// Add fields, then stack using proper chaining
	loglet1 := loglet.WithField("before", testValue)
	loglet2 := loglet1.WithStack(1)
	loglet3 := loglet2.WithField("after", testValue)

	// Both fields should be present
	core.AssertEqual(t, 2, loglet3.FieldsCount(), "field count")

	// Stack should be present
	core.AssertNotNil(t, loglet3.CallStack(), "stack present")

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
	t.Helper()
	var loglet *Loglet
	core.AssertTrue(t, loglet.IsZero(), "nil loglet should be zero")
}

func testLogletIsZeroTrue(t *testing.T) {
	t.Helper()
	var loglet Loglet
	core.AssertTrue(t, loglet.IsZero(), "empty loglet should be zero")
}

func testLogletIsZeroFalse(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// Test with field
	loglet1 := loglet.WithField("key", testValue)
	core.AssertFalse(t, loglet1.IsZero(), "loglet with field should not be zero")

	// Test with level
	loglet2 := loglet.WithLevel(slog.Info)
	core.AssertFalse(t, loglet2.IsZero(), "loglet with level should not be zero")

	// Test with stack
	loglet3 := loglet.WithStack(1)
	core.AssertFalse(t, loglet3.IsZero(), "loglet with stack should not be zero")
}

// TestLogletGetParent tests the GetParent method for all branches
func TestLogletGetParent(t *testing.T) {
	t.Run("NilLoglet", testLogletGetParentNil)
	t.Run("SelfReference", testLogletGetParentSelf)
	t.Run("NormalParent", testLogletGetParentNormal)
}

func testLogletGetParentNil(t *testing.T) {
	t.Helper()
	// Test GetParent method indirectly
	// We'll test indirectly through FieldsCount which uses GetParent
	var nilLoglet *Loglet
	count := nilLoglet.FieldsCount()
	core.AssertEqual(t, 0, count, "nil loglet field count")
}

func testLogletGetParentSelf(t *testing.T) {
	t.Helper()
	t.Run("CircularReference", testCircularReference)
	t.Run("ReassignmentPattern", testReassignmentPattern)
	t.Run("ProperChaining", testProperChaining)
	t.Run("DirectSelfReference", testDirectSelfReference)
}

func testCircularReference(t *testing.T) {
	t.Helper()
	// Test the problematic reassignment pattern
	var loglet Loglet
	loglet = loglet.WithField("key1", "value1")
	loglet = loglet.WithField("key2", "value2")

	// Should work without infinite loop due to GetParent protection
	count := loglet.FieldsCount()
	core.AssertEqual(t, 1, count, "circular reference field count")

	// Test field iteration
	iter := loglet.Fields()
	fieldCount := 0
	for iter.Next() {
		fieldCount++
		_ = iter.Key()
		_ = iter.Value()
	}

	core.AssertEqual(t, 1, fieldCount, "circular reference iterator count")
}

func testReassignmentPattern(t *testing.T) {
	t.Helper()
	// Test reassignment with non-zero base
	var base Loglet
	base = base.WithLevel(slog.Info)
	base = base.WithField("service", "test")
	base = base.WithField("version", "1.0")

	core.AssertEqual(t, 1, base.FieldsCount(), "reassignment field count")
}

func testProperChaining(t *testing.T) {
	t.Helper()
	// Test proper chaining creates field chains
	var chain Loglet
	chain1 := chain.WithField("key1", "value1")
	chain2 := chain1.WithField("key2", "value2")
	chain3 := chain2.WithField("key3", "value3")

	core.AssertEqual(t, 3, chain3.FieldsCount(), "proper chaining field count")

	// Test field iteration
	iter := chain3.Fields()
	fieldCount := 0
	for iter.Next() {
		fieldCount++
	}

	core.AssertEqual(t, 3, fieldCount, "proper chain iterator count")
}

func testDirectSelfReference(t *testing.T) {
	t.Helper()
	// Manually create self-reference
	selfRef := Loglet{parent: nil}
	selfRef.parent = &selfRef

	// Should be caught by GetParent() protection
	core.AssertNil(t, selfRef.GetParent(), "GetParent with self-reference")

	// Should not cause infinite loop in FieldsCount
	core.AssertEqual(t, 0, selfRef.FieldsCount(), "self-reference field count")
}

func testLogletGetParentNormal(t *testing.T) {
	t.Helper()
	var loglet Loglet
	loglet1 := loglet.WithField("key", testValue)

	// Test that parent relationship works correctly
	core.AssertEqual(t, 1, loglet1.FieldsCount(), "normal parent field count")
}

// TestFilterFields tests the filterFields function through WithFields
func TestFilterFields(t *testing.T) {
	t.Run("EmptyMap", testFilterFieldsEmpty)
	t.Run("EmptyKeys", testFilterFieldsEmptyKeys)
	t.Run("MixedKeys", testFilterFieldsMixed)
}

func testFilterFieldsEmpty(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// Test with empty map
	emptyFields := map[string]any{}
	newLoglet := loglet.WithFields(emptyFields)

	core.AssertEqual(t, 0, newLoglet.FieldsCount(), "empty fields count")
}

func testFilterFieldsEmptyKeys(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// Test with map containing only empty keys
	fieldsWithEmptyKeys := map[string]any{
		"":  "value1",
		" ": "value2", // Non-empty key
	}

	newLoglet := loglet.WithFields(fieldsWithEmptyKeys)

	// Should only include the non-empty key
	core.AssertEqual(t, 1, newLoglet.FieldsCount(), "filtered empty keys count")
}

func testFilterFieldsMixed(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// Test with mix of valid and invalid keys
	mixedFields := map[string]any{
		"":      "filtered_out",
		"valid": "kept",
		"also":  "kept",
	}

	newLoglet := loglet.WithFields(mixedFields)

	// Should only include valid keys
	core.AssertEqual(t, 2, newLoglet.FieldsCount(), "mixed fields valid count")
}

// TestWithFieldsEdgeCases tests additional edge cases for WithFields
func TestWithFieldsEdgeCases(t *testing.T) {
	t.Run("ZeroLogletParent", testWithFieldsZeroLogletParent)
	t.Run("NonZeroLogletParent", testWithFieldsNonZeroLogletParent)
}

func testWithFieldsZeroLogletParent(t *testing.T) {
	t.Helper()
	var loglet Loglet // Zero loglet

	fields := map[string]any{"key": testValue}
	newLoglet := loglet.WithFields(fields)

	// Should not set parent for zero loglet
	core.AssertEqual(t, 1, newLoglet.FieldsCount(), "zero loglet parent field count")
}

func testWithFieldsNonZeroLogletParent(t *testing.T) {
	t.Helper()
	var loglet Loglet
	loglet1 := loglet.WithLevel(slog.Info) // Make it non-zero

	fields := map[string]any{"key": testValue}
	newLoglet := loglet1.WithFields(fields)

	// Should set parent for non-zero loglet
	core.AssertEqual(t, 1, newLoglet.FieldsCount(), "non-zero loglet parent field count")
}

func TestLogletCopy(t *testing.T) {
	t.Run("NilLoglet", testLogletCopyNil)
	t.Run("EmptyLoglet", testLogletCopyEmpty)
	t.Run("WithFields", testLogletCopyWithFields)
	t.Run("IndependentCopies", testLogletCopyIndependent)
}

func testLogletCopyNil(t *testing.T) {
	t.Helper()
	var ll *Loglet
	copied := ll.Copy()

	// Should return zero Loglet for nil input
	core.AssertTrue(t, copied.IsZero(), "nil loglet copy should be zero")
}

func testLogletCopyEmpty(t *testing.T) {
	t.Helper()
	var ll Loglet
	copied := ll.Copy()

	// Should be equal but independent
	core.AssertTrue(t, copied.IsZero(), "zero loglet copy should be zero")
	core.AssertEqual(t, ll.Level(), copied.Level(), "copy level")
}

func testLogletCopyWithFields(t *testing.T) {
	t.Helper()
	var base Loglet
	l1 := base.WithField("key1", "value1")
	l2 := l1.WithField("key2", "value2")
	original := l2.WithLevel(slog.Info)
	copied := original.Copy()

	// Verify all fields are preserved
	core.AssertEqual(t, original.Level(), copied.Level(), "level")
	core.AssertEqual(t, original.FieldsCount(), copied.FieldsCount(), "field count")
}

func testLogletCopyIndependent(t *testing.T) {
	t.Helper()
	var base Loglet
	original := base.WithField("original", "value")
	copied := original.Copy()

	// Create new loglets from each
	originalChild := original.WithField("child", "original")
	copyChild := copied.WithField("child", "copy")

	// They should be independent - verify field counts
	core.AssertEqual(t, 2, originalChild.FieldsCount(), "original child field count")
	core.AssertEqual(t, 2, copyChild.FieldsCount(), "copy child field count")

	// Verify fields via iterator
	originalIter := originalChild.Fields()
	originalFields := make(map[string]any)
	for originalIter.Next() {
		k, v := originalIter.Field()
		originalFields[k] = v
	}

	copyIter := copyChild.Fields()
	copyFields := make(map[string]any)
	for copyIter.Next() {
		k, v := copyIter.Field()
		copyFields[k] = v
	}

	core.AssertNotEqual(t, originalFields["child"], copyFields["child"], "independent children")
	core.AssertEqual(t, "original", originalFields["child"], "original child value")
	core.AssertEqual(t, "copy", copyFields["child"], "copy child value")
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
	t.Helper()
	var loglet *Loglet
	fields := loglet.FieldsMap()
	core.AssertNil(t, fields, "nil loglet FieldsMap")
}

func testFieldsMapEmpty(t *testing.T) {
	t.Helper()
	var loglet Loglet
	fields := loglet.FieldsMap()
	core.AssertNotNil(t, fields, "empty loglet FieldsMap")
	core.AssertEqual(t, 0, len(fields), "empty loglet field count")
}

func testFieldsMapSingle(t *testing.T) {
	t.Helper()
	var loglet Loglet
	loglet1 := loglet.WithField("key", testValue)

	fields := loglet1.FieldsMap()
	core.AssertMustNotNil(t, fields, "single field FieldsMap")
	core.AssertEqual(t, 1, len(fields), "single field count")
	core.AssertEqual(t, "value", fields["key"], "field value")
}

func testFieldsMapMultiple(t *testing.T) {
	t.Helper()
	var loglet Loglet
	inputFields := map[string]any{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	loglet1 := loglet.WithFields(inputFields)
	fields := loglet1.FieldsMap()

	core.AssertMustNotNil(t, fields, "multi-field FieldsMap")
	core.AssertEqual(t, len(inputFields), len(fields), "field count")
	for k, v := range inputFields {
		core.AssertEqual(t, v, fields[k], "field %q", k)
	}
}

func testFieldsMapChained(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// Create chain of loglets using proper chaining
	loglet1 := loglet.WithField("root", "rootValue")
	loglet2 := loglet1.WithField("child", "childValue")
	loglet3 := loglet2.WithFields(map[string]any{
		"grand": "grandValue",
		"leaf":  "leafValue",
	})

	fields := loglet3.FieldsMap()
	core.AssertMustNotNil(t, fields, "chained FieldsMap")

	expected := map[string]any{
		"root":  "rootValue",
		"child": "childValue",
		"grand": "grandValue",
		"leaf":  "leafValue",
	}

	core.AssertEqual(t, len(expected), len(fields), "chained field count")
	for k, v := range expected {
		core.AssertEqual(t, v, fields[k], "chained field %q", k)
	}
}

func testFieldsMapOverride(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// Create chain where child overrides parent field
	loglet1 := loglet.WithField("key", "parentValue")
	loglet2 := loglet1.WithField("key", "childValue") // Same key

	fields := loglet2.FieldsMap()
	core.AssertMustNotNil(t, fields, "override FieldsMap")
	core.AssertEqual(t, "childValue", fields["key"], "overridden field value")
	core.AssertEqual(t, 1, len(fields), "override field count")
}

func testFieldsMapCaching(t *testing.T) {
	t.Helper()
	var loglet Loglet
	loglet1 := loglet.WithField("key", testValue)

	// First call should build the cache
	fields1 := loglet1.FieldsMap()
	core.AssertMustNotNil(t, fields1, "first call FieldsMap")

	// Second call should return same cached map
	fields2 := loglet1.FieldsMap()
	core.AssertMustNotNil(t, fields2, "second call FieldsMap")

	// Should be the same map instance (cached)
	core.AssertMustSame(t, fields1, fields2, "cached map instance")
}

func testFieldsMapImmutableCache(t *testing.T) {
	t.Helper()
	var loglet Loglet
	loglet1 := loglet.WithField("original", testValue)

	// Cache the fields map
	fields1 := loglet1.FieldsMap()
	core.AssertMustNotNil(t, fields1, "original FieldsMap")

	// Create new loglet with additional field - should not have cache
	loglet2 := loglet1.WithField("additional", "value2")
	fields2 := loglet2.FieldsMap()

	core.AssertMustNotNil(t, fields2, "new FieldsMap")
	core.AssertEqual(t, 2, len(fields2), "new loglet field count")
	core.AssertEqual(t, 1, len(fields1), "original loglet field count")

	// Different map instances
	core.AssertNotSame(t, fields1, fields2, "different map instances")
}

func testFieldsMapSharedWarning(t *testing.T) {
	t.Helper()
	// Test that demonstrates why the returned map should not be modified
	var loglet Loglet
	loglet1 := loglet.WithField("key", "original")

	// Get the cached map
	fields1 := loglet1.FieldsMap()
	core.AssertMustNotNil(t, fields1, "cached FieldsMap")
	core.AssertEqual(t, "original", fields1["key"], "field value")

	// Get the map again (should be same cached instance)
	fields2 := loglet1.FieldsMap()
	core.AssertMustSame(t, fields1, fields2, "same cached instance")

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

// TestFieldsMapParentDelegation tests the new parent delegation optimisation
func TestFieldsMapParentDelegation(t *testing.T) {
	t.Helper()
	var base Loglet
	parent := base.WithField("parent_key", "parent_value")
	child := parent.WithLevel(slog.Info) // No fields, only level change

	parentMap := parent.FieldsMap()
	childMap := child.FieldsMap()

	// Should return same map reference (delegation, not copy)
	core.AssertMustSame(t, parentMap, childMap, "child delegation")
	core.AssertEqual(t, "parent_value", childMap["parent_key"], "delegated value")
}

// TestFieldsMapMultiLevelDelegation tests delegation through multiple levels
func TestFieldsMapMultiLevelDelegation(t *testing.T) {
	t.Helper()
	var base Loglet
	l1 := base.WithField("key1", "value1")
	l2 := l1.WithLevel(slog.Info)        // No fields, delegates
	l3 := l2.WithStack(1)                // No fields, delegates
	l4 := l3.WithField("key2", "value2") // Has fields, builds map

	// l2 and l3 should delegate to l1
	map1 := l1.FieldsMap()
	map2 := l2.FieldsMap()
	map3 := l3.FieldsMap()

	// l2 and l3 should delegate to l1 (same map instance)
	core.AssertMustSame(t, map1, map2, "l2 delegates to l1")
	core.AssertMustSame(t, map1, map3, "l3 delegates to l1")

	// l4 should have its own map with both keys
	map4 := l4.FieldsMap()
	core.AssertNotSame(t, map1, map4, "l4 has own map")
	core.AssertEqual(t, 2, len(map4), "l4 field count")
	core.AssertEqual(t, "value1", map4["key1"], "l4 key1 value")
	core.AssertEqual(t, "value2", map4["key2"], "l4 key2 value")
}

// TestFieldsMapDelegationCaching tests caching behaviour with delegation
func TestFieldsMapDelegationCaching(t *testing.T) {
	t.Helper()
	var base Loglet
	parent := base.WithField("key", "value")
	child := parent.WithLevel(slog.Info)

	// First call should delegate
	map1 := child.FieldsMap()

	// Second call should still delegate (not cache the delegated result)
	map2 := child.FieldsMap()

	// Should be same reference (delegates each time)
	core.AssertMustSame(t, map1, map2, "delegation consistency")

	// Should delegate to parent
	parentMap := parent.FieldsMap()
	core.AssertMustSame(t, parentMap, map1, "child delegates to parent")
}

// TestFieldsMapCopy tests the FieldsMapCopy method
func TestFieldsMapCopy(t *testing.T) {
	t.Run("NilLoglet", testFieldsMapCopyNil)
	t.Run("EmptyLoglet", testFieldsMapCopyEmpty)
	t.Run("SingleField", testFieldsMapCopySingle)
	t.Run("MultipleFields", testFieldsMapCopyMultiple)
	t.Run("ChainedFields", testFieldsMapCopyChained)
	t.Run("WithExcess", testFieldsMapCopyWithExcess)
	t.Run("CachedSource", testFieldsMapCopyCachedSource)
	t.Run("ModifiableCopy", testFieldsMapCopyModifiable)
	t.Run("NegativeExcess", testFieldsMapCopyNegativeExcess)
}

func testFieldsMapCopyNil(t *testing.T) {
	t.Helper()
	var loglet *Loglet
	fields := loglet.FieldsMapCopy(0)
	core.AssertNil(t, fields, "nil loglet FieldsMapCopy")
}

func testFieldsMapCopyEmpty(t *testing.T) {
	t.Helper()
	var loglet Loglet
	fields := loglet.FieldsMapCopy(0)
	core.AssertNotNil(t, fields, "empty loglet FieldsMapCopy")
	core.AssertEqual(t, 0, len(fields), "empty loglet field count")
}

func testFieldsMapCopySingle(t *testing.T) {
	t.Helper()
	var loglet Loglet
	loglet1 := loglet.WithField("key", testValue)

	fields := loglet1.FieldsMapCopy(0)
	core.AssertMustNotNil(t, fields, "single field FieldsMapCopy")
	core.AssertEqual(t, 1, len(fields), "single field count")
	core.AssertEqual(t, testValue, fields["key"], "field value")
}

func testFieldsMapCopyMultiple(t *testing.T) {
	t.Helper()
	var loglet Loglet
	inputFields := map[string]any{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	loglet1 := loglet.WithFields(inputFields)
	fields := loglet1.FieldsMapCopy(0)

	core.AssertMustNotNil(t, fields, "multi-field FieldsMapCopy")
	core.AssertEqual(t, len(inputFields), len(fields), "field count")
	for k, v := range inputFields {
		core.AssertEqual(t, v, fields[k], "field %q", k)
	}
}

func testFieldsMapCopyChained(t *testing.T) {
	t.Helper()
	var loglet Loglet

	// Create chain of loglets using proper chaining
	loglet1 := loglet.WithField("root", "rootValue")
	loglet2 := loglet1.WithField("child", "childValue")
	loglet3 := loglet2.WithFields(map[string]any{
		"grand": "grandValue",
		"leaf":  "leafValue",
	})

	fields := loglet3.FieldsMapCopy(0)
	core.AssertMustNotNil(t, fields, "chained FieldsMapCopy")

	expected := map[string]any{
		"root":  "rootValue",
		"child": "childValue",
		"grand": "grandValue",
		"leaf":  "leafValue",
	}

	core.AssertEqual(t, len(expected), len(fields), "chained field count")
	for k, v := range expected {
		core.AssertEqual(t, v, fields[k], "chained field %q", k)
	}
}

func testFieldsMapCopyWithExcess(t *testing.T) {
	var loglet Loglet
	loglet1 := loglet.WithField("key", testValue)

	testCases := []int{0, 1, 5, 10}
	for _, excess := range testCases {
		validateExcessCapacity(t, &loglet1, excess)
	}
}

func validateExcessCapacity(t *testing.T, loglet *Loglet, excess int) {
	t.Helper()
	fields := loglet.FieldsMapCopy(excess)
	core.AssertMustNotNil(t, fields, "excess %d map", excess)

	// Add excess fields to verify capacity
	for i := range excess {
		key := fmt.Sprintf("extra%d", i)
		fields[key] = fmt.Sprintf("value%d", i)
	}

	expectedLen := 1 + excess
	core.AssertEqual(t, expectedLen, len(fields), "excess %d field count", excess)
	core.AssertEqual(t, testValue, fields["key"], "excess %d original field", excess)
}

func testFieldsMapCopyCachedSource(t *testing.T) {
	t.Helper()
	var loglet Loglet
	loglet1 := loglet.WithField("key", testValue)

	// Prime the cache with FieldsMap
	cached := loglet1.FieldsMap()
	core.AssertMustNotNil(t, cached, "cached map")

	// Now get copy
	copyMap := loglet1.FieldsMapCopy(2)
	core.AssertMustNotNil(t, copyMap, "copy from cached source")

	// Should be different instances
	core.AssertNotSame(t, cached, copyMap, "different instances")

	// But should have same content
	core.AssertEqual(t, len(cached), len(copyMap), "same field count")
	for k, v := range cached {
		core.AssertEqual(t, v, copyMap[k], "field %q content", k)
	}
}

func testFieldsMapCopyModifiable(t *testing.T) {
	t.Helper()
	var loglet Loglet
	loglet1 := loglet.WithField("original", testValue)

	copyMap := loglet1.FieldsMapCopy(1)
	core.AssertMustNotNil(t, copyMap, "modifiable copy")

	// Modify the copy
	copyMap["original"] = "modified"
	copyMap["new"] = "added"

	// Get original via FieldsMap
	original := loglet1.FieldsMap()
	core.AssertMustNotNil(t, original, "original map")

	// Original should be unchanged
	core.AssertEqual(t, testValue, original["original"], "original unchanged")
	_, exists := original["new"]
	core.AssertFalse(t, exists, "new field not in original")

	// Copy should be modified
	core.AssertEqual(t, "modified", copyMap["original"], "copy modified")
	core.AssertEqual(t, "added", copyMap["new"], "copy added field")
}

func testFieldsMapCopyNegativeExcess(t *testing.T) {
	t.Helper()
	var loglet Loglet
	loglet1 := loglet.WithField("key", testValue)

	// Test negative excess - should handle gracefully
	copyMap := loglet1.FieldsMapCopy(-5)
	core.AssertMustNotNil(t, copyMap, "negative excess handled")
	core.AssertEqual(t, testValue, copyMap["key"], "original field preserved")
}
