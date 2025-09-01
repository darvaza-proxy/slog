package filter_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
	"darvaza.org/slog/handlers/mock"
	slogtest "darvaza.org/slog/internal/testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = noopPrintTestCase{}

// noopPrintTestCase tests no-op Print methods on Logger
type noopPrintTestCase struct {
	printMethod func(slog.Logger, string)
	methodName  string
	name        string
	message     string
}

func (tc noopPrintTestCase) Name() string {
	return tc.name
}

func (tc noopPrintTestCase) Test(t *testing.T) {
	t.Helper()

	// Create filter with parent to verify no messages are passed through
	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Info)

	// Call the print method
	tc.printMethod(logger, tc.message)

	// Logger.Print is a no-op - should not pass to parent without a level
	messages := parent.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 0)

	// Also test with noop logger
	noopLogger := filter.NewNoop()
	core.AssertNoPanic(t, func() {
		tc.printMethod(noopLogger, tc.message)
	}, "noop logger %s", tc.methodName)
}

// Factory function for noopPrintTestCase
func newNoopPrintTestCase(name, methodName string,
	printMethod func(slog.Logger, string), message string) noopPrintTestCase {
	return noopPrintTestCase{
		name:        name,
		methodName:  methodName,
		printMethod: printMethod,
		message:     message,
	}
}

// TestLoggerPrintNoOp tests that Logger.Print methods behave like UndefinedLevel entries
func TestLoggerPrintNoOp(t *testing.T) {
	testCases := []noopPrintTestCase{
		newNoopPrintTestCase("Print with simple message", "Print",
			func(l slog.Logger, msg string) { l.Print(msg) },
			"test message"),
		newNoopPrintTestCase("Print with empty message", "Print",
			func(l slog.Logger, msg string) { l.Print(msg) },
			""),
		newNoopPrintTestCase("Println with simple message", "Println",
			func(l slog.Logger, msg string) { l.Println(msg) },
			"test message"),
		newNoopPrintTestCase("Println with empty message", "Println",
			func(l slog.Logger, msg string) { l.Println(msg) },
			""),
		newNoopPrintTestCase("Printf with format string", "Printf",
			func(l slog.Logger, msg string) { l.Printf("%s %d", msg, 42) },
			"test"),
		newNoopPrintTestCase("Printf with empty format", "Printf",
			func(l slog.Logger, _ string) { l.Printf("") },
			""),
	}

	core.RunTestCases(t, testCases)
}

// TestLoggerPrintVariadic tests Print methods with multiple arguments
func TestLoggerPrintVariadic(t *testing.T) {
	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Info)

	// Test Print with multiple arguments - should be no-op without level
	logger.Print("arg1", "arg2", "arg3")
	slogtest.AssertMustMessageCount(t, parent.GetMessages(), 0)
	parent.Clear()

	// Test Println with multiple arguments
	logger.Println("arg1", "arg2", "arg3")
	slogtest.AssertMustMessageCount(t, parent.GetMessages(), 0)
	parent.Clear()

	// Test Printf with multiple format arguments
	logger.Printf("%s %s %d %v", "arg1", "arg2", 3, true)
	slogtest.AssertMustMessageCount(t, parent.GetMessages(), 0)
}

// TestLoggerPrintWithFields tests Print methods behaviour after WithField
func TestLoggerPrintWithFields(t *testing.T) {
	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Info)

	// Test that Logger.Print methods are no-ops
	// Logger.Print is a no-op - doesn't pass to parent
	logger.Print("message")
	logger.Println("message")
	logger.Printf("message %d", 1)
	messages := parent.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 0)
	parent.Clear()

	// WithField creates a LogEntry without a level
	// LogEntry.Print without level is also a no-op
	loggerWithFields := logger.
		WithField("key1", "value1").
		WithField("key2", "value2")

	loggerWithFields.Print("message1")
	loggerWithFields.Println("message2")
	loggerWithFields.Printf("message %d", 3)

	messages = parent.GetMessages()
	// LogEntry without a level doesn't log
	slogtest.AssertMustMessageCount(t, messages, 0)
}

// TestLoggerPrintAfterLevelMethods tests Print methods don't interfere with level methods
func TestLoggerPrintAfterLevelMethods(t *testing.T) {
	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Info)

	// Use a level method first
	logger.Info().Print("info message")
	slogtest.AssertMessageCount(t, parent.GetMessages(), 1)

	parent.Clear()

	// Now use logger Print (no-op without level)
	logger.Print("direct print")
	messages := parent.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 0)
	parent.Clear()

	// Use another level method
	logger.Error().Print("error message")
	slogtest.AssertMessageCount(t, parent.GetMessages(), 1)
}

