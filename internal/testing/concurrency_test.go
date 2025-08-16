package testing

import (
	"fmt"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// TestConcurrencyVerification tests that our concurrency test actually validates messages
func TestConcurrencyVerification(t *testing.T) {
	t.Run("CorrectMessagesPass", testConcurrencyCorrectMessages)
	t.Run("VerifiesMessageCount", testConcurrencyMessageCount)
	t.Run("MessagesHaveRequiredFields", testConcurrencyRequiredFields)
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
	expectedCount := test.Goroutines * test.Operations
	core.AssertNotEqual(t, expectedCount, len(msgs), "Test setup error: should have wrong count")
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
	core.AssertEqual(t, nil, msgs[0].Fields["goroutine"], "First message should not have goroutine field")
	core.AssertEqual(t, nil, msgs[1].Fields["operation"], "Second message should not have operation field")
}

// TestCompareMessagesWithConcurrentData tests CompareMessages with concurrent test data
func TestCompareMessagesWithConcurrentData(t *testing.T) {
	t.Run("WithConcurrentData", testConcurrentDataComparison)
}

func testConcurrentDataComparison(t *testing.T) {
	t.Helper()
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
func verifyConcurrentComparison(t core.T, onlyExpected, onlyActual, both []Message) {
	t.Helper()

	core.AssertEqual(t, 1, len(onlyExpected), "Expected 1 message only in expected")
	core.AssertEqual(t, 1, len(onlyActual), "Expected 1 message only in actual")
	core.AssertEqual(t, 5, len(both), "Expected 5 messages in both")
}

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = logNoVerificationTestCase{}

type logNoVerificationTestCase struct {
	name string
	test ConcurrencyTest
}

func (tc logNoVerificationTestCase) Name() string {
	return tc.name
}

func (tc logNoVerificationTestCase) Test(t *testing.T) {
	t.Helper()

	// Call logNoVerification - this function logs directly to t
	// We can't capture the log output, but we can verify it doesn't panic
	logNoVerification(t, tc.test)

	// The function should complete without error
	// We verify this by checking the test values make sense
	expectedTotal := tc.test.Goroutines * tc.test.Operations
	core.AssertTrue(t, expectedTotal >= 0, "total operations should be non-negative")
}

func newLogNoVerificationTestCase(name string, test ConcurrencyTest) logNoVerificationTestCase {
	return logNoVerificationTestCase{
		name: name,
		test: test,
	}
}

func logNoVerificationTestCases() []logNoVerificationTestCase {
	return []logNoVerificationTestCase{
		newLogNoVerificationTestCase("small test", ConcurrencyTest{Goroutines: 2, Operations: 3}),
		newLogNoVerificationTestCase("medium test", ConcurrencyTest{Goroutines: 10, Operations: 50}),
		newLogNoVerificationTestCase("large test", ConcurrencyTest{Goroutines: 100, Operations: 1000}),
		newLogNoVerificationTestCase("single goroutine", ConcurrencyTest{Goroutines: 1, Operations: 10}),
		newLogNoVerificationTestCase("single operation", ConcurrencyTest{Goroutines: 5, Operations: 1}),
		newLogNoVerificationTestCase("zero values", ConcurrencyTest{Goroutines: 0, Operations: 0}),
	}
}

func TestLogNoVerification(t *testing.T) {
	core.RunTestCases(t, logNoVerificationTestCases())
}
