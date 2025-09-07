// Package testing provides shared test utilities for slog handler testing.
package testing

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// StressTest represents a stress test scenario for loggers.
type StressTest struct {
	// Custom stress function (if set, default behaviour is overridden)
	StressFunc func(logger slog.Logger, id int)

	// Concurrency settings
	Goroutines int
	Operations int

	// Duration-based stress test (if set, Operations is ignored)
	Duration time.Duration

	// Memory pressure settings
	FieldsPerMessage     int // Number of fields to add to each message
	MessageSize          int // Size of message content in bytes
	EnableMemoryPressure bool
}

// DefaultStressTest returns a standard stress test configuration.
func DefaultStressTest() StressTest {
	return StressTest{
		Goroutines: 10,
		Operations: 100,
	}
}

// HighVolumeStressTest returns a high-volume stress test configuration.
func HighVolumeStressTest() StressTest {
	return StressTest{
		Goroutines: 50,
		Operations: 1000,
	}
}

// MemoryPressureStressTest returns a memory pressure stress test configuration.
func MemoryPressureStressTest() StressTest {
	return StressTest{
		Goroutines:           20,
		Operations:           500,
		EnableMemoryPressure: true,
		FieldsPerMessage:     50,
		MessageSize:          1024, // 1KB messages
	}
}

// DurationBasedStressTest returns a duration-based stress test configuration.
func DurationBasedStressTest(duration time.Duration) StressTest {
	return StressTest{
		Goroutines: 20,
		Duration:   duration,
	}
}

// StressTestOptions provides options for stress testing.
type StressTestOptions struct {
	// GetMessages provides a way to retrieve messages for verification
	GetMessages func() []Message

	// PreStressFunc is called before stress test starts
	PreStressFunc func()

	// PostStressFunc is called after stress test completes
	PostStressFunc func()

	// Custom verification function
	VerifyFunc func(t core.T, messages []Message, test StressTest)

	// VerifyResults indicates whether to verify logged messages
	VerifyResults bool
}

// RunStressTest executes a stress test against a logger.
func RunStressTest(t core.T, logger slog.Logger, test StressTest) {
	t.Helper()
	RunStressTestWithOptions(t, logger, test, nil)
}

// RunStressTestWithOptions executes a stress test with custom options.
func RunStressTestWithOptions(t core.T, logger slog.Logger, test StressTest, opts *StressTestOptions) {
	t.Helper()

	// Execute pre-stress function if provided
	executePreStress(opts)

	// Run the stress test and measure duration
	duration := measureStressExecution(logger, test)

	// Execute post-stress function if provided
	executePostStress(opts)

	// Report performance metrics
	reportStressMetrics(t, test, duration)

	// Verify results if requested
	verifyStressTestResults(t, test, opts)
}

// executePreStress runs the pre-stress function if provided.
func executePreStress(opts *StressTestOptions) {
	if opts != nil && opts.PreStressFunc != nil {
		opts.PreStressFunc()
	}
}

// measureStressExecution runs stress operations and returns the duration.
func measureStressExecution(logger slog.Logger, test StressTest) time.Duration {
	startTime := time.Now()
	runStressOperations(logger, test)
	return time.Since(startTime)
}

// executePostStress runs the post-stress function if provided.
func executePostStress(opts *StressTestOptions) {
	if opts != nil && opts.PostStressFunc != nil {
		opts.PostStressFunc()
	}
}

// verifyStressTestResults performs result verification if configured.
func verifyStressTestResults(t core.T, test StressTest, opts *StressTestOptions) {
	t.Helper()

	if !shouldVerifyResults(opts) {
		return
	}

	messages := opts.GetMessages()
	if opts.VerifyFunc != nil {
		opts.VerifyFunc(t, messages, test)
	} else {
		verifyStressResults(t, messages, test)
	}
}

// shouldVerifyResults checks if result verification is enabled.
func shouldVerifyResults(opts *StressTestOptions) bool {
	return opts != nil && opts.VerifyResults && opts.GetMessages != nil
}

// runStressOperations performs the stress test operations.
func runStressOperations(logger slog.Logger, test StressTest) {
	var wg sync.WaitGroup

	for i := 0; i < test.Goroutines; i++ {
		wg.Add(1)
		go runStressWorker(&wg, logger, i, test)
	}

	wg.Wait()
}

