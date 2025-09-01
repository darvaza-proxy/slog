package filter_test

import (
	"testing"
	"time"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
	"darvaza.org/slog/handlers/mock"
	slogtest "darvaza.org/slog/internal/testing"
)

// TestFilterStress runs comprehensive stress tests on the filter handler.
func TestFilterStress(t *testing.T) {
	suite := slogtest.StressTestSuite{
		NewLogger: func() slog.Logger {
			backend := mock.NewLogger()
			return filter.New(backend, slog.Debug)
		},
		NewLoggerWithRecorder: func(recorder slog.Logger) slog.Logger {
			return filter.New(recorder, slog.Debug)
		},
	}

	suite.Run(t)
}

// TestFilterHighVolumeStress tests filter under high volume conditions.
func TestFilterHighVolumeStress(t *testing.T) {
	backend := mock.NewLogger()
	logger := filter.New(backend, slog.Debug)

	stress := slogtest.HighVolumeStressTest()
	slogtest.RunStressTest(t, logger, stress)
}

// TestFilterMemoryPressureStress tests filter under memory pressure.
func TestFilterMemoryPressureStress(t *testing.T) {
	backend := mock.NewLogger()
	logger := filter.New(backend, slog.Debug)

	stress := slogtest.MemoryPressureStressTest()
	slogtest.RunStressTest(t, logger, stress)
}

// TestFilterDurationBasedStress tests filter for a specific duration.
func TestFilterDurationBasedStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping duration-based stress test in short mode")
	}

	backend := mock.NewLogger()
	logger := filter.New(backend, slog.Debug)

	// Run for 2 seconds
	stress := slogtest.DurationBasedStressTest(2 * time.Second)
	slogtest.RunStressTest(t, logger, stress)
}

// createCustomFilterLogger creates a filter with custom field and message filters.
func createCustomFilterLogger(parent slog.Logger) slog.Logger {
	return &filter.Logger{
		Parent:    parent,
		Threshold: slog.Debug,
		// Complex field filter that transforms fields
		FieldFilter: func(key string, val any) (string, any, bool) {
			// Filter out internal fields
			if len(key) > 0 && key[0] == '_' {
				return "", nil, false
			}
			// Transform sensitive fields
			if key == sensitiveKey1 || key == sensitiveKey2 {
				return key, redactedValue, true
			}
			return key, val, true
		},
		// Complex fields filter
		FieldsFilter: func(fields slog.Fields) (slog.Fields, bool) {
			if len(fields) == 0 {
				return fields, true
			}
			// Add timestamp to all field sets
			result := make(slog.Fields, len(fields)+1)
			for k, v := range fields {
				if k != "_internal" {
					result[k] = v
				}
			}
			result["timestamp"] = time.Now().Unix()
			return result, true
		},
		// Message filter that adds context
		MessageFilter: func(msg string) (string, bool) {
			if msg == "" {
				return msg, false // Filter out empty messages
			}
			return "[APP] " + msg, true
		},
	}
}

// TestFilterWithCustomFiltersStress tests filter with custom filters under stress.
func TestFilterWithCustomFiltersStress(t *testing.T) {
	suite := slogtest.StressTestSuite{
		NewLogger: func() slog.Logger {
			backend := mock.NewLogger()
			return createCustomFilterLogger(backend)
		},
		NewLoggerWithRecorder: createCustomFilterLogger,
	}

	suite.Run(t)
}

// TestFilterChainedStress tests multiple chained filters under stress.
func TestFilterChainedStress(t *testing.T) {
	suite := slogtest.StressTestSuite{
		NewLogger: func() slog.Logger {
			// Create a chain of filters
			backend := mock.NewLogger()

			// First filter: Level filtering
			filter1 := filter.New(backend, slog.Info)

			// Second filter: Field filtering
			filter2 := &filter.Logger{
				Parent:    filter1,
				Threshold: slog.Debug, // Allow all through this layer
				FieldFilter: func(key string, val any) (string, any, bool) {
					// Filter out debug fields
					if key == "debug" || key == "trace" {
						return "", nil, false
					}
					return key, val, true
				},
			}

			// Third filter: Message prefixing
			filter3 := &filter.Logger{
				Parent:    filter2,
				Threshold: slog.Debug,
				MessageFilter: func(msg string) (string, bool) {
					return "[CHAIN] " + msg, true
				},
			}

			return filter3
		},
		NewLoggerWithRecorder: nil, // Complex chains don't need recorder variant
	}

	// Run stress tests on the chained filters
	t.Run("HighVolume", func(t *testing.T) {
		logger := suite.NewLogger()
		stress := slogtest.HighVolumeStressTest()
		slogtest.RunStressTest(t, logger, stress)
	})

	t.Run("MemoryPressure", func(t *testing.T) {
		logger := suite.NewLogger()
		stress := slogtest.MemoryPressureStressTest()
		slogtest.RunStressTest(t, logger, stress)
	})
}

// TestNoopFilterStress tests the noop filter under stress.
func TestNoopFilterStress(t *testing.T) {
	suite := slogtest.StressTestSuite{
		NewLogger: func() slog.Logger {
			return filter.NewNoop()
		},
		// Noop can't have recorder
		NewLoggerWithRecorder: nil,
		// Skip tests that require message verification
		SkipHighVolume:      false, // Can still test performance
		SkipMemoryPressure:  false, // Can still test memory
		SkipDurationBased:   false, // Can still test duration
		SkipConcurrentField: true,  // Skip since fields aren't collected
	}

	suite.Run(t)
}
