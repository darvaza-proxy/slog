# logrus

[![Go Reference][godoc-badge]][godoc]

[godoc]: https://pkg.go.dev/darvaza.org/slog/handlers/logrus
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/slog/handlers/logrus.svg

[Logrus](https://github.com/sirupsen/logrus) adapter for
[darvaza.org/slog](https://github.com/darvaza-proxy/slog).
Wraps a `*logrus.Logger` to implement the slog.Logger interface, enabling use of
logrus as a backend for slog-based applications.

## Installation

```bash
go get darvaza.org/slog/handlers/logrus
```

## Quick Start

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

## Features

- Full compatibility with logrus features (hooks, formatters, outputs)
- Preserves structured fields through the adapter
- Stack trace support via `WithStack()`
- Correct level mapping between slog and logrus
- Immutable logger instances ensure thread-safe field management

## Important Notes

- Disable `SetReportCaller()` to avoid duplicate method fields
- `WithStack()` adds both method info and full call stack
- Fatal and Panic levels maintain logrus behavior (exit/panic)

## Documentation

- [API Reference](https://pkg.go.dev/darvaza.org/slog/handlers/logrus)
- [slog Documentation](https://github.com/darvaza-proxy/slog)
- [Development Guide](AGENT.md)
- [Logrus Documentation](https://github.com/sirupsen/logrus)
