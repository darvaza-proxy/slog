# `slog`

[![Go Reference][godoc-badge]][godoc]
[![Go Report Card][goreport-badge]][goreport]
[![codecov][codecov-badge]][codecov]
[![Socket Badge][socket-badge]][socket-link]

`darvaza.org/slog` provides a backend-agnostic interface for structured logging
in Go. It defines a simple, standardised API that libraries can use without
forcing a specific logging implementation on their users.

[godoc]: https://pkg.go.dev/darvaza.org/slog
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/slog.svg
[goreport]: https://goreportcard.com/report/darvaza.org/slog
[goreport-badge]: https://goreportcard.com/badge/darvaza.org/slog
[codecov]: https://codecov.io/gh/darvaza-proxy/slog
[codecov-badge]: https://codecov.io/github/darvaza-proxy/slog/graph/badge.svg
[socket-badge]: https://socket.dev/api/badge/go/package/darvaza.org/slog
[socket-link]: https://socket.dev/go/package/darvaza.org/slog

## Features

- **Backend-agnostic**: Define logging interfaces without forcing
  implementation choices.
- **Structured logging**: Support for typed fields with string keys.
- **Method chaining**: Fluent API for composing log entries.
- **Six log levels**: Debug, Info, Warn, Error, Fatal, and Panic.
- **Context integration**: Store and retrieve loggers from context values.
- **Standard library compatible**: Adapters for Go's standard `log` package.
- **Multiple handlers**: Pre-built integrations with popular logging libraries.
- **Immutable logger instances**: Each modification creates a new logger,
  enabling safe concurrent use and proper branching behaviour.

## Installation

```bash
go get darvaza.org/slog
```

## Quick Start

```go
package main

import (
    "darvaza.org/slog"
    "darvaza.org/slog/handlers/discard"
)

func main() {
    // Create a logger (using discard handler for example)
    logger := discard.New()

    // Log with different levels
    logger.Info().Print("Application started")

    // Add fields
    logger.Debug().
        WithField("user", "john").
        WithField("action", "login").
        Print("User logged in")

    // Use Printf-style formatting
    logger.Warn().
        WithField("retry_count", 3).
        Printf("Connection failed, will retry")
}
```

## Interface

The `slog.Logger` interface provides a fluent API where most methods return a
`Logger` for method chaining. A log entry is composed by:

1. Setting the log level
2. Optionally adding fields and call stack information
3. Emitting the entry with a Print method

Disabled log entries incur minimal overhead as string formatting and field
collection are skipped.

## Log Levels

The library supports six log levels with clear semantics:

1. **Debug**: Detailed information for developers.
2. **Info**: General informational messages.
3. **Warn**: Warning messages for potentially harmful situations.
4. **Error**: Error conditions that allow continued operation.
5. **Fatal**: Critical errors that terminate the program (like `log.Fatal()`).
6. **Panic**: Errors that trigger a recoverable panic (like `log.Panic()`).

Create log entries using named methods (`Debug()`, `Info()`, etc.) or
`WithLevel(level)`.

## Enabled State

A log entry is "enabled" if the handler will actually emit it. Operating on
disabled loggers is safe and efficient - string formatting and field collection
are skipped.

Use `WithEnabled()` to check if a level is enabled:

```go
if log, ok := logger.Debug().WithEnabled(); ok {
    // Expensive debug logging
    log.WithField("details", expensiveOperation()).Print("Debug info")
} else if log, ok := logger.Info().WithEnabled(); ok {
    // Simpler info logging
    log.Print("Operation completed")
}
```

**Note**: Fatal and Panic levels always execute regardless of enabled state.

## Fields

Fields are key/value pairs for structured logging:

- Keys must be non-empty strings
- Values can be any type
- Fields are attached using `WithField(key, value)`
- Multiple fields can be attached by chaining calls

```go
import "time"

start := time.Now()
// ... perform some work ...

logger.Info().
    WithField("user_id", 123).
    WithField("duration", time.Since(start)).
    Print("Request processed")
```

## Branching Behaviour

Each logger instance is immutable. When you call methods like `WithField()` or
`WithLevel()`, you get a new logger instance that inherits from the parent:

```go
// Create a base logger with common fields
baseLogger := logger.WithField("service", "api")

// Branch off for different request handlers
userLogger := baseLogger.WithField("handler", "user")
adminLogger := baseLogger.WithField("handler", "admin")

// Each logger maintains its own field chain
userLogger.Info().Print("Processing user request")
// Output includes: service=api, handler=user

adminLogger.Info().Print("Processing admin request")
// Output includes: service=api, handler=admin

// Original logger is unchanged
baseLogger.Info().Print("Base logger message") // only has service=api
```

This design ensures:

- Thread-safe concurrent use without locks
- No unintended field pollution between different code paths
- Clear ownership and lifecycle of logger configurations

## Call Stack

Attach stack traces to log entries using `WithStack(skip)`:

```go
logger.Error().
    WithStack(0).  // 0 = current function
    WithField("error", err).
    Print("Operation failed")
```

The `skip` parameter specifies how many stack frames to skip (0 = current
function).

## Print Methods

Three print methods match the `fmt` package conventions:

- `Print(v ...any)`: Like `fmt.Print`
- `Println(v ...any)`: Like `fmt.Println`
- `Printf(format string, v ...any)`: Like `fmt.Printf`

These methods emit the log entry with all attached fields.

## Standard Library Integration

