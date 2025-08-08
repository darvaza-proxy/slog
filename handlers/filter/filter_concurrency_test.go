package filter_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
	"darvaza.org/slog/handlers/mock"
	slogtest "darvaza.org/slog/internal/testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = filterConcurrentFieldTestCase{}
var _ core.TestCase = filterConcurrentLoggingTestCase{}
var _ core.TestCase = filterImmutabilityTestCase{}

// filterConcurrentFieldTestCase tests concurrent field attachment
type filterConcurrentFieldTestCase struct {
	workers int
	fields  int
	name    string
}

func (tc filterConcurrentFieldTestCase) Name() string {
	return tc.name
}

func (tc filterConcurrentFieldTestCase) Test(t *testing.T) {
	t.Helper()

	base := mock.NewLogger()
	logger := filter.New(base, slog.Debug)

	var wg sync.WaitGroup
	wg.Add(tc.workers)

	// Each worker adds fields concurrently
	for i := 0; i < tc.workers; i++ {
		workerID := i
		go func() {
			defer wg.Done()

			// Create a new entry for each worker
			entry := logger.Info()

			// Add multiple fields
			for j := 0; j < tc.fields; j++ {
				fieldKey := fmt.Sprintf("worker_%d_field_%d", workerID, j)
				fieldValue := fmt.Sprintf("value_%d_%d", workerID, j)
				entry = entry.WithField(fieldKey, fieldValue)
			}

			// Log the message
			entry.Printf("Worker %d message", workerID)
		}()
	}

	wg.Wait()

	// Verify all messages were recorded
	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, tc.workers)

	// Verify each message has the expected fields
	for i, msg := range msgs {
		// Count fields - should have tc.fields per message
		fieldCount := len(msg.Fields)
		if fieldCount != tc.fields {
			t.Errorf("Message %d: expected %d fields, got %d. Fields: %+v", i, tc.fields, fieldCount, msg.Fields)
		}
	}
}

func newFilterConcurrentFieldTestCase(name string, workers, fields int) filterConcurrentFieldTestCase {
	return filterConcurrentFieldTestCase{
		name:    name,
		workers: workers,
		fields:  fields,
	}
}

func filterConcurrentFieldTestCases() []filterConcurrentFieldTestCase {
	return []filterConcurrentFieldTestCase{
		newFilterConcurrentFieldTestCase("Few workers few fields", 2, 3),
		newFilterConcurrentFieldTestCase("Many workers few fields", 10, 2),
		newFilterConcurrentFieldTestCase("Few workers many fields", 2, 10),
		newFilterConcurrentFieldTestCase("Many workers many fields", 10, 10),
	}
}

func TestFilterConcurrentFields(t *testing.T) {
	core.RunTestCases(t, filterConcurrentFieldTestCases())
}

// filterConcurrentLoggingTestCase tests parallel logging from multiple goroutines
type filterConcurrentLoggingTestCase struct {
	workers  int
	messages int
	name     string
}

func (tc filterConcurrentLoggingTestCase) Name() string {
	return tc.name
}

func (tc filterConcurrentLoggingTestCase) Test(t *testing.T) {
	t.Helper()

	base := mock.NewLogger()

	// Add filters to ensure thread safety
	filterCalls := 0
	var filterMu sync.Mutex

	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		MessageFilter: func(msg string) (string, bool) {
			filterMu.Lock()
			filterCalls++
			filterMu.Unlock()
			return "[filtered] " + msg, true
		},
		FieldFilter: func(key string, val any) (string, any, bool) {
			filterMu.Lock()
			filterCalls++
			filterMu.Unlock()
			return "filtered_" + key, val, true
		},
	}

	var wg sync.WaitGroup
	wg.Add(tc.workers)

	// Each worker logs multiple messages concurrently
	for i := 0; i < tc.workers; i++ {
		workerID := i
		go func() {
			defer wg.Done()

			for j := 0; j < tc.messages; j++ {
				// Mix different log levels and operations
				switch j % 4 {
				case 0:
					logger.Debug().
						WithField("worker", workerID).
						Printf("Debug message %d", j)
				case 1:
					logger.Info().
						WithField("worker", workerID).
						Printf("Info message %d", j)
				case 2:
					logger.Warn().
						WithField("worker", workerID).
						Printf("Warn message %d", j)
				case 3:
					logger.Error().
						WithField("worker", workerID).
						Printf("Error message %d", j)
				}
			}
		}()
	}

	wg.Wait()

	// Verify all messages were recorded
	msgs := base.GetMessages()
	expectedMessages := tc.workers * tc.messages
	slogtest.AssertMessageCount(t, msgs, expectedMessages)

	// Verify filters were called for each message and field
	filterMu.Lock()
	actualCalls := filterCalls
	filterMu.Unlock()

	// Each message has 1 message filter call + 1 field filter call
	expectedCalls := expectedMessages * 2
	core.AssertEqual(t, expectedCalls, actualCalls, "filter call count")

	// Verify all messages have filtered content
	for _, msg := range msgs {
		core.AssertContains(t, msg.Message, "[filtered]", "message filtering")

		// Check that worker field was filtered - we don't care about the value, just that it exists
		if _, exists := msg.Fields["filtered_worker"]; !exists {
			t.Error("Worker field not properly filtered")
		}
	}
}

