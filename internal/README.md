# slog/internal

This package provides internal implementation details for the slog library,
including the core `Loglet` structure that manages immutable logging contexts.

## Loglet

The `Loglet` type is the foundation of slog's immutable logger design. It
represents a link in a chain of logging contexts, where each link contains
fields, log level, and call stack information.

### Key Features

- **Immutable Design**: Each `With*` method returns a new `Loglet` instance
- **Chain Structure**: Loglets form a parent-child chain for context inheritance
- **Thread-Safe**: Immutability ensures safe concurrent access
- **Memory Efficient**: Shared parent chains avoid duplication
- **Circular Reference Protection**: Built-in safeguards against infinite loops

### Core Methods

#### Field Management

```go
// Add single field
loglet := base.WithField("user_id", 123)

// Add multiple fields
loglet := base.WithFields(map[string]any{
    "service": "api",
    "version": "1.0",
})

// Count total fields in chain
count := loglet.FieldsCount()
```

#### Field Access

Two methods are provided for accessing fields, each optimized for different
use cases:

##### Fields() Iterator - For Modification

Use when you need to modify, transform, or selectively process fields:

```go
// Build modifiable copy
fields := make(map[string]any, loglet.FieldsCount())
iter := loglet.Fields()
for iter.Next() {
    k, v := iter.Field()
    fields[k] = transformValue(v)  // Safe to modify
}
```

##### FieldsMap() - For Read-Only Access

Use when you only need to read fields without modification:

```go
// Get cached read-only map
if fields := loglet.FieldsMap(); fields != nil {
    for k, v := range fields {
        sendToLogger(k, v)  // Read-only operations only
    }
}
```

**⚠️ CRITICAL WARNING**: The map returned by `FieldsMap()` is cached and shared.
**NEVER modify it directly** as this will break immutability and affect all
future calls on the same loglet instance.

##### FieldsMapCopy() - For Modifiable Maps

Use when you need to modify or extend the fields map:

```go
// Get modifiable copy with extra capacity for new fields
modifiable := loglet.FieldsMapCopy(5)  // Extra capacity for 5 more fields
if modifiable != nil {
    modifiable["new_field"] = "safe to modify"
    modifiable["timestamp"] = time.Now()
    // Process or pass to functions that need to modify the map
}

// Get exact copy without extra capacity
exactCopy := loglet.FieldsMapCopy(0)
```

**Key Benefits**:

- Always returns a new, modifiable map
- Optional excess capacity for efficient field additions
- Safe to modify without affecting the original loglet
- Consistent nil behaviour with FieldsMap()

#### Parent Delegation Optimization

When a loglet has no fields of its own but has a parent with fields,
`FieldsMap()` delegates directly to the parent instead of building an
empty map and iterating through the chain. This provides better
performance for intermediate loglets created for level/stack changes.

```go
var base Loglet
parent := base.WithField("service", "api")
child := parent.WithLevel(slog.Info) // No fields, only level

// child.FieldsMap() delegates to parent.FieldsMap()
// Returns same map reference - no iteration needed
parentMap := parent.FieldsMap()
childMap := child.FieldsMap()
// parentMap and childMap point to the same map instance
```

**Delegation Rules**:

- If loglet has fields: builds and caches its own map
- If loglet has no fields but has parent: delegates to parent's `FieldsMap()`
- If loglet has no fields and no parent: returns empty cached map

**Performance Benefits**:

- Eliminates unnecessary field iteration for intermediate loglets
- Reduces memory allocations for field-less loglets
- Maintains O(1) performance for delegation calls

#### Level and Stack Management

```go
// Set log level
loglet := base.WithLevel(slog.Info)

// Attach call stack (skip frames as needed)
loglet := base.WithStack(1)

// Check current level
level := loglet.Level()

// Get call stack
stack := loglet.CallStack()
```

### Usage Patterns

#### ✅ Correct: Proper Chaining

```go
var base Loglet
step1 := base.WithField("service", "api")
step2 := step1.WithLevel(slog.Info)
step3 := step2.WithField("request_id", "123")
```

#### ❌ Incorrect: Reassignment Pattern

```go
var loglet Loglet
loglet = loglet.WithField("key1", "value1")  // Can cause circular references
loglet = loglet.WithField("key2", "value2")  // Avoid this pattern
```

#### ✅ Correct: Read-Only Field Access

```go
// For handlers that only read fields
if fields := loglet.FieldsMap(); fields != nil {
    json.Marshal(fields)  // Safe - no modification
}
```

#### ❌ Incorrect: Modifying Cached Map

```go
if fields := loglet.FieldsMap(); fields != nil {
    fields["new"] = "value"  // ❌ Breaks immutability!
}
```

#### ✅ Correct: Building Modifiable Copy

```go
// Method 1: Use FieldsMapCopy() (recommended)
modifiable := loglet.FieldsMapCopy(3)  // Extra capacity for 3 more fields
if modifiable != nil {
    for k, v := range modifiable {
        modifiable[k] = processValue(v)  // Safe to modify
    }
    modifiable["extra"] = "new field"  // Safe to add
}

// Method 2: Manual iteration (when custom processing needed)
modifiable := make(map[string]any)
iter := loglet.Fields()
for iter.Next() {
    k, v := iter.Field()
    modifiable[k] = processValue(v)  // Safe to modify
}
```

