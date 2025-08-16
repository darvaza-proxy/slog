# Mock Logger Handler

The mock handler provides a Logger implementation that records log messages for
testing and verification purposes. It is designed to help test other slog
handlers and applications that use slog.

## Overview

The mock handler consists of two main components:

- **Logger**: A fully functional slog.Logger implementation that records
  messages instead of outputting them.
- **Recorder**: A thread-safe storage system for capturing and retrieving log
  messages.

## Usage

### Basic Testing

```go
import (
    "testing"
    "darvaza.org/slog"
    "darvaza.org/slog/handlers/mock"
)

func TestMyCode(t *testing.T) {
    // Create a mock logger.
    logger := mock.NewLogger()

    // Use it in your code.
    myFunction(logger)

    // Verify the logged messages.
    messages := logger.GetMessages()
    core.AssertMustEqual(t, 1, len(messages), "message count")

    msg := messages[0]
    core.AssertEqual(t, slog.Info, msg.Level, "log level")
    core.AssertEqual(t, "expected message", msg.Message, "log message")
}
```

### Testing with Fields

```go
func TestWithFields(t *testing.T) {
    logger := mock.NewLogger()

    logger.Info().
        WithField("user", "john").
        WithField("action", "login").
        Print("User logged in")

    messages := logger.GetMessages()
    msg := messages[0]

    // Check fields
    if msg.Fields["user"] != "john" {
        t.Errorf("expected user=john, got %v", msg.Fields["user"])
    }
    if msg.Fields["action"] != "login" {
        t.Errorf("expected action=login, got %v", msg.Fields["action"])
    }
}
```

### Testing with Level Filtering

The mock logger supports threshold-based filtering for testing handlers that
only process certain log levels:

```go
package mock_test

import (
    "testing"

    "darvaza.org/core"
    "darvaza.org/slog"
    "darvaza.org/slog/handlers/mock"
)

func TestLevelFiltering(t *testing.T) {
    // Create logger that only records Info and above
    logger := mock.NewLoggerWithThreshold(slog.Info)

    logger.Debug().Print("debug message")  // Not recorded
    logger.Info().Print("info message")    // Recorded
    logger.Error().Print("error message")  // Recorded

    messages := logger.GetMessages()
    core.AssertEqual(t, 2, len(messages), "message count")
}
```

### Testing Handler Adapters

The mock logger is particularly useful for testing adapter handlers that wrap
other loggers:

```go
func TestMyAdapter(t *testing.T) {
    // Create mock as backend
    backend := mock.NewLogger()

    // Create your adapter using the mock backend
    adapter := myadapter.New(backend)

    // Use the adapter
    adapter.Error().Print("Something went wrong")

    // Verify through the backend
    messages := backend.GetMessages()
    // ... verify messages were properly forwarded
}
```

## API Reference

### Logger

- `NewLogger() *Logger` - Creates a new mock logger that records all levels.
- `NewLoggerWithThreshold(threshold slog.LogLevel) *Logger` - Creates a new mock
  logger with level filtering.
- `GetMessages() []Message` - Returns all recorded messages.
- `Clear()` - Removes all recorded messages from this logger.
- Implements all slog.Logger methods (Debug, Info, WithField, etc.).

### Recorder

- `NewRecorder() *Recorder` - Creates a new message recorder.
- `Record(msg Message)` - Stores a message.
- `GetMessages() []Message` - Returns all recorded messages.
- `Clear()` - Removes all stored messages.

### Message

- `Message` struct contains:
  - `Message string` - The log message text.
  - `Level slog.LogLevel` - The log level.
  - `Fields map[string]any` - Attached fields.
  - `Stack bool` - Whether stack trace was requested.
- `String() string` - Formatted string representation.

## Features

- **Thread-safe**: All operations are safe for concurrent use.
- **Immutable**: Logger methods return new instances, preserving immutability.
- **Complete**: Implements the full slog.Logger interface.
- **Field preservation**: Maintains field chains correctly.
- **Stack tracking**: Records when WithStack() was called.
- **Level filtering**: Optional threshold filtering for testing specific log
  levels.
- **Backward compatible**: Existing code using NewLogger() continues to work
  unchanged.

## Design Notes

The mock logger is designed to be a faithful implementation of the slog.Logger
interface whilst capturing all the information needed for testing. It preserves
the immutable nature of loggers by creating new instances for each modification.

The recorder component is separate to allow for advanced testing scenarios where
you might want to share a recorder between multiple logger instances or create
custom recording strategies.

Level filtering uses the same logic as the filter handler: the zero value
(UndefinedLevel) means no filtering occurs, maintaining backward compatibility.
When a threshold is set, only messages at or above that level are recorded.
