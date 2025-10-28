# AGENTS.md

This file provides guidance to AI agents when working with code in this
repository. For developers and general project information, please refer to
[README.md](README.md) first.

## Handler Overview

`darvaza.org/slog/handlers/zap` provides bidirectional integration between
[Uber's zap](https://github.com/uber-go/zap) and the slog interface:

1. **zap → slog**: Wraps a zap logger to implement `slog.Logger`
2. **slog → zap**: Implements `zapcore.Core` to create zap loggers backed by
   slog

## Key Features

- **Bidirectional conversion**: Use either API with either backend
- **High performance**: Leverages zap's zero-allocation design
- **Full compatibility**: Supports all features of both logging systems
- **Structured logging**: Natural fit for both APIs' field-based approach
- **Level mapping**: Bidirectional mapping between slog and zap levels

## Implementation Notes

### zap → slog (Logger adapter)

- Wraps `*zap.Logger` with configurable setup.
- Fatal and Panic levels trigger zap's Fatal/Panic behaviour.
- Fields are converted to zap's field types for efficiency.
- Defers to zap for all formatting and output handling.
- Provides `NewWithCallback` method to create derived loggers with hooks.
- Includes `NewNoop` for creating no-op loggers (useful for testing).

### slog → zap (Core implementation)

- Implements `zapcore.Core` interface.
- Converts zap fields using `MapObjectEncoder`.
- Maintains field accumulation through `With()` calls.
- Preserves zap's expectations for Fatal/Panic behaviour.

## Usage Patterns

### Using zap as slog

```go
import (
    "fmt"
    "time"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "darvaza.org/slog"
    slogzap "darvaza.org/slog/handlers/zap"
)

// Create slog adapter from zap config
zapConfig := slogzap.NewDefaultConfig()
slogLogger, err := slogzap.New(zapConfig)
if err != nil {
    // handle error
    return err
}

// Or with custom zap options
slogLogger, err := slogzap.New(zapConfig,
    zap.Hooks(func(entry zapcore.Entry) error {
        // Custom hook for processing log entries
        return nil
    }),
    zap.Fields(zap.String("service", "api")), // Global fields
)
if err != nil {
    // handle error
}

// Use with slog interface
start := time.Now()
// ... perform some work ...
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

### Using slog as zap

```go
import (
    "time"

    "go.uber.org/zap"
    "darvaza.org/slog"
    slogzap "darvaza.org/slog/handlers/zap"
    "darvaza.org/slog/handlers/filter"
)

// Start with any slog logger
baseLogger := getSlogLogger()
filteredLogger := filter.New(baseLogger, filter.MinLevel(slog.Info))

// Create zap logger backed by slog
zapLogger := slogzap.NewZapLogger(filteredLogger)

// Use with zap interface
start := time.Now()
// ... perform some work ...
zapLogger.Info("Request completed",
    zap.Duration("latency", time.Since(start)),
    zap.Int("status", 200),
)
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
[slog AGENTS.md](../../AGENTS.md).
