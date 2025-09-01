# `filter`

[![Go Reference][godoc-badge]][godoc]
[![Go Report Card][goreportcard-badge]][goreportcard-link]
[![codecov][codecov-badge]][codecov]
[![Socket Badge][socket-badge]][socket-link]

[godoc]: https://pkg.go.dev/darvaza.org/slog/handlers/filter
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/slog/handlers/filter.svg
[goreportcard-badge]: https://goreportcard.com/badge/darvaza.org/slog/handlers/filter
[goreportcard-link]: https://goreportcard.com/report/darvaza.org/slog/handlers/filter
[codecov]: https://codecov.io/gh/darvaza-proxy/slog?flag=filter
[codecov-badge]: https://codecov.io/github/darvaza-proxy/slog/graph/badge.svg?flag=filter
[socket-badge]: https://socket.dev/api/badge/go/package/darvaza.org/slog/handlers/filter
[socket-link]: https://socket.dev/go/package/darvaza.org/slog/handlers/filter

Filtering and transformation handler for
[darvaza.org/slog](https://github.com/darvaza-proxy/slog). Wraps any
slog.Logger to provide level-based filtering and custom log entry
transformation.

## Architecture

The filter handler implements a two-tier architecture:

- **Logger**: Factory and configuration layer that creates LogEntry instances.
  Print methods are no-ops. WithField/WithStack create LogEntry instances.
- **LogEntry**: Working logger that handles actual logging operations with
  immutable field chains. Only logs when a level is set and within threshold.

## Installation

```bash
go get darvaza.org/slog/handlers/filter
```

## Key Behaviors

### Logger vs LogEntry

- **Logger.Print()/Printf()/Println()**: No-ops that don't log anything.
- **Logger.WithField()/WithStack()**: Create LogEntry with fields
  (parentless don't collect).
- **Logger.Debug()/Info()/etc**: Create LogEntry with specified level.

- **LogEntry without level**: NOT enabled, collects fields speculatively.
- **LogEntry with level**: Enabled if level ≤ threshold, fields only
  collected when enabled.

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
- **Selective field collection**: Fields only collected when potentially
  enabled.
- **Parent delegation**: Final output handled by wrapped parent logger.
- **Filter application**: Applied at field attachment time for efficiency.
- **Level-less entries**: Can accumulate fields speculatively but are NOT
  enabled for logging.
- **Level is terminal**: Once a level is set (via `.Debug()`, `.Info()`,
  etc.), it should not be changed. The level is intended to be set at the
  point of logging, not as an intermediate transformation.
- **Logger.Print() methods**: No-ops that do not log without an explicit
  level.

## Special Behaviours

### Parentless Loggers

Created with `filter.NewNoop()` or `filter.New(nil, threshold)`:

- Only Fatal and Panic levels are enabled (for termination).
- Bypass parent delegation and use standard library logging.
- Fields are NOT collected (nowhere to forward them).
- Used primarily for termination-only behaviour.

### Filter Chaining

Multiple filter loggers can be chained together:

```go
filter1 := filter.New(baseLogger, slog.Error)  // Only Error and above
filter2 := filter.New(filter1, slog.Warn)     // Further restricted
filter3 := filter.New(filter2, slog.Info)     // Most restrictive wins
```

The most restrictive threshold in the chain determines final behaviour.

## Performance Considerations

- **Efficient field collection**: Only occurs when entry is potentially
  enabled.
- **Optimised disabled entries**: Fields NOT collected when level exceeds
  threshold.
- **Immutable design**: Safe for concurrent use without synchronisation.
- **Filter overhead**: Transformation functions called during field
  attachment.
- **Speculative collection**: Level-less entries collect fields that will be
  used when a level is eventually set.

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
