package filter_test

import (
	"strings"
	"sync"
	"testing"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
	slogtest "darvaza.org/slog/internal/testing"
)

const (
	redacted = "[REDACTED]"
)

func TestFilterWithTransforms(t *testing.T) {
	t.Run("FieldFilter", testFilterFieldFilter)
	t.Run("MessageFilter", testFilterMessageFilter)
	t.Run("FieldOverride", testFilterFieldOverride)
}

func testFilterFieldFilter(t *testing.T) {
	base := newTestLogger()

	filterCalled := false
	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Info,
		FieldFilter: func(key string, val any) (string, any, bool) {
			filterCalled = true
			// Redact sensitive fields
			if key == "password" || key == "secret" {
				return strings.ToUpper(key), redacted, true
			}
			// Remove internal fields
			if strings.HasPrefix(key, "_") {
				return "", nil, false
			}
			// Transform key to uppercase
			return strings.ToUpper(key), val, true
		},
	}

	logger.Info().
		WithField("username", "john").
		WithField("password", "secret123").
		WithField("_internal", "hidden").
		Print("test message")

	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, 1)

	msg := msgs[0]
	if !filterCalled {
		t.Error("FieldFilter was not called")
	}

	// Check transformed fields
	slogtest.AssertField(t, msg, "USERNAME", "john")
	slogtest.AssertField(t, msg, "PASSWORD", redacted)
	slogtest.AssertNoField(t, msg, "_internal")
}

func testFilterMessageFilter(t *testing.T) {
	base := newTestLogger()

	filterCalled := false
	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Info,
		MessageFilter: func(msg string) (string, bool) {
			filterCalled = true
			// Filter out debug messages
			if strings.Contains(msg, "[DEBUG]") {
				return "", false
			}
			// Add prefix to all messages
			return "[FILTERED] " + msg, true
		},
	}

	// This should be filtered out
	logger.Info().Print("[DEBUG] internal message")

	// This should go through with prefix
	logger.Info().Print("normal message")

	msgs := base.GetMessages()
	if !filterCalled {
		t.Error("MessageFilter was not called")
	}

	slogtest.AssertMessageCount(t, msgs, 1)
	slogtest.AssertMessage(t, msgs[0], slog.Info, "[FILTERED] normal message")
}

func testFilterFieldOverride(t *testing.T) {
	base := newTestLogger()

	overrideCalled := false
	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Info,
		FieldOverride: func(entry slog.Logger, key string, val any) {
			overrideCalled = true
			// Add prefix to all field keys
			entry.WithField("custom_"+key, val).Print("")
		},
	}

	logger.Info().WithField("test", "value").Print("message")

	msgs := base.GetMessages()
	if !overrideCalled {
		t.Error("FieldOverride was not called")
	}

	// Should have 2 messages - one from override, one from actual print
	slogtest.AssertMessageCount(t, msgs, 2)

	// First message from FieldOverride
	slogtest.AssertMessage(t, msgs[0], slog.Info, "")
	slogtest.AssertField(t, msgs[0], "custom_test", "value")

	// Second message is the actual one
	slogtest.AssertMessage(t, msgs[1], slog.Info, "message")
}

func TestFilterConcurrency(t *testing.T) {
	t.Run("BasicConcurrency", testFilterBasicConcurrency)
	t.Run("ConcurrencyWithFieldFilter", testFilterConcurrencyWithFieldFilter)
	t.Run("ConcurrentFields", testFilterConcurrentFields)
	t.Run("ConcurrencyWithThresholds", testFilterConcurrencyWithThresholds)
}

func testFilterBasicConcurrency(t *testing.T) {
	base := newTestLogger()
	logger := filter.New(base, slog.Info)

	// Use RunConcurrentTestWithOptions to handle filtered Debug messages
	opts := slogtest.ConcurrencyTestOptions{
		AdapterOptions: slogtest.AdapterOptions{
			LevelExceptions: map[slog.LogLevel]slog.LogLevel{
				slog.Debug: slog.UndefinedLevel, // Debug messages are filtered out
			},
		},
		FactoryOptions: slogtest.FactoryOptions{
			NewLoggerWithRecorder: func(recorder slog.Logger) slog.Logger {
				return filter.New(recorder, slog.Info)
			},
		},
	}

	slogtest.RunConcurrentTestWithOptions(t, logger, slogtest.DefaultConcurrencyTest(), &opts)
}

