package zap

import (
	"testing"

	"go.uber.org/zap"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// TestMapToZapLevel_InvalidLevel tests invalid level handling using internal access
func TestMapToZapLevel_InvalidLevel(t *testing.T) {
	// This tests the default case in mapToZapLevel (lines 249-251)
	// by creating a logger with an invalid level and testing Enabled()

	core.AssertPanic(t, func() {
		logger, _ := New(nil)
		zapLogger := core.AssertMustTypeIs[*Logger](t, logger, "logger type")

		// Create a new logger with an invalid level by directly setting the internal field
		invalidLogger := &Logger{
			loglet: zapLogger.loglet.WithLevel(slog.LogLevel(100)), // Invalid level
			logger: zapLogger.logger,
			config: zapLogger.config,
		}

		// This should panic when calling Enabled()
		invalidLogger.Enabled()
	}, nil, "invalid level panic")
}

// newWithInvalidLogger passes NewWithLogger the given logger, as the
// caller its panic names.
func newWithInvalidLogger(logger *zap.Logger) {
	NewWithLogger(logger)
}

var _ core.TestCase = invalidLoggerTestCase{}

type invalidLoggerTestCase struct {
	logger *zap.Logger
	name   string
}

func (tc invalidLoggerTestCase) Name() string {
	return tc.name
}

func (tc invalidLoggerTestCase) Test(t *testing.T) {
	t.Helper()
	err := core.Catch(func() error {
		newWithInvalidLogger(tc.logger)
		return nil
	})
	core.AssertErrorIs(t, err, errInvalidLogger, "error")
	core.AssertTopFrame(t, err, "newWithInvalidLogger", "caller frame")
}

// newInvalidLoggerTestCase creates a case whose logger NewWithLogger
// rejects with errInvalidLogger.
func newInvalidLoggerTestCase(name string,
	logger *zap.Logger) invalidLoggerTestCase {
	return invalidLoggerTestCase{
		logger: logger,
		name:   name,
	}
}

func invalidLoggerTestCases() []invalidLoggerTestCase {
	return []invalidLoggerTestCase{
		newInvalidLoggerTestCase("Nil", nil),
		newInvalidLoggerTestCase("NoCore", new(zap.Logger)),
	}
}

func TestNewWithLoggerInvalid(t *testing.T) {
	core.RunTestCases(t, invalidLoggerTestCases())
}
