# `discard`

[![Go Reference][godoc-badge]][godoc]

[godoc]: https://pkg.go.dev/darvaza.org/slog/handlers/discard
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/slog/handlers/discard.svg

No-op logging handler for
[darvaza.org/slog](https://github.com/darvaza-proxy/slog).
Discards all log entries except Fatal and Panic levels, perfect for testing and
optional logging scenarios.

## Installation

```bash
go get darvaza.org/slog/handlers/discard
```

## Quick Start

```go
// Create a discard logger
logger := discard.New()

// Use anywhere a logger is needed
service := NewService(logger) // Won't produce any output

// Safe to use - never returns nil
logger.Debug().WithField("data", complexObject).Print("This is discarded")
logger.Info().Print("This too")

// Fatal and Panic still work as expected
logger.Fatal().Print("This will exit") // Calls log.Fatal()
logger.Panic().Print("This will panic") // Calls log.Panic()
```

## Features

- Zero overhead for disabled log levels
- Immutable logger instances for safe concurrent use
- Always returns a valid logger (never nil)
- Fatal and Panic levels still trigger appropriate exits

## Use Cases

- **Testing**: Silence logs during test execution
- **Optional logging**: When a component doesn't require logging
- **Benchmarking**: Remove logging overhead from performance tests
- **Default logger**: Safe placeholder before configuration

## Documentation

- [API Reference](https://pkg.go.dev/darvaza.org/slog/handlers/discard)
- [slog Documentation](https://github.com/darvaza-proxy/slog)
- [Development Guide](AGENT.md)
