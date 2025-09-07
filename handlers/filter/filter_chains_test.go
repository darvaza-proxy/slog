package filter_test

import (
	"fmt"
	"strings"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
	"darvaza.org/slog/handlers/mock"
	slogtest "darvaza.org/slog/internal/testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = filterChainTestCase{}
var _ core.TestCase = nestedFilterTestCase{}
var _ core.TestCase = dynamicThresholdTestCase{}

// filterChainTestCase tests complex filter chains with varying thresholds
type filterChainTestCase struct {
	setupChain     func() slog.Logger
	name           string
	description    string
	testOperations []operation
	expectedCount  int
}

type operation struct {
	fields    map[string]any
	message   string
	level     slog.LogLevel
	shouldLog bool
}

func newOperation(level slog.LogLevel, message string, fields map[string]any, shouldLog bool) operation {
	return operation{
		level:     level,
		message:   message,
		fields:    fields,
		shouldLog: shouldLog,
	}
}

func (tc filterChainTestCase) Name() string {
	return tc.name
}

func (tc filterChainTestCase) Test(t *testing.T) {
	t.Helper()

	logger := tc.setupChain()

	// Extract the base mock logger
	var base *mock.Logger
	current := logger
	for {
		if m, ok := current.(*mock.Logger); ok {
			base = m
			break
		}
		if f, ok := current.(*filter.Logger); ok && f.Parent != nil {
			current = f.Parent
		} else {
			break
		}
	}

	if base == nil {
		t.Fatal("Could not find base mock logger in chain")
	}

	// Execute operations
	for _, op := range tc.testOperations {
		entry := logger.WithLevel(op.level)
		if len(op.fields) > 0 {
			entry = entry.WithFields(op.fields)
		}
		entry.Print(op.message)
	}

	// Verify results
	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, tc.expectedCount)

	// Verify which operations were logged
	msgIdx := 0
	for i, op := range tc.testOperations {
		if op.shouldLog {
			if msgIdx >= len(msgs) {
				t.Errorf("Operation %d should have logged but didn't", i)
				continue
			}
			msg := msgs[msgIdx]
			core.AssertContains(t, msg.Message, op.message, "message content")
			msgIdx++
		}
	}
}

func newFilterChainTestCase(name, description string,
	setupChain func() slog.Logger,
	testOperations []operation,
	expectedCount int) filterChainTestCase {
	return filterChainTestCase{
		name:           name,
		description:    description,
		setupChain:     setupChain,
		testOperations: testOperations,
		expectedCount:  expectedCount,
	}
}

func filterChainTestCases() []filterChainTestCase {
	return []filterChainTestCase{
		newFilterChainTestCase(
			"Three-tier filter chain",
			"Multiple filters with different thresholds",
			func() slog.Logger {
				base := mock.NewLogger()
				// First filter: Debug and above
				filter1 := filter.New(base, slog.Debug)
				// Second filter: Info and above
				filter2 := filter.New(filter1, slog.Info)
				// Third filter: Warn and above
				filter3 := filter.New(filter2, slog.Warn)
				return filter3
			},
			[]operation{
				newOperation(slog.Debug, "debug message", nil, false), // Blocked by filter3
				newOperation(slog.Info, "info message", nil, false),   // Blocked by filter3
				newOperation(slog.Warn, "warn message", nil, true),    // Passes all
				newOperation(slog.Error, "error message", nil, true),  // Passes all
				newOperation(slog.Fatal, "fatal message", nil, true),  // Passes all
			},
			3, // warn, error, fatal
		),
		newFilterChainTestCase(
			"Filter with transformations",
			"Chain with field and message transformations",
			func() slog.Logger {
				base := mock.NewLogger()

				// First filter: Add prefix to fields
				filter1 := &filter.Logger{
					Parent:    base,
					Threshold: slog.Debug,
					FieldFilter: func(key string, val any) (string, any, bool) {
						return "f1_" + key, val, true
					},
				}

				// Second filter: Add another prefix and filter some fields
				filter2 := &filter.Logger{
					Parent:    filter1,
					Threshold: slog.Info,
					FieldFilter: func(key string, val any) (string, any, bool) {
						if key == sensitiveKey2 {
							return "", nil, false // Drop secret fields
						}
						return "f2_" + key, val, true
					},
					MessageFilter: func(msg string) (string, bool) {
						return "[FILTERED] " + msg, true
					},
				}

				return filter2
			},
			[]operation{
				newOperation(slog.Debug, "debug", map[string]any{"key": "value"}, false),
				newOperation(slog.Info, "info", map[string]any{"public": "yes", "secret": "no"}, true),
				newOperation(slog.Error, "error", map[string]any{"level": "high"}, true),
			},
			2, // info and error pass
		),
		newFilterChainTestCase(
			"Mixed threshold chain",
			"Filters with non-monotonic thresholds",
			func() slog.Logger {
				base := mock.NewLogger()
				// Permissive filter
				filter1 := filter.New(base, slog.Debug)
				// Restrictive filter
				filter2 := filter.New(filter1, slog.Error)
				// Permissive again (but limited by filter2)
				filter3 := filter.New(filter2, slog.Debug)
				return filter3
			},
			[]operation{
				newOperation(slog.Debug, "debug", nil, false), // Blocked by filter2
				newOperation(slog.Info, "info", nil, false),   // Blocked by filter2
				newOperation(slog.Warn, "warn", nil, false),   // Blocked by filter2
				newOperation(slog.Error, "error", nil, true),  // Passes all
				newOperation(slog.Fatal, "fatal", nil, true),  // Passes all
			},
			2, // Only error and fatal pass the restrictive middle filter
		),
	}
}

