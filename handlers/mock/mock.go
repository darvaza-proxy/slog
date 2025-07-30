// Package mock provides a mock logger implementation for testing slog handlers.
package mock

import (
	"fmt"
	"strings"
	"sync"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/internal"
)

// Message represents a recorded log message for testing.
type Message struct {
	Message string
	Level   slog.LogLevel
	Fields  map[string]any
	Stack   bool
}

// String returns a string representation of the message with sorted fields.
func (m Message) String() string {
	var b strings.Builder

	_, _ = fmt.Fprintf(&b, "[%v] %q", m.Level, m.Message)

	if len(m.Fields) > 0 {
		keys := core.SortedKeys(m.Fields)
		for _, k := range keys {
			_, _ = fmt.Fprintf(&b, " %s=%v", k, m.Fields[k])
		}
	}

	if m.Stack {
		_, _ = b.WriteString(" [stack]")
	}

	return b.String()
}

// Recorder provides thread-safe recording of log messages for testing.
type Recorder struct {
	mu       sync.Mutex
	messages []Message
}

// NewRecorder creates a new message recorder for testing.
func NewRecorder() *Recorder {
	return &Recorder{
		messages: make([]Message, 0),
	}
}

// Record stores a log message in the recorder.
func (r *Recorder) Record(msg Message) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages = append(r.messages, msg)
}

// GetMessages returns a copy of all recorded messages.
func (r *Recorder) GetMessages() []Message {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]Message, len(r.messages))
	copy(result, r.messages)
	return result
}

// Clear removes all recorded messages.
func (r *Recorder) Clear() {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages = r.messages[:0]
}

// Logger implements slog.Logger for testing purposes.
type Logger struct {
	internal.Loglet

	recorder *Recorder
	enabled  bool
}

// NewLogger creates a new test logger with a recorder.
func NewLogger() *Logger {
	return &Logger{
		recorder: NewRecorder(),
		enabled:  true,
	}
}

// GetMessages returns all recorded messages from this logger.
func (l *Logger) GetMessages() []Message {
	if l == nil {
		return nil
	}
	return l.recorder.GetMessages()
}

// Clear removes all recorded messages from this logger.
func (l *Logger) Clear() {
	if l == nil {
		return
	}
	l.recorder.Clear()
}

// Enabled returns whether this logger is enabled.
func (l *Logger) Enabled() bool {
	if l == nil {
		return false
	}
	return l.enabled
}

// WithEnabled returns the logger and its enabled state.
func (l *Logger) WithEnabled() (slog.Logger, bool) {
	if l == nil {
		return nil, false
	}
	return l, l.enabled
}

// Print implements slog.Logger.
func (l *Logger) Print(args ...any) {
	if l == nil {
		return
	}
	l.record(fmt.Sprint(args...))
}

// Println implements slog.Logger.
func (l *Logger) Println(args ...any) {
	if l == nil {
		return
	}
	l.record(fmt.Sprintln(args...))
}

// Printf implements slog.Logger.
func (l *Logger) Printf(format string, args ...any) {
	if l == nil {
		return
	}
	l.record(fmt.Sprintf(format, args...))
}

func (l *Logger) record(msg string) {
	// Collect all fields from the loglet chain
	fields := make(map[string]any)
	iter := l.Fields()
	for iter.Next() {
		k, v := iter.Field()
		fields[k] = v
	}

	l.recorder.Record(Message{
		Message: msg,
		Level:   l.Level(),
		Fields:  fields,
		Stack:   l.CallStack() != nil,
	})
}

// Debug returns a logger with Debug level.
func (l *Logger) Debug() slog.Logger {
	if l == nil {
		return nil
	}
	return l.WithLevel(slog.Debug)
}

// Info returns a logger with Info level.
func (l *Logger) Info() slog.Logger {
	if l == nil {
		return nil
	}
	return l.WithLevel(slog.Info)
}

// Warn returns a logger with Warn level.
func (l *Logger) Warn() slog.Logger {
	if l == nil {
		return nil
	}
	return l.WithLevel(slog.Warn)
}

// Error returns a logger with Error level.
func (l *Logger) Error() slog.Logger {
	if l == nil {
		return nil
	}
	return l.WithLevel(slog.Error)
}

// Fatal returns a logger with Fatal level.
func (l *Logger) Fatal() slog.Logger {
	if l == nil {
		return nil
	}
	return l.WithLevel(slog.Fatal)
}

// Panic returns a logger with Panic level.
func (l *Logger) Panic() slog.Logger {
	if l == nil {
		return nil
	}
	return l.WithLevel(slog.Panic)
}

// WithLevel returns a logger with the specified level.
func (l *Logger) WithLevel(level slog.LogLevel) slog.Logger {
	if l == nil {
		return nil
	}
	return l.withLoglet(l.Loglet.WithLevel(level))
}

// WithStack implements slog.Logger.
func (l *Logger) WithStack(skip int) slog.Logger {
	if l == nil {
		return nil
	}
	return l.withLoglet(l.Loglet.WithStack(skip + 1))
}

// WithField implements slog.Logger.
func (l *Logger) WithField(label string, value any) slog.Logger {
	if l == nil {
		return nil
	}
	if label == "" {
		return l
	}
	return l.withLoglet(l.Loglet.WithField(label, value))
}

// WithFields implements slog.Logger.
func (l *Logger) WithFields(fields map[string]any) slog.Logger {
	if l == nil {
		return nil
	}
	if !internal.HasFields(fields) {
		return l
	}
	return l.withLoglet(l.Loglet.WithFields(fields))
}

// withLoglet creates a new Logger with the given loglet.
func (l *Logger) withLoglet(loglet internal.Loglet) *Logger {
	return &Logger{
		Loglet:   loglet,
		recorder: l.recorder,
		enabled:  l.enabled,
	}
}