func newFilterConcurrentLoggingTestCase(name string, workers, messages int) filterConcurrentLoggingTestCase {
	return filterConcurrentLoggingTestCase{
		name:     name,
		workers:  workers,
		messages: messages,
	}
}

func filterConcurrentLoggingTestCases() []filterConcurrentLoggingTestCase {
	return []filterConcurrentLoggingTestCase{
		newFilterConcurrentLoggingTestCase("Few workers few messages", 2, 5),
		newFilterConcurrentLoggingTestCase("Many workers few messages", 10, 2),
		newFilterConcurrentLoggingTestCase("Few workers many messages", 2, 20),
		newFilterConcurrentLoggingTestCase("Many workers many messages", 10, 10),
	}
}

func TestFilterConcurrentLogging(t *testing.T) {
	core.RunTestCases(t, filterConcurrentLoggingTestCases())
}

// filterImmutabilityTestCase verifies immutability guarantees
type filterImmutabilityTestCase struct {
	name string
}

func (tc filterImmutabilityTestCase) Name() string {
	return tc.name
}

func (tc filterImmutabilityTestCase) Test(t *testing.T) {
	t.Helper()

	base := mock.NewLogger()
	logger := filter.New(base, slog.Debug)

	// Create a base logger with some fields
	baseLogger := logger.WithField("base", "value1")

	var wg sync.WaitGroup
	wg.Add(3)

	// Worker 1: Add more fields to baseLogger
	go func() {
		defer wg.Done()
		worker1 := baseLogger.WithField("worker1", "data1")
		worker1.Info().Print("Worker 1 message")
	}()

	// Worker 2: Add different fields to baseLogger
	go func() {
		defer wg.Done()
		worker2 := baseLogger.WithField("worker2", "data2")
		worker2.Info().Print("Worker 2 message")
	}()

	// Worker 3: Use baseLogger directly
	go func() {
		defer wg.Done()
		baseLogger.Info().Print("Worker 3 message")
	}()

	wg.Wait()

	// Verify all messages were recorded
	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, 3)

	// Verify each worker's fields are independent
	worker1Msg := findMessageByContent(msgs, "Worker 1")
	worker2Msg := findMessageByContent(msgs, "Worker 2")
	worker3Msg := findMessageByContent(msgs, "Worker 3")

	// Worker 1 should have base + worker1 fields
	core.AssertEqual(t, "value1", worker1Msg.Fields["base"], "worker1 base field")
	core.AssertEqual(t, "data1", worker1Msg.Fields["worker1"], "worker1 specific field")
	core.AssertNil(t, worker1Msg.Fields["worker2"], "worker1 should not have worker2 field")

	// Worker 2 should have base + worker2 fields
	core.AssertEqual(t, "value1", worker2Msg.Fields["base"], "worker2 base field")
	core.AssertEqual(t, "data2", worker2Msg.Fields["worker2"], "worker2 specific field")
	core.AssertNil(t, worker2Msg.Fields["worker1"], "worker2 should not have worker1 field")

	// Worker 3 should only have base field
	core.AssertEqual(t, "value1", worker3Msg.Fields["base"], "worker3 base field")
	core.AssertNil(t, worker3Msg.Fields["worker1"], "worker3 should not have worker1 field")
	core.AssertNil(t, worker3Msg.Fields["worker2"], "worker3 should not have worker2 field")
}

