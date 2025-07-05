# filter

[![Go Reference][godoc-badge]][godoc]

[godoc]: https://pkg.go.dev/darvaza.org/slog/handlers/filter
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/slog/handlers/filter.svg

Filtering and transformation handler for
[darvaza.org/slog](https://github.com/darvaza-proxy/slog).
Wraps any slog.Logger to provide level-based filtering and custom log entry
transformation.

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

// Custom filtering with field and message transformations
filterLogger := &filter.Logger{
    Parent:    baseLogger,
    Threshold: slog.Debug,
    FieldFilter: func(key string, val any) (string, any, bool) {
        // Add prefix to all field keys
        return "app_" + key, val, true
    },
    MessageFilter: func(msg string) (string, bool) {
        // Filter out health check logs
        if strings.Contains(msg, "/health") {
            return "", false // drop this entry
        }
        return msg, true // keep this entry
    },
}
```

## Features

- Level-based filtering (minimum level threshold)
- Custom transformation functions
- Field enrichment and modification
- Conditional log filtering
- Composable with any slog.Logger
- Immutable logger instances ensure thread-safe field management

## Documentation

- [API Reference](https://pkg.go.dev/darvaza.org/slog/handlers/filter)
- [slog Documentation](https://github.com/darvaza-proxy/slog)
- [Development Guide](AGENT.md)
