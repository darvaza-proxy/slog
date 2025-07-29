package slog_test

import (
	"log"
	"strings"
	"testing"

	"darvaza.org/slog"
	slogtest "darvaza.org/slog/internal/testing"
)

func TestNewStdLogger(t *testing.T) {
	t.Run("BasicCreation", testNewStdLoggerBasic)
	t.Run("WithPrefix", testNewStdLoggerPrefix)
	t.Run("WithFlags", testNewStdLoggerFlags)
	t.Run("EmptyPrefix", testNewStdLoggerEmptyPrefix)
	t.Run("Integration", testNewStdLoggerIntegration)
}

func testNewStdLoggerBasic(t *testing.T) {
	logger := slogtest.NewLogger()
	stdLog := slog.NewStdLogger(logger, "", 0)

	if stdLog == nil {
		t.Fatal("NewStdLogger should not return nil")
	}

	// Test basic functionality
	stdLog.Print("test message")

	msgs := logger.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Message != "test message" {
		t.Errorf("Message should be 'test message', got %q", msgs[0].Message)
	}
}

func testNewStdLoggerPrefix(t *testing.T) {
	logger := slogtest.NewLogger()
	prefix := "TEST"
	stdLog := slog.NewStdLogger(logger, prefix, 0)

	stdLog.Print("message")

	msgs := logger.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(msgs))
	}

	expected := "TEST: message"
	if msgs[0].Message != expected {
		t.Errorf("Expected %q, got %q", expected, msgs[0].Message)
	}
}

func testNewStdLoggerFlags(t *testing.T) {
	logger := slogtest.NewLogger()

	// Test various flag combinations
	testCases := []struct {
		name  string
		flags int
	}{
		{"NoFlags", 0},
		{"DateFlag", log.Ldate},
		{"TimeFlag", log.Ltime},
		{"DateTimeFlags", log.Ldate | log.Ltime},
		{"AllFlags", log.LstdFlags},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger.Clear()
			stdLog := slog.NewStdLogger(logger, "", tc.flags)

			if stdLog == nil {
				t.Fatal("NewStdLogger should not return nil")
			}

			stdLog.Print("test")

			msgs := logger.GetMessages()
			if len(msgs) != 1 {
				t.Fatalf("Expected 1 message, got %d", len(msgs))
			}
			// Note: The actual timestamp formatting would require more complex
			// parsing to validate, but we can at least ensure it doesn't crash
		})
	}
}

func testNewStdLoggerEmptyPrefix(t *testing.T) {
	logger := slogtest.NewLogger()
	stdLog := slog.NewStdLogger(logger, "", 0)

	stdLog.Print("no prefix")

	msgs := logger.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Message != "no prefix" {
		t.Errorf("Expected 'no prefix', got %q", msgs[0].Message)
	}
}

func testNewStdLoggerIntegration(t *testing.T) {
	logger := slogtest.NewLogger()
	stdLog := slog.NewStdLogger(logger, "APP", 0)

	// Test multiple logging methods
	stdLog.Print("print message")
	stdLog.Println("println message")
	stdLog.Printf("printf %s %d", "message", 42)

	msgs := logger.GetMessages()
	if len(msgs) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(msgs))
	}

	expected := []string{
		"APP: print message",
		"APP: println message", // Println adds newline but it gets trimmed by LogWriter
		"APP: printf message 42",
	}

	for i, exp := range expected {
		if msgs[i].Message != exp {
			t.Errorf("Message %d: expected %q, got %q", i, exp, msgs[i].Message)
		}
	}
}

func TestStdLogSink(t *testing.T) {
	t.Run("PrefixHandling", testStdLogSinkPrefix)
	t.Run("NoPrefixHandling", testStdLogSinkNoPrefix)
	t.Run("FlagsHandling", testStdLogSinkFlags)
}