func testFilterConcurrencyWithFieldFilter(t *testing.T) {
	base := newTestLogger()

	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Info,
		FieldFilter: func(key string, val any) (string, any, bool) {
			// Simulate some processing
			if key == "remove" {
				return "", nil, false
			}
			return key, val, true
		},
	}

	const goroutines = 10
	const messagesPerGoroutine = 100

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				logger.Info().
					WithField("goroutine", id).
					WithField("message", j).
					WithField("remove", "this").
					Printf("message %d from %d", j, id)
			}
		}(i)
	}

	wg.Wait()

	// Verify all messages were sent
	msgs := base.GetMessages()
	if len(msgs) != goroutines*messagesPerGoroutine {
		t.Errorf("Expected %d messages, got %d", goroutines*messagesPerGoroutine, len(msgs))
	}

	// Verify filter was applied
	for _, msg := range msgs {
		if _, exists := msg.Fields["remove"]; exists {
			t.Error("Field 'remove' should have been filtered out")
		}
		if msg.Fields["goroutine"] == nil {
			t.Error("Missing goroutine field")
		}
	}
}

func testFilterConcurrentFields(t *testing.T) {
	slogtest.TestConcurrentFields(t, func() slog.Logger {
		base := newTestLogger()
		return filter.New(base, slog.Info)
	})
}

func testFilterConcurrencyWithThresholds(t *testing.T) {
	// Test concurrency with thresholds using custom goroutine tests
	// since the standard test infrastructure hardcodes Info level

	thresholds := []struct {
		name      string
		level     slog.LogLevel
		testLevel slog.LogLevel // Level to actually test with
	}{
		{
			name:      "WarnThreshold",
			level:     slog.Warn,
			testLevel: slog.Warn,
		},
		{
			name:      "ErrorThreshold",
			level:     slog.Error,
			testLevel: slog.Error,
		},
	}

	for _, tc := range thresholds {
		t.Run(tc.name, func(t *testing.T) {
			base := newTestLogger()
			logger := filter.New(base, tc.level)

			const goroutines = 10
			const messagesPerGoroutine = 10

			var wg sync.WaitGroup
			for i := 0; i < goroutines; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < messagesPerGoroutine; j++ {
						// Use the test level that should pass the threshold
						logger.WithLevel(tc.testLevel).
							WithField("goroutine", id).
							WithField("operation", j).
							Printf("concurrent message %d-%d", id, j)
					}
				}(i)
			}

			wg.Wait()

			// Verify all messages were sent
			msgs := base.GetMessages()
			expected := goroutines * messagesPerGoroutine
			if len(msgs) != expected {
				t.Errorf("Expected %d messages, got %d", expected, len(msgs))
			}

			// Verify all messages have required fields
			for i, msg := range msgs {
				if msg.Fields["goroutine"] == nil {
					t.Errorf("message %d missing goroutine field", i)
				}
				if msg.Fields["operation"] == nil {
					t.Errorf("message %d missing operation field", i)
				}
			}
		})
	}
}

func TestFilterChainingBehaviour(t *testing.T) {
	base := newTestLogger()

	// Create a chain of filters
	filter1 := filter.New(base, slog.Error)
	filter2 := filter.New(filter1, slog.Warn)
	filter3 := filter.New(filter2, slog.Info)

	// Test that the most restrictive filter wins
	filter3.Debug().Print("debug - should not appear")
	filter3.Info().Print("info - should not appear")
	filter3.Warn().Print("warn - should not appear")
	filter3.Error().Print("error - should appear")
	filter3.Fatal().Print("fatal - should appear")

	msgs := base.GetMessages()
	if len(msgs) != 2 {
		t.Errorf("Expected 2 messages (Error and Fatal), got %d", len(msgs))
		for i, msg := range msgs {
			t.Logf("Message %d: level=%v, text=%q", i, msg.Level, msg.Message)
		}
	}
}
