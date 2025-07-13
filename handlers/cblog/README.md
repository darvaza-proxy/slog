# cblog

[![Go Reference][godoc-badge]][godoc]
[![codecov][codecov-badge]][codecov]

[godoc]: https://pkg.go.dev/darvaza.org/slog/handlers/cblog
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/slog/handlers/cblog.svg
[codecov]: https://codecov.io/gh/darvaza-proxy/slog
[codecov-badge]: https://codecov.io/github/darvaza-proxy/slog/graph/badge.svg?flag=cblog

Channel-based logging handler for
[darvaza.org/slog](https://github.com/darvaza-proxy/slog).
Sends log entries through Go channels for custom processing, buffering, or
asynchronous handling.

## Installation

```bash
go get darvaza.org/slog/handlers/cblog
```

## Quick Start

```go
import (
    "fmt"

    "darvaza.org/slog/handlers/cblog"
)

// Create a buffered channel for log entries
ch := make(chan cblog.LogMsg, 100)
logger := cblog.New(ch)

// Process entries in a background goroutine
go func() {
    for entry := range ch {
        // Custom processing logic
        fmt.Printf("[%s] %s\n",
            entry.Level,
            entry.Message)

        // Print fields if any
        for k, v := range entry.Fields {
            fmt.Printf("  %s: %v\n", k, v)
        }
    }
}()

// Use like any slog.Logger
logger.Info().WithField("component", "api").Print("Server started")
```

## Features

- Channel-based delivery for flexible log processing
- Configurable buffering and non-blocking modes
- Worker management for concurrent processing
- Suitable for custom log aggregation or filtering
- Immutable logger instances ensure thread-safe field management

## Documentation

- [API Reference](https://pkg.go.dev/darvaza.org/slog/handlers/cblog)
- [slog Documentation](https://github.com/darvaza-proxy/slog)
- [Development Guide](AGENT.md)
