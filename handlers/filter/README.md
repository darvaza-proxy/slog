# `filter`

[![Go Reference][godoc-badge]][godoc]
[![codecov][codecov-badge]][codecov]

[godoc]: https://pkg.go.dev/darvaza.org/slog/handlers/filter
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/slog/handlers/filter.svg
[codecov]: https://codecov.io/gh/darvaza-proxy/slog
[codecov-badge]: https://codecov.io/github/darvaza-proxy/slog/graph/badge.svg?flag=filter

Filtering and transformation handler for
[darvaza.org/slog](https://github.com/darvaza-proxy/slog). Wraps any
slog.Logger to provide level-based filtering and custom log entry
transformation.

## Architecture

The filter handler implements a two-tier architecture:

- **Logger**: Factory and configuration layer that creates LogEntry instances.
- **LogEntry**: Working logger that handles actual logging operations with
  immutable field chains.

## Installation

```bash
go get darvaza.org/slog/handlers/filter
```

## Quick Start

```go
import (
    "strings"

    "darvaza.org/slog"
    "darvaza.org/slog/handlers/filter"
)

// Create a filter that only passes Info and above
baseLogger := getSomeLogger() // Any slog.Logger
filtered := filter.New(baseLogger, slog.Info)

// Debug and below are filtered out
filtered.Debug().Print("This won't appear")
filtered.Info().Print("This will appear")
```

### Proper Usage Patterns

```go
// CORRECT: Set level at the point of logging
logger.WithField("user", "alice").Info().Print("login successful")
logger.WithField("error", err).Error().Print("operation failed")

// INCORRECT: Don't change levels after setting
logger.Debug().Error().Print("confusing")  // Wrong: level set twice

// CORRECT: Branch for different levels
base := logger.WithField("request_id", "123")
base.Debug().Print("processing started")
base.Info().Print("processing complete")
base.Error().Print("processing failed")
```

## Advanced Configuration

### Custom Filter Configuration

```go
filterLogger := &filter.Logger{
    Parent:    baseLogger,
    Threshold: slog.Debug,
    FieldFilter: func(key string, val any) (string, any, bool) {
        // Redact sensitive fields
        if key == "password" {
            return key, "[REDACTED]", true
        }
        // Remove internal fields
        if strings.HasPrefix(key, "_") {
            return "", nil, false
        }
        return key, val, true
    },
    FieldsFilter: func(fields map[string]any) (map[string]any, bool) {
        // Add common prefix to all fields
        result := make(map[string]any, len(fields))
        for k, v := range fields {
            result["app_"+k] = v
        }
        return result, true
    },
    MessageFilter: func(msg string) (string, bool) {
        // Filter out health check logs
        if strings.Contains(msg, "/health") {
            return "", false // Drop this entry
        }
        return "[APP] " + msg, true // Add prefix and keep
    },
}
```

## Filter Hierarchy

The filter handler applies transformations using a specific hierarchy to
ensure predictable behaviour:

### For `WithField()` (Single Field Operations)

1. **FieldFilter** (most specific) - Designed for single field transformation.
2. **FieldsFilter** (fallback) - Handle as `{key: value}` → `map[string]any`.
3. **No filter** - Store field as-is in the loglet.

### For `WithFields()` (Multiple Field Operations)

1. **FieldsFilter** (most specific) - Designed for map transformation.
2. **FieldFilter** (fallback) - Apply to each field individually.
3. **No filter** - Store fields as-is in the loglet.

## Features

### Level-Based Filtering

- Configurable minimum level threshold.
- Parent logger consulted for final enablement decision.
- Special handling for Fatal/Panic levels in parentless configurations.

### Field Transformation

- **FieldFilter**: Transform individual fields with key/value pairs.
- **FieldsFilter**: Transform entire field maps for bulk operations.
- **Rejection**: Return `false` to drop fields entirely.

### Message Transformation

- **MessageFilter**: Modify or filter log messages before output.
- **Conditional logging**: Return `false` to drop entire log entries.

### Design Principles

- **Immutable logger instances**: Ensure thread-safe field management.
- **Lazy evaluation**: Fields collected at print time, not attachment time.
- **Parent delegation**: Final output handled by wrapped parent logger.
- **Filter application**: Applied at field attachment time for efficiency.
- **Level-less entries**: LogEntry instances without a level can accumulate
  fields.
- **Level is terminal**: Once a level is set (via `.Debug()`, `.Info()`,
  etc.), it should not be changed. The level is intended to be set at the
  point of logging, not as an intermediate transformation.

## Special Behaviours

### Parentless Loggers

Created with `filter.NewNoop()` or `filter.New(nil, threshold)`:

- Only Fatal and Panic levels are enabled.
- Bypass parent delegation and use standard library logging.
- Fields are collected but not forwarded (termination-only behaviour).

### Filter Chaining

Multiple filter loggers can be chained together:

```go
filter1 := filter.New(baseLogger, slog.Error)  // Only Error and above
filter2 := filter.New(filter1, slog.Warn)     // Further restricted
filter3 := filter.New(filter2, slog.Info)     // Most restrictive wins
```

The most restrictive threshold in the chain determines final behaviour.

## Performance Considerations

- **Efficient field collection**: Only occurs at print time when needed.
- **Conservative enablement**: May collect fields for ultimately disabled
  entries.
- **Immutable design**: Safe for concurrent use without synchronisation.
- **Filter overhead**: Transformation functions called during field attachment.

## Examples

### Security Field Filtering

```go
secureLogger := &filter.Logger{
    Parent:    baseLogger,
    Threshold: slog.Info,
    FieldFilter: func(key string, val any) (string, any, bool) {
        // Redact sensitive data
        sensitive := []string{"password", "token", "secret", "key"}
        for _, s := range sensitive {
            if strings.Contains(strings.ToLower(key), s) {
                return key, "[REDACTED]", true
            }
        }
        return key, val, true
    },
}
```

### Development vs Production Filtering

```go
func createLogger(isDevelopment bool) slog.Logger {
    baseLogger := getSomeLogger()

    if isDevelopment {
        // Development: Allow all levels, no filtering
        return filter.New(baseLogger, slog.Debug)
    }

    // Production: Filter out debug, redact sensitive fields
    return &filter.Logger{
        Parent:    baseLogger,
        Threshold: slog.Info,
        FieldFilter: func(key string, val any) (string, any, bool) {
            if strings.HasPrefix(key, "debug_") {
                return "", nil, false // Remove debug fields
            }
            return key, val, true
        },
    }
}
```

## Documentation

- [API Reference](https://pkg.go.dev/darvaza.org/slog/handlers/filter)
- [slog Documentation](https://github.com/darvaza-proxy/slog)
- [Development Guide](../../AGENT.md)
