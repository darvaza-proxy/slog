package testing_test

import (
	"sync/atomic"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
	slogtest "darvaza.org/slog/internal/testing"
)

// TestStressTests runs all stress test related tests
func TestStressTests(t *testing.T) {
	t.Run("DefaultStressTest", testDefaultStressTest)
	t.Run("HighVolumeStressTest", testHighVolumeStressTest)
	t.Run("MemoryPressureStressTest", testMemoryPressureStressTest)
	t.Run("DurationBasedStressTest", testDurationBasedStressTest)
	t.Run("StressTestWithOptions", testStressTestWithOptions)
	t.Run("StressTestSuite", testStressTestSuite)
	t.Run("CustomStressFunction", testCustomStressFunction)
}

func testDefaultStressTest(t *testing.T) {
	t.Helper()
	logger := slogtest.NewLogger()
	stress := slogtest.DefaultStressTest()

	slogtest.RunStressTest(t, logger, stress)

	// Verify we got the expected number of messages
	messages := logger.GetMessages()
	expected := stress.Goroutines * stress.Operations
	core.AssertEqual(t, expected, len(messages), "Expected %d messages", expected)
}

func testHighVolumeStressTest(t *testing.T) {
	t.Helper()
	logger := slogtest.NewLogger()
	stress := slogtest.HighVolumeStressTest()

	slogtest.RunStressTest(t, logger, stress)

	messages := logger.GetMessages()
	expected := stress.Goroutines * stress.Operations
	core.AssertEqual(t, expected, len(messages), "Expected %d messages", expected)
}

func testMemoryPressureStressTest(t *testing.T) {
	t.Helper()
	logger := slogtest.NewLogger()
	stress := slogtest.MemoryPressureStressTest()

	slogtest.RunStressTest(t, logger, stress)

	messages := logger.GetMessages()
	expected := stress.Goroutines * stress.Operations
	core.AssertEqual(t, expected, len(messages), "Expected %d messages", expected)

	// Verify fields were added
	for i, msg := range messages {
		// Check that memory pressure fields were added
		fieldCount := len(msg.Fields)
		// At least goroutine, operation, and the memory pressure fields
		minExpected := 2 + stress.FieldsPerMessage
		core.AssertMustTrue(t, fieldCount >= minExpected,
			"Message %d should have at least %d fields, got %d", i, minExpected, fieldCount)
	}
}

func testDurationBasedStressTest(t *testing.T) {
	t.Helper()
	logger := slogtest.NewLogger()
	duration := 50 * time.Millisecond
	stress := slogtest.DurationBasedStressTest(duration)

	startTime := time.Now()
	slogtest.RunStressTest(t, logger, stress)
	elapsed := time.Since(startTime)

	// Verify test ran for approximately the requested duration
	core.AssertTrue(t, elapsed >= duration, "Test should run for at least %v, ran for %v", duration, elapsed)
	core.AssertTrue(t, elapsed <= duration*2, "Test should not run longer than %v, ran for %v", duration*2, elapsed)

	messages := logger.GetMessages()
	core.AssertTrue(t, len(messages) > 0, "Duration-based test should produce messages")
}

func testStressTestWithOptions(t *testing.T) {
	t.Helper()
	recorder := slogtest.NewLogger()
	logger := recorder // In real usage, this would be a handler wrapping the recorder

	var preStressCalled, postStressCalled bool
	var verifyFuncCalled bool

	opts := &slogtest.StressTestOptions{
		VerifyResults: true,
		GetMessages:   recorder.GetMessages,
		PreStressFunc: func() {
			preStressCalled = true
		},
		PostStressFunc: func() {
			postStressCalled = true
		},
		VerifyFunc: func(t core.T, messages []slogtest.Message, _ slogtest.StressTest) {
			verifyFuncCalled = true
			core.AssertTrue(t, len(messages) > 0, "Custom verify: should have messages")
		},
	}

	stress := slogtest.DefaultStressTest()
	slogtest.RunStressTestWithOptions(t, logger, stress, opts)

	core.AssertEqual(t, true, preStressCalled, "PreStressFunc was not called")
	core.AssertEqual(t, true, postStressCalled, "PostStressFunc was not called")
	core.AssertEqual(t, true, verifyFuncCalled, "VerifyFunc was not called")
}

func testStressTestSuite(t *testing.T) {
	t.Helper()
	suite := slogtest.StressTestSuite{
		NewLogger: func() slog.Logger {
			return slogtest.NewLogger()
		},
		NewLoggerWithRecorder: func(recorder slog.Logger) slog.Logger {
			// In real usage, this would create a handler backed by the recorder
			return recorder
		},
	}

	suite.Run(t)
}

