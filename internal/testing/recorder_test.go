package testing

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog/handlers/mock"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = recorderFactoryTestCase{}

type recorderFactoryTestCase struct {
	name            string
	expectNonNil    bool
	expectTypeValid bool
	expectEmpty     bool
}

func (tc recorderFactoryTestCase) Name() string {
	return tc.name
}

func (tc recorderFactoryTestCase) Test(t *testing.T) {
	t.Helper()

	recorder := NewRecorder()

	if !tc.expectNonNil {
		core.AssertNil(t, recorder, "recorder")
		return
	}
	core.AssertNotNil(t, recorder, "recorder")

	if tc.expectTypeValid {
		mockRecorder, ok := core.AssertTypeIs[*mock.Recorder](t, recorder, "type cast")
		if ok && tc.expectEmpty {
			messages := mockRecorder.GetMessages()
			core.AssertEqual(t, 0, len(messages), "initial message count")

			// Clear should not panic
			mockRecorder.Clear()
			messages = mockRecorder.GetMessages()
			core.AssertEqual(t, 0, len(messages), "cleared message count")
		}
	}

	// Test uniqueness - each call creates fresh instance
	recorder2 := NewRecorder()
	core.AssertNotEqual(t, recorder, recorder2, "each call creates new instance")
}

func newRecorderFactoryTestCase(name string, expectNonNil, expectTypeValid, expectEmpty bool) recorderFactoryTestCase {
	return recorderFactoryTestCase{
		name:            name,
		expectNonNil:    expectNonNil,
		expectTypeValid: expectTypeValid,
		expectEmpty:     expectEmpty,
	}
}

func recorderFactoryTestCases() []recorderFactoryTestCase {
	return []recorderFactoryTestCase{
		newRecorderFactoryTestCase("creates non-nil recorder", true, true, true),
	}
}

func TestNewRecorder(t *testing.T) {
	core.RunTestCases(t, recorderFactoryTestCases())
}

var _ core.TestCase = loggerFactoryTestCase{}

type loggerFactoryTestCase struct {
	name             string
	expectNonNil     bool
	expectTypeValid  bool
	testMessage      string
	expectedMsgCount int
}

func (tc loggerFactoryTestCase) Name() string {
	return tc.name
}

func (tc loggerFactoryTestCase) Test(t *testing.T) {
	t.Helper()

	logger := NewLogger()

	if !tc.expectNonNil {
		core.AssertNil(t, logger, "logger")
		return
	}
	core.AssertNotNil(t, logger, "logger")

	if tc.expectTypeValid {
		mockLogger, ok := core.AssertTypeIs[*mock.Logger](t, logger, "type cast")
		if ok {
			core.AssertNotNil(t, mockLogger, "mock logger")

			// Test logging functionality if test message provided
			if tc.testMessage != "" {
				mockLogger.Info().Print(tc.testMessage)
				messages := mockLogger.GetMessages()
				core.AssertEqual(t, tc.expectedMsgCount, len(messages), "message count after logging")
			}
		}
	}

	// Test uniqueness - each call creates fresh instance
	logger2 := NewLogger()
	core.AssertNotEqual(t, logger, logger2, "each call creates new instance")
}

func newLoggerFactoryTestCase(name string, expectNonNil, expectTypeValid bool,
	testMessage string, expectedMsgCount int) loggerFactoryTestCase {
	return loggerFactoryTestCase{
		name:             name,
		expectNonNil:     expectNonNil,
		expectTypeValid:  expectTypeValid,
		testMessage:      testMessage,
		expectedMsgCount: expectedMsgCount,
	}
}

func loggerFactoryTestCases() []loggerFactoryTestCase {
	return []loggerFactoryTestCase{
		newLoggerFactoryTestCase("creates functional logger", true, true, "test message", 1),
		newLoggerFactoryTestCase("creates valid type", true, true, "", 0),
	}
}

func TestNewLogger(t *testing.T) {
	core.RunTestCases(t, loggerFactoryTestCases())
}
