package logrus_test

import (
	"testing"

	"github.com/sirupsen/logrus"

	"darvaza.org/slog"
	sloglogrus "darvaza.org/slog/handlers/logrus"
	slogtest "darvaza.org/slog/internal/testing"
)

// TestSlogHook tests the basic functionality of the slog hook
func TestSlogHook(t *testing.T) {
	// Create a test logger to capture output
	recorder := slogtest.NewLogger()

	// Create a logrus logger that outputs to our slog recorder
	logrusLogger := sloglogrus.NewLogrusLogger(recorder)

	// Test basic logging
	logrusLogger.Info("test message")

	// Check the output
	messages := recorder.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(messages))
	}

	msg := messages[0]
	if msg.Level != slog.Info {
		t.Errorf("Expected Info level, got %v", msg.Level)
	}
	if msg.Message != "test message" {
		t.Errorf("Expected 'test message', got %q", msg.Message)
	}
}

// TestSlogHookWithFields tests field handling
func TestSlogHookWithFields(t *testing.T) {
	recorder := slogtest.NewLogger()
	logrusLogger := sloglogrus.NewLogrusLogger(recorder)

	// Log with fields
	logrusLogger.WithFields(logrus.Fields{
		"key1": "value1",
		"key2": 42,
	}).Info("test with fields")

	// Check the output
	messages := recorder.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(messages))
	}

	msg := messages[0]
	if msg.Fields["key1"] != "value1" {
		t.Errorf("Expected field key1=value1, got %v", msg.Fields["key1"])
	}
	if msg.Fields["key2"] != 42 {
		t.Errorf("Expected field key2=42, got %v", msg.Fields["key2"])
	}
}

// TestSlogHookLevels tests level mapping
func TestSlogHookLevels(t *testing.T) {
	recorder := slogtest.NewLogger()
	logrusLogger := sloglogrus.NewLogrusLogger(recorder)

	// Set to trace level to capture all
	logrusLogger.SetLevel(logrus.TraceLevel)

	tests := []struct {
		logFunc  func(args ...interface{})
		logLevel logrus.Level
		expected slog.LogLevel
	}{
		{logrusLogger.Trace, logrus.TraceLevel, slog.Debug},
		{logrusLogger.Debug, logrus.DebugLevel, slog.Debug},
		{logrusLogger.Info, logrus.InfoLevel, slog.Info},
		{logrusLogger.Warn, logrus.WarnLevel, slog.Warn},
		{logrusLogger.Error, logrus.ErrorLevel, slog.Error},
	}

	for _, test := range tests {
		recorder.Clear()
		test.logFunc("test")

		messages := recorder.GetMessages()
		if len(messages) != 1 {
			t.Errorf("Expected 1 message for level %v, got %d", test.logLevel, len(messages))
			continue
		}

		if messages[0].Level != test.expected {
			t.Errorf("Expected slog level %v for logrus level %v, got %v",
				test.expected, test.logLevel, messages[0].Level)
		}
	}
}

// TestSlogHookWithError tests error field handling
func TestSlogHookWithError(t *testing.T) {
	recorder := slogtest.NewLogger()
	logrusLogger := sloglogrus.NewLogrusLogger(recorder)

	// Create a test error
	testErr := &testError{msg: "test error"}

	// Log with error
	logrusLogger.WithError(testErr).Error("operation failed")

	// Check the output
	messages := recorder.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(messages))
	}

	msg := messages[0]
	if msg.Fields["error"] != testErr {
		t.Errorf("Expected error field to contain test error, got %v", msg.Fields["error"])
	}
}

// TestBidirectionalAdapter tests using both adapters together
func TestBidirectionalAdapter(t *testing.T) {
	// Test slog -> logrus -> slog round trip
	slogtest.TestBidirectional(t, "SlogToLogrus", func(backend slog.Logger) slog.Logger {
		// Create a logrus logger that outputs to the backend
		logrusLogger := sloglogrus.NewLogrusLogger(backend)
		// Wrap it back as slog
		return sloglogrus.New(logrusLogger)
	})
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
