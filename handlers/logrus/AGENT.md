# AGENT.md

This file provides guidance to AI agents when working with code in this
repository. For developers and general project information, please refer to
[README.md](README.md) first.

## Handler Overview

`darvaza.org/slog/handlers/logrus` is an adapter that implements the
`slog.Logger` interface using
[Sirupsen/logrus](https://github.com/sirupsen/logrus) as the backend logging
library.

## Key Features

- **Full logrus compatibility**: Leverages all logrus features including hooks,
  formatters, and outputs.
- **Level mapping**: Correctly maps slog levels to logrus levels.
- **Field preservation**: Maintains structured fields through the adapter.
- **Context support**: Preserves logrus.Entry context through operations.

## Implementation Notes

- Uses `logrus.Entry` internally to maintain field context.
- Fatal and Panic levels behave like logrus (exit/panic after logging).
- Field values are passed directly to logrus for serialization.
- Stack traces are added as fields when requested.

## Usage Patterns

```go
import (
    "github.com/sirupsen/logrus"
    slogrus "darvaza.org/slog/handlers/logrus"
)

// Wrap existing logrus logger
logrusLogger := logrus.New()
logrusLogger.SetLevel(logrus.DebugLevel)
logrusLogger.SetFormatter(&logrus.JSONFormatter{})

slogLogger := slogrus.New(logrusLogger)

// Use with slog interface
slogLogger.Info().
    WithField("component", "api").
    WithField("method", "GET").
    Print("Request received")
```

## Level Mapping

| slog Level | logrus Level |
|------------|--------------|
| Debug      | Debug        |
| Info       | Info         |
| Warn       | Warning      |
| Error      | Error        |
| Fatal      | Fatal        |
| Panic      | Panic        |

## Testing Considerations

- Use logrus test hooks to verify log output.
- Test that Fatal and Panic behave correctly (in separate processes).
- Verify field propagation through the adapter.
- Ensure logrus configuration is respected.

## Development Notes

- Maintain compatibility with logrus API changes.
- Preserve logrus-specific features where possible.
- Document any behavioural differences from native logrus.
- Keep the adapter layer as thin as possible.

For general development guidelines, see the main
[slog AGENT.md](../../AGENT.md).
