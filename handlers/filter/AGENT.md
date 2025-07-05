# AGENT.md

This file provides guidance to AI agents when working with code in this
repository. For developers and general project information, please refer to
[README.md](README.md) first.

## Handler Overview

`darvaza.org/slog/handlers/filter` is a middleware logging handler that filters
and transforms log entries before passing them to another `slog.Logger`. It
enables level-based filtering and custom entry transformation.

## Key Features

- **Level filtering**: Only pass entries meeting minimum level requirements.
- **Entry transformation**: Modify log entries before forwarding.
- **Chaining support**: Can wrap any `slog.Logger` implementation.
- **Custom filters**: Support for user-defined filtering logic.

## Architecture

The handler consists of:

1. **Filter implementation** (filter.go): Main filter logic and logger wrapping.
2. **Entry type** (entry.go): Represents a log entry that can be modified
   before forwarding.

## Usage Patterns

```go
import (
    "darvaza.org/slog"
    "darvaza.org/slog/handlers/filter"
    slogzap "darvaza.org/slog/handlers/zap"
)

// Create a filter that only passes Info and above
baseLogger := slogzap.New(zapLogger.Sugar())
filtered := filter.New(baseLogger, slog.Info)

// Debug won't be logged, Info will
filtered.Debug().Print("This won't appear")
filtered.Info().Print("This will appear")

// With custom transformation
transformer := filter.NewWithTransform(baseLogger, slog.Debug,
    func(e *filter.Entry) bool {
        // Add common fields
        e.Fields["app"] = "myapp"
        e.Fields["version"] = "1.0"
        return true // forward the entry
    })
```

## Filter Functions

Custom filter functions can:

- Modify the entry's level, message, or fields.
- Return false to drop the entry entirely.
- Add, remove, or transform fields.
- Change the message format.

## Design Principles

1. **Composability**: Filters can be chained for complex behaviors.
2. **Efficiency**: Minimize overhead for filtered-out entries.
3. **Transparency**: Preserve original logger behavior when not filtering.
4. **Flexibility**: Support both simple level filtering and complex
   transformations.

## Testing Considerations

- Test both filtering (dropping) and forwarding paths.
- Verify field modifications are applied correctly.
- Ensure filter functions are called with correct entries.
- Test edge cases like nil loggers or filters.

## Common Use Cases

1. **Environment-based filtering**: Different log levels for dev/prod.
2. **Field enrichment**: Add common fields to all log entries.
3. **Privacy filtering**: Remove sensitive data from logs.
4. **Conditional logging**: Complex rules for what to log.

## Development Notes

- Filter functions should be efficient as they're called for every log entry.
- Maintain immutability where possible to avoid race conditions.
- Document any performance implications of complex filters.
- Ensure the filtered logger remains thread-safe.

For general development guidelines, see the main
[slog AGENT.md](../../AGENT.md).