// runStressWorker executes stress operations in a single goroutine.
func runStressWorker(wg *sync.WaitGroup, logger slog.Logger, id int, test StressTest) {
	defer wg.Done()

	// Select appropriate stress function
	stressFunc := selectStressFunction(test)
	stressFunc(logger, id, test)
}

// selectStressFunction returns the appropriate stress function based on test configuration.
func selectStressFunction(test StressTest) func(slog.Logger, int, StressTest) {
	if test.StressFunc != nil {
		return func(logger slog.Logger, id int, _ StressTest) {
			test.StressFunc(logger, id)
		}
	}

	if test.Duration > 0 {
		return runDurationBasedStress
	}

	return runCountBasedStress
}

// runCountBasedStress runs a fixed number of operations.
func runCountBasedStress(logger slog.Logger, id int, test StressTest) {
	for j := 0; j < test.Operations; j++ {
		l := logger.Info().
			WithField("goroutine", id).
			WithField("operation", j)

		if test.EnableMemoryPressure {
			l = addMemoryPressure(l, test)
		}

		msg := fmt.Sprintf("stress test message %d-%d", id, j)
		if test.MessageSize > 0 {
			msg = generateLargeMessage(msg, test.MessageSize)
		}

		l.Print(msg)
	}
}

// runDurationBasedStress runs operations for a specified duration.
func runDurationBasedStress(logger slog.Logger, id int, test StressTest) {
	endTime := time.Now().Add(test.Duration)
	operation := 0

	for time.Now().Before(endTime) {
		l := logger.Info().
			WithField("goroutine", id).
			WithField("operation", operation)

		if test.EnableMemoryPressure {
			l = addMemoryPressure(l, test)
		}

		msg := fmt.Sprintf("stress test message %d-%d", id, operation)
		if test.MessageSize > 0 {
			msg = generateLargeMessage(msg, test.MessageSize)
		}

		l.Print(msg)
		operation++
	}
}

// addMemoryPressure adds many fields to create memory pressure.
func addMemoryPressure(logger slog.Logger, test StressTest) slog.Logger {
	for i := 0; i < test.FieldsPerMessage; i++ {
		logger = logger.WithField(
			fmt.Sprintf("field_%d", i),
			fmt.Sprintf("value_%d_with_some_extra_content_to_increase_memory_usage", i),
		)
	}
	return logger
}

// generateLargeMessage creates a message of specified size.
func generateLargeMessage(prefix string, size int) string {
	if len(prefix) >= size {
		return prefix[:size]
	}

	// Create a message of the specified size
	padding := size - len(prefix)
	return prefix + string(make([]byte, padding))
}

// reportStressMetrics reports performance metrics from the stress test.
func reportStressMetrics(t core.T, test StressTest, duration time.Duration) {
	t.Helper()

	var totalOps int
	if test.Duration > 0 {
		t.Logf("Stress test completed: %d goroutines for %v (duration: %v)",
			test.Goroutines, test.Duration, duration)
	} else {
		totalOps = test.Goroutines * test.Operations
		opsPerSecond := float64(totalOps) / duration.Seconds()
		t.Logf("Stress test completed: %d goroutines × %d operations = %d total messages in %v (%.0f ops/sec)",
			test.Goroutines, test.Operations, totalOps, duration, opsPerSecond)
	}

	if test.EnableMemoryPressure {
		t.Logf("Memory pressure enabled: %d fields per message, %d bytes per message",
			test.FieldsPerMessage, test.MessageSize)
	}
}

// verifyStressResults performs basic verification of stress test results.
func verifyStressResults(t core.T, messages []Message, test StressTest) {
	t.Helper()

	// Verify message count
	verifyMessageCount(t, messages, test)

	// Verify message structure
	verifyMessageStructure(t, messages)
}

// verifyMessageCount checks if the correct number of messages were generated.
func verifyMessageCount(t core.T, messages []Message, test StressTest) {
	t.Helper()

	if test.Duration > 0 {
		verifyDurationBasedCount(t, messages)
	} else {
		verifyOperationBasedCount(t, messages, test)
	}
}

