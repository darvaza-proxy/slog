# Mock Logger Handler

The mock handler provides a Logger implementation that records log messages for
testing and verification purposes. It's designed to help test other slog
handlers and applications that use slog.

## Overview

The mock handler consists of two main components:

- **Logger**: A fully functional slog.Logger implementation that records
  messages instead of outputting them
- **Recorder**: A thread-safe storage system for capturing and retrieving log
  messages

## Usage

### Basic Testing

```go
import (
    "testing"
    "darvaza.org/slog"
    "darvaza.org/slog/handlers/mock"
)

func TestMyCode(t *testing.T) {
    // Create a mock logger
    logger := mock.NewLogger()

    // Use it in your code
    myFunction(logger)

    // Verify the logged messages
    messages := logger.GetMessages()
    if len(messages) != 1 {
        t.Fatalf("expected 1 message, got %d", len(messages))
    }

    msg := messages[0]
    if msg.Level != slog.Info {
        t.Errorf("expected Info level, got %v", msg.Level)
    }
    if msg.Message != "expected message" {
        t.Errorf("expected 'expected message', got '%s'", msg.Message)
    }
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

- `NewLogger() *Logger` - Creates a new mock logger
- `GetMessages() []Message` - Returns all recorded messages
- Implements all slog.Logger methods (Debug, Info, WithField, etc.)

### Recorder

- `NewRecorder() *Recorder` - Creates a new message recorder
- `Record(msg Message)` - Stores a message
- `GetMessages() []Message` - Returns all recorded messages
- `Clear()` - Removes all stored messages

### Message

- `Message` struct contains:
  - `Message string` - The log message text
  - `Level slog.LogLevel` - The log level
  - `Fields map[string]any` - Attached fields
  - `Stack bool` - Whether stack trace was requested
- `String() string` - Formatted string representation

## Features

- **Thread-safe**: All operations are safe for concurrent use
- **Immutable**: Logger methods return new instances, preserving immutability
- **Complete**: Implements the full slog.Logger interface
- **Field preservation**: Maintains field chains correctly
- **Stack tracking**: Records when WithStack() was called

## Design Notes

The mock logger is designed to be a faithful implementation of the slog.Logger
interface while capturing all the information needed for testing. It preserves
the immutable nature of loggers by creating new instances for each modification.

The recorder component is separate to allow for advanced testing scenarios where
you might want to share a recorder between multiple logger instances or create
custom recording strategies.
