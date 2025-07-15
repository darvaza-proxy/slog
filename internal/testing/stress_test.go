package testing_test

import (
	"sync/atomic"
	"testing"
	"time"

	"darvaza.org/slog"
	slogtest "darvaza.org/slog/internal/testing"
)

func TestDefaultStressTest(t *testing.T) {
	logger := slogtest.NewLogger()
	stress := slogtest.DefaultStressTest()

	slogtest.RunStressTest(t, logger, stress)

	// Verify we got the expected number of messages
	messages := logger.GetMessages()
	expected := stress.Goroutines * stress.Operations
	if len(messages) != expected {
		t.Errorf("Expected %d messages, got %d", expected, len(messages))
	}
}

func TestHighVolumeStressTest(t *testing.T) {
	logger := slogtest.NewLogger()
	stress := slogtest.HighVolumeStressTest()

	slogtest.RunStressTest(t, logger, stress)

	messages := logger.GetMessages()
	expected := stress.Goroutines * stress.Operations
	if len(messages) != expected {
		t.Errorf("Expected %d messages, got %d", expected, len(messages))
	}
}

func TestMemoryPressureStressTest(t *testing.T) {
	logger := slogtest.NewLogger()
	stress := slogtest.MemoryPressureStressTest()

	slogtest.RunStressTest(t, logger, stress)

	messages := logger.GetMessages()
	expected := stress.Goroutines * stress.Operations
	if len(messages) != expected {
		t.Errorf("Expected %d messages, got %d", expected, len(messages))
	}

	// Verify fields were added
	for i, msg := range messages {
		// Check that memory pressure fields were added
		fieldCount := len(msg.Fields)
		// At least goroutine, operation, and the memory pressure fields
		minExpected := 2 + stress.FieldsPerMessage
		if fieldCount < minExpected {
			t.Errorf("Message %d has %d fields, expected at least %d", i, fieldCount, minExpected)
			break
		}
	}
}

func TestDurationBasedStressTest(t *testing.T) {
	logger := slogtest.NewLogger()
	duration := 50 * time.Millisecond
	stress := slogtest.DurationBasedStressTest(duration)

	startTime := time.Now()
	slogtest.RunStressTest(t, logger, stress)
	elapsed := time.Since(startTime)

	// Verify test ran for approximately the requested duration
	if elapsed < duration || elapsed > duration*2 {
		t.Errorf("Test ran for %v, expected approximately %v", elapsed, duration)
	}

	messages := logger.GetMessages()
	if len(messages) == 0 {
		t.Error("Duration-based test produced no messages")
	}
}

func TestStressTestWithOptions(t *testing.T) {
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
		VerifyFunc: func(t *testing.T, messages []slogtest.Message, _ slogtest.StressTest) {
			verifyFuncCalled = true
			if len(messages) == 0 {
				t.Error("Custom verify: no messages")
			}
		},
	}

	stress := slogtest.DefaultStressTest()
	slogtest.RunStressTestWithOptions(t, logger, stress, opts)

	if !preStressCalled {
		t.Error("PreStressFunc was not called")
	}
	if !postStressCalled {
		t.Error("PostStressFunc was not called")
	}
	if !verifyFuncCalled {
		t.Error("VerifyFunc was not called")
	}
}

func TestStressTestSuite(t *testing.T) {
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

func TestCustomStressFunction(t *testing.T) {
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
	if len(messages) != 50 {
		t.Errorf("Expected 50 messages, got %d", len(messages))
	}
}
