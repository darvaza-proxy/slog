# logrus

[![Go Reference][godoc-badge]][godoc]

[godoc]: https://pkg.go.dev/darvaza.org/slog/handlers/logrus
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/slog/handlers/logrus.svg

Bidirectional adapter between [Logrus](https://github.com/sirupsen/logrus) and
[darvaza.org/slog](https://github.com/darvaza-proxy/slog).

This package provides two-way integration:

- **logrus → slog**: Use a logrus logger as the backend for slog.Logger
  interface
- **slog → logrus**: Use logrus hooks to forward logs to any slog implementation

## Installation

```bash
go get darvaza.org/slog/handlers/logrus
```

## Quick Start

### Using logrus as slog backend (logrus → slog)

```go
import (
    "github.com/sirupsen/logrus"
    slogrus "darvaza.org/slog/handlers/logrus"
)

// Configure logrus
logrusLogger := logrus.New()
logrusLogger.SetLevel(logrus.DebugLevel)
logrusLogger.SetFormatter(&logrus.JSONFormatter{})

// Important: Disable ReportCaller to avoid duplicate caller info
logrusLogger.SetReportCaller(false)

// Create slog adapter
slogLogger := slogrus.New(logrusLogger)

// Use with slog interface
slogLogger.Info().
    WithField("user", "alice").
    WithField("action", "login").
    Print("User authenticated")
```

### Using slog as logrus backend (slog → logrus)

```go
import (
    "github.com/sirupsen/logrus"
    slogrus "darvaza.org/slog/handlers/logrus"
    "darvaza.org/slog/handlers/discard"
)

// Create any slog implementation
slogLogger := discard.New() // or any other slog handler

// Create a logrus logger that forwards to slog
logrusLogger := slogrus.NewLogrusLogger(slogLogger)

// Use with logrus interface
logrusLogger.WithFields(logrus.Fields{
    "component": "api",
    "version":   "1.0",
}).Info("Service started")

// Or configure an existing logrus logger to use slog
existingLogger := logrus.StandardLogger()
slogrus.SetupLogrusToSlog(existingLogger, slogLogger)
```

## Features

- **Bidirectional Integration**: Convert between logrus and slog in both
  directions.
- **Hook-based Design**: Uses logrus hooks for clean integration.
- **Full Compatibility**: Works with existing logrus code and configurations.
- **Preserves Fields**: Structured fields are maintained across conversions.
- **Stack Trace Support**: Via `WithStack()` when using logrus → slog.
- **Level Mapping**: Automatic conversion between slog and logrus levels.
- **Thread-Safe**: Immutable logger instances for concurrent use.

## Important Notes

### When using logrus → slog

- Disable `SetReportCaller()` to avoid duplicate method fields.
- `WithStack()` adds both method info and full call stack.
- Fatal and Panic levels maintain logrus behaviour (exit/panic).

### When using slog → logrus

- The hook disables logrus's native output to prevent double logging.
- Trace level is mapped to Debug (slog doesn't have Trace).
- All logrus features (formatters, other hooks) still work normally.

## Documentation

- [API Reference](https://pkg.go.dev/darvaza.org/slog/handlers/logrus)
- [slog Documentation](https://github.com/darvaza-proxy/slog)
- [Development Guide](AGENT.md)
- [Logrus Documentation](https://github.com/sirupsen/logrus)
