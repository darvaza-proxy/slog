package internal

import (
	"maps"
	"sync"

	"darvaza.org/core"
	"darvaza.org/slog"
)

var (
	_ core.CallStacker = (*Loglet)(nil)
)

// Loglet represents an immutable link in a logger context chain. Each Loglet
// contains fields, log level, and call stack information, with a reference to
// its parent Loglet forming a chain of contexts.
//
// Loglets are immutable - each With* method returns a new Loglet instance.
// This design ensures thread safety and prevents accidental modification.
//
// Usage pattern:
//
//	var base internal.Loglet
//	l1 := base.WithField("user_id", 123)
//	l2 := l1.WithField("action", "login")
//
// IMPORTANT: Avoid reassignment patterns like:
//
//	loglet = loglet.WithField(...) // Can cause circular references
//
// Instead, use proper chaining with new variable names.
type Loglet struct {
	parent    *Loglet
	level     slog.LogLevel
	keys      []string
	values    []any
	stack     core.Stack
	fieldsMap map[string]any // cached fields map
	fieldsMu  sync.Mutex     // protects fieldsMap access and computation
}

// IsZero returns true if the loglet has no meaningful content.
// A zero loglet has no parent, no fields, no stack, and default level.
func (ll *Loglet) IsZero() bool {
	if ll == nil {
		return true
	}
	return ll.parent == nil && len(ll.keys) == 0 && ll.stack == nil && ll.level == 0
}

// GetParent returns the parent loglet with circular reference protection.
// Returns nil if parent points to self, preventing infinite loops during traversal.
func (ll *Loglet) GetParent() *Loglet {
	switch ll {
	case nil, ll.parent:
		return nil
	default:
		return ll.parent
	}
}

// Copy returns a copy of the loglet without copying the sync.Once field.
// This avoids copy-locks warnings while preserving the cached fieldsMap.
func (ll *Loglet) Copy() Loglet {
	if ll == nil {
		return Loglet{}
	}
	return Loglet{
		parent:    ll.parent,
		level:     ll.level,
		keys:      ll.keys,
		values:    ll.values,
		stack:     ll.stack,
		fieldsMap: ll.fieldsMap,
		// fieldsMu is intentionally omitted - will be zero value
	}
}

// Level returns the LogLevel of a Loglet
func (ll *Loglet) Level() slog.LogLevel {
	return ll.level
}

// WithLevel creates a new Loglet with the specified log level.
// Returns the same loglet if the level is unchanged.
func (ll *Loglet) WithLevel(level slog.LogLevel) Loglet {
	if level == ll.level {
		return ll.Copy()
	}

	return Loglet{
		parent: ll,
		level:  level,
		stack:  ll.stack,
	}
}

// CallStack returns the callstack associated to a Loglet
func (ll *Loglet) CallStack() core.Stack {
	return ll.stack
}

// WithStack creates a new Loglet with an attached call stack.
// The skip parameter indicates how many stack frames to skip.
func (ll *Loglet) WithStack(skip int) Loglet {
	return Loglet{
		parent: ll,
		level:  ll.level,
		stack:  core.StackTrace(skip + 1),
	}
}

// WithField attaches a field to a new Loglet.
// Returns the same loglet if label is empty.
// Only sets parent if current loglet has meaningful content.
func (ll *Loglet) WithField(label string, value any) Loglet {
	if label == "" {
		return ll.Copy()
	}

	var parent *Loglet
	if !ll.IsZero() {
		parent = ll
	}

	return Loglet{
		parent: parent,
		level:  ll.level,
		stack:  ll.stack,
		keys:   []string{label},
		values: []any{value},
	}
}

// WithFields attaches a set of fields to a new Loglet.
// Returns the same loglet if no valid fields are provided.
// Empty keys are filtered out. Only sets parent if current loglet has meaningful content.
func (ll *Loglet) WithFields(fields map[string]any) Loglet {
	keys, values, count := filterFields(fields)
	if count == 0 {
		return ll.Copy()
	}

	var parent *Loglet
	if !ll.IsZero() {
		parent = ll
	}

	return Loglet{
		parent: parent,
		level:  ll.level,
		stack:  ll.stack,
		keys:   keys,
		values: values,
	}
}

// filterFields extracts non-empty keys and their values from a map
func filterFields(fields map[string]any) ([]string, []any, int) {
	count := len(fields)
	if count == 0 {
		return nil, nil, 0
	}

	keys := make([]string, 0, len(fields))
	values := make([]any, 0, len(fields))

	for k, v := range fields {
		if k != "" {
			keys = append(keys, k)
			values = append(values, v)
		}
	}

	return keys, values, len(keys)
}

// FieldsCount return the number of fields on a Log context
func (ll *Loglet) FieldsCount() int {
	if ll == nil {
		return 0
	}

	count := 0
	current := ll
	for current != nil {
		count += len(current.keys)
		current = current.GetParent()
	}
	return count
}

// Fields returns a FieldsIterator for traversing all fields in the context chain.
// The iterator walks from the current loglet up through the parent chain.
func (ll *Loglet) Fields() (iter *FieldsIterator) {
	return &FieldsIterator{
		ll: ll,
		i:  0,
	}
}