// TestNoopLoggerPrintMethods tests Print methods on noop logger
func TestNoopLoggerPrintMethods(t *testing.T) {
	logger := filter.NewNoop()

	// Logger.Print methods create UndefinedLevel entries
	// For parentless loggers, these won't log (no parent to delegate to)
	core.AssertNoPanic(t, func() {
		logger.Print("test")
		logger.Println("test")
		logger.Printf("test %d", 42)
	}, "noop Logger.Print methods don't panic")

	// WithField creates a LogEntry at UndefinedLevel
	// With the bug fix: UndefinedLevel is NOT enabled for parentless loggers
	// So Print should be a no-op, not panic
	loggerWithFields := logger.WithField("key", "value")

	// After fix: This should NOT panic (UndefinedLevel is disabled)
	core.AssertNoPanic(t, func() {
		loggerWithFields.Print("test")
	}, "parentless LogEntry at UndefinedLevel should not panic")

	// To safely use fields on a noop logger, you must use Fatal or Panic level
	fatalWithFields := logger.WithField("key", "value").Fatal()
	// Fatal is enabled on noop, but we can't test Print without exiting
	// Just verify it doesn't panic during creation
	core.AssertNotNil(t, fatalWithFields, "fatal with fields created")

	panicWithFields := logger.WithField("key", "value").Panic()
	// Panic will actually panic when Print is called
	core.AssertPanic(t, func() {
		panicWithFields.Print("test panic")
	}, nil, "panic with fields panics as expected")

	// Test completed
	core.AssertTrue(t, true, "noop print methods completed")
}

// TestPrintMethodsEdgeCases tests edge cases for Print methods
func TestPrintMethodsEdgeCases(t *testing.T) {
	t.Run("Print with nil arguments", runTestPrintWithNilArguments)

	t.Run("Printf with mismatched format", runTestPrintfWithMismatchedFormat)

	t.Run("Print methods in goroutines", runTestPrintMethodsInGoroutines)
}

// TestPrintMethodCombinations tests various combinations of Print method usage
func TestPrintMethodCombinations(t *testing.T) {
	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Debug)

	// Chain of operations mixing Print methods and level methods
	logger.Print("print 1")               // No-op
	logger.Debug().Print("debug message") // Debug
	logger.Println("println 1")           // No-op
	logger.Info().Print("info message")   // Info
	logger.Printf("printf %d", 1)         // No-op
	logger.Warn().Print("warn message")   // Warn

	messages := parent.GetMessages()
	// Only level methods log, Logger.Print is no-op
	slogtest.AssertMustMessageCount(t, messages, 3)

	// Verify the logged messages have correct levels
	slogtest.AssertMessage(t, messages[0], slog.Debug, "debug message")
	slogtest.AssertMessage(t, messages[1], slog.Info, "info message")
	slogtest.AssertMessage(t, messages[2], slog.Warn, "warn message")
}

func runTestPrintWithNilArguments(t *testing.T) {
	t.Helper()
	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Info)

	// Should not panic, but are no-ops without level
	logger.Print(nil)
	logger.Println(nil)
	logger.Printf("%v", nil)

	// Logger.Print methods are no-ops
	slogtest.AssertMustMessageCount(t, parent.GetMessages(), 0)
}

func runTestPrintfWithMismatchedFormat(t *testing.T) {
	t.Helper()
	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Info)

	// Should not panic even with mismatched format
	logger.Printf("%s %d", "only one arg")
	logger.Printf("%d", "not a number")

	// Logger.Printf is a no-op
	slogtest.AssertMustMessageCount(t, parent.GetMessages(), 0)
}

func runTestPrintMethodsInGoroutines(t *testing.T) {
	t.Helper()
	parent := mock.NewLogger()
	logger := filter.New(parent, slog.Info)

	// Run Print methods concurrently
	done := make(chan bool, 3)

	go func() {
		logger.Print("goroutine 1")
		done <- true
	}()

	go func() {
		logger.Println("goroutine 2")
		done <- true
	}()

	go func() {
		logger.Printf("goroutine %d", 3)
		done <- true
	}()

	// Wait for all goroutines
	for range 3 {
		<-done
	}

	// Logger.Print methods are no-ops even in goroutines
	slogtest.AssertMustMessageCount(t, parent.GetMessages(), 0)
}
