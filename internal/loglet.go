package internal

import "github.com/darvaza-proxy/slog"

// Loglet represents a link on the Logger context chain
type Loglet struct {
	parent *Loglet
	level  slog.LogLevel
	keys   []string
	values []any
	stack  Stack
}

// Level returns the LogLevel of a Loglet
func (ll *Loglet) Level() slog.LogLevel {
	return ll.level
}

// WithLevel sets the LogLevel for a new Loglet
func (ll *Loglet) WithLevel(level slog.LogLevel) Loglet {
	if level == ll.level {
		return *ll
	}

	return Loglet{
		parent: ll,
		level:  level,
		stack:  ll.stack,
	}
}

// Stack returns the callstack associated to a Loglet
func (ll *Loglet) Stack() Stack {
	return ll.stack
}

// WithStack attaches a call stack to a new Loglet
func (ll *Loglet) WithStack(skip int) Loglet {
	return Loglet{
		parent: ll,
		level:  ll.level,
		stack:  StackTrace(skip + 1),
	}
}

// WithField attaches a field to a new Loglet
func (ll *Loglet) WithField(label string, value any) Loglet {
	return Loglet{
		parent: ll,
		level:  ll.level,
		stack:  ll.stack,
		keys:   []string{label},
		values: []any{value},
	}
}

// WithFields attaches a set of fields to a new Loglet
func (ll *Loglet) WithFields(fields map[string]any) Loglet {
	if n := len(fields); n > 0 {
		keys := make([]string, n)
		values := make([]any, n)

		i := 0
		for k, v := range fields {
			keys[i] = k
			values[i] = v
			i++
		}

		return Loglet{
			parent: ll,
			level:  ll.level,
			stack:  ll.stack,
			keys:   keys[:i],
			values: values[:i],
		}
	}
	return *ll
}

// FieldsCount return the number of fields on a Log context
func (ll *Loglet) FieldsCount() int {
	count := 0
	for ll != nil {
		count += len(ll.keys)
		ll = ll.parent
	}
	return count
}

// Fields returns a FieldsIterator
func (ll *Loglet) Fields() (iter *FieldsIterator) {
	return &FieldsIterator{
		ll: ll,
		i:  0,
	}
}

// FieldsIterator iterates over fields on a Log context
type FieldsIterator struct {
	ll *Loglet
	i  int
	k  string
	v  any
}

// Next advances iterator to next value. it returns false to indicate
// end of iteration, or true when the next (or first) field
// is ready to be accessed using Key(), Value(), or Field()
// when there are no new ones
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
		iter.ll = iter.ll.parent
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