### Field Precedence

When the same key appears in multiple loglets in a chain, the **most recent**
(child) value takes precedence:

```go
parent := base.WithField("env", "development")
child := parent.WithField("env", "production")

// child.FieldsMap()["env"] == "production" (child overrides parent)
```

### Performance Characteristics

- **FieldsCount()**: O(n) where n is chain depth
- **Fields()**: O(1) to create iterator, O(total fields) to iterate
- **FieldsMap()**: O(total fields) first call, O(1) subsequent calls (cached)
- **FieldsMapCopy()**: O(total fields) to build copy, O(1) if delegating to
  cached parent
- **With* methods**: O(1) for field/level/stack operations

### Advanced Delegation System

The FieldsMap implementation includes a sophisticated delegation and caching
system that optimizes performance for complex loglet chains through intelligent
ancestor traversal and selective caching.

#### Ancestor Caching Strategy

The system automatically identifies opportunities to cache field maps at
strategic points in the loglet chain:

```go
parent := base.WithField("service", "api")
child1 := parent.WithLevel(slog.Info)     // No fields - delegates to parent
child2 := child1.WithStack(1)             // No fields - delegates through chain
child3 := child2.WithField("user", "123") // Has fields - builds own map

// child1.FieldsMap() and child2.FieldsMap() both delegate to parent
// child3.FieldsMap() builds a new map incorporating all ancestor fields
```

#### Delegation Decision Logic

The system uses the following rules to determine optimal field access strategy:

1. **Direct Delegation**: Loglet has no fields but parent has cached map
   - Returns parent's cached map directly (O(1) performance)
   - No memory allocation or field iteration required

2. **Ancestor Caching**: Loglet has fields but ancestor in chain has many fields
   - Proactively cache ancestor's map for future delegation
   - Build current map from cached ancestor + remaining path

3. **Fresh Construction**: No suitable cached ancestor found
   - Build map from scratch by traversing entire chain
   - Cache result for future calls

#### Intelligent Caching Heuristics

The system employs heuristics to decide when ancestor caching is beneficial:

- **Field Threshold**: Ancestors with multiple fields are more likely to be
  cached
- **Chain Depth**: Deeper chains benefit more from intermediate caching
- **Access Patterns**: Frequently accessed intermediate loglets get priority

```go
// Example: Strategic caching in middleware chains
base := loglet.WithField("service", "api")
request := base.WithField("request_id", "123")
user := request.WithField("user_id", "456")

// Intermediate loglets for different contexts
trace := user.WithLevel(slog.Debug)    // Delegates to user
error := trace.WithStack(1)            // Delegates through chain
final := error.WithField("error", err) // Builds from cached user + path
```

#### Memory Efficiency

The delegation system provides significant memory benefits:

- **Shared Maps**: Multiple loglets can reference the same cached map
- **Reduced Allocations**: Delegation eliminates redundant map construction
- **Garbage Collection**: Fewer temporary maps reduce GC pressure
- **Chain Optimization**: Long chains benefit from strategic cache points

### Thread Safety

Loglets are immutable after creation, making them inherently thread-safe:

- Multiple goroutines can safely call methods on the same loglet
- No synchronization primitives needed
- Cached data (like `FieldsMap()`) is safe to share

### Memory Management

- Parent chains are shared between loglets to minimize memory usage
- Cached maps are only created when `FieldsMap()` is called
- Circular reference protection prevents memory leaks
- Zero-value loglets consume minimal memory

### Integration with Handlers

Handler implementations should contain `Loglet` as a field and delegate field
management. Since `Loglet` contains `sync.Once`, use the `Copy()` method when
creating new handler instances:

```go
type Handler struct {
    loglet  internal.Loglet
    backend SomeLogger
}

func (h *Handler) WithField(key string, value any) slog.Logger {
    return &Handler{
        loglet:  h.loglet.WithField(key, value).Copy(),
        backend: h.backend,
    }
}

func (h *Handler) print(msg string) {
    // Use FieldsMap() for read-only access
    if fields := h.loglet.FieldsMap(); fields != nil {
        h.backend.LogWithFields(h.loglet.Level(), msg, fields)
    } else {
        h.backend.Log(h.loglet.Level(), msg)
    }
}
```

### Testing

The package includes comprehensive tests covering:

- Basic field operations and chaining
- Circular reference protection
- Field precedence and overrides
- Iterator functionality
- Caching behaviour and thread safety
- Integration scenarios

See `loglet_test.go` for detailed examples and edge case handling.

## Helper Functions

### HasFields()

Utility function to check if a field map contains any non-empty keys:

```go
fields := map[string]any{"": "empty", "valid": "value"}
hasValid := internal.HasFields(fields)  // true - has "valid" key
```

This is used internally by `WithFields()` to avoid creating loglets for
maps that contain only empty keys.
