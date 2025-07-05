package filter_test

import (
	slogtest "darvaza.org/slog/internal/testing"
)

// Type alias for backward compatibility
type testLogger = slogtest.Logger

// Helper function to create a new test logger
func newTestLogger() *testLogger {
	return slogtest.NewLogger()
}
