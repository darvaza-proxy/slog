# `zap`

[![Go Reference][godoc-badge]][godoc]
[![codecov][codecov-badge]][codecov]

[godoc]: https://pkg.go.dev/darvaza.org/slog/handlers/zap
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/slog/handlers/zap.svg
[codecov]: https://codecov.io/gh/darvaza-proxy/slog
[codecov-badge]: https://codecov.io/github/darvaza-proxy/slog/graph/badge.svg?flag=zap

Bidirectional adapter between
[Uber's zap][zap-github] and
[darvaza.org/slog][slog-github].

This package provides two-way integration:

- **zap → slog**: Use a zap logger as the backend for `slog.Logger`
  interface.
- **slog → zap**: Implement `zapcore.Core` to create zap loggers backed by
  any slog implementation.

## Installation

```bash
go get darvaza.org/slog/handlers/zap
```

## Quick Start

### Using zap as slog backend (zap → slog)

```go
import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    slogzap "darvaza.org/slog/handlers/zap"
)

// Create a zap config
zapConfig := slogzap.NewDefaultConfig() // or zap.NewProductionConfig()

// Create slog adapter
slogLogger, err := slogzap.New(zapConfig)
if err != nil {
    panic(err)
}

// Or with zap options (e.g., for testing with a custom core)
slogLogger, err := slogzap.New(zapConfig,
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
```

### Using slog as zap backend (slog → zap)

```go
import (
    "darvaza.org/slog"
    "darvaza.org/slog/handlers/filter"
    slogzap "darvaza.org/slog/handlers/zap"
    "go.uber.org/zap"
)

// Start with any slog implementation
baseLogger := getSlogLogger() // your existing slog logger
filtered := filter.New(baseLogger, filter.MinLevel(slog.Info))

// Create a zap logger backed by slog
zapLogger := slogzap.NewZapLogger(filtered)

// Use with zap interface
zapLogger.Info("Request completed",
    zap.Int("latency_ms", 42),
    zap.Int("status", 200),
    zap.String("method", "GET"),
)

// Or create with custom options
core := slogzap.NewCore(filtered, zap.InfoLevel)
zapLogger = zap.New(core,
    zap.AddCaller(),
    zap.AddStacktrace(zap.ErrorLevel),
)
```

## Breaking Changes

**v0.7.0**: The `New()` function now returns `(slog.Logger, error)` instead of
just `slog.Logger` to properly handle configuration build errors. Additionally,
it now accepts variadic `zap.Option` parameters for customizing the logger.

## Features

- **Bidirectional Integration**: Convert between zap and slog in both
  directions.
- **High Performance**: Leverages zap's zero-allocation design.
- **Full Compatibility**: Supports all features of both logging systems.
- **Flexible Architecture**: Use zap's API with any slog backend, or vice
  versa.
- **Thread-Safe**: Immutable logger instances with safe field management.
- **Production Ready**: Battle-tested with comprehensive test coverage.

## Use Cases

### When to use zap → slog

- You have existing zap infrastructure but want to use slog's interface
- You need to integrate with libraries that expect slog.Logger
- You want to apply slog middleware (filters, transformers) to zap output

### When to use slog → zap

- You have an slog backend but need to use libraries that require zap
- You want zap's rich API while writing to a custom slog implementation
- You're migrating between logging systems and need to support both APIs

## Level Mapping

| slog Level | zap Level    | Behaviour |
|------------|--------------|-----------|
| Debug      | Debug        | Standard mapping |
| Info       | Info         | Standard mapping |
| Warn       | Warn         | Standard mapping |
| Error      | Error        | Standard mapping |
| Fatal      | Fatal        | Calls os.Exit(1) |
| Panic      | Panic/DPanic | Calls panic() |

**Note:** DPanic (Development Panic) maps to slog.Panic and triggers panic()
behaviour.

## Performance Tips

- Use `zap.NewProductionConfig()` for best performance.
- Avoid console encoders in production.
- Pre-compute expensive field values.
- Consider using `zap.Logger` directly for hot paths.

## Documentation

- [API Reference][godoc]
- [slog Documentation][slog-github]
- [Development Guide](AGENTS.md)
- [Zap Documentation][zap-docs]

[zap-github]: https://github.com/uber-go/zap
[slog-github]: https://github.com/darvaza-proxy/slog
[zap-docs]: https://pkg.go.dev/go.uber.org/zap