Integrate with Go's standard `log` package:

```go
import (
    "log"
    "net/http"

    "darvaza.org/slog"
)

// Assuming you have a slog logger instance
// var logger slog.Logger

// Create a standard logger that writes to slog
stdLogger := slog.NewStdLogger(logger, "[HTTP]", log.LstdFlags)

// Use with libraries expecting *log.Logger
server := &http.Server{
    ErrorLog: stdLogger,
}
```

For custom parsing, use `NewLogWriter()` with a handler function.

## Architecture Overview

```text
┌────────────────────────────────────────────────────────────────────┐
│                        External Dependencies                       │
├─────────────────────────────┬──────────────────────────────────────┤
│    darvaza.org/core         │      Go Standard Library             │
└─────────────┬───────────────┴───────────┬──────────────────────────┘
              │                           │
              ▼                           ▼
┌────────────────────────────────────────────────────────────────────┐
│                            slog Core                               │
├──────────────────────┬─────────────────────┬───────────────────────┤
│   Logger Interface   │ Context Integration │ Std Library Adapter   │
├──────────────────────┴─────────────────────┴───────────────────────┤
│             internal.Loglet (field chain management)               │
└──────────┬─────────────────────────────────────────────────────────┘
           │
           ▼
┌────────────────────────────────────────────────────────────────────┐
│                            Handlers                                │
├─────────┬─────────┬─────────┬─────────┬─────────┬────────┬─────────┤
│  logr   │ logrus  │   zap   │ zerolog │  cblog  │ filter │ discard │
└─────────┴─────────┴─────────┴─────────┴─────────┴────────┴─────────┘
```

All handlers use the `internal.Loglet` type for consistent field chain
management and immutable logger behaviour.

## Available Handlers

### Adapter Types

Handlers fall into two categories based on their integration capabilities:

#### Bidirectional Adapters

These handlers allow conversion in both directions - you can use the external
logging library as a slog backend, OR use slog as a backend for the external
library:

- **[logr](https://pkg.go.dev/darvaza.org/slog/handlers/logr)**:
  Full bidirectional adapter for go-logr/logr interface.
  - `logr.Logger` → `slog.Logger` (use logr as slog backend)
  - `slog.Logger` → `logr.Logger` (use slog as logr backend)
- **[`logrus`](https://pkg.go.dev/darvaza.org/slog/handlers/logrus)**:
  Bidirectional adapter for Sirupsen/logrus.
  - `logrus.Logger` → `slog.Logger` (use logrus as slog backend)
  - `slog.Logger` → `logrus.Logger` (use slog as logrus backend)
- **[`zap`](https://pkg.go.dev/darvaza.org/slog/handlers/zap)**:
  Bidirectional adapter between Uber's zap and slog. Use zap as a slog backend
  or create zap loggers backed by any slog implementation.

#### Unidirectional Adapters

These handlers only allow using the external logging library as a backend for
slog. They wrap existing loggers but don't provide the reverse conversion:

- **[`zerolog`](https://pkg.go.dev/darvaza.org/slog/handlers/zerolog)**:
  Wraps rs/zerolog as a slog backend.

### Utility Handlers

These handlers provide additional functionality without external dependencies:

- **[`cblog`](https://pkg.go.dev/darvaza.org/slog/handlers/cblog)**:
  Channel-based handler for custom log processing.
- **[`filter`](https://pkg.go.dev/darvaza.org/slog/handlers/filter)**:
  Middleware to filter and transform log entries.
- **[`mock`](https://pkg.go.dev/darvaza.org/slog/handlers/mock)**:
  Mock logger implementation that records messages for testing and verification.
  Provides a fully functional slog.Logger that captures all log entries with
  their levels, messages, and fields for programmatic inspection.

  ```go
  import (
      "testing"

      "darvaza.org/slog/handlers/mock"
  )

  func TestMyCode(t *testing.T) {
      logger := mock.NewLogger()

      // Use logger in your code
      myFunction(logger)

      // Verify what was logged
      messages := logger.GetMessages()
      if len(messages) != 1 {
          t.Fatalf("expected 1 message, got %d", len(messages))
      }

      msg := messages[0]
      if msg.Level != slog.Info || msg.Message != "expected message" {
          t.Errorf("unexpected log entry: %v", msg)
      }
  }
  ```

- **[`discard`](https://pkg.go.dev/darvaza.org/slog/handlers/discard)**:
  No-op handler for testing and optional logging.

### Adapter Differences

**Bidirectional adapters** are valuable when:

- Integration with libraries that expect a specific logger interface is
  required.
- Gradual migration between logging systems is in progress.
- A common interface is desired across different application components while
  maintaining compatibility with existing code.

**Unidirectional adapters** are simpler and suitable when:

- An existing logger serves as the slog backend without reverse integration.
- New applications can adopt slog as the primary logging interface.
- Libraries expecting the backend's specific interface are not a concern.

## Testing

The package provides comprehensive test utilities for handler implementations in
`internal/testing`. These utilities help ensure consistent testing patterns and
reduce code duplication across handlers.

See [internal/testing/README.md](internal/testing/README.md) for detailed
documentation on using the test utilities, including:

- Test logger for recording and verifying log messages
- Assertion helpers for message verification
- Compliance test suite for interface conformance
- Concurrency testing utilities

## Development

See [AGENTS.md](AGENTS.md) for development guidelines and
[LICENCE.txt](LICENCE.txt) for licensing information.
