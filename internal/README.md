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
// For handlers that need to modify fields
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
- **With* methods**: O(1) for field/level/stack operations

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
