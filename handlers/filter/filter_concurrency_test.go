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
	name    string
	workers int
	fields  int
}

func (tc filterConcurrentFieldTestCase) Name() string {
	return tc.name
}

func (tc filterConcurrentFieldTestCase) Test(t *testing.T) {
	t.Helper()

	base := mock.NewLogger()
	logger := filter.New(base, slog.Debug)

	runConcurrentFieldWorkers(logger, tc.workers, tc.fields)

	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, tc.workers)
	verifyEachMessageHasFields(t, msgs, tc.fields)
}

func runConcurrentFieldWorkers(logger *filter.Logger, workers, fields int) {
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := range workers {
		workerID := i
		go func() {
			defer wg.Done()
			emitFieldStorm(logger, workerID, fields)
		}()
	}
	wg.Wait()
}

func emitFieldStorm(logger *filter.Logger, workerID, fields int) {
	entry := logger.Info()
	for j := range fields {
		fieldKey := fmt.Sprintf("worker_%d_field_%d", workerID, j)
		fieldValue := fmt.Sprintf("value_%d_%d", workerID, j)
		entry = entry.WithField(fieldKey, fieldValue)
	}
	entry.Printf("Worker %d message", workerID)
}

func verifyEachMessageHasFields(t *testing.T, msgs []slogtest.Message, want int) {
	t.Helper()
	for i, msg := range msgs {
		if len(msg.Fields) != want {
			t.Errorf("Message %d: expected %d fields, got %d. Fields: %+v",
				i, want, len(msg.Fields), msg.Fields)
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
	name     string
	workers  int
	messages int
}

func (tc filterConcurrentLoggingTestCase) Name() string {
	return tc.name
}

func (tc filterConcurrentLoggingTestCase) Test(t *testing.T) {
	t.Helper()

	base := mock.NewLogger()
	counter := &filterCallCounter{}
	logger := newCountingFilterLogger(base, counter)

	runConcurrentLevelMix(logger, tc.workers, tc.messages)

	expectedMessages := tc.workers * tc.messages
	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, expectedMessages)

	// Each message has 1 message filter call + 1 field filter call.
	core.AssertEqual(t, expectedMessages*2, counter.value(), "filter call count")
	verifyFilteredMessages(t, msgs)
}

// filterCallCounter tracks how many times filter callbacks fire under
// concurrent load.
type filterCallCounter struct {
	mu    sync.Mutex
	count int
}

func (c *filterCallCounter) inc() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

func (c *filterCallCounter) value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

func newCountingFilterLogger(parent slog.Logger, counter *filterCallCounter) *filter.Logger {
	return &filter.Logger{
		Parent:    parent,
		Threshold: slog.Debug,
		MessageFilter: func(msg string) (string, bool) {
			counter.inc()
			return "[filtered] " + msg, true
		},
		FieldFilter: func(key string, val any) (string, any, bool) {
			counter.inc()
			return "filtered_" + key, val, true
		},
	}
}

func runConcurrentLevelMix(logger *filter.Logger, workers, messages int) {
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := range workers {
		workerID := i
		go func() {
			defer wg.Done()
			emitMixedLevelMessages(logger, workerID, messages)
		}()
	}
	wg.Wait()
}

func emitMixedLevelMessages(logger *filter.Logger, workerID, messages int) {
	for j := range messages {
		switch j % 4 {
		case 0:
			logger.Debug().WithField("worker", workerID).Printf("Debug message %d", j)
		case 1:
			logger.Info().WithField("worker", workerID).Printf("Info message %d", j)
		case 2:
			logger.Warn().WithField("worker", workerID).Printf("Warn message %d", j)
		case 3:
			logger.Error().WithField("worker", workerID).Printf("Error message %d", j)
		default:
			// Skip unexpected dispatch values rather than crash the worker.
		}
	}
}

func verifyFilteredMessages(t *testing.T, msgs []slogtest.Message) {
	t.Helper()
	for _, msg := range msgs {
		core.AssertContains(t, msg.Message, "[filtered]", "message filtering")
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

func (filterImmutabilityTestCase) Test(t *testing.T) {
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

	counters := &filterCallTallies{}
	logger := newWorkSimulatingFilter(base, counters)

	const workers = 20
	const operations = 10

	runFilterModificationWorkers(logger, workers, operations)

	expectedMessages := workers * operations
	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, expectedMessages)

	field, message := counters.snapshot()
	core.AssertEqual(t, expectedMessages, message, "message filter calls")
	core.AssertEqual(t, expectedMessages, field, "field filter calls")
}

// filterCallTallies counts FieldFilter and MessageFilter invocations
// under concurrent load.
type filterCallTallies struct {
	mu      sync.Mutex
	field   int
	message int
}

func (c *filterCallTallies) incField() {
	c.mu.Lock()
	c.field++
	c.mu.Unlock()
}

func (c *filterCallTallies) incMessage() {
	c.mu.Lock()
	c.message++
	c.mu.Unlock()
}

func (c *filterCallTallies) snapshot() (field, message int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.field, c.message
}

// simulateFilterWork burns a few cycles to widen the race window
// without touching shared state.
func simulateFilterWork() {
	for i := range 100 {
		_ = i * 2
	}
}

func newWorkSimulatingFilter(parent slog.Logger, counters *filterCallTallies) *filter.Logger {
	return &filter.Logger{
		Parent:    parent,
		Threshold: slog.Debug,
		FieldFilter: func(key string, val any) (string, any, bool) {
			counters.incField()
			simulateFilterWork()
			return key, val, true
		},
		MessageFilter: func(msg string) (string, bool) {
			counters.incMessage()
			simulateFilterWork()
			return msg, true
		},
	}
}

func runFilterModificationWorkers(logger *filter.Logger, workers, operations int) {
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := range workers {
		go func(id int) {
			defer wg.Done()
			for j := range operations {
				logger.Info().
					WithField(fmt.Sprintf("field_%d_%d", id, j), "value").
					Printf("Message from worker %d operation %d", id, j)
			}
		}(i)
	}
	wg.Wait()
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
	for i := range workers {
		go func(id int) {
			defer wg.Done()
			for j := range iterations {
				runSharedLoggerOperation(logger, id, j)
			}
		}(i)
	}
	wg.Wait()

	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, workers*iterations)
	verifySharedLoggerMessages(t, msgs)
}

// runSharedLoggerOperation runs one of five distinct logging shapes
// against the shared logger, picked by iteration index.
func runSharedLoggerOperation(logger *filter.Logger, id, j int) {
	switch j % 5 {
	case 0:
		entry := logger.Debug()
		entry = entry.WithField("id", id)
		entry = entry.WithField("iteration", j)
		entry.Print("debug message")
	case 1:
		logger.Info().
			WithField("worker", id).
			WithField("op", j).
			Print("info message")
	case 2:
		logger.Warn().WithFields(map[string]any{
			"worker": id,
			"iter":   j,
			"type":   "warning",
		}).Print("warn message")
	case 3:
		logger.Debug().
			WithField("start", "debug").
			Error().
			WithField("end", "error").
			Printf("Level change %d", id)
	case 4:
		logger.Error().
			WithStack(0).
			WithField("worker", id).
			Print("error with stack")
	default:
		// Skip unexpected dispatch values rather than crash the worker.
	}
}

func verifySharedLoggerMessages(t *testing.T, msgs []slogtest.Message) {
	t.Helper()
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
