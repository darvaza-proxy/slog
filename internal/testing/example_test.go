package testing_test

import (
	"testing"

	"darvaza.org/slog"
	slogtest "darvaza.org/slog/internal/testing"
)

// TestRecorderExample demonstrates using the test recorder for handler testing
func TestRecorderExample(t *testing.T) {
	// Create a test logger that records messages
	recorder := slogtest.NewLogger()

	// Use it as you would any logger
	recorder.Info().
		WithField("user", "john").
		WithField("action", "login").
		Print("User logged in")

	// Verify the recorded messages
	messages := recorder.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	// Use helper assertions
	msg := messages[0]
	slogtest.AssertMessage(t, msg, slog.Info, "User logged in")
	slogtest.AssertField(t, msg, "user", "john")
	slogtest.AssertField(t, msg, "action", "login")
}

// TestComplianceExample demonstrates running the compliance test suite
func TestComplianceExample(t *testing.T) {
	// Define how to create your logger
	compliance := slogtest.ComplianceTest{
		NewLogger: func() slog.Logger {
			// For a real handler, return a new instance
			return slogtest.NewLogger()
		},
		// Skip tests that might not apply
		SkipPanicTests: true, // if your logger exits on panic
	}

	// Run the full compliance suite
	compliance.Run(t)
}

// TestCustomHandlerExample demonstrates testing a custom handler
func TestCustomHandlerExample(t *testing.T) {
	// Test level methods with a fresh logger each time
	slogtest.TestLevelMethods(t, func() slog.Logger {
		return slogtest.NewLogger()
	})

	// Test field methods with a fresh logger each time
	slogtest.TestFieldMethods(t, func() slog.Logger {
		return slogtest.NewLogger()
	})

	// Test concurrency with a fresh logger
	concurrentLogger := slogtest.NewLogger()
	slogtest.RunConcurrentTest(t, concurrentLogger, slogtest.DefaultConcurrencyTest())

	// Verify specific behavior with a separate logger
	testLogger := slogtest.NewLogger()
	testLogger.Info().WithField("test", "value").Print("message")

	msgs := testLogger.GetMessages()
	slogtest.AssertMessageCount(t, msgs, 1)
	slogtest.AssertMessage(t, msgs[0], slog.Info, "message")
	slogtest.AssertField(t, msgs[0], "test", "value")
}