func TestComplexFilterChains(t *testing.T) {
	core.RunTestCases(t, filterChainTestCases())
}

// nestedFilterTestCase tests deeply nested filter scenarios
type nestedFilterTestCase struct {
	name       string
	depth      int
	threshold  slog.LogLevel
	testLevel  slog.LogLevel
	shouldPass bool
}

func (tc nestedFilterTestCase) Name() string {
	return tc.name
}

func (tc nestedFilterTestCase) Test(t *testing.T) {
	t.Helper()

	// Create a chain of filters with specified depth
	base := mock.NewLogger()
	var current slog.Logger = base

	for i := 0; i < tc.depth; i++ {
		current = filter.New(current, tc.threshold)
	}

	// Test logging at specified level
	current.WithLevel(tc.testLevel).Print("nested test message")

	msgs := base.GetMessages()
	if tc.shouldPass {
		slogtest.AssertMessageCount(t, msgs, 1)
		if len(msgs) > 0 {
			core.AssertEqual(t, tc.testLevel, msgs[0].Level, "level preserved")
		}
	} else {
		slogtest.AssertMessageCount(t, msgs, 0)
	}
}

func newNestedFilterTestCase(name string, depth int, threshold, testLevel slog.LogLevel,
	shouldPass bool) nestedFilterTestCase {
	return nestedFilterTestCase{
		name:       name,
		depth:      depth,
		threshold:  threshold,
		testLevel:  testLevel,
		shouldPass: shouldPass,
	}
}

func nestedFilterTestCases() []nestedFilterTestCase {
	return []nestedFilterTestCase{
		newNestedFilterTestCase("Deep chain allows Info", 10, slog.Info, slog.Info, true),
		newNestedFilterTestCase("Deep chain blocks Debug", 10, slog.Info, slog.Debug, false),
		newNestedFilterTestCase("Deep chain allows Error", 10, slog.Info, slog.Error, true),
		newNestedFilterTestCase("Very deep chain", 50, slog.Warn, slog.Warn, true),
		newNestedFilterTestCase("Single filter", 1, slog.Error, slog.Warn, false),
	}
}

func TestNestedFilters(t *testing.T) {
	core.RunTestCases(t, nestedFilterTestCases())
}

// dynamicThresholdTestCase tests changing thresholds dynamically
type dynamicThresholdTestCase struct {
	name string
}

func (tc dynamicThresholdTestCase) Name() string {
	return tc.name
}

func (tc dynamicThresholdTestCase) Test(t *testing.T) {
	t.Helper()

	base := mock.NewLogger()

	// Create filter with initial threshold
	filterLogger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Error,
	}

	// Log at various levels with initial threshold
	filterLogger.Debug().Print("debug 1")
	filterLogger.Info().Print("info 1")
	filterLogger.Error().Print("error 1")

	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, 1)
	base.Clear()

	// Change threshold to Info
	filterLogger.Threshold = slog.Info

	filterLogger.Debug().Print("debug 2")
	filterLogger.Info().Print("info 2")
	filterLogger.Error().Print("error 2")

	msgs = base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, 2)
	base.Clear()

	// Change threshold to Debug
	filterLogger.Threshold = slog.Debug

	filterLogger.Debug().Print("debug 3")
	filterLogger.Info().Print("info 3")

	msgs = base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, 2)
}

func newDynamicThresholdTestCase(name string) dynamicThresholdTestCase {
	return dynamicThresholdTestCase{
		name: name,
	}
}