func newFilterImmutabilityTestCase(name string) filterImmutabilityTestCase {
	return filterImmutabilityTestCase{
		name: name,
	}
}

func filterImmutabilityTestCases() []filterImmutabilityTestCase {
	return []filterImmutabilityTestCase{
		newFilterImmutabilityTestCase("Logger immutability"),
	}
}

func TestFilterImmutability(t *testing.T) {
	core.RunTestCases(t, filterImmutabilityTestCases())
}

// Helper function to find a message by content
func findMessageByContent(msgs []mock.Message, content string) *mock.Message {
	for i := range msgs {
		if strings.Contains(msgs[i].Message, content) {
			return &msgs[i]
		}
	}
	return nil
}

func runTestConcurrentFilterModification(t *testing.T) {
	t.Helper()
	base := mock.NewLogger()

	// Shared counters protected by mutex
	var mu sync.Mutex
	fieldFilterCalls := 0
	messageFilterCalls := 0

	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		FieldFilter: func(key string, val any) (string, any, bool) {
			mu.Lock()
			fieldFilterCalls++
			mu.Unlock()
			// Simulate some work
			for i := 0; i < 100; i++ {
				_ = i * 2
			}
			return key, val, true
		},
		MessageFilter: func(msg string) (string, bool) {
			mu.Lock()
			messageFilterCalls++
			mu.Unlock()
			// Simulate some work
			for i := 0; i < 100; i++ {
				_ = i * 2
			}
			return msg, true
		},
	}

	const workers = 20
	const operations = 10

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				logger.Info().
					WithField(fmt.Sprintf("field_%d_%d", id, j), "value").
					Printf("Message from worker %d operation %d", id, j)
			}
		}(i)
	}

	wg.Wait()

	// Verify counts
	expectedMessages := workers * operations
	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, expectedMessages)

	mu.Lock()
	core.AssertEqual(t, expectedMessages, messageFilterCalls, "message filter calls")
	core.AssertEqual(t, expectedMessages, fieldFilterCalls, "field filter calls")
	mu.Unlock()
}

// Test concurrent filter modifications
func TestConcurrentFilterModification(t *testing.T) {
	t.Run("concurrent modification", runTestConcurrentFilterModification)
}

func runTestSharedLoggerRaceConditions(t *testing.T) {
	t.Helper()
	base := mock.NewLogger()
	logger := filter.New(base, slog.Debug)

	const workers = 50
	const iterations = 20

	var wg sync.WaitGroup
	wg.Add(workers)

	// All workers share the same logger instance
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				// Perform various operations on the shared logger
				switch j % 5 {
				case 0:
					// Create new entry and add fields
					entry := logger.Debug()
					entry = entry.WithField("id", id)
					entry = entry.WithField("iteration", j)
					entry.Print("debug message")
				case 1:
					// Chain operations
					logger.Info().
						WithField("worker", id).
						WithField("op", j).
						Print("info message")
				case 2:
					// Use WithFields
					logger.Warn().WithFields(map[string]any{
						"worker": id,
						"iter":   j,
						"type":   "warning",
					}).Print("warn message")
				case 3:
					// Change level mid-chain
					logger.Debug().
						WithField("start", "debug").
						Error().
						WithField("end", "error").
						Printf("Level change %d", id)
				case 4:
					// Add stack trace
					logger.Error().
						WithStack(0).
						WithField("worker", id).
						Print("error with stack")
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all messages were recorded without panic
	msgs := base.GetMessages()
	expectedMessages := workers * iterations
	slogtest.AssertMessageCount(t, msgs, expectedMessages)

	// Verify message integrity (no corruption)
	for _, msg := range msgs {
		core.AssertNotNil(t, msg.Message, "message should not be nil")
		core.AssertTrue(t, msg.Level >= slog.Panic && msg.Level <= slog.Debug,
			"valid log level")
	}
}

// Test race conditions with shared logger instance
func TestSharedLoggerRaceConditions(t *testing.T) {
	t.Run("shared logger race conditions", runTestSharedLoggerRaceConditions)
}
