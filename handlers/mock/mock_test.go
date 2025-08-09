package mock

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// levelTest represents a test case for log level methods.
type levelTest struct {
	name     string
	logFunc  func(*Logger) slog.Logger
	expected slog.LogLevel
}

func newLevelTest(name string, logFunc func(*Logger) slog.Logger, expected slog.LogLevel) levelTest {
	return levelTest{
		name:     name,
		logFunc:  logFunc,
		expected: expected,
	}
}

func (tc levelTest) test(t *testing.T) {
	logger := NewLogger()
	tc.logFunc(logger).Print("test")

	messages := logger.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	if messages[0].Level != tc.expected {
		t.Errorf("expected level %v, got %v", tc.expected, messages[0].Level)
	}
}

// printTest represents a test case for print methods.
type printTest struct {
	name     string
	logFunc  func(*Logger)
	expected string
}

func newPrintTest(name string, logFunc func(*Logger), expected string) printTest {
	return printTest{
		name:     name,
		logFunc:  logFunc,
		expected: expected,
	}
}

func (tc printTest) test(t *testing.T) {
	logger := NewLogger()
	tc.logFunc(logger)

	messages := logger.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	if messages[0].Message != tc.expected {
		t.Errorf("expected %q, got %q", tc.expected, messages[0].Message)
	}
}

func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}

	if !logger.Enabled() {
		t.Error("NewLogger should be enabled by default")
	}

	messages := logger.GetMessages()
	if len(messages) != 0 {
		t.Errorf("NewLogger should start with no messages, got %d", len(messages))
	}
}

func TestNewRecorder(t *testing.T) {
	recorder := NewRecorder()
	if recorder == nil {
		t.Fatal("NewRecorder returned nil")
	}

	messages := recorder.GetMessages()
	if len(messages) != 0 {
		t.Errorf("NewRecorder should start with no messages, got %d", len(messages))
	}
}

func TestLoggerBasicLogging(t *testing.T) {
	logger := NewLogger()

	logger.Info().Print("test message")

	messages := logger.GetMessages()
	assertMessageCount(t, messages, 1)

	msg := messages[0]
	assertMessage(t, msg, slog.Info, "test message")
	assertFieldCount(t, msg, 0)
	assertStack(t, msg, false)
}

