package mock

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
)

func TestNewPublicMethods(t *testing.T) {
	// Test nil receiver for Loglet shortcut methods
	var nilLogger *Logger
	core.AssertEqual(t, slog.UndefinedLevel, nilLogger.Level(), "nil logger level")
	core.AssertNil(t, nilLogger.CallStack(), "nil logger call stack")
	core.AssertEqual(t, 0, nilLogger.FieldsCount(), "nil logger fields count")
	core.AssertNil(t, nilLogger.Fields(), "nil logger fields iterator")

	// Test normal logger
	logger := NewLogger()

	// Test Level method
	core.AssertEqual(t, slog.UndefinedLevel, logger.Level(), "default level")
	infoLogger := core.AssertMustTypeIs[*Logger](t, logger.Info(), "info logger type")
	core.AssertEqual(t, slog.Info, infoLogger.Level(), "info level")

	// Test CallStack method
	core.AssertEqual(t, 0, len(logger.CallStack()), "empty stack")
	stackLogger := core.AssertMustTypeIs[*Logger](t, logger.WithStack(0), "stack logger type")
	core.AssertTrue(t, len(stackLogger.CallStack()) > 0, "has stack")

	// Test FieldsCount method
	core.AssertEqual(t, 0, logger.FieldsCount(), "no fields")
	fieldLogger := core.AssertMustTypeIs[*Logger](t, logger.WithField("key", "value"), "field logger type")
	core.AssertEqual(t, 1, fieldLogger.FieldsCount(), "one field")

	// Test Fields method
	iter := logger.Fields()
	core.AssertNotNil(t, iter, "fields iterator")
	core.AssertFalse(t, iter.Next(), "no fields to iterate")

	iter2 := fieldLogger.Fields()
	core.AssertNotNil(t, iter2, "has fields iterator")
	core.AssertTrue(t, iter2.Next(), "iterator has next")
	k, v := iter2.Field()
	core.AssertEqual(t, "key", k, "field key")
	core.AssertEqual(t, "value", v, "field value")
}

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = levelTestCase{}
var _ core.TestCase = printTestCase{}
var _ core.TestCase = fieldCopyingTestCase{}

// levelTestCase represents a test case for log level methods.
type levelTestCase struct {
	logFunc  func(*Logger) slog.Logger
	expected slog.LogLevel
	name     string
}

func newLevelTestCase(name string, logFunc func(*Logger) slog.Logger, expected slog.LogLevel) levelTestCase {
	return levelTestCase{
		name:     name,
		logFunc:  logFunc,
		expected: expected,
	}
}

func (tc levelTestCase) Name() string {
	return tc.name
}

func (tc levelTestCase) Test(t *testing.T) {
	t.Helper()
	logger := NewLogger()
	tc.logFunc(logger).Print("test")

	messages := logger.GetMessages()
	core.AssertMustEqual(t, 1, len(messages), "message count")
	core.AssertEqual(t, tc.expected, messages[0].Level, "log level")
}

// printTestCase represents a test case for print methods.
type printTestCase struct {
	logFunc  func(*Logger)
	expected string
	name     string
}

func newPrintTestCase(name string, logFunc func(*Logger), expected string) printTestCase {
	return printTestCase{
		name:     name,
		logFunc:  logFunc,
		expected: expected,
	}
}

func (tc printTestCase) Name() string {
	return tc.name
}

func (tc printTestCase) Test(t *testing.T) {
	t.Helper()
	logger := NewLogger()
	tc.logFunc(logger)

	messages := logger.GetMessages()
	core.AssertMustEqual(t, 1, len(messages), "message count")
	core.AssertEqual(t, tc.expected, messages[0].Message, "log message")
}

func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	core.AssertMustNotNil(t, logger, "logger")

	core.AssertTrue(t, logger.Enabled(), "enabled by default")

	messages := logger.GetMessages()
	core.AssertEqual(t, 0, len(messages), "initial message count")
}

func TestNewRecorder(t *testing.T) {
	recorder := NewRecorder()
	core.AssertMustNotNil(t, recorder, "recorder")

	messages := recorder.GetMessages()
	core.AssertEqual(t, 0, len(messages), "initial message count")
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
	tests := []levelTestCase{
		newLevelTestCase("Debug", (*Logger).Debug, slog.Debug),
		newLevelTestCase("Info", (*Logger).Info, slog.Info),
		newLevelTestCase("Warn", (*Logger).Warn, slog.Warn),
		newLevelTestCase("Error", (*Logger).Error, slog.Error),
		newLevelTestCase("Fatal", (*Logger).Fatal, slog.Fatal),
		newLevelTestCase("Panic", (*Logger).Panic, slog.Panic),
	}

	core.RunTestCases(t, tests)
}

func testLoggerPrintMethods(t *testing.T) {
	tests := []printTestCase{
		newPrintTestCase("Print", func(l *Logger) { l.Info().Print("hello", " ", "world") }, "hello world"),
		newPrintTestCase("Println", func(l *Logger) { l.Info().Println("hello", "world") }, "hello world\n"),
		newPrintTestCase("Printf", func(l *Logger) { l.Info().Printf("hello %s", "world") }, "hello world"),
	}

	core.RunTestCases(t, tests)
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
	core.AssertEqual(t, 2, len(messages), "message count")

	// First message should have no fields (empty key filtered out)
	core.AssertEqual(t, 0, len(messages[0].Fields), "first message field count")

	// Second message should have the "key" field
	core.AssertEqual(t, 1, len(messages[1].Fields), "second message field count")
	core.AssertEqual(t, "value", messages[1].Fields["key"], "second message key field")

	core.AssertTrue(t, logger2 != logger, "WithField returns new instance")
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
	core.AssertTrue(t, enabled, "logger enabled")
	core.AssertTrue(t, l == logger, "WithEnabled returns same instance")
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
	core.AssertEqual(t, expected, str, "message string without stack")

	msg.Stack = true
	str = msg.String()
	expected = `[5] "test message" key1=value1 key2=value2 [stack]`
	core.AssertEqual(t, expected, str, "message string with stack")
}

