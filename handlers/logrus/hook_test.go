package logrus_test

import (
	"testing"

	"github.com/sirupsen/logrus"

	"darvaza.org/core"
	"darvaza.org/slog"
	slogrus "darvaza.org/slog/handlers/logrus"
	slogtest "darvaza.org/slog/internal/testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = slogHookLevelTestCase{}

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
	slogtest.AssertMustMessageCount(t, messages, 1)

	msg := messages[0]
	slogtest.AssertMessage(t, msg, slog.Info, "test message")
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
	slogtest.AssertMustMessageCount(t, messages, 1)

	msg := messages[0]
	slogtest.AssertField(t, msg, "key1", "value1")
	slogtest.AssertField(t, msg, "key2", 42)
}

// TestSlogHookLevels tests level mapping
func TestSlogHookLevels(t *testing.T) {
	recorder := slogtest.NewLogger()
	logrusLogger := slogrus.NewLogrusLogger(recorder)

	// Set to trace level to capture all
	logrusLogger.SetLevel(logrus.TraceLevel)

	// Test level mapping using TestCase pattern
	core.RunTestCases(t, slogHookLevelTestCases(logrusLogger, recorder))
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
	slogtest.AssertMustMessageCount(t, messages, 1)

	msg := messages[0]
	slogtest.AssertField(t, msg, "error", testErr)
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
	core.AssertNotNil(t, err, "panic error")

	core.AssertContains(t, err.Error(), "test panic", "error message")
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// mockPanicLogger is a mock slog.Logger that panics on Print
type mockPanicLogger struct{}

func (m *mockPanicLogger) WithLevel(_ slog.LogLevel) slog.Logger   { return m }
func (m *mockPanicLogger) WithStack(_ int) slog.Logger             { return m }
func (m *mockPanicLogger) WithField(_ string, _ any) slog.Logger   { return m }
func (m *mockPanicLogger) WithFields(_ map[string]any) slog.Logger { return m }
func (m *mockPanicLogger) Print(_ ...any)                          { panic("test panic") }
func (m *mockPanicLogger) Println(_ ...any)                        { panic("test panic") }
func (m *mockPanicLogger) Printf(_ string, _ ...any)               { panic("test panic") }
func (m *mockPanicLogger) Debug() slog.Logger                      { return m }
func (m *mockPanicLogger) Info() slog.Logger                       { return m }
func (m *mockPanicLogger) Warn() slog.Logger                       { return m }
func (m *mockPanicLogger) Error() slog.Logger                      { return m }
func (m *mockPanicLogger) Fatal() slog.Logger                      { return m }
func (m *mockPanicLogger) Panic() slog.Logger                      { return m }
func (m *mockPanicLogger) Enabled() bool                           { return true }
func (m *mockPanicLogger) WithEnabled() (slog.Logger, bool)        { return m, true }

type slogHookLevelTestCase struct {
	logFunc      func(args ...any)
	logrusLogger *logrus.Logger
	recorder     *slogtest.Logger
	name         string
	logLevel     logrus.Level
	expected     slog.LogLevel
}

func (tc slogHookLevelTestCase) Name() string {
	return tc.name
}

func (tc slogHookLevelTestCase) Test(t *testing.T) {
	t.Helper()
	tc.recorder.Clear()
	tc.logFunc("test")

	messages := tc.recorder.GetMessages()
	core.AssertMustEqual(t, 1, len(messages), "message count")

	core.AssertEqual(t, tc.expected, messages[0].Level, "slog level")
}

func newSlogHookLevelTestCase(
	name string, logFunc func(args ...any), logLevel logrus.Level,
	expected slog.LogLevel, logrusLogger *logrus.Logger, recorder *slogtest.Logger,
) slogHookLevelTestCase {
	return slogHookLevelTestCase{
		name:         name,
		logFunc:      logFunc,
		logLevel:     logLevel,
		expected:     expected,
		logrusLogger: logrusLogger,
		recorder:     recorder,
	}
}

func slogHookLevelTestCases(logrusLogger *logrus.Logger, recorder *slogtest.Logger) []slogHookLevelTestCase {
	return []slogHookLevelTestCase{
		newSlogHookLevelTestCase("Trace", logrusLogger.Trace, logrus.TraceLevel, slog.Debug, logrusLogger, recorder),
		newSlogHookLevelTestCase("Debug", logrusLogger.Debug, logrus.DebugLevel, slog.Debug, logrusLogger, recorder),
		newSlogHookLevelTestCase("Info", logrusLogger.Info, logrus.InfoLevel, slog.Info, logrusLogger, recorder),
		newSlogHookLevelTestCase("Warn", logrusLogger.Warn, logrus.WarnLevel, slog.Warn, logrusLogger, recorder),
		newSlogHookLevelTestCase("Error", logrusLogger.Error, logrus.ErrorLevel, slog.Error, logrusLogger, recorder),
	}
}
