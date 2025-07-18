package logrus_test

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"darvaza.org/slog"
	slogrus "darvaza.org/slog/handlers/logrus"
	slogtest "darvaza.org/slog/internal/testing"
)

// TestSlogHook tests the basic functionality of the slog hook
func TestSlogHook(t *testing.T) {
	// Create a test logger to capture output
	recorder := slogtest.NewLogger()

	// Create a logrus logger that outputs to our slog recorder
	logrusLogger := slogrus.NewLogrusLogger(recorder)

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
	logrusLogger := slogrus.NewLogrusLogger(recorder)

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
	logrusLogger := slogrus.NewLogrusLogger(recorder)

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
	logrusLogger := slogrus.NewLogrusLogger(recorder)

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
		logrusLogger := slogrus.NewLogrusLogger(backend)
		// Wrap it back as slog
		return slogrus.New(logrusLogger)
	})
}

// TestSlogHookPanicRecovery tests that panics in slog operations are recovered
func TestSlogHookPanicRecovery(t *testing.T) {
	// Create a mock slog logger that panics
	panicLogger := &mockPanicLogger{}
	hook := slogrus.NewSlogHook(panicLogger)

	// Create a logrus entry
	entry := &logrus.Entry{
		Level:   logrus.InfoLevel,
		Message: "test message",
		Data:    logrus.Fields{},
	}

	// Fire the hook - should recover from panic and return error
	err := hook.Fire(entry)
	if err == nil {
		t.Error("Expected error from panicking slog logger")
	}

	if !strings.Contains(err.Error(), "test panic") {
		t.Errorf("Expected panic error message, got: %v", err)
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// mockPanicLogger is a mock slog.Logger that panics on Print
type mockPanicLogger struct{}

func (m *mockPanicLogger) WithLevel(_ slog.LogLevel) slog.Logger         { return m }
func (m *mockPanicLogger) WithStack(_ int) slog.Logger                   { return m }
func (m *mockPanicLogger) WithField(_ string, _ interface{}) slog.Logger { return m }
func (m *mockPanicLogger) WithFields(_ map[string]any) slog.Logger       { return m }
func (m *mockPanicLogger) Print(_ ...any)                                { panic("test panic") }
func (m *mockPanicLogger) Println(_ ...any)                              { panic("test panic") }
func (m *mockPanicLogger) Printf(_ string, _ ...any)                     { panic("test panic") }
func (m *mockPanicLogger) Debug() slog.Logger                            { return m }
func (m *mockPanicLogger) Info() slog.Logger                             { return m }
func (m *mockPanicLogger) Warn() slog.Logger                             { return m }
func (m *mockPanicLogger) Error() slog.Logger                            { return m }
func (m *mockPanicLogger) Fatal() slog.Logger                            { return m }
func (m *mockPanicLogger) Panic() slog.Logger                            { return m }
func (m *mockPanicLogger) Enabled() bool                                 { return true }
func (m *mockPanicLogger) WithEnabled() (slog.Logger, bool)              { return m, true }