// FieldsMap returns a map containing all fields from the loglet chain.
// The map is built once and cached for subsequent calls, providing
// better performance than iterating through Fields() multiple times.
//
// Fields from parent loglets are included, with child fields overriding
// parent fields when keys collide. Returns nil only for nil loglets,
// otherwise returns an empty map for loglets with no fields.
//
// WARNING: The returned map is shared and MUST NOT be modified.
// If you need to modify fields, use FieldsMapCopy() instead.
//
// Note: The returned map is immutable by design for performance reasons.
// Any attempt to modify it may cause undefined behaviour in concurrent
// environments and break the caching mechanism.
func (ll *Loglet) FieldsMap() map[string]any {
	fieldsMap, _ := ll.getFieldsMap()
	return fieldsMap
}

// getFieldsMap computes and caches the fields map if needed.
// Returns the fields map and whether it was freshly computed.
func (ll *Loglet) getFieldsMap() (fields map[string]any, fresh bool) {
	if ll == nil {
		return nil, false
	}

	ll.fieldsMu.Lock()
	defer ll.fieldsMu.Unlock()

	fields = ll.fieldsMap
	if fields == nil {
		// Compute and cache
		fields, fresh = ll.buildFieldsMap()
		ll.fieldsMap = fields
	}

	return fields, fresh
}

// peekFieldsMap returns the cached fields map without computing it.
// If needsLock is true, it acquires the mutex for thread-safe access.
// If needsLock is false, it assumes the caller already holds the lock.
//
//revive:disable-next-line:flag-parameter
func (ll *Loglet) peekFieldsMap(needsLock bool) (map[string]any, bool) {
	if ll == nil {
		return nil, false
	}

	if needsLock {
		ll.fieldsMu.Lock()
		defer ll.fieldsMu.Unlock()
	}

	m := ll.fieldsMap
	return m, m != nil
}

// FieldsMapCopy returns a modifiable copy of the fields map with optional
// excess capacity. Unlike FieldsMap(), this always returns a new map that
// can be safely modified without affecting the original loglet.
//
// The excess parameter specifies additional capacity beyond the current
// field count, useful when you plan to add more fields to the returned map.
// Negative excess values are normalized to zero.
//
// Returns nil for nil loglets, consistent with FieldsMap().
func (ll *Loglet) FieldsMapCopy(excess int) map[string]any {
	if ll == nil {
		return nil
	}

	// Normalize excess to avoid negative values
	if excess < 0 {
		excess = 0
	}

	return copyFieldsMap(ll.FieldsMap(), excess)
}

// buildFieldsMap builds and returns a fields map, either by delegating to
// a cached ancestor or by building from the traversal path
func (ll *Loglet) buildFieldsMap() (map[string]any, bool) {
	// Find cached ancestor and build from there
	info := ll.findCachedAncestor()

	// Get base map, ensuring ancestors with fields get cached if needed
	baseMap := info.getBaseMap()

	// Check if we can delegate to parent
	if info.canDelegate() {
		return baseMap, false
	}

	// Build our own map
	return ll.buildFieldsMapFromPath(info.pathToRoot, baseMap, info.totalFields)
}

// buildFieldsMapFromPath builds a fields map from the given path and base map
func (ll *Loglet) buildFieldsMapFromPath(
	pathToRoot []*Loglet, baseMap map[string]any, totalFields int,
) (map[string]any, bool) {
	result := make(map[string]any, totalFields)
	if baseMap != nil {
		maps.Copy(result, baseMap)
	}
	ll.populateFromPath(pathToRoot, result)
	return result, true
}

// cacheIfNeeded caches this loglet's field map if it has fields but no cache
func (ll *Loglet) cacheIfNeeded() (map[string]any, bool) {
	if len(ll.keys) == 0 {
		return nil, false
	}

	return ll.getFieldsMap()
}

// populateFromPath adds fields from the path to the result map
func (*Loglet) populateFromPath(pathToRoot []*Loglet, result map[string]any) {
	for i := len(pathToRoot) - 1; i >= 0; i-- {
		node := pathToRoot[i]
		node.addOwnFields(result)
	}
}

// addOwnFields adds only this loglet's own fields to the map
func (ll *Loglet) addOwnFields(fields map[string]any) {
	for i, key := range ll.keys {
		if key != "" {
			fields[key] = ll.values[i]
		}
	}
}

// copyFieldsMap creates a copy of the fields map with optional excess capacity
func copyFieldsMap(source map[string]any, excess int) map[string]any {
	if source == nil {
		return nil
	}

	totalCap := len(source) + excess
	result := make(map[string]any, totalCap)
	maps.Copy(result, source)
	return result
}

// FieldsIterator iterates over fields in a Loglet context chain.
// Use Next() to advance and Key(), Value(), or Field() to access current values.
type FieldsIterator struct {
	ll *Loglet
	i  int
	k  string
	v  any
}

// Next advances the iterator to the next field in the chain.
// Returns false when iteration is complete, true when a field is available.
// Call Key(), Value(), or Field() to access the current field after Next() returns true.
func (iter *FieldsIterator) Next() bool {
	for iter.ll != nil {
		ll := iter.ll

		if i := iter.i; i < len(ll.keys) {
			iter.k = ll.keys[i]
			iter.v = ll.values[i]
			iter.i = i + 1
			return true
		}

		// up
		iter.ll = iter.ll.GetParent()
		iter.i = 0
	}
	return false
}

// Key returns the label of the current field
func (iter *FieldsIterator) Key() string {
	return iter.k
}

// Value returns the value of the current field
func (iter *FieldsIterator) Value() any {
	return iter.v
}

// Field returns key and value of the current field
func (iter *FieldsIterator) Field() (key string, value any) {
	return iter.k, iter.v
}
