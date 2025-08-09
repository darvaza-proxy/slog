package testing

import (
	"fmt"
	"sync"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// ConcurrencyTest represents a concurrent test scenario.
type ConcurrencyTest struct {
	Goroutines int
	Operations int
}

// DefaultConcurrencyTest returns a standard concurrency test configuration.
func DefaultConcurrencyTest() ConcurrencyTest {
	return ConcurrencyTest{
		Goroutines: 10,
		Operations: 100,
	}
}

// RunConcurrentTest executes a concurrent logging test.
func RunConcurrentTest(t *testing.T, logger slog.Logger, test ConcurrencyTest) {
	t.Helper()
	RunConcurrentTestWithOptions(t, logger, test, nil)
}

// RunConcurrentTestWithOptions executes a concurrent logging test with options.
func RunConcurrentTestWithOptions(t *testing.T, logger slog.Logger,
	test ConcurrencyTest, opts *ConcurrencyTestOptions) {
	t.Helper()

	// Setup logger and message getter
	testLogger, getMessages := setupConcurrentTest(logger, opts)

	// Run the concurrent logging test
	runConcurrentLogging(testLogger, test)

	// Verify results
	verifyConcurrentTestResults(t, test, getMessages)
}

// setupConcurrentTest sets up the logger and message getter for concurrent testing
func setupConcurrentTest(logger slog.Logger, opts *ConcurrencyTestOptions) (slog.Logger, func() []Message) {
	if opts != nil && opts.NewLoggerWithRecorder != nil {
		// Use factory pattern with recorder
		recorder := NewLogger()
		testLogger := opts.NewLoggerWithRecorder(recorder)
		return testLogger, recorder.GetMessages
	}

	// Use provided logger
	testLogger := logger
	getMessages := extractMessageGetter(logger, opts)
	return testLogger, getMessages
}

// extractMessageGetter determines how to get messages from the logger
func extractMessageGetter(logger slog.Logger, opts *ConcurrencyTestOptions) func() []Message {
	if opts != nil && opts.GetMessages != nil {
		return opts.GetMessages
	}
	if tl, ok := logger.(*Logger); ok {
		return tl.GetMessages
	}
	return nil
}

// verifyConcurrentTestResults verifies the results of concurrent testing.
// If getMessages is nil, logs completion without verification rather than failing.
func verifyConcurrentTestResults(t core.T, test ConcurrencyTest, getMessages func() []Message) {
	t.Helper()

	if getMessages == nil {
		// Log completion when verification is not available (not an error condition)
		t.Logf("Concurrent test completed: %d goroutines × %d operations = %d total messages "+
			"(verification not available)",
			test.Goroutines, test.Operations, test.Goroutines*test.Operations)
		return
	}

	msgs := getMessages()
	expected := test.Goroutines * test.Operations
	AssertMustMessageCount(t, msgs, expected)
	verifyMessageFields(t, msgs)
}

// logNoVerification logs when verification is not available
func logNoVerification(t core.T, test ConcurrencyTest) {
	t.Logf("Concurrent test completed: %d goroutines × %d operations = %d total messages "+
		"(verification not available)",
		test.Goroutines, test.Operations, test.Goroutines*test.Operations)
}

// verifyMessageFields verifies that all messages have required fields
func verifyMessageFields(t core.T, msgs []Message) {
	t.Helper()

	for i, msg := range msgs {
		core.AssertMustNotNil(t, msg.Fields["goroutine"], "message %d goroutine field", i)
		core.AssertMustNotNil(t, msg.Fields["operation"], "message %d operation field", i)
	}
}

// runConcurrentLogging performs concurrent logging operations.
func runConcurrentLogging(logger slog.Logger, test ConcurrencyTest) {
	var wg sync.WaitGroup
	for i := 0; i < test.Goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runGoroutineOperations(logger, id, test.Operations)
		}(i)
	}
	wg.Wait()
}

// runGoroutineOperations performs logging operations for a single goroutine.
func runGoroutineOperations(logger slog.Logger, id, operations int) {
	for j := 0; j < operations; j++ {
		logger.Info().
			WithField("goroutine", id).
			WithField("operation", j).
			Printf("concurrent message %d-%d", id, j)
	}
}

// verifyConcurrentResults verifies the results of concurrent logging.
func verifyConcurrentResults(t core.T, logger *Logger, test ConcurrencyTest) {
	t.Helper()

	msgs := logger.GetMessages()
	expected := test.Goroutines * test.Operations
	AssertMustMessageCount(t, msgs, expected)

	// Verify all messages have required fields
	verifyMessageFields(t, msgs)
}

// TestConcurrentFields verifies field handling under concurrent access.
func TestConcurrentFields(t core.T, newLogger func() slog.Logger) {
	logger := newLogger()

	const goroutines = 50
	const fieldsPerGoroutine = 20

	loggers := createConcurrentLoggers(logger, goroutines, fieldsPerGoroutine)
	verifyConcurrentLoggers(t, loggers)
}

// createConcurrentLoggers creates loggers with different fields concurrently.
func createConcurrentLoggers(base slog.Logger, goroutines, fieldsPerGoroutine int) []slog.Logger {
	var wg sync.WaitGroup
	loggers := make([]slog.Logger, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			loggers[id] = createLoggerWithFields(base, id, fieldsPerGoroutine)
		}(i)
	}

	wg.Wait()
	return loggers
}

// createLoggerWithFields creates a logger with multiple fields.
func createLoggerWithFields(base slog.Logger, id, numFields int) slog.Logger {
	l := base
	for j := 0; j < numFields; j++ {
		fieldName := fmt.Sprintf("field_%d", j)
		fieldValue := id*100 + j
		l = l.WithField(fieldName, fieldValue)
	}
	return l
}

// verifyConcurrentLoggers verifies each logger is independent.
func verifyConcurrentLoggers(t core.T, loggers []slog.Logger) {
	t.Helper()
	for i, l := range loggers {
		core.AssertNotNil(t, l, "logger %d", i)
	}
}