func TestRecorderDirectUsage(t *testing.T) {
	recorder := NewRecorder()

	msg1 := Message{Message: "message 1", Level: slog.Info}
	msg2 := Message{Message: "message 2", Level: slog.Error}

	recorder.Record(msg1)
	recorder.Record(msg2)

	messages := recorder.GetMessages()
	assertMessageCount(t, messages, 2)

	core.AssertEqual(t, "message 1", messages[0].Message, "first message")
	core.AssertEqual(t, "message 2", messages[1].Message, "second message")

	recorder.Clear()
	messages = recorder.GetMessages()
	assertMessageCount(t, messages, 0)
}

type fieldCopyingTestCase struct {
	setupLogger    func() *Logger
	expectedFields map[string]any
	name           string
}

func (tc fieldCopyingTestCase) Name() string {
	return tc.name
}

func (tc fieldCopyingTestCase) Test(t *testing.T) {
	t.Helper()
	logger := tc.setupLogger()

	// Test that fields are preserved through WithLevel
	loggerWithLevel := logger.WithLevel(slog.Info)
	core.AssertMustNotNil(t, loggerWithLevel, "logger with level")

	// Test that fields are preserved
	loggerWithLevel.Print("test message")
	messages := logger.GetMessages()
	core.AssertMustTrue(t, len(messages) > 0, "has messages")

	msg := messages[len(messages)-1]
	tc.verifyFields(t, msg)
}

func (tc fieldCopyingTestCase) verifyFields(t *testing.T, msg Message) {
	t.Helper()
	core.AssertEqual(t, len(tc.expectedFields), len(msg.Fields), "field count")

	for key, expectedValue := range tc.expectedFields {
		actualValue, exists := msg.Fields[key]
		core.AssertMustTrue(t, exists, "field %q exists", key)
		core.AssertEqual(t, expectedValue, actualValue, "field %q", key)
	}
}

func fieldCopyingTestCases() []fieldCopyingTestCase {
	return []fieldCopyingTestCase{
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
	core.RunTestCases(t, fieldCopyingTestCases())
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
	core.AssertFalse(t, logger.Enabled(), "nil logger enabled")

	messages := logger.GetMessages()
	core.AssertNil(t, messages, "nil logger messages")

	// Test void methods - these should not panic
	logger.Clear()
	logger.Print("test")
	logger.Println("test")
	logger.Printf("test %s", "format")

	// Test builder methods
	testNilLoggerBuilders(t, logger)

	l, enabled := logger.WithEnabled()
	core.AssertNil(t, l, "nil logger WithEnabled result")
	core.AssertFalse(t, enabled, "nil logger WithEnabled status")
}

func testNilLoggerBuilders(t *testing.T, logger *Logger) {
	testNilLoggerLevels(t, logger)
	testNilLoggerModifiers(t, logger)
}

func testNilLoggerLevels(t *testing.T, logger *Logger) {
	t.Helper()
	core.AssertNil(t, logger.Debug(), "nil logger Debug")
	core.AssertNil(t, logger.Info(), "nil logger Info")
	core.AssertNil(t, logger.Warn(), "nil logger Warn")
	core.AssertNil(t, logger.Error(), "nil logger Error")
	core.AssertNil(t, logger.Fatal(), "nil logger Fatal")
	core.AssertNil(t, logger.Panic(), "nil logger Panic")
}

func testNilLoggerModifiers(t *testing.T, logger *Logger) {
	t.Helper()
	core.AssertNil(t, logger.WithField("key", "value"), "nil logger WithField")
	core.AssertNil(t, logger.WithFields(map[string]any{"key": "value"}), "nil logger WithFields")
	core.AssertNil(t, logger.WithStack(0), "nil logger WithStack")
	core.AssertNil(t, logger.WithLevel(slog.Info), "nil logger WithLevel")
}

func testNilRecorder(t *testing.T) {
	var recorder *Recorder

	messages := recorder.GetMessages()
	core.AssertNil(t, messages, "nil recorder messages")

	// These should not panic
	recorder.Clear()
	recorder.Record(Message{Message: "test", Level: slog.Info})
}

// Test helper functions

func assertMessageCount(t *testing.T, messages []Message, expected int) {
	t.Helper()
	core.AssertEqual(t, expected, len(messages), "message count")
}

func assertMessage(t *testing.T, msg Message, level slog.LogLevel, text string) {
	t.Helper()
	core.AssertEqual(t, level, msg.Level, "log level")
	core.AssertEqual(t, text, msg.Message, "log message")
}

func assertFieldCount(t *testing.T, msg Message, expected int) {
	t.Helper()
	core.AssertEqual(t, expected, len(msg.Fields), "field count")
}

func assertFieldValue(t *testing.T, msg Message, key string, expected any) {
	t.Helper()
	core.AssertEqual(t, expected, msg.Fields[key], "field %s", key)
}

//revive:disable-next-line:flag-parameter
func assertStack(t *testing.T, msg Message, expected bool) {
	t.Helper()
	core.AssertEqual(t, expected, msg.Stack, "stack trace")
}
