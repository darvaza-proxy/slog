# zerolog

[![Go Reference][godoc-badge]][godoc]
[![codecov][codecov-badge]][codecov]

[godoc]: https://pkg.go.dev/darvaza.org/slog/handlers/zerolog
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/slog/handlers/zerolog.svg
[codecov]: https://codecov.io/gh/darvaza-proxy/slog
[codecov-badge]: https://codecov.io/github/darvaza-proxy/slog/graph/badge.svg?flag=zerolog

[Zerolog](https://github.com/rs/zerolog) adapter for
[darvaza.org/slog](https://github.com/darvaza-proxy/slog).
Wraps a `*zerolog.Logger` to implement the slog.Logger interface, providing
blazing-fast, zero-allocation JSON logging.

## Installation

```bash
go get darvaza.org/slog/handlers/zerolog
```

## Quick Start

```go
import (
    "os"
    "github.com/rs/zerolog"
    slogzerolog "darvaza.org/slog/handlers/zerolog"
)

// Create zerolog logger
zlogger := zerolog.New(os.Stdout).With().Timestamp().Logger()

// For pretty console output during development
// console := zerolog.ConsoleWriter{Out: os.Stdout}
// zlogger := zerolog.New(console).With().Timestamp().Logger()

// Create slog adapter
slogLogger := slogzerolog.New(&zlogger)

// Use with slog interface
slogLogger.Info().
    WithField("service", "api").
    WithField("version", "1.0.0").
    WithField("request_id", "abc-123").
    Print("Service started")
```

## Features

- Zero-allocation JSON encoding for maximum performance
- One of the fastest structured loggers for Go
- Clean, minimal JSON output
- Console writer available for development
- Excellent for high-throughput applications
- Immutable logger instances ensure thread-safe field management

## Output Example

```json
{
  "level":"info",
  "service":"api",
  "version":"1.0.0",
  "request_id":"abc-123",
  "time":"2023-01-01T12:00:00Z",
  "message":"Service started"
}
```

## Performance Notes

- Use JSON output (not ConsoleWriter) in production
- Zerolog is optimized for JSON encoding speed
- Fields are encoded directly without intermediate allocations
- Ideal for services with high log volumes

## Documentation

- [API Reference](https://pkg.go.dev/darvaza.org/slog/handlers/zerolog)
- [slog Documentation](https://github.com/darvaza-proxy/slog)
- [Development Guide](AGENT.md)
- [Zerolog Documentation](https://github.com/rs/zerolog)
