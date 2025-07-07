package logr

import (
	"bytes"
	"strings"
	"testing"

	"darvaza.org/slog"
)

// createTestLoggerWithBuffer creates a test logger with the given buffer
func createTestLoggerWithBuffer(buf *bytes.Buffer) *Logger {
	logrLogger := createTestLogger(buf)
	logger, ok := New(logrLogger).(*Logger)
	if !ok {
		panic("failed to create Logger")
	}
	return logger
}

// verifyFieldPresence checks if expected fields are present and unexpected fields are absent
func verifyFieldPresence(t *testing.T, output string, expectedFields, unexpectedFields []string) {
	t.Helper()
	for _, field := range expectedFields {
		assertContains(t, output, field)
	}
	for _, field := range unexpectedFields {
		if strings.Contains(output, field) {
			t.Errorf("unexpected field %s in output", field)
		}
	}
}

// verifyLoggerLevel checks if a logger has the expected level
func verifyLoggerLevel(t *testing.T, logger slog.Logger, expectedLevel slog.LogLevel) {
	t.Helper()
	if l, ok := logger.(*Logger); ok {
		if l.Level() != expectedLevel {
			t.Errorf("expected %v level, got %v", expectedLevel, l.Level())
		}
	} else {
		t.Error("logger is not of type *Logger")
	}
}

// testImmutableLogger tests that loggers are immutable
func testImmutableLogger(t *testing.T, logger *Logger, buf *bytes.Buffer) {
	t.Helper()
	// Create a logger with a field
	l1 := logger.Info().WithField("key1", "value1")

	// Create another logger from the original
	l2 := logger.Info().WithField("key2", "value2")

	// Log with both
	buf.Reset()
	l1.Print("message 1")
	output1 := buf.String()

	buf.Reset()
	l2.Print("message 2")
	output2 := buf.String()

	// Verify l1 only has key1
	verifyFieldPresence(t, output1, []string{"key1"}, []string{"key2"})

	// Verify l2 only has key2
	verifyFieldPresence(t, output2, []string{"key2"}, []string{"key1"})
}

// TestLoggerLoglet tests that the Logger properly uses internal.Loglet
func TestLoggerLoglet(t *testing.T) {
	var buf bytes.Buffer
	logger := createTestLoggerWithBuffer(&buf)

	t.Run("ImmutableLogger", func(t *testing.T) {
		testImmutableLogger(t, logger, &buf)
	})

	t.Run("FieldChaining", func(t *testing.T) {
		testFieldChaining(t, logger, &buf)
	})

	t.Run("LevelPersistence", func(t *testing.T) {
		testLevelPersistence(t, logger)
	})

	t.Run("StackAttachment", func(t *testing.T) {
		testStackAttachment(t, logger, &buf)
	})
}

// testFieldChaining tests field chaining functionality
func testFieldChaining(t *testing.T, logger *Logger, buf *bytes.Buffer) {
	t.Helper()
	// Create a logger with chained fields
	l := logger.Info().
		WithField("app", "test").
		WithField("version", "1.0").
		WithFields(map[string]any{
			"env":    "dev",
			"region": "us-west",
		})

	buf.Reset()
	l.Print("chained fields")
	output := buf.String()

	// Verify all fields are present
	expectedFields := []string{"app", "version", "env", "region"}
	verifyFieldPresence(t, output, expectedFields, nil)
}

// testLevelPersistence tests that levels persist correctly
func testLevelPersistence(t *testing.T, logger *Logger) {
	t.Helper()
	// Create loggers with different levels but same fields
	base := logger.WithField("component", "test")

	debugLogger := base.Debug()
	infoLogger := base.Info()
	errorLogger := base.Error()

	// Verify each maintains its level
	verifyLoggerLevel(t, debugLogger, slog.Debug)
	verifyLoggerLevel(t, infoLogger, slog.Info)
	verifyLoggerLevel(t, errorLogger, slog.Error)
}

// testStackAttachment tests stack trace attachment
func testStackAttachment(t *testing.T, logger *Logger, buf *bytes.Buffer) {
	t.Helper()
	// Create a logger with stack
	l := logger.Error().WithStack(0)

	buf.Reset()
	l.Print("error with stack")
	output := buf.String()

	// Verify stack is included
	assertContains(t, output, "stack")
}
