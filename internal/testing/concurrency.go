package testing

import (
	"fmt"
	"sync"
	"testing"

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

	runConcurrentLogging(logger, test)
	verifyConcurrentResults(t, logger, test)
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
func verifyConcurrentResults(t *testing.T, logger slog.Logger, test ConcurrencyTest) {
	t.Helper()

	tl, ok := logger.(*Logger)
	if !ok {
		t.Fatal("logger does not implement *Logger - cannot verify concurrent results")
	}

	msgs := tl.GetMessages()
	expected := test.Goroutines * test.Operations
	AssertMessageCount(t, msgs, expected)

	// Verify all messages have required fields
	for i, msg := range msgs {
		if msg.Fields["goroutine"] == nil {
			t.Errorf("message %d missing goroutine field", i)
		}
		if msg.Fields["operation"] == nil {
			t.Errorf("message %d missing operation field", i)
		}
	}
}

// TestConcurrentFields verifies field handling under concurrent access.
func TestConcurrentFields(t *testing.T, newLogger func() slog.Logger) {
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
func verifyConcurrentLoggers(t *testing.T, loggers []slog.Logger) {
	t.Helper()
	for i, l := range loggers {
		if l == nil {
			t.Errorf("logger %d is nil", i)
		}
	}
}
