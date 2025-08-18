package slog_test

import (
	"errors"
	"strings"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/mock"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = logWriterNewlineTestCase{}

func TestNewLogWriter(t *testing.T) {
	t.Run("BasicCreation", testNewLogWriterBasic)
	t.Run("NilLogger", testNewLogWriterNilLogger)
	t.Run("NilHandler", testNewLogWriterNilHandler)
	t.Run("CustomHandler", testNewLogWriterCustomHandler)
}

func testNewLogWriterBasic(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	writer := slog.NewLogWriter(logger, nil)

	core.AssertNotNil(t, writer, "log writer")
}

func testNewLogWriterNilLogger(t *testing.T) {
	t.Helper()
	writer := slog.NewLogWriter(nil, nil)

	core.AssertNil(t, writer, "log writer with nil logger")
}

func testNewLogWriterNilHandler(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	writer := slog.NewLogWriter(logger, nil)

	// Should use default handler
	core.AssertMustNotNil(t, writer, "log writer with nil handler")

	// Test default behaviour
	n, err := writer.Write([]byte("test message"))
	core.AssertMustNoError(t, err, "write with default handler")
	core.AssertMustEqual(t, 12, n, "byte count")

	msgs := logger.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "message count")
	core.AssertEqual(t, "test message", msgs[0].Message, "message content")
}

func testNewLogWriterCustomHandler(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	called := false

	customHandler := func(l slog.Logger, s string) error {
		called = true
		l.Printf("custom: %s", s)
		return nil
	}

	writer := slog.NewLogWriter(logger, customHandler)
	core.AssertMustNotNil(t, writer, "log writer with custom handler")

	_, err := writer.Write([]byte("test"))
	core.AssertMustNoError(t, err, "write with custom handler")

	core.AssertMustTrue(t, called, "custom handler called")

	msgs := logger.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "message count")
	core.AssertEqual(t, "custom: test", msgs[0].Message, "custom message")
}

func TestLogWriterWrite(t *testing.T) {
	t.Run("BasicWrite", testLogWriterWriteBasic)
	t.Run("NewlineHandling", testLogWriterWriteNewlines)
	t.Run("EmptyMessage", testLogWriterWriteEmpty)
	t.Run("HandlerError", testLogWriterWriteError)
	t.Run("MultipleWrites", testLogWriterWriteMultiple)
}

func testLogWriterWriteBasic(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	writer := slog.NewLogWriter(logger, nil)

	message := "basic test message"
	n, err := writer.Write([]byte(message))

	core.AssertMustNoError(t, err, "write operation")
	core.AssertMustEqual(t, len(message), n, "bytes written")

	msgs := logger.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "message count")
	core.AssertEqual(t, message, msgs[0].Message, "message content")
}

// logWriterNewlineTestCase represents a test case for log writer newline handling.
type logWriterNewlineTestCase struct {
	input    string
	expected string
	name     string
}

func (tc logWriterNewlineTestCase) Name() string {
	return tc.name
}

func (tc logWriterNewlineTestCase) Test(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	writer := slog.NewLogWriter(logger, nil)

	n, err := writer.Write([]byte(tc.input))
	core.AssertMustNoError(t, err, "write operation")
	core.AssertMustEqual(t, len(tc.input), n, "bytes written")

	msgs := logger.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "message count")
	core.AssertEqual(t, tc.expected, msgs[0].Message, "message content")
}

func newLogWriterNewlineTestCase(name, input, expected string) logWriterNewlineTestCase {
	return logWriterNewlineTestCase{
		name:     name,
		input:    input,
		expected: expected,
	}
}

func logWriterNewlineTestCases() []logWriterNewlineTestCase {
	return []logWriterNewlineTestCase{
		newLogWriterNewlineTestCase("SingleNewline", "message\n", "message"),
		newLogWriterNewlineTestCase("MultipleNewlines", "message\n\n\n", "message"),
		newLogWriterNewlineTestCase("NoNewline", "message", "message"),
		newLogWriterNewlineTestCase("OnlyNewlines", "\n\n", ""),
		newLogWriterNewlineTestCase("MiddleNewlines", "line1\nline2\n", "line1\nline2"),
	}
}