func testLoggerLevels(t *testing.T) {
	tests := []levelTest{
		newLevelTest("Debug", (*Logger).Debug, slog.Debug),
		newLevelTest("Info", (*Logger).Info, slog.Info),
		newLevelTest("Warn", (*Logger).Warn, slog.Warn),
		newLevelTest("Error", (*Logger).Error, slog.Error),
		newLevelTest("Fatal", (*Logger).Fatal, slog.Fatal),
		newLevelTest("Panic", (*Logger).Panic, slog.Panic),
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func testLoggerPrintMethods(t *testing.T) {
	tests := []printTest{
		newPrintTest("Print", func(l *Logger) { l.Info().Print("hello", " ", "world") }, "hello world"),
		newPrintTest("Println", func(l *Logger) { l.Info().Println("hello", "world") }, "hello world\n"),
		newPrintTest("Printf", func(l *Logger) { l.Info().Printf("hello %s", "world") }, "hello world"),
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func TestLoggerMethods(t *testing.T) {
	t.Run("levels", testLoggerLevels)
	t.Run("print", testLoggerPrintMethods)
}

func TestLoggerWithField(t *testing.T) {
	logger := NewLogger()

	logger.Info().
		WithField("key1", "value1").
		WithField("key2", 42).
		Print("test message")

	messages := logger.GetMessages()
	assertMessageCount(t, messages, 1)

	msg := messages[0]
	assertFieldCount(t, msg, 2)
	assertFieldValue(t, msg, "key1", "value1")
	assertFieldValue(t, msg, "key2", 42)
}

func TestLoggerWithFields(t *testing.T) {
	logger := NewLogger()

	fields := map[string]any{
		"key1": "value1",
		"key2": 42,
		"":     "ignored",
	}

	logger.Info().WithFields(fields).Print("test")

	messages := logger.GetMessages()
	assertMessageCount(t, messages, 1)

	msg := messages[0]
	assertFieldCount(t, msg, 2)
	assertFieldValue(t, msg, "key1", "value1")
	assertFieldValue(t, msg, "key2", 42)
}

func TestLoggerWithFieldEmptyKey(t *testing.T) {
	logger := NewLogger()

	logger1 := logger.WithField("", "value")
	logger2 := logger.WithField("key", "value")

	// Empty keys are filtered out by internal.Loglet, so logger1 should be functionally
	// equivalent to logger but may be a different instance
	logger1.Info().Print("test message 1")
	logger2.Info().Print("test message 2")

	messages := logger.GetMessages()

	// Should have 2 messages, one from logger1 (no extra empty field) and one from logger2 (with field)
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	// First message should have no fields (empty key filtered out)
	if len(messages[0].Fields) != 0 {
		t.Errorf("Expected no fields in first message, got %v", messages[0].Fields)
	}

	// Second message should have the "key" field
	if len(messages[1].Fields) != 1 || messages[1].Fields["key"] != "value" {
		t.Errorf("Expected one field 'key'='value' in second message, got %v", messages[1].Fields)
	}

	if logger2 == logger {
		t.Error("WithField with valid key should return new logger instance")
	}
}

func TestLoggerWithStack(t *testing.T) {
	logger := NewLogger()

	logger.Error().WithStack(0).Print("error message")

	messages := logger.GetMessages()
	assertMessageCount(t, messages, 1)
	assertStack(t, messages[0], true)
}

func TestLoggerImmutability(t *testing.T) {
	logger := NewLogger()

	logger1 := logger.WithField("key1", "value1")
	logger2 := logger1.WithField("key2", "value2")

	logger.Info().Print("base")
	logger1.Info().Print("with key1")
	logger2.Info().Print("with key1 and key2")

	messages := logger.GetMessages()
	assertMessageCount(t, messages, 3)

	assertFieldCount(t, messages[0], 0)
	assertFieldCount(t, messages[1], 1)
	assertFieldValue(t, messages[1], "key1", "value1")
	assertFieldCount(t, messages[2], 2)
	assertFieldValue(t, messages[2], "key1", "value1")
	assertFieldValue(t, messages[2], "key2", "value2")
}

func TestLoggerClear(t *testing.T) {
	logger := NewLogger()

	logger.Info().Print("message 1")
	logger.Info().Print("message 2")

	messages := logger.GetMessages()
	assertMessageCount(t, messages, 2)

	logger.Clear()

	messages = logger.GetMessages()
	assertMessageCount(t, messages, 0)
}

func TestLoggerWithEnabled(t *testing.T) {
	logger := NewLogger()

	l, enabled := logger.WithEnabled()
	if !enabled {
		t.Error("logger should be enabled")
	}
	if l != logger {
		t.Error("WithEnabled should return same logger instance")
	}
}

func TestMessageString(t *testing.T) {
	msg := Message{
		Message: "test message",
		Level:   slog.Info,
		Fields: map[string]any{
			"key2": "value2",
			"key1": "value1",
		},
		Stack: false,
	}

	str := msg.String()
	expected := `[5] "test message" key1=value1 key2=value2`
	if str != expected {
		t.Errorf("expected %q, got %q", expected, str)
	}

	msg.Stack = true
	str = msg.String()
	expected = `[5] "test message" key1=value1 key2=value2 [stack]`
	if str != expected {
		t.Errorf("expected %q, got %q", expected, str)
	}
}

func TestRecorderDirectUsage(t *testing.T) {
	recorder := NewRecorder()

	msg1 := Message{Message: "message 1", Level: slog.Info}
	msg2 := Message{Message: "message 2", Level: slog.Error}

	recorder.Record(msg1)
	recorder.Record(msg2)

	messages := recorder.GetMessages()
	assertMessageCount(t, messages, 2)

	if messages[0].Message != "message 1" {
		t.Errorf("expected first message to be 'message 1', got %q", messages[0].Message)
	}
	if messages[1].Message != "message 2" {
		t.Errorf("expected second message to be 'message 2', got %q", messages[1].Message)
	}

	recorder.Clear()
	messages = recorder.GetMessages()
	assertMessageCount(t, messages, 0)
}

type fieldCopyingTest struct {
	name           string
	setupLogger    func() *Logger
	expectedFields map[string]any
}

func (tc fieldCopyingTest) test(t *testing.T) {
	logger := tc.setupLogger()

	// Test that fields are preserved through WithLevel
	loggerWithLevel := logger.WithLevel(slog.Info)
	if loggerWithLevel == nil {
		t.Error("expected non-nil logger")
		return
	}

	// Test that fields are preserved
	loggerWithLevel.Print("test message")
	messages := logger.GetMessages()
	if len(messages) == 0 {
		t.Fatal("expected at least one message")
	}

	msg := messages[len(messages)-1]
	tc.verifyFields(t, msg)
}

func (tc fieldCopyingTest) verifyFields(t *testing.T, msg Message) {
	t.Helper()
	if len(msg.Fields) != len(tc.expectedFields) {
		t.Errorf("expected %d fields, got %d", len(tc.expectedFields), len(msg.Fields))
	}

	for key, expectedValue := range tc.expectedFields {
		actualValue, exists := msg.Fields[key]
		core.AssertMustTrue(t, exists, "field %q exists", key)
		core.AssertEqual(t, expectedValue, actualValue, "field %q", key)
	}
}

func fieldCopyingTestCases() []fieldCopyingTest {
	return []fieldCopyingTest{
		{
			name: "logger with no fields",
			setupLogger: func() *Logger {
				return NewLogger()
			},
			expectedFields: map[string]any{},
		},
		{
			name: "logger with single field",
			setupLogger: func() *Logger {
				logger, ok := NewLogger().WithField("key1", "value1").(*Logger)
				if !ok {
					panic("type assertion failed")
				}
				return logger
			},
			expectedFields: map[string]any{"key1": "value1"},
		},
		{
			name: "logger with multiple fields",
			setupLogger: func() *Logger {
				logger, ok := NewLogger().WithField("key1", "value1").WithField("key2", "value2").(*Logger)
				if !ok {
					panic("type assertion failed")
				}
				return logger
			},
			expectedFields: map[string]any{"key1": "value1", "key2": "value2"},
		},
	}
}

func TestFieldCopying(t *testing.T) {
	for _, tc := range fieldCopyingTestCases() {
		t.Run(tc.name, tc.test)
	}
}

func TestThresholdFiltering(t *testing.T) {
	t.Run("NoThreshold", testThresholdFilteringNoThreshold)
	t.Run("WithThreshold", testThresholdFilteringWithThreshold)
}

func testThresholdFilteringNoThreshold(t *testing.T) {
	// Default NewLogger() should have no threshold (backward compatible)
	logger := NewLogger()

	// All levels should be enabled
	core.AssertTrue(t, logger.Debug().Enabled(), "Debug enabled without threshold")
	core.AssertTrue(t, logger.Info().Enabled(), "Info enabled without threshold")
	core.AssertTrue(t, logger.Error().Enabled(), "Error enabled without threshold")

	// Should record all levels
	logger.Debug().Print("debug message")
	logger.Info().Print("info message")
	logger.Error().Print("error message")

	messages := logger.GetMessages()
	core.AssertEqual(t, 3, len(messages), "all messages recorded without threshold")
}

func testThresholdFilteringWithThreshold(t *testing.T) {
	// Logger with Info threshold should filter out Debug
	logger := NewLoggerWithThreshold(slog.Info)

	// Check enablement
	core.AssertFalse(t, logger.Debug().Enabled(), "Debug disabled with Info threshold")
	core.AssertTrue(t, logger.Info().Enabled(), "Info enabled with Info threshold")
	core.AssertTrue(t, logger.Error().Enabled(), "Error enabled with Info threshold")

	// Try to record all levels
	logger.Debug().Print("debug message")
	logger.Info().Print("info message")
	logger.Error().Print("error message")

	messages := logger.GetMessages()
	core.AssertEqual(t, 2, len(messages), "only Info and above recorded")
	core.AssertEqual(t, slog.Info, messages[0].Level, "first message is Info")
	core.AssertEqual(t, slog.Error, messages[1].Level, "second message is Error")
}

func TestNewLoggerWithThreshold(t *testing.T) {
	logger := NewLoggerWithThreshold(slog.Warn)

	core.AssertNotNil(t, logger, "logger created")
	core.AssertTrue(t, logger.enabled, "logger enabled")
	core.AssertEqual(t, slog.Warn, logger.threshold, "threshold set correctly")
	core.AssertNotNil(t, logger.recorder, "recorder created")
}

func TestNilReceiver(t *testing.T) {
	t.Run("Logger", testNilLogger)
	t.Run("Recorder", testNilRecorder)
}

func testNilLogger(t *testing.T) {
	var logger *Logger

	// Test getter methods
	if logger.Enabled() {
		t.Error("nil logger should return false for Enabled()")
	}

	messages := logger.GetMessages()
	if messages != nil {
		t.Error("nil logger should return nil for GetMessages()")
	}

	// Test void methods - these should not panic
	logger.Clear()
	logger.Print("test")
	logger.Println("test")
	logger.Printf("test %s", "format")

	// Test builder methods
	testNilLoggerBuilders(t, logger)

	l, enabled := logger.WithEnabled()
	if l != nil || enabled {
		t.Error("nil logger should return nil, false for WithEnabled()")
	}
}

func testNilLoggerBuilders(t *testing.T, logger *Logger) {
	testNilLoggerLevels(t, logger)
	testNilLoggerModifiers(t, logger)
}

func testNilLoggerLevels(t *testing.T, logger *Logger) {
	t.Helper()
	if result := logger.Debug(); result != nil {
		t.Error("nil logger should return nil for Debug()")
	}
	if result := logger.Info(); result != nil {
		t.Error("nil logger should return nil for Info()")
	}
	if result := logger.Warn(); result != nil {
		t.Error("nil logger should return nil for Warn()")
	}
	if result := logger.Error(); result != nil {
		t.Error("nil logger should return nil for Error()")
	}
	if result := logger.Fatal(); result != nil {
		t.Error("nil logger should return nil for Fatal()")
	}
	if result := logger.Panic(); result != nil {
		t.Error("nil logger should return nil for Panic()")
	}
}

func testNilLoggerModifiers(t *testing.T, logger *Logger) {
	t.Helper()
	if result := logger.WithField("key", "value"); result != nil {
		t.Error("nil logger should return nil for WithField()")
	}
	if result := logger.WithFields(map[string]any{"key": "value"}); result != nil {
		t.Error("nil logger should return nil for WithFields()")
	}
	if result := logger.WithStack(0); result != nil {
		t.Error("nil logger should return nil for WithStack()")
	}
	if result := logger.WithLevel(slog.Info); result != nil {
		t.Error("nil logger should return nil for WithLevel()")
	}
}

func testNilRecorder(t *testing.T) {
	var recorder *Recorder

	messages := recorder.GetMessages()
	if messages != nil {
		t.Error("nil recorder should return nil for GetMessages()")
	}

	// These should not panic
	recorder.Clear()
	recorder.Record(Message{Message: "test", Level: slog.Info})
}

// Test helper functions

func assertMessageCount(t *testing.T, messages []Message, expected int) {
	t.Helper()
	if len(messages) != expected {
		t.Errorf("expected %d messages, got %d", expected, len(messages))
	}
}

func assertMessage(t *testing.T, msg Message, level slog.LogLevel, text string) {
	t.Helper()
	if msg.Level != level {
		t.Errorf("expected level %v, got %v", level, msg.Level)
	}
	if msg.Message != text {
		t.Errorf("expected message %q, got %q", text, msg.Message)
	}
}

func assertFieldCount(t *testing.T, msg Message, expected int) {
	t.Helper()
	if len(msg.Fields) != expected {
		t.Errorf("expected %d fields, got %d", expected, len(msg.Fields))
	}
}

func assertFieldValue(t *testing.T, msg Message, key string, expected any) {
	t.Helper()
	if msg.Fields[key] != expected {
		t.Errorf("expected %s=%v, got %v", key, expected, msg.Fields[key])
	}
}

//revive:disable-next-line:flag-parameter
func assertStack(t *testing.T, msg Message, expected bool) {
	t.Helper()
	if msg.Stack != expected {
		t.Errorf("expected Stack to be %v, got %v", expected, msg.Stack)
	}
}
