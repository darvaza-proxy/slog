// Package testing provides shared test utilities for slog handler testing.
package testing

import (
	"darvaza.org/slog/handlers/mock"
)

// Backward compatibility aliases for the moved types.
// These allow existing code to continue working while encouraging
// migration to the public handlers/mock package.

// Message represents a recorded log message for testing.
// Deprecated: Use darvaza.org/slog/handlers/mock.Message instead.
type Message = mock.Message

// Recorder provides thread-safe recording of log messages for testing.
// Deprecated: Use darvaza.org/slog/handlers/mock.Recorder instead.
type Recorder = mock.Recorder

// Logger implements slog.Logger for testing purposes.
// Deprecated: Use darvaza.org/slog/handlers/mock.Logger instead.
type Logger = mock.Logger

// NewRecorder creates a new message recorder for testing.
// Deprecated: Use darvaza.org/slog/handlers/mock.NewRecorder instead.
func NewRecorder() *Recorder {
	return mock.NewRecorder()
}

// NewLogger creates a new test logger with a recorder.
// Deprecated: Use darvaza.org/slog/handlers/mock.NewLogger instead.
func NewLogger() *Logger {
	return mock.NewLogger()
}
