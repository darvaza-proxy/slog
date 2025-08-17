package slog_test

import (
	"log"
	"strings"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/mock"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = stdLoggerFlagsTestCase{}
var _ core.TestCase = stdLogSinkPrefixTestCase{}
var _ core.TestCase = stdLogSinkNoPrefixTestCase{}
var _ core.TestCase = stdLogSinkFlagsTestCase{}

func TestNewStdLogger(t *testing.T) {
	t.Run("BasicCreation", testNewStdLoggerBasic)
	t.Run("WithPrefix", testNewStdLoggerPrefix)
	t.Run("WithFlags", testNewStdLoggerFlags)
	t.Run("EmptyPrefix", testNewStdLoggerEmptyPrefix)
	t.Run("Integration", testNewStdLoggerIntegration)
}

func testNewStdLoggerBasic(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	stdLog := slog.NewStdLogger(logger, "", 0)

	core.AssertMustNotNil(t, stdLog, "std logger")

	// Test basic functionality
	stdLog.Print("test message")

	msgs := logger.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "message count")
	core.AssertEqual(t, "test message", msgs[0].Message, "message content")
}

func testNewStdLoggerPrefix(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	prefix := "TEST"
	stdLog := slog.NewStdLogger(logger, prefix, 0)

	stdLog.Print("message")

	msgs := logger.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "message count")

	expected := "TEST: message"
	core.AssertEqual(t, expected, msgs[0].Message, "prefixed message")
}

// stdLoggerFlagsTestCase represents a test case for std logger flags.
type stdLoggerFlagsTestCase struct {
	flags int
	name  string
}

func (tc stdLoggerFlagsTestCase) Name() string {
	return tc.name
}

func (tc stdLoggerFlagsTestCase) Test(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	stdLog := slog.NewStdLogger(logger, "", tc.flags)

	core.AssertMustNotNil(t, stdLog, "std logger with flags")

	stdLog.Print("test")

	msgs := logger.GetMessages()
	core.AssertEqual(t, 1, len(msgs), "message count")
	// Note: The actual timestamp formatting would require more complex
	// parsing to validate, but we can at least ensure it doesn't crash
}

func newStdLoggerFlagsTestCase(name string, flags int) stdLoggerFlagsTestCase {
	return stdLoggerFlagsTestCase{
		name:  name,
		flags: flags,
	}
}

func stdLoggerFlagsTestCases() []stdLoggerFlagsTestCase {
	return []stdLoggerFlagsTestCase{
		newStdLoggerFlagsTestCase("NoFlags", 0),
		newStdLoggerFlagsTestCase("DateFlag", log.Ldate),
		newStdLoggerFlagsTestCase("TimeFlag", log.Ltime),
		newStdLoggerFlagsTestCase("DateTimeFlags", log.Ldate|log.Ltime),
		newStdLoggerFlagsTestCase("AllFlags", log.LstdFlags),
	}
}

func testNewStdLoggerFlags(t *testing.T) {
	core.RunTestCases(t, stdLoggerFlagsTestCases())
}

func testNewStdLoggerEmptyPrefix(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	stdLog := slog.NewStdLogger(logger, "", 0)

	stdLog.Print("no prefix")

	msgs := logger.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "message count")
	core.AssertEqual(t, "no prefix", msgs[0].Message, "message without prefix")
}

func testNewStdLoggerIntegration(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	stdLog := slog.NewStdLogger(logger, "APP", 0)

	// Test multiple logging methods
	stdLog.Print("print message")
	stdLog.Println("println message")
	stdLog.Printf("printf %s %d", "message", 42)

	msgs := logger.GetMessages()
	core.AssertMustEqual(t, 3, len(msgs), "message count")

	expected := []string{
		"APP: print message",
		"APP: println message", // Println adds newline but it gets trimmed by LogWriter
		"APP: printf message 42",
	}

	for i, exp := range expected {
		core.AssertEqual(t, exp, msgs[i].Message, "message %d", i)
	}
}

func TestStdLogSink(t *testing.T) {
	t.Run("PrefixHandling", testStdLogSinkPrefix)
	t.Run("NoPrefixHandling", testStdLogSinkNoPrefix)
	t.Run("FlagsHandling", testStdLogSinkFlags)
}

// stdLogSinkPrefixTestCase represents a test case for std log sink with prefix.
type stdLogSinkPrefixTestCase struct {
	input    string
	expected string
	name     string
}

func (tc stdLogSinkPrefixTestCase) Name() string {
	return tc.name
}

func (tc stdLogSinkPrefixTestCase) Test(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	stdLog := slog.NewStdLogger(logger, "PREFIX", 0)

	stdLog.Print(tc.input)

	msgs := logger.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "message count")
	core.AssertEqual(t, tc.expected, msgs[0].Message, "prefixed message")
}

