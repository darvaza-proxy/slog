# darvaza.org/slog/handlers/logr

`darvaza.org/slog/handlers/logr` provides bidirectional adapters between
`slog.Logger` and the [go-logr/logr][logr-github] logging interface.

This package enables:

- Using any logr implementation as a slog.Logger backend
- Using any slog.Logger as a logr.LogSink implementation

## Installation

```bash
go get darvaza.org/slog/handlers/logr
```

## Usage

### Using logr as slog.Logger

```go
package main

import (
    "github.com/go-logr/logr"
    "github.com/go-logr/zapr"
    "go.uber.org/zap"

    sloglogr "darvaza.org/slog/handlers/logr"
)

func main() {
    // Create a logr logger (using zap as example)
    zapLogger, _ := zap.NewProduction()
    logrLogger := zapr.NewLogger(zapLogger)

    // Wrap it as slog.Logger
    logger := sloglogr.New(logrLogger)

    // Use as normal slog.Logger
    logger.Info().
        WithField("component", "main").
        WithField("version", "1.0").
        Print("Application started")
}
```

### Using slog.Logger as logr

```go
package main

import (
    "darvaza.org/slog"
    slogzap "darvaza.org/slog/handlers/zap"
    sloglogr "darvaza.org/slog/handlers/logr"
)

func main() {
    // Create any slog.Logger
    slogLogger := slogzap.New(nil)

    // Convert to logr.Logger
    logrLogger := sloglogr.NewLogr(slogLogger)

    // Use as normal logr.Logger
    logrLogger.Info("starting server", "port", 8080, "env", "production")
    logrLogger.Error(err, "failed to connect", "host", "db.example.com")
}
```

## Level Mapping

### slog to logr (when using logr as backend)

| slog Level | logr Method          |
|------------|----------------------|
| Debug      | V(1).Info()          |
| Info       | V(0).Info()          |
| Warn       | V(0).Info()          |
| Error      | Error()              |
| Fatal      | Error() + os.Exit(1) |
| Panic      | Error() + panic      |

Note: logr doesn't have a warn level, so Warn is mapped to V(0) like Info.

### logr to slog (when using slog as backend)

| logr V-Level | slog Level |
|--------------|------------|
| V(1+)        | Debug      |
| V(0)         | Info       |
| Error()      | Error      |

## Features

- **Bidirectional Conversion**: Convert between slog and logr in both directions
- **Field Preservation**: Structured fields are preserved across conversions
- **Level Mapping**: Intelligent mapping between slog levels and logr V-levels
- **Immutable Loggers**: Both adapters follow the immutable logger pattern
- **Stack Traces**: Call stacks are preserved when using WithStack()
- **Named Loggers**: logr's WithName() is preserved as a "logger" field

## Implementation Details

The package provides two main types:

1. **Logger**: Adapts a logr.Logger to implement slog.Logger
   - Uses `internal.Loglet` for field chain management
   - Converts slog levels to appropriate logr calls
   - Maintains structured logging fields

2. **Sink**: Implements logr.LogSink using a slog.Logger
   - Supports all logr.LogSink methods
   - Implements optional interfaces (CallDepthLogSink, CallStackHelperLogSink)
   - Uses `core.SortedKeys` for consistent field ordering

## See Also

- [darvaza.org/slog][slog-pkg] - Main slog interface
- [go-logr/logr][logr-github] - The logr interface
- [logr implementations][logr-impl] - List of logr backends

[slog-pkg]: https://pkg.go.dev/darvaza.org/slog
[logr-github]: https://github.com/go-logr/logr
[logr-impl]: https://github.com/go-logr/logr#implementations