// verifyDurationBasedCount verifies message count for duration-based tests.
func verifyDurationBasedCount(t core.T, messages []Message) {
	t.Helper()

	if len(messages) == 0 {
		t.Error("No messages recorded during duration-based stress test")
	}
	t.Logf("Duration-based stress test produced %d messages", len(messages))
}

// verifyOperationBasedCount verifies message count for operation-based tests.
func verifyOperationBasedCount(t core.T, messages []Message, test StressTest) {
	t.Helper()

	expected := test.Goroutines * test.Operations
	if len(messages) != expected {
		t.Errorf("Expected %d messages, got %d", expected, len(messages))
	}
}

// verifyMessageStructure checks that all messages have required fields.
func verifyMessageStructure(t core.T, messages []Message) {
	t.Helper()

	requiredFields := []string{"goroutine", "operation"}

	for i, msg := range messages {
		for _, field := range requiredFields {
			if msg.Fields[field] == nil {
				t.Errorf("Message %d missing %s field", i, field)
			}
		}
	}
}

// StressTestSuite runs a comprehensive suite of stress tests.
type StressTestSuite struct {
	// Logger factory for creating test instances
	NewLogger func() slog.Logger

	// Optional: Logger factory that creates logger with recorder
	NewLoggerWithRecorder func(slog.Logger) slog.Logger

	// Skip specific stress tests
	SkipHighVolume      bool
	SkipMemoryPressure  bool
	SkipDurationBased   bool
	SkipConcurrentField bool
}

// stressTestCase defines a stress test case.
type stressTestCase struct {
	stressTestFactory func() StressTest
	name              string
	skip              bool
}

// Run executes the stress test suite.
func (sts StressTestSuite) Run(t *testing.T) {
	// Define all stress test cases
	tests := sts.getStressTestCases()

	// Run each test case
	for _, tc := range tests {
		if !tc.skip {
			t.Run(tc.name, func(t *testing.T) {
				sts.runStressTest(t, tc.stressTestFactory())
			})
		}
	}

	// Run special test cases
	sts.runSpecialTests(t)
}

// getStressTestCases returns all standard stress test cases.
func (sts StressTestSuite) getStressTestCases() []stressTestCase {
	return []stressTestCase{
		{
			name:              "BasicStress",
			skip:              false,
			stressTestFactory: DefaultStressTest,
		},
		{
			name:              "HighVolumeStress",
			skip:              sts.SkipHighVolume,
			stressTestFactory: HighVolumeStressTest,
		},
		{
			name:              "MemoryPressureStress",
			skip:              sts.SkipMemoryPressure,
			stressTestFactory: MemoryPressureStressTest,
		},
		{
			name: "DurationBasedStress",
			skip: sts.SkipDurationBased,
			stressTestFactory: func() StressTest {
				return DurationBasedStressTest(100 * time.Millisecond)
			},
		},
	}
}

// runStressTest runs a single stress test with a new logger.
func (sts StressTestSuite) runStressTest(t *testing.T, test StressTest) {
	t.Helper()
	logger := sts.NewLogger()
	RunStressTest(t, logger, test)
}

// runSpecialTests runs special test cases that require different handling.
func (sts StressTestSuite) runSpecialTests(t *testing.T) {
	// Run concurrent field stress test if not skipped
	if !sts.SkipConcurrentField {
		t.Run("ConcurrentFieldStress", func(t *testing.T) {
			TestConcurrentFields(t, sts.NewLogger)
		})
	}

	// Run verified stress tests if recorder is available
	if sts.NewLoggerWithRecorder != nil {
		t.Run("VerifiedStress", func(t *testing.T) {
			sts.runVerifiedStressTests(t)
		})
	}
}

// runVerifiedStressTests runs stress tests with result verification.
func (sts StressTestSuite) runVerifiedStressTests(t *testing.T) {
	t.Helper()

	t.Run("BasicVerified", func(t *testing.T) {
		recorder := NewLogger()
		logger := sts.NewLoggerWithRecorder(recorder)

		opts := &StressTestOptions{
			VerifyResults: true,
			GetMessages:   recorder.GetMessages,
		}

		RunStressTestWithOptions(t, logger, DefaultStressTest(), opts)
	})
}
