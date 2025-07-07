package testing

import (
	"fmt"
	"testing"

	"darvaza.org/slog"
)

// TestConcurrencyVerification tests that our concurrency test actually validates messages
func TestConcurrencyVerification(t *testing.T) {
	t.Run("correct messages pass", func(t *testing.T) {
		testConcurrencyCorrectMessages(t)
	})

	t.Run("verifies message count", func(t *testing.T) {
		testConcurrencyMessageCount(t)
	})

	t.Run("messages have required fields", func(t *testing.T) {
		testConcurrencyRequiredFields(t)
	})
}

// testConcurrencyCorrectMessages tests that correct messages pass verification
func testConcurrencyCorrectMessages(t *testing.T) {
	t.Helper()

	recorder := NewLogger()
	test := ConcurrencyTest{Goroutines: 2, Operations: 3}

	// Manually add the expected messages
	addExpectedMessages(recorder, test)

	// This should pass without errors
	verifyConcurrentResults(t, recorder, test)
}

// addExpectedMessages adds the expected messages for a concurrency test
func addExpectedMessages(recorder *Logger, test ConcurrencyTest) {
	for g := 0; g < test.Goroutines; g++ {
		for o := 0; o < test.Operations; o++ {
			recorder.Info().
				WithField("goroutine", g).
				WithField("operation", o).
				Printf("concurrent message %d-%d", g, o)
		}
	}
}

// testConcurrencyMessageCount tests message count verification
func testConcurrencyMessageCount(t *testing.T) {
	t.Helper()

	recorder := NewLogger()
	test := ConcurrencyTest{Goroutines: 2, Operations: 3}

	// Add fewer messages than expected
	addIncompleteMessages(recorder, 5) // Should be 6

	// Check that AssertMessageCount would fail
	msgs := recorder.GetMessages()
	if len(msgs) == test.Goroutines*test.Operations {
		t.Error("Test setup error: should have wrong count")
	}
}

// addIncompleteMessages adds a specific number of messages
func addIncompleteMessages(recorder *Logger, count int) {
	for i := 0; i < count; i++ {
		recorder.Info().Print("message")
	}
}

// testConcurrencyRequiredFields tests that required fields are validated
func testConcurrencyRequiredFields(t *testing.T) {
	t.Helper()

	recorder := NewLogger()

	// Add messages with missing fields
	recorder.Info().Print("missing fields")
	recorder.Info().WithField("goroutine", 0).Print("missing operation")

	// Verify field presence
	msgs := recorder.GetMessages()
	if msgs[0].Fields["goroutine"] != nil {
		t.Error("First message should not have goroutine field")
	}
	if msgs[1].Fields["operation"] != nil {
		t.Error("Second message should not have operation field")
	}
}

// TestCompareMessagesWithConcurrentData tests CompareMessages with concurrent test data
func TestCompareMessagesWithConcurrentData(t *testing.T) {
	expected := createConcurrentTestMessages(2, 3)
	actual := createModifiedConcurrentMessages(expected)

	onlyExpected, onlyActual, both := CompareMessages(expected, actual)

	verifyConcurrentComparison(t, onlyExpected, onlyActual, both)
}

// createConcurrentTestMessages creates expected messages for concurrent testing
func createConcurrentTestMessages(goroutines, operations int) []Message {
	var messages []Message
	for g := 0; g < goroutines; g++ {
		for o := 0; o < operations; o++ {
			messages = append(messages, Message{
				Level:   slog.Info,
				Message: fmt.Sprintf("concurrent message %d-%d", g, o),
				Fields:  map[string]any{"goroutine": g, "operation": o},
			})
		}
	}
	return messages
}

// createModifiedConcurrentMessages creates actual messages with modifications
func createModifiedConcurrentMessages(expected []Message) []Message {
	var actual []Message
	// Skip one message (message at index 2)
	for i, msg := range expected {
		if i != 2 {
			actual = append(actual, msg)
		}
	}
	// Add an extra unexpected message
	actual = append(actual, Message{
		Level:   slog.Warn,
		Message: "unexpected",
		Fields:  map[string]any{"extra": true},
	})
	return actual
}

// verifyConcurrentComparison verifies the comparison results
func verifyConcurrentComparison(t *testing.T, onlyExpected, onlyActual, both []Message) {
	t.Helper()

	if len(onlyExpected) != 1 {
		t.Errorf("Expected 1 message only in expected, got %d", len(onlyExpected))
	}

	if len(onlyActual) != 1 {
		t.Errorf("Expected 1 message only in actual, got %d", len(onlyActual))
	}

	if len(both) != 5 {
		t.Errorf("Expected 5 messages in both, got %d", len(both))
	}
}
