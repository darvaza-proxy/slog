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
- Provides `NewWithCallback` method to create derived loggers with hooks.
- Includes `NewNoop` for creating no-op loggers (useful for testing).

## Usage Patterns

```go
import (
    "time"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    slogzap "darvaza.org/slog/handlers/zap"
)

// Create zap config
zapConfig := zap.NewProductionConfig()

// Wrap with slog interface
slogLogger, err := slogzap.New(&zapConfig)
if err != nil {
    // handle error
}

// Or with custom zap options
slogLogger, err := slogzap.New(&zapConfig,
    zap.Hooks(func(entry zapcore.Entry) error {
        // Custom hook for processing log entries
        return nil
    }),
    zap.Fields(zap.String("service", "api")), // Global fields
)
if err != nil {
    // handle error
}

// Use with high-performance logging
slogLogger.Info().
    WithField("latency", time.Since(start)).
    WithField("status", 200).
    Print("Request completed")

// Create a derived logger with a hook
zapLogger := slogLogger.(*slogzap.Logger)
hookedLogger := zapLogger.NewWithCallback(func(entry zapcore.Entry) error {
    // Process log entries (e.g., send metrics, filter, etc.)
    fmt.Printf("Log entry: %s at level %s\n", entry.Message, entry.Level)
    return nil
})
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
