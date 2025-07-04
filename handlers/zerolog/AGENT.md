# AGENT.md

This file provides guidance to AI agents when working with code in this
repository. For developers and general project information, please refer to
[README.md](README.md) first.

## Handler Overview

`darvaza.org/slog/handlers/zerolog` is an adapter that implements the
`slog.Logger` interface using [rs/zerolog](https://github.com/rs/zerolog) as
the backend JSON logging library.

## Key Features

- **Zero allocation**: Leverages zerolog's zero-allocation JSON encoding.
- **Blazing fast**: One of the fastest JSON loggers for Go.
- **Clean JSON output**: Produces well-structured JSON logs.
- **Context pattern**: Uses zerolog's context pattern for field accumulation.

## Implementation Notes

- Uses `zerolog.Context` for maintaining fields between calls.
- Converts slog levels to zerolog levels.
- Fatal level calls `os.Exit(1)` after logging (zerolog behavior).
- Panic level triggers a panic after logging.
- Maintains zerolog's performance characteristics.

## Usage Patterns

```go
import (
    "os"

    "github.com/rs/zerolog"
    slogzerolog "darvaza.org/slog/handlers/zerolog"
)

// Create zerolog logger
output := zerolog.ConsoleWriter{Out: os.Stdout}
zLogger := zerolog.New(output).With().Timestamp().Logger()

// Wrap with slog interface
slogLogger := slogzerolog.New(&zLogger)

// Use for structured JSON logging
slogLogger.Info().
    WithField("service", "api").
    WithField("request_id", requestID).
    Print("Processing request")
```

## Level Mapping

| slog Level | zerolog Level |
|------------|---------------|
| Debug      | Debug         |
| Info       | Info          |
| Warn       | Warn          |
| Error      | Error         |
| Fatal      | Fatal         |
| Panic      | Panic         |

## Performance Considerations

- Zerolog is optimized for JSON output performance.
- Avoid console writer in production for best performance.
- Use zerolog's context pattern efficiently.
- Consider batching for high-frequency logging.

## Testing Considerations

- Use bytes.Buffer as output for testing.
- Parse JSON output to verify fields.
- Test level filtering behavior.
- Verify Fatal doesn't exit in tests (use subprocess).

## Development Notes

- Maintain zerolog's zero-allocation guarantees.
- Keep the adapter layer minimal.
- Document JSON field naming conventions.
- Ensure compatibility with zerolog's API.

For general development guidelines, see the main
[slog AGENT.md](../../AGENT.md).
