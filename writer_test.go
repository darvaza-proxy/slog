package slog_test

import (
	"errors"
	"strings"
	"testing"

	"darvaza.org/slog"
	slogtest "darvaza.org/slog/internal/testing"
)

func TestNewLogWriter(t *testing.T) {
	t.Run("BasicCreation", testNewLogWriterBasic)
	t.Run("NilLogger", testNewLogWriterNilLogger)
	t.Run("NilHandler", testNewLogWriterNilHandler)
	t.Run("CustomHandler", testNewLogWriterCustomHandler)
}

func testNewLogWriterBasic(t *testing.T) {
	logger := slogtest.NewLogger()
	writer := slog.NewLogWriter(logger, nil)

	if writer == nil {
		t.Fatal("NewLogWriter should not return nil for valid logger")
	}
}

func testNewLogWriterNilLogger(t *testing.T) {
	writer := slog.NewLogWriter(nil, nil)

	if writer != nil {
		t.Error("NewLogWriter should return nil for nil logger")
	}
}

func testNewLogWriterNilHandler(t *testing.T) {
	logger := slogtest.NewLogger()
	writer := slog.NewLogWriter(logger, nil)

	// Should use default handler
	if writer == nil {
		t.Fatal("NewLogWriter should handle nil handler with default")
	}

	// Test default behaviour
	n, err := writer.Write([]byte("test message"))
	if err != nil {
		t.Errorf("Write should not error with default handler: %v", err)
	}
	if n != 12 {
		t.Errorf("Write should return correct byte count: got %d, want 12", n)
	}

	msgs := logger.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Message != "test message" {
		t.Errorf("Message should be 'test message', got %q", msgs[0].Message)
	}
}

func testNewLogWriterCustomHandler(t *testing.T) {
	logger := slogtest.NewLogger()
	called := false

	customHandler := func(l slog.Logger, s string) error {
		called = true
		l.Printf("custom: %s", s)
		return nil
	}

	writer := slog.NewLogWriter(logger, customHandler)
	if writer == nil {
		t.Fatal("NewLogWriter should not return nil")
	}

	_, err := writer.Write([]byte("test"))
	if err != nil {
		t.Errorf("Write should not error: %v", err)
	}

	if !called {
		t.Error("Custom handler should have been called")
	}

	msgs := logger.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Message != "custom: test" {
		t.Errorf("Expected 'custom: test', got %q", msgs[0].Message)
	}
}

func TestLogWriterWrite(t *testing.T) {
	t.Run("BasicWrite", testLogWriterWriteBasic)
	t.Run("NewlineHandling", testLogWriterWriteNewlines)
	t.Run("EmptyMessage", testLogWriterWriteEmpty)
	t.Run("HandlerError", testLogWriterWriteError)
	t.Run("MultipleWrites", testLogWriterWriteMultiple)
}

func testLogWriterWriteBasic(t *testing.T) {
	logger := slogtest.NewLogger()
	writer := slog.NewLogWriter(logger, nil)

	message := "basic test message"
	n, err := writer.Write([]byte(message))

	if err != nil {
		t.Errorf("Write should not error: %v", err)
	}
	if n != len(message) {
		t.Errorf("Write should return message length: got %d, want %d", n, len(message))
	}

	msgs := logger.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Message != message {
		t.Errorf("Message mismatch: got %q, want %q", msgs[0].Message, message)
	}
}

func testLogWriterWriteNewlines(t *testing.T) {
	logger := slogtest.NewLogger()
	writer := slog.NewLogWriter(logger, nil)

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"SingleNewline", "message\n", "message"},
		{"MultipleNewlines", "message\n\n\n", "message"},
		{"NoNewline", "message", "message"},
		{"OnlyNewlines", "\n\n", ""},
		{"MiddleNewlines", "line1\nline2\n", "line1\nline2"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger.Clear()

			n, err := writer.Write([]byte(tc.input))
			if err != nil {
				t.Errorf("Write should not error: %v", err)
			}
			if n != len(tc.input) {
				t.Errorf("Write should return input length: got %d, want %d", n, len(tc.input))
			}

			msgs := logger.GetMessages()
			if len(msgs) != 1 {
				t.Fatalf("Expected 1 message, got %d", len(msgs))
			}
			if msgs[0].Message != tc.expected {
				t.Errorf("Message mismatch: got %q, want %q", msgs[0].Message, tc.expected)
			}
		})
	}
}