func testLogWriterWriteNewlines(t *testing.T) {
	core.RunTestCases(t, logWriterNewlineTestCases())
}

func testLogWriterWriteEmpty(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	writer := slog.NewLogWriter(logger, nil)

	n, err := writer.Write([]byte(""))
	core.AssertMustNoError(t, err, "write empty input")
	core.AssertMustEqual(t, 0, n, "bytes written for empty input")

	msgs := logger.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "message count")
	core.AssertEqual(t, "", msgs[0].Message, "empty message")
}

func testLogWriterWriteError(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	expectedErr := errors.New("handler error")

	errorHandler := func(slog.Logger, string) error {
		return expectedErr
	}

	writer := slog.NewLogWriter(logger, errorHandler)

	n, err := writer.Write([]byte("test message"))
	core.AssertMustErrorIs(t, err, expectedErr, "handler error")
	core.AssertEqual(t, 0, n, "bytes written on error")
}

func testLogWriterWriteMultiple(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	writer := slog.NewLogWriter(logger, nil)

	messages := []string{"first", "second", "third"}

	for _, msg := range messages {
		n, err := writer.Write([]byte(msg))
		core.AssertMustNoError(t, err, "write message %q", msg)
		core.AssertMustEqual(t, len(msg), n, "bytes written for %q", msg)
	}

	msgs := logger.GetMessages()
	core.AssertMustEqual(t, len(messages), len(msgs), "total message count")

	for i, expected := range messages {
		core.AssertEqual(t, expected, msgs[i].Message, "message %d", i)
	}
}

func TestLogWriterHandlerTypes(t *testing.T) {
	t.Run("DefaultHandler", testLogWriterDefaultHandler)
	t.Run("CustomTransform", testLogWriterCustomTransform)
	t.Run("FilteringHandler", testLogWriterFilteringHandler)
}

func testLogWriterDefaultHandler(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	writer := slog.NewLogWriter(logger, nil)

	// Test that default handler just calls Print
	_, _ = writer.Write([]byte("default test"))

	msgs := logger.GetMessages()
	if core.AssertEqual(t, 1, len(msgs), "message count") {
		core.AssertEqual(t, "default test", msgs[0].Message, "default handler message")
	}
}

func testLogWriterCustomTransform(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()

	transformHandler := func(l slog.Logger, s string) error {
		// Transform message to uppercase with prefix
		l.Printf("[TRANSFORM] %s", strings.ToUpper(s))
		return nil
	}

	writer := slog.NewLogWriter(logger, transformHandler)
	_, _ = writer.Write([]byte("transform this"))

	msgs := logger.GetMessages()
	if core.AssertEqual(t, 1, len(msgs), "message count") {
		expected := "[TRANSFORM] TRANSFORM THIS"
		core.AssertEqual(t, expected, msgs[0].Message, "transformed message")
		return
	}
}

func testLogWriterFilteringHandler(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	filtered := 0

	filteringHandler := func(l slog.Logger, s string) error {
		if strings.Contains(s, "ignore") {
			filtered++
			return nil // Don't log
		}
		l.Print(s)
		return nil
	}

	writer := slog.NewLogWriter(logger, filteringHandler)
	_, _ = writer.Write([]byte("keep this"))
	_, _ = writer.Write([]byte("ignore this"))
	_, _ = writer.Write([]byte("keep this too"))

	core.AssertMustEqual(t, 1, filtered, "filtered message count")

	msgs := logger.GetMessages()
	core.AssertMustEqual(t, 2, len(msgs), "kept message count")
	core.AssertEqual(t, "keep this", msgs[0].Message, "first kept message")
	core.AssertEqual(t, "keep this too", msgs[1].Message, "second kept message")
}
