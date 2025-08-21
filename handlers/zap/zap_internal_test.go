package zap

import (
	"testing"

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