func testLogWriterWriteEmpty(t *testing.T) {
	logger := slogtest.NewLogger()
	writer := slog.NewLogWriter(logger, nil)

	n, err := writer.Write([]byte(""))
	if err != nil {
		t.Errorf("Write should not error on empty input: %v", err)
	}
	if n != 0 {
		t.Errorf("Write should return 0 for empty input: got %d", n)
	}

	msgs := logger.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Message != "" {
		t.Errorf("Message should be empty, got %q", msgs[0].Message)
	}
}

func testLogWriterWriteError(t *testing.T) {
	logger := slogtest.NewLogger()
	expectedErr := errors.New("handler error")

	errorHandler := func(slog.Logger, string) error {
		return expectedErr
	}

	writer := slog.NewLogWriter(logger, errorHandler)

	n, err := writer.Write([]byte("test message"))
	if err != expectedErr {
		t.Errorf("Write should return handler error: got %v, want %v", err, expectedErr)
	}
	if n != 0 {
		t.Errorf("Write should return 0 on error: got %d", n)
	}
}

func testLogWriterWriteMultiple(t *testing.T) {
	logger := slogtest.NewLogger()
	writer := slog.NewLogWriter(logger, nil)

	messages := []string{"first", "second", "third"}

	for _, msg := range messages {
		n, err := writer.Write([]byte(msg))
		if err != nil {
			t.Errorf("Write should not error: %v", err)
		}
		if n != len(msg) {
			t.Errorf("Write should return message length: got %d, want %d", n, len(msg))
		}
	}

	msgs := logger.GetMessages()
	if len(msgs) != len(messages) {
		t.Fatalf("Expected %d messages, got %d", len(messages), len(msgs))
	}

	for i, expected := range messages {
		if msgs[i].Message != expected {
			t.Errorf("Message %d mismatch: got %q, want %q", i, msgs[i].Message, expected)
		}
	}
}

func TestLogWriterHandlerTypes(t *testing.T) {
	t.Run("DefaultHandler", testLogWriterDefaultHandler)
	t.Run("CustomTransform", testLogWriterCustomTransform)
	t.Run("FilteringHandler", testLogWriterFilteringHandler)
}

func testLogWriterDefaultHandler(t *testing.T) {
	logger := slogtest.NewLogger()
	writer := slog.NewLogWriter(logger, nil)

	// Test that default handler just calls Print
	_, _ = writer.Write([]byte("default test"))

	msgs := logger.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(msgs))
	}
	// Default handler should use Print method
	if msgs[0].Message != "default test" {
		t.Errorf("Default handler message: got %q, want %q", msgs[0].Message, "default test")
	}
}

func testLogWriterCustomTransform(t *testing.T) {
	logger := slogtest.NewLogger()

	transformHandler := func(l slog.Logger, s string) error {
		// Transform message to uppercase with prefix
		l.Printf("[TRANSFORM] %s", strings.ToUpper(s))
		return nil
	}

	writer := slog.NewLogWriter(logger, transformHandler)
	_, _ = writer.Write([]byte("transform this"))

	msgs := logger.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(msgs))
	}
	expected := "[TRANSFORM] TRANSFORM THIS"
	if msgs[0].Message != expected {
		t.Errorf("Transform handler message: got %q, want %q", msgs[0].Message, expected)
	}
}

func testLogWriterFilteringHandler(t *testing.T) {
	logger := slogtest.NewLogger()
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

	if filtered != 1 {
		t.Errorf("Expected 1 filtered message, got %d", filtered)
	}

	msgs := logger.GetMessages()
	if len(msgs) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Message != "keep this" {
		t.Errorf("First message: got %q, want %q", msgs[0].Message, "keep this")
	}
	if msgs[1].Message != "keep this too" {
		t.Errorf("Second message: got %q, want %q", msgs[1].Message, "keep this too")
	}
}