func newStdLogSinkPrefixTestCase(name, input, expected string) stdLogSinkPrefixTestCase {
	return stdLogSinkPrefixTestCase{
		name:     name,
		input:    input,
		expected: expected,
	}
}

func stdLogSinkPrefixTestCases() []stdLogSinkPrefixTestCase {
	return []stdLogSinkPrefixTestCase{
		newStdLogSinkPrefixTestCase("Print", "test", "PREFIX: test"),
		newStdLogSinkPrefixTestCase("EmptyMessage", "", "PREFIX: "),
		newStdLogSinkPrefixTestCase("WithNewline", "test\n", "PREFIX: test"),
	}
}

func testStdLogSinkPrefix(t *testing.T) {
	core.RunTestCases(t, stdLogSinkPrefixTestCases())
}

// stdLogSinkNoPrefixTestCase represents a test case for std log sink without prefix.
type stdLogSinkNoPrefixTestCase struct {
	action   func(*log.Logger)
	expected string
	name     string
}

func (tc stdLogSinkNoPrefixTestCase) Name() string {
	return tc.name
}

func (tc stdLogSinkNoPrefixTestCase) Test(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	stdLog := slog.NewStdLogger(logger, "", 0)

	tc.action(stdLog)

	msgs := logger.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "message count")
	core.AssertEqual(t, tc.expected, msgs[0].Message, "message content")
}

func newStdLogSinkNoPrefixTestCase(name, expected string, action func(*log.Logger)) stdLogSinkNoPrefixTestCase {
	return stdLogSinkNoPrefixTestCase{
		name:     name,
		expected: expected,
		action:   action,
	}
}

func stdLogSinkNoPrefixTestCases() []stdLogSinkNoPrefixTestCase {
	return []stdLogSinkNoPrefixTestCase{
		newStdLogSinkNoPrefixTestCase("Print", "direct", func(stdLog *log.Logger) {
			stdLog.Print("direct")
		}),
		newStdLogSinkNoPrefixTestCase("PrintMultiple", "multiple args", func(stdLog *log.Logger) {
			stdLog.Print("multiple", " ", "args")
		}),
	}
}

func testStdLogSinkNoPrefix(t *testing.T) {
	core.RunTestCases(t, stdLogSinkNoPrefixTestCases())
}

// stdLogSinkFlagsTestCase represents a test case for std log sink with flags.
type stdLogSinkFlagsTestCase struct {
	flags int
	name  string
}

func (tc stdLogSinkFlagsTestCase) Name() string {
	return tc.name
}

func (tc stdLogSinkFlagsTestCase) Test(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	stdLog := slog.NewStdLogger(logger, "TEST", tc.flags)

	// Should not panic with any flag combination
	stdLog.Print("test message")

	msgs := logger.GetMessages()
	core.AssertMustEqual(t, 1, len(msgs), "message count")
	// Message should contain the prefix and text, but format may vary with flags
	message := msgs[0].Message
	core.AssertContains(t, message, "TEST:", "message contains prefix")
	core.AssertContains(t, message, "test message", "message contains text")
}

func newStdLogSinkFlagsTestCase(name string, flags int) stdLogSinkFlagsTestCase {
	return stdLogSinkFlagsTestCase{
		name:  name,
		flags: flags,
	}
}

func stdLogSinkFlagsTestCases() []stdLogSinkFlagsTestCase {
	return []stdLogSinkFlagsTestCase{
		newStdLogSinkFlagsTestCase("NoFlags", 0),
		newStdLogSinkFlagsTestCase("DateFlag", log.Ldate),
		newStdLogSinkFlagsTestCase("TimeFlag", log.Ltime),
		newStdLogSinkFlagsTestCase("MicrosecondsFlag", log.Lmicroseconds),
		newStdLogSinkFlagsTestCase("LongFileFlag", log.Llongfile),
		newStdLogSinkFlagsTestCase("ShortFileFlag", log.Lshortfile),
		newStdLogSinkFlagsTestCase("UTCFlag", log.LUTC),
		newStdLogSinkFlagsTestCase("StdFlags", log.LstdFlags),
	}
}

func testStdLogSinkFlags(t *testing.T) {
	core.RunTestCases(t, stdLogSinkFlagsTestCases())
}

func TestStdLoggerCompatibility(t *testing.T) {
	t.Run("StdLogInterface", testStdLoggerInterface)
	t.Run("ConcurrentUse", testStdLoggerConcurrent)
}

func testStdLoggerInterface(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
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
	core.AssertEqual(t, 3, len(msgs), "Expected 3 messages")
}

func testStdLoggerConcurrent(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
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
	core.AssertMustEqual(t, 2, len(msgs), "Expected 2 messages")

	// Both messages should have prefix
	for i, msg := range msgs {
		core.AssertTrue(t, strings.Contains(msg.Message, "CONCURRENT:"), "Message %d should have prefix: %q", i, msg.Message)
	}
}