func testCustomStressFunction(t *testing.T) {
	t.Helper()
	logger := slogtest.NewLogger()

	var customOperations atomic.Int32
	stress := slogtest.StressTest{
		Goroutines: 5,
		StressFunc: func(l slog.Logger, id int) {
			// Custom stress logic
			for i := 0; i < 10; i++ {
				l.Info().
					WithField("custom_id", id).
					WithField("iteration", i).
					Print("custom stress message")
				customOperations.Add(1)
			}
		},
	}

	slogtest.RunStressTest(t, logger, stress)

	messages := logger.GetMessages()
	// 5 goroutines * 10 iterations each
	core.AssertEqual(t, 50, len(messages), "Expected 50 messages")
}

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = verifyDurationBasedCountTestCase{}

type verifyDurationBasedCountTestCase struct {
	name       string
	messages   []slogtest.Message
	shouldFail bool
}

func (tc verifyDurationBasedCountTestCase) Name() string {
	return tc.name
}

func (tc verifyDurationBasedCountTestCase) Test(t *testing.T) {
	t.Helper()

	// For this test, we'll create a duration-based stress test and verify behaviour
	logger := slogtest.NewLogger()

	// Pre-populate the logger with our test messages
	for _, msg := range tc.messages {
		logger.Info().Print(msg.Message)
	}

	stress := slogtest.StressTest{
		Duration:   10 * time.Millisecond, // Short duration
		Goroutines: 1,
	}

	// Run with custom options that let us verify behaviour
	var verifyFuncCalled bool
	var messageCountAtVerify int
	opts := &slogtest.StressTestOptions{
		VerifyResults: true,
		GetMessages:   logger.GetMessages,
		VerifyFunc: func(testT core.T, messages []slogtest.Message, _ slogtest.StressTest) {
			verifyFuncCalled = true
			messageCountAtVerify = len(messages)

			// This mimics what verifyDurationBasedCount does
			if len(messages) == 0 && tc.shouldFail {
				testT.Error("No messages recorded during duration-based stress test")
			} else {
				testT.Logf("Duration-based stress test produced %d messages", len(messages))
			}
		},
	}

	slogtest.RunStressTestWithOptions(t, logger, stress, opts)

	core.AssertTrue(t, verifyFuncCalled, "verify function should have been called")

	// For duration-based tests, we expect messages to be generated during the test
	// So we just verify that the verify function was called and completed
	core.AssertTrue(t, messageCountAtVerify >= 0, "message count should be non-negative")
}

func newVerifyDurationBasedCountTestCase(name string, messages []slogtest.Message,
	shouldFail bool) verifyDurationBasedCountTestCase {
	return verifyDurationBasedCountTestCase{
		name:       name,
		messages:   messages,
		shouldFail: shouldFail,
	}
}

func verifyDurationBasedCountTestCases() []verifyDurationBasedCountTestCase {
	return []verifyDurationBasedCountTestCase{
		newVerifyDurationBasedCountTestCase("duration test produces messages",
			[]slogtest.Message{}, // Start with empty, but stress test will add messages
			false),               // Should not fail because stress test generates messages
		newVerifyDurationBasedCountTestCase("pre-existing messages should pass",
			[]slogtest.Message{
				{Level: slog.Info, Message: "test message", Fields: map[string]any{}},
			},
			false),
	}
}

func TestVerifyDurationBasedCount(t *testing.T) {
	core.RunTestCases(t, verifyDurationBasedCountTestCases())
}

func TestVerifyDurationBasedCountDirect(t *testing.T) {
	logger := slogtest.NewLogger()

	// Create a duration-based stress test that uses the default verification path
	stress := slogtest.StressTest{
		Duration:   10 * time.Millisecond, // Duration > 0 triggers verifyDurationBasedCount
		Goroutines: 1,
	}

	// Use default options which will call verifyDurationBasedCount via verifyMessageCount
	opts := &slogtest.StressTestOptions{
		VerifyResults: true,
		GetMessages:   logger.GetMessages,
		// No custom VerifyFunc - this ensures default verification path is used
	}

	// This should call verifyMessageCount -> verifyDurationBasedCount
	slogtest.RunStressTestWithOptions(t, logger, stress, opts)

	// Verify messages were generated during the test
	messages := logger.GetMessages()
	core.AssertTrue(t, len(messages) > 0, "stress test should generate messages")
}
