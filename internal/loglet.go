package internal

import (
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
	parent *Loglet
	level  slog.LogLevel
	keys   []string
	values []any
	stack  core.Stack
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

// Copy creates a shallow copy of the Loglet, preserving all fields.
func (ll *Loglet) Copy() Loglet {
	if ll == nil {
		return Loglet{}
	}
	return Loglet{
		parent: ll.parent,
		level:  ll.level,
		keys:   ll.keys,
		values: ll.values,
		stack:  ll.stack,
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