func testStdLogSinkPrefix(t *testing.T) {
	logger := slogtest.NewLogger()

	// Create std logger with prefix
	stdLog := slog.NewStdLogger(logger, "PREFIX", 0)

	// Test various message types
	testCases := []struct {
		name     string
		action   func()
		expected string
	}{
		{
			"Print",
			func() { stdLog.Print("test") },
			"PREFIX: test",
		},
		{
			"EmptyMessage",
			func() { stdLog.Print("") },
			"PREFIX: ",
		},
		{
			"WithNewline",
			func() { stdLog.Print("test\n") },
			"PREFIX: test", // LogWriter trims trailing newlines
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger.Clear()
			tc.action()

			msgs := logger.GetMessages()
			if len(msgs) != 1 {
				t.Fatalf("Expected 1 message, got %d", len(msgs))
			}
			if msgs[0].Message != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, msgs[0].Message)
			}
		})
	}
}

func testStdLogSinkNoPrefix(t *testing.T) {
	logger := slogtest.NewLogger()
	stdLog := slog.NewStdLogger(logger, "", 0)

	testCases := []struct {
		name     string
		action   func()
		expected string
	}{
		{
			"Print",
			func() { stdLog.Print("direct") },
			"direct",
		},
		{
			"PrintMultiple",
			func() { stdLog.Print("multiple", " ", "args") },
			"multiple args",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger.Clear()
			tc.action()

			msgs := logger.GetMessages()
			if len(msgs) != 1 {
				t.Fatalf("Expected 1 message, got %d", len(msgs))
			}
			if msgs[0].Message != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, msgs[0].Message)
			}
		})
	}
}

func testStdLogSinkFlags(t *testing.T) {
	logger := slogtest.NewLogger()

	// Test that flags are properly passed to standard logger
	// The actual flag behaviour is handled by Go's log package
	testCases := []int{
		0,
		log.Ldate,
		log.Ltime,
		log.Lmicroseconds,
		log.Llongfile,
		log.Lshortfile,
		log.LUTC,
		log.LstdFlags,
	}

	for _, flags := range testCases {
		t.Run("", func(t *testing.T) {
			logger.Clear()
			stdLog := slog.NewStdLogger(logger, "TEST", flags)

			// Should not panic with any flag combination
			stdLog.Print("test message")

			msgs := logger.GetMessages()
			if len(msgs) != 1 {
				t.Fatalf("Expected 1 message, got %d", len(msgs))
			}
			// Message should contain the prefix and text, but format may vary with flags
			message := msgs[0].Message
			if !strings.Contains(message, "TEST:") || !strings.Contains(message, "test message") {
				t.Errorf("Message should contain prefix and text: %q", message)
			}
		})
	}
}

func TestStdLoggerCompatibility(t *testing.T) {
	t.Run("StdLogInterface", testStdLoggerInterface)
	t.Run("ConcurrentUse", testStdLoggerConcurrent)
}

func testStdLoggerInterface(t *testing.T) {
	logger := slogtest.NewLogger()
	stdLog := slog.NewStdLogger(logger, "", 0)

	// Verify it implements the standard logger interface
	var _ interface {
		Print(...any)
		Printf(string, ...any)
		Println(...any)
	} = stdLog

	// Test all interface methods work
	stdLog.Print("print")
	stdLog.Printf("printf %d", 1)
	stdLog.Println("println")

	msgs := logger.GetMessages()
	if len(msgs) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(msgs))
	}
}

func testStdLoggerConcurrent(t *testing.T) {
	logger := slogtest.NewLogger()
	stdLog := slog.NewStdLogger(logger, "CONCURRENT", 0)

	// Simple concurrent usage test
	done := make(chan bool, 2)

	go func() {
		stdLog.Print("goroutine 1")
		done <- true
	}()

	go func() {
		stdLog.Print("goroutine 2")
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	msgs := logger.GetMessages()
	if len(msgs) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(msgs))
	}

	// Both messages should have prefix
	for i, msg := range msgs {
		if !strings.Contains(msg.Message, "CONCURRENT:") {
			t.Errorf("Message %d should have prefix: %q", i, msg.Message)
		}
	}
}
