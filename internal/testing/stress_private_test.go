package testing

import (
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = verifyDurationBasedCountPrivateTestCase{}

type verifyDurationBasedCountPrivateTestCase struct {
	name     string
	messages []Message
	wantErr  bool
}

func (tc verifyDurationBasedCountPrivateTestCase) Name() string {
	return tc.name
}

func (tc verifyDurationBasedCountPrivateTestCase) Test(t *testing.T) {
	t.Helper()

	// Create a mock T to capture errors
	mockT := &core.MockT{}

	// Call the private function directly
	verifyDurationBasedCount(mockT, tc.messages)

	// Verify the expected behaviour
	if tc.wantErr {
		core.AssertTrue(t, mockT.Failed(), "should record error")
		core.AssertTrue(t, mockT.HasErrors(), "should have error messages")

		lastError, ok := mockT.LastError()
		core.AssertTrue(t, ok, "should have captured error")
		core.AssertContains(t, lastError, "No messages recorded during duration-based stress test", "error message")
	} else {
		core.AssertFalse(t, mockT.Failed(), "should not record error")
		core.AssertTrue(t, mockT.HasLogs(), "should have log messages")

		lastLog, ok := mockT.LastLog()
		core.AssertTrue(t, ok, "should have captured log")
		core.AssertContains(t, lastLog, "Duration-based stress test produced", "log message")
	}
}

func newVerifyDurationBasedCountPrivateTestCase(
	name string, messages []Message, wantErr bool,
) verifyDurationBasedCountPrivateTestCase {
	return verifyDurationBasedCountPrivateTestCase{
		name:     name,
		messages: messages,
		wantErr:  wantErr,
	}
}

func verifyDurationBasedCountPrivateTestCases() []verifyDurationBasedCountPrivateTestCase {
	return []verifyDurationBasedCountPrivateTestCase{
		newVerifyDurationBasedCountPrivateTestCase("empty messages should error",
			[]Message{}, true),
		newVerifyDurationBasedCountPrivateTestCase("single message should pass",
			[]Message{
				{Level: 5, Message: "test message", Fields: map[string]any{}},
			}, false),
		newVerifyDurationBasedCountPrivateTestCase("multiple messages should pass",
			[]Message{
				{Level: 5, Message: "test message 1", Fields: map[string]any{}},
				{Level: 3, Message: "test message 2", Fields: map[string]any{"key": "value"}},
			}, false),
	}
}

func TestVerifyDurationBasedCountPrivate(t *testing.T) {
	core.RunTestCases(t, verifyDurationBasedCountPrivateTestCases())
}
