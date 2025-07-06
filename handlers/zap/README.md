# `zap`

[![Go Reference][godoc-badge]][godoc]
[![codecov][codecov-badge]][codecov]

[godoc]: https://pkg.go.dev/darvaza.org/slog/handlers/zap
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/slog/handlers/zap.svg
[codecov]: https://codecov.io/gh/darvaza-proxy/slog
[codecov-badge]: https://codecov.io/github/darvaza-proxy/slog/graph/badge.svg?flag=zap

[Uber's zap](https://github.com/uber-go/zap) adapter for
[darvaza.org/slog](https://github.com/darvaza-proxy/slog).
Uses a `*zap.Config` to create a slog.Logger interface, enabling
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

// Create zap config (production config for performance)
zapConfig := zap.NewProductionConfig()

// Create slog adapter
slogLogger, err := slogzap.New(&zapConfig)
if err != nil {
    panic(err)
}

// Or with zap options (e.g., for testing with a custom core)
slogLogger, err := slogzap.New(&zapConfig,
    zap.WrapCore(func(core zapcore.Core) zapcore.Core {
        // Replace or wrap the core as needed
        return core
    }),
)
if err != nil {
    panic(err)
}

// Use with slog interface
slogLogger.Info().
    WithField("latency_ms", 42).
    WithField("status", 200).
    WithField("method", "GET").
    Print("Request completed")

// Development-friendly console output
devConfig := zap.NewDevelopmentConfig()
slogDev, err := slogzap.New(&devConfig)
if err != nil {
    log.Fatalf("cannot create dev slog logger: %v", err)
}
```

## Breaking Changes

**v0.7.0**: The `New()` function now returns `(slog.Logger, error)` instead of
just `slog.Logger` to properly handle configuration build errors. Additionally,
it now accepts variadic `zap.Option` parameters for customizing the logger.

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
