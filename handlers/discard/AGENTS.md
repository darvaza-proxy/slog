# AGENTS.md

This file provides guidance to AI agents when working with code in this
repository. For developers and general project information, please refer to
[README.md](README.md) first.

## Handler Overview

`darvaza.org/slog/handlers/discard` is a no-op logging handler that implements
the `slog.Logger` interface but discards all log entries. It's designed for
scenarios where logging is optional or for testing purposes.

## Key Features

- **Zero overhead**: All operations are no-ops with minimal CPU cost.
- **Safe to use**: Never returns nil, preventing nil pointer dereferences.
- **Testing friendly**: Ideal for unit tests where logging is not under test.
- **Optional logging**: Allows code to work with or without active logging.

## Implementation Details

The discard handler:

- Returns itself for all chained method calls.
- Performs no allocations for log operations.
- Ignores all fields, levels, and messages.
- Always reports as "not enabled" for any level.

## Usage Patterns

```go
import (
    "testing"

    "darvaza.org/slog"
    "darvaza.org/slog/handlers/discard"
)

// Use when logger is optional
func ProcessData(data []byte, logger slog.Logger) error {
    if logger == nil {
        logger = discard.New()
    }

    logger.Debug().Print("Starting processing")
    // ... rest of function
}

// In tests
func TestSomething(t *testing.T) {
    service := NewService(discard.New())
    // Test service without log output
}
```

## Design Principles

1. **Null object pattern**: Provides a valid object that does nothing.
2. **Interface compliance**: Fully implements `slog.Logger` interface.
3. **Immutable**: All operations return the same instance.
4. **Thread-safe**: Safe for concurrent use (does nothing).

## Testing Considerations

- Useful for benchmarks to eliminate logging overhead.
- It helps isolate code under test from logging concerns by eliminating
  log output.
- Can be used to verify code works without a logger.

## Common Use Cases

1. **Optional dependencies**: When logging is not required.
2. **Testing**: Silence logs during test execution.
3. **Benchmarking**: Remove logging overhead from measurements.
4. **Configuration**: Default logger before real logger is configured.

## Development Notes

- Keep implementation as simple as possible.
- Avoid any allocations or side effects.
- Maintain compatibility with the `slog.Logger` interface.
- Document that this handler discards everything.

For general development guidelines, see the main
[slog AGENTS.md](../../AGENTS.md).
