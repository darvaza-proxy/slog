package logr

import (
	"bytes"
	"strings"
	"testing"

	"darvaza.org/slog"
	slogtest "darvaza.org/slog/internal/testing"
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

	t.Run("DisabledLogger", func(t *testing.T) {
		testDisabledLoggerLoglet(t, &buf)
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

// testDisabledLoggerLoglet tests disabled logger behaviour with Loglet
func testDisabledLoggerLoglet(t *testing.T, buf *bytes.Buffer) {
	t.Helper()
	// Create a disabled logger with verbosity -1 (nothing enabled)
	disabledLogger := createDisabledLogger(buf)

	// Test that disabled logger methods don't produce output
	buf.Reset()
	disabledLogger.Debug().Print("debug message")
	if buf.Len() > 0 {
		t.Error("Disabled Debug logger should not produce output")
	}

	buf.Reset()
	disabledLogger.Info().Print("info message")
	if buf.Len() > 0 {
		t.Error("Disabled Info logger should not produce output")
	}

	buf.Reset()
	disabledLogger.Warn().Print("warn message")
	if buf.Len() > 0 {
		t.Error("Disabled Warn logger should not produce output")
	}

	// Error should still work (always enabled in logr)
	buf.Reset()
	disabledLogger.Error().Print("error message")
	if !strings.Contains(buf.String(), "error message") {
		t.Error("Error logger should work even when disabled")
	}
}

// RunWithRecorder is a helper that runs test functions with a logger and recorder setup
func RunWithRecorder(
	name string,
	t *testing.T,
	testFunc func(t *testing.T, logger slog.Logger, recorder *slogtest.Logger),
) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		recorder := slogtest.NewLogger()

		// Create logr logger that outputs to the recorder
		logrLogger := NewLogr(recorder)
		logger := New(logrLogger)

		testFunc(t, logger, recorder)
	})
}

// testBasicLoggingWithRecorder tests basic logging functionality with recorder
func testBasicLoggingWithRecorder(t *testing.T, logger slog.Logger, recorder *slogtest.Logger) {
	t.Helper()
	recorder.Clear()
	logger.Info().Print("test message")

	messages := recorder.GetMessages()
	slogtest.AssertMessageCount(t, messages, 1)
	slogtest.AssertMessage(t, messages[0], slog.Info, "test message")
}

// testFieldPreservation tests field preservation with recorder
func testFieldPreservation(t *testing.T, logger slog.Logger, recorder *slogtest.Logger) {
	t.Helper()
	recorder.Clear()
	logger.Info().
		WithField("key1", "value1").
		WithField("key2", 42).
		Print("message with fields")

	messages := recorder.GetMessages()
	slogtest.AssertMessageCount(t, messages, 1)
	slogtest.AssertMessage(t, messages[0], slog.Info, "message with fields")
	slogtest.AssertField(t, messages[0], "key1", "value1")
	slogtest.AssertField(t, messages[0], "key2", 42)
}

// testLevelMapping tests level mapping functionality with recorder
func testLevelMapping(t *testing.T, logger slog.Logger, recorder *slogtest.Logger) {
	t.Helper()
	testCases := []struct {
		name     string
		level    slog.LogLevel
		expected slog.LogLevel
	}{
		{"Debug", slog.Debug, slog.Debug},
		{"Info", slog.Info, slog.Info},
		{"Warn", slog.Warn, slog.Info}, // logr maps Warn to Info
		{"Error", slog.Error, slog.Error},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder.Clear()
			logger.WithLevel(tc.level).Print("test message")

			messages := recorder.GetMessages()
			slogtest.AssertMessageCount(t, messages, 1)
			slogtest.AssertMessage(t, messages[0], tc.expected, "test message")
		})
	}
}

// TestLoggerWithRecorder tests Logger using test recorder for better verification
func TestLoggerWithRecorder(t *testing.T) {
	RunWithRecorder("BasicLogging", t, testBasicLoggingWithRecorder)
	RunWithRecorder("FieldPreservation", t, testFieldPreservation)
	RunWithRecorder("LevelMapping", t, testLevelMapping)
}

// TestLoggerFieldIsolation tests that field modifications don't affect other loggers
func TestLoggerFieldIsolation(t *testing.T) {
	recorder := slogtest.NewLogger()
	logrLogger := NewLogr(recorder)
	logger := New(logrLogger)

	// Create base logger with one field
	base := logger.WithField("base", "value")

	// Create two branches from base
	branch1 := base.WithField("branch", "1")
	branch2 := base.WithField("branch", "2")

	// Test that branches are independent
	recorder.Clear()
	branch1.Info().Print("branch 1 message")
	branch2.Info().Print("branch 2 message")

	messages := recorder.GetMessages()
	slogtest.AssertMessageCount(t, messages, 2)

	// Verify first message has correct fields
	slogtest.AssertMessage(t, messages[0], slog.Info, "branch 1 message")
	slogtest.AssertField(t, messages[0], "base", "value")
	slogtest.AssertField(t, messages[0], "branch", "1")

	// Verify second message has correct fields
	slogtest.AssertMessage(t, messages[1], slog.Info, "branch 2 message")
	slogtest.AssertField(t, messages[1], "base", "value")
	slogtest.AssertField(t, messages[1], "branch", "2")
}

// TestLoggerEmptyFields tests handling of empty field keys
func TestLoggerEmptyFields(t *testing.T) {
	recorder := slogtest.NewLogger()
	logrLogger := NewLogr(recorder)
	logger := New(logrLogger)

	// Test WithField with empty key
	recorder.Clear()
	logger.Info().WithField("", "should-be-ignored").WithField("valid", "value").Print("test")

	messages := recorder.GetMessages()
	slogtest.AssertMessageCount(t, messages, 1)
	slogtest.AssertMessage(t, messages[0], slog.Info, "test")
	slogtest.AssertField(t, messages[0], "valid", "value")
	slogtest.AssertNoField(t, messages[0], "")

	// Test WithFields with empty keys
	recorder.Clear()
	fields := map[string]any{
		"":      "ignored",
		"valid": "kept",
	}
	logger.Info().WithFields(fields).Print("test fields")

	messages = recorder.GetMessages()
	slogtest.AssertMessageCount(t, messages, 1)
	slogtest.AssertMessage(t, messages[0], slog.Info, "test fields")
	slogtest.AssertField(t, messages[0], "valid", "kept")
	slogtest.AssertNoField(t, messages[0], "")
}
