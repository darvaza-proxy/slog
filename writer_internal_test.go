package slog

import (
	"fmt"
	"testing"

	"darvaza.org/core"
)

// Simple mock logger for testing without import cycles
type mockLogger struct {
	messages []string
}

func (m *mockLogger) Debug() Logger                     { return m }
func (m *mockLogger) Info() Logger                      { return m }
func (m *mockLogger) Warn() Logger                      { return m }
func (m *mockLogger) Error() Logger                     { return m }
func (m *mockLogger) Fatal() Logger                     { return m }
func (m *mockLogger) Panic() Logger                     { return m }
func (m *mockLogger) Print(args ...any)                 { m.messages = append(m.messages, fmt.Sprint(args...)) }
func (m *mockLogger) Println(args ...any)               { m.Print(args...) }
func (m *mockLogger) Printf(format string, args ...any) { m.Print(fmt.Sprintf(format, args...)) }
func (m *mockLogger) WithLevel(LogLevel) Logger         { return m }
func (m *mockLogger) WithStack(int) Logger              { return m }
func (m *mockLogger) WithField(string, any) Logger      { return m }
func (m *mockLogger) WithFields(map[string]any) Logger  { return m }
func (*mockLogger) Enabled() bool                       { return true }
func (m *mockLogger) WithEnabled() (Logger, bool)       { return m, m.Enabled() }

func TestLogWriterPrivate(t *testing.T) {
	t.Run("NilHandlerField", testLogWriterWriteNilHandlerField)
}

func testLogWriterWriteNilHandlerField(t *testing.T) {
	t.Helper()
	logger := &mockLogger{}

	// Create LogWriter directly with nil fn field (bypassing NewLogWriter)
	// This tests the specific case where fn is nil inside Write method
	writer := &LogWriter{
		l:  logger,
		fn: nil, // Explicitly nil to test the nil check in Write
	}

	message := "test with nil handler field"
	n, err := writer.Write([]byte(message))

	core.AssertMustNoError(t, err, "write with nil handler field")
	core.AssertMustEqual(t, len(message), n, "bytes written")
	core.AssertMustEqual(t, 1, len(logger.messages), "message count")
	core.AssertEqual(t, message, logger.messages[0], "message content")
}