func dynamicThresholdTestCases() []dynamicThresholdTestCase {
	return []dynamicThresholdTestCase{
		newDynamicThresholdTestCase("Dynamic threshold changes"),
	}
}

func TestDynamicThresholds(t *testing.T) {
	core.RunTestCases(t, dynamicThresholdTestCases())
}

func runTestCompleteTransformationChain(t *testing.T) {
	t.Helper()
	base := mock.NewLogger()

	// Track all transformations
	fieldTransforms := 0
	fieldsTransforms := 0
	messageTransforms := 0

	filterLogger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		FieldFilter: func(key string, val any) (string, any, bool) {
			fieldTransforms++
			if strings.HasPrefix(key, "drop_") {
				return "", nil, false
			}
			if strings.HasPrefix(key, "rename_") {
				return strings.Replace(key, "rename_", "renamed_", 1), val, true
			}
			if key == "redact" {
				return key, "[REDACTED]", true
			}
			return key, val, true
		},
		FieldsFilter: func(fields slog.Fields) (slog.Fields, bool) {
			fieldsTransforms++
			// Add a tracking field
			result := make(map[string]any)
			for k, v := range fields {
				result[k] = v
			}
			result["fields_processed"] = true
			return result, true
		},
		MessageFilter: func(msg string) (string, bool) {
			messageTransforms++
			if msg == "drop_message" {
				return "", false
			}
			return fmt.Sprintf("[%d] %s", messageTransforms, msg), true
		},
	}

	// Test various scenarios

	// Scenario 1: WithField transformations
	filterLogger.Info().
		WithField("normal", "value").
		WithField("drop_this", "gone").
		WithField("rename_me", "value").
		WithField("redact", "secret").
		Print("message 1")

	// Scenario 2: WithFields transformations
	filterLogger.Info().
		WithFields(map[string]any{
			"batch1": "value1",
			"batch2": "value2",
		}).
		Print("message 2")

	// Scenario 3: Message drop
	filterLogger.Info().
		WithField("test", "value").
		Print("drop_message")

	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, 2)

	// Verify first message transformations
	msg1 := msgs[0]
	slogtest.AssertMessage(t, msg1, slog.Info, "[1] message 1")
	slogtest.AssertField(t, msg1, "normal", "value")
	slogtest.AssertNoField(t, msg1, "drop_this")
	slogtest.AssertField(t, msg1, "renamed_me", "value")
	slogtest.AssertField(t, msg1, "redact", "[REDACTED]")

	// Verify second message transformations
	msg2 := msgs[1]
	slogtest.AssertMessage(t, msg2, slog.Info, "[2] message 2")
	slogtest.AssertField(t, msg2, "fields_processed", true)

	// Verify transformation counts
	core.AssertTrue(t, fieldTransforms > 0, "field filter called")
	core.AssertTrue(t, fieldsTransforms > 0, "fields filter called")
	core.AssertEqual(t, 3, messageTransforms, "message filter called 3 times")
}

// Test filter chain with all transformation types
func TestCompleteTransformationChain(t *testing.T) {
	t.Run("transformation chain", runTestCompleteTransformationChain)
}

func runTestFilterNilAndEmptyHandling(t *testing.T) {
	t.Helper()
	base := mock.NewLogger()

	filterLogger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Debug,
		FieldFilter: func(key string, val any) (string, any, bool) {
			// Accept nil values
			return key, val, true
		},
		MessageFilter: func(msg string) (string, bool) {
			// Accept empty messages
			return msg, true
		},
	}

	// Test with nil field value
	filterLogger.Info().
		WithField("nil_value", nil).
		Print("message with nil field")

	// Test with empty message
	filterLogger.Info().
		WithField("key", "value").
		Print("")

	// Test with empty field key (should be ignored by WithField)
	filterLogger.Info().
		WithField("", "ignored").
		WithField("valid", "included").
		Print("message")

	msgs := base.GetMessages()
	slogtest.AssertMessageCount(t, msgs, 3)

	// Check nil field
	msg1 := msgs[0]
	slogtest.AssertField(t, msg1, "nil_value", nil)

	// Check empty message
	msg2 := msgs[1]
	slogtest.AssertMessage(t, msg2, slog.Info, "")

	// Check empty key ignored
	msg3 := msgs[2]
	slogtest.AssertField(t, msg3, "valid", "included")
	_, hasEmpty := msg3.Fields[""]
	core.AssertFalse(t, hasEmpty, "empty key field not present")
}

// Test filter behaviour with nil and empty values
func TestFilterNilAndEmptyHandling(t *testing.T) {
	t.Run("nil and empty handling", runTestFilterNilAndEmptyHandling)
}
