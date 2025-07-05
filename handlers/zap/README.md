# zap

[![Go Reference][godoc-badge]][godoc]

[godoc]: https://pkg.go.dev/darvaza.org/slog/handlers/zap
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/slog/handlers/zap.svg

[Uber's zap](https://github.com/uber-go/zap) adapter for
[darvaza.org/slog](https://github.com/darvaza-proxy/slog).
Wraps a `*zap.SugaredLogger` to implement the slog.Logger interface, enabling
high-performance structured logging with zap as the backend.

## Installation

```bash
go get darvaza.org/slog/handlers/zap
```

## Quick Start

```go
import (
    "go.uber.org/zap"
    slogzap "darvaza.org/slog/handlers/zap"
)

// Create zap logger (production config for performance)
zapConfig := zap.NewProductionConfig()
zapLogger, err := zapConfig.Build()
if err != nil {
    panic(err)
}
sugar := zapLogger.Sugar()

// Create slog adapter
slogLogger := slogzap.New(sugar)

// Use with slog interface
slogLogger.Info().
    WithField("latency_ms", 42).
    WithField("status", 200).
    WithField("method", "GET").
    Print("Request completed")

// Development-friendly console output
devLogger, err := zap.NewDevelopment()
if err != nil {
    log.Fatalf("cannot create dev zap logger: %v", err)
}
defer devLogger.Sync()           // flush any buffered logs
slogDev := slogzap.New(devLogger.Sugar())
```

## Features

- Leverages zap's high-performance, zero-allocation design
- Natural fit for structured logging with fields
- Supports both production and development configurations
- Preserves zap's excellent performance characteristics
- Immutable logger instances ensure thread-safe field management

## Performance Tips

- Use `zap.NewProductionConfig()` for best performance
- Avoid console encoders in production
- Pre-compute expensive field values
- Consider using `zap.Logger` directly for hot paths

## Documentation

- [API Reference](https://pkg.go.dev/darvaza.org/slog/handlers/zap)
- [slog Documentation](https://github.com/darvaza-proxy/slog)
- [Development Guide](AGENT.md)
- [Zap Documentation](https://pkg.go.dev/go.uber.org/zap)
