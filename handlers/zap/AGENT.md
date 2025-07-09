# AGENT.md

This file provides guidance to AI agents when working with code in this
repository. For developers and general project information, please refer to
[README.md](README.md) first.

## Handler Overview

`darvaza.org/slog/handlers/zap` is an adapter that implements the `slog.Logger`
interface using [Uber's zap](https://github.com/uber-go/zap) as the backend
high-performance logging library.

## Key Features

- **High performance**: Leverages zap's zero-allocation design.
- **Structured logging**: Natural fit for slog's field-based approach.
- **Level mapping**: Maps slog levels to zap levels appropriately.
- **SugaredLogger support**: Uses zap.SugaredLogger for flexibility.

## Implementation Notes

- Uses `*zap.SugaredLogger` internally for easier field handling.
- Fatal and Panic levels trigger zap's Fatal/Panic behaviour.
- Fields are converted to zap's field types for efficiency.
- Defers to zap for all formatting and output handling.

## Usage Patterns

```go
import (
    "time"

    "go.uber.org/zap"
    slogzap "darvaza.org/slog/handlers/zap"
)

// Create zap logger
zapConfig := zap.NewProductionConfig()
zapLogger, _ := zapConfig.Build()
sugar := zapLogger.Sugar()

// Wrap with slog interface
slogLogger := slogzap.New(sugar)

// Use with high-performance logging
slogLogger.Info().
    WithField("latency", time.Since(start)).
    WithField("status", 200).
    Print("Request completed")
```

## Level Mapping

| slog Level | zap Level |
|------------|-----------|
| Debug      | Debug     |
| Info       | Info      |
| Warn       | Warn      |
| Error      | Error     |
| Fatal      | Fatal     |
| Panic      | Panic     |

## Performance Considerations

- Use zap's production configuration for best performance.
- Consider using zap.Logger directly for hot paths.
- Be aware of allocation costs when using complex field values.
- Benchmark critical paths to ensure performance meets requirements.

## Testing Considerations

- Use zap's observer for testing log output.
- Test with both development and production configurations.
- Verify level filtering works correctly.
- Ensure fields are properly typed for zap.

## Development Notes

- Preserve zap's performance characteristics.
- Avoid unnecessary allocations in the adapter layer.
- Document any performance trade-offs.
- Keep up with zap API changes and optimizations.

For general development guidelines, see the main
[slog AGENT.md](../../AGENT.md).
