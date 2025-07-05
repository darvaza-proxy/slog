# AGENT.md

This file provides guidance to AI agents when working with code in this
repository. For developers and general project information, please refer to
[README.md](README.md) first.

## Handler Overview

`darvaza.org/slog/handlers/cblog` is a channel-based logging handler that
implements the `slog.Logger` interface. It sends log entries through Go
channels, allowing custom processing, buffering, or asynchronous handling of
log messages.

## Key Features

- **Channel-based delivery**: Log entries are sent through channels for
  flexible processing.
- **Worker management**: Supports background workers for asynchronous log
  handling.
- **Configurable buffering**: Control channel buffer size for performance
  tuning.
- **Non-blocking option**: Can be configured to drop messages instead of
  blocking when full.

## Architecture

The handler consists of two main components:

1. **Logger implementation** (cblog.go): Implements the `slog.Logger`
   interface and sends entries to channels.
2. **Worker system** (worker.go): Manages background workers that process log
   entries from channels.

## Usage Patterns

```go
import (
    "fmt"

    "darvaza.org/slog/handlers/cblog"
)

// Create a channel-based logger
ch := make(chan cblog.LogMsg, 100)
logger := cblog.New(ch)

// Process entries in a separate goroutine
go func() {
    for entry := range ch {
        // Custom processing
        fmt.Printf("[%s] %s\n", entry.Level, entry.Message)
    }
}()

// Use like any slog logger
logger.Info().WithField("user", "john").Print("Login successful")
```

## Testing Considerations

When testing with cblog:

- Remember to drain channels to avoid goroutine leaks.
- Use buffered channels to prevent blocking in tests.
- Close channels properly when done.
- Consider using select with timeout for test assertions.
- Note: The actual type is `LogMsg`, not `Entry`.

## Common Issues

1. **Channel blocking**: Unbuffered channels will block if no reader is active.
2. **Goroutine leaks**: Always ensure channels are drained and closed.
3. **Message ordering**: Channel delivery order is guaranteed but processing
   order depends on worker implementation.

## Development Notes

- The handler should never panic due to closed channels.
- All methods must be safe for concurrent use.
- Consider performance implications of channel operations.
- Document any blocking behavior clearly.

For general development guidelines, see the main
[slog AGENT.md](../../AGENT.md).
