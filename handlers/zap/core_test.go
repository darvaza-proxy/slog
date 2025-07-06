package zap_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"darvaza.org/slog"
	slogzap "darvaza.org/slog/handlers/zap"
	slogtest "darvaza.org/slog/internal/testing"
)

// getFieldValue is a generic helper to safely extract and cast field values
func getFieldValue[T any](fields map[string]any, key string) (T, bool) {
	var zero T
	value, exists := fields[key]
	if !exists {
		return zero, false
	}
	typed, ok := value.(T)
	return typed, ok
}

// assertField is a test helper that checks if a field exists and has the expected value
func assertField[T comparable](t *testing.T, fields map[string]any, key string, expected T) {
	t.Helper()
	actual, ok := getFieldValue[T](fields, key)
	if !ok {
		t.Errorf("Field %q not found or wrong type, got %v (type: %T)", key, fields[key], fields[key])
		return
	}
	if actual != expected {
		t.Errorf("Field %q = %v, want %v", key, actual, expected)
	}
}

// assertAnySliceField is a test helper for []any fields (interface slices)
func assertAnySliceField(t *testing.T, fields map[string]any, key string, expected ...any) {
	t.Helper()
	actual, ok := fields[key].([]any)
	if !ok {
		t.Errorf("Field %q not []any, got %v (type: %T)", key, fields[key], fields[key])
		return
	}
	if len(actual) != len(expected) {
		t.Errorf("Field %q length = %d, want %d", key, len(actual), len(expected))
		return
	}
	for i, v := range expected {
		if actual[i] != v {
			t.Errorf("Field %q[%d] = %v, want %v", key, i, actual[i], v)
		}
	}
}

func TestSlogCore(t *testing.T) {
	t.Run("BasicFunctionality", testSlogCoreBasic)
	t.Run("WithFields", testSlogCoreWithFields)
	t.Run("With", testSlogCoreWith)
	t.Run("WithEmpty", testSlogCoreWithEmpty)
	t.Run("Levels", testSlogCoreLevels)
	t.Run("Enabled", testSlogCoreEnabled)
}

func TestNewZapLogger(t *testing.T) {
	t.Run("ConvenienceConstructor", testNewZapLogger)
}

func TestSlogCoreCaller(t *testing.T) {
	t.Run("WithCaller", testSlogCoreWithCaller)
	t.Run("CallerUndefined", testSlogCoreCallerUndefined)
}

func TestSlogCoreStack(t *testing.T) {
	t.Run("WithStack", testSlogCoreWithStack)
}

func TestSlogCoreFields(t *testing.T) {
	t.Run("ComplexFields", testSlogCoreComplexFields)
	t.Run("ConvertFieldsEmpty", testConvertFieldsEmpty)
}

func TestSlogCoreErrorCases(t *testing.T) {
	t.Run("NilLogger", testSlogCoreNilLogger)
	t.Run("NilLevel", testSlogCoreNilLevel)
}

func TestSlogCoreSync(t *testing.T) {
	t.Run("Sync", testSlogCoreSync)
}

func TestBidirectionalIntegration(t *testing.T) {
	t.Run("slog_to_zap_direction", testBidirectionalSlogToZap)
	t.Run("zap_to_slog_direction", testBidirectionalZapToSlog)
	t.Run("both_directions_compatibility", testBidirectionalCompatibility)
}

func TestSlogCoreConcurrent(t *testing.T) {
	t.Run("ConcurrentAccess", testSlogCoreConcurrent)
}

func TestSlogCoreCheck(t *testing.T) {
	t.Run("CheckDisabled", testSlogCoreCheckDisabled)
}

func TestSlogCoreWrite(t *testing.T) {
	t.Run("FatalWrite", testSlogCoreFatalWrite)
	t.Run("PanicWrite", testSlogCorePanicWrite)
}

func TestMapZapToSlogLevel(t *testing.T) {
	// This uses test case types, so we keep it as is
	testCases := []mapTestCase{
		{"DPanicLevel", zapcore.DPanicLevel, slog.Panic, true},
		{"UnknownLevel", zapcore.Level(99), slog.Info, false},
		{"InvalidLevel", zapcore.InvalidLevel, slog.Info, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

func TestSlogCoreWithConfigurations(t *testing.T) {
	// This uses test case types, so we keep it as is
	tests := []configTestCase{
		{"Development", zap.NewDevelopmentConfig()},
		{"Production", zap.NewProductionConfig()},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

// Test functions

func testSlogCoreBasic(t *testing.T) {
	recorder := slogtest.NewLogger()

	// Create a zap logger using our SlogCore
	core := slogzap.NewCore(recorder, zap.DebugLevel)
	zapLogger := zap.New(core)

	// Test basic logging
	zapLogger.Info("test message")

	// Check the output
	messages := recorder.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(messages))
	}

	msg := messages[0]
	if msg.Level != slog.Info {
		t.Errorf("Expected Info level, got %v", msg.Level)
	}
	if msg.Message != "test message" {
		t.Errorf("Expected 'test message', got %q", msg.Message)
	}
}

func testSlogCoreWithFields(t *testing.T) {
	recorder := slogtest.NewLogger()

	core := slogzap.NewCore(recorder, zap.DebugLevel)
	zapLogger := zap.New(core)

	// Log with fields
	zapLogger.Info("test with fields",
		zap.String("key1", "value1"),
		zap.Int("key2", 42),
	)

	// Check the output
	messages := recorder.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(messages))
	}

	msg := messages[0]
	assertField(t, msg.Fields, "key1", "value1")
	assertField(t, msg.Fields, "key2", int64(42))
}

func testSlogCoreWith(t *testing.T) {
	recorder := slogtest.NewLogger()

	core := slogzap.NewCore(recorder, zap.DebugLevel)
	zapLogger := zap.New(core)

	// Create a logger with persistent fields
	childLogger := zapLogger.With(
		zap.String("persistent", "field"),
		zap.Int("request_id", 123),
	)

	// Log with additional fields
	childLogger.Info("child message", zap.String("extra", "value"))

	// Check the output
	messages := recorder.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(messages))
	}

	msg := messages[0]
	assertField(t, msg.Fields, "persistent", "field")
	assertField(t, msg.Fields, "request_id", int64(123))
	assertField(t, msg.Fields, "extra", "value")
}

func testSlogCoreWithEmpty(t *testing.T) {
	recorder := slogtest.NewLogger()
	core := slogzap.NewCore(recorder, zap.DebugLevel)

	// With() should return the same core when no fields are provided
	sameCore := core.With([]zapcore.Field{})
	if sameCore != core {
		t.Error("With() should return the same core when no fields are provided")
	}
}

// levelTestCase represents a test case for level mapping
type levelTestCase struct {
	name      string
	zapLevel  zapcore.Level
	slogLevel slog.LogLevel
	logFunc   func(*zap.Logger, string, ...zap.Field)
}

// test runs the level mapping test
func (tc levelTestCase) test(t *testing.T) {
	recorder := slogtest.NewLogger()
	core := slogzap.NewCore(recorder, tc.zapLevel)
	zapLogger := zap.New(core)

	// Log at the test level
	tc.logFunc(zapLogger, "test")

	// Check output
	messages := recorder.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(messages))
	}

	msg := messages[0]
	if msg.Level != tc.slogLevel {
		t.Errorf("Expected slog level %v, got %v", tc.slogLevel, msg.Level)
	}
}

func testSlogCoreLevels(t *testing.T) {
	testCases := []levelTestCase{
		{"Debug", zap.DebugLevel, slog.Debug, (*zap.Logger).Debug},
		{"Info", zap.InfoLevel, slog.Info, (*zap.Logger).Info},
		{"Warn", zap.WarnLevel, slog.Warn, (*zap.Logger).Warn},
		{"Error", zap.ErrorLevel, slog.Error, (*zap.Logger).Error},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

func testSlogCoreEnabled(t *testing.T) {
	recorder := slogtest.NewLogger()

	// Create core with Info level
	core := slogzap.NewCore(recorder, zap.InfoLevel)

	// Debug should be disabled
	if core.Enabled(zap.DebugLevel) {
		t.Error("Debug level should be disabled when core is at Info level")
	}

	// Info and above should be enabled
	if !core.Enabled(zap.InfoLevel) {
		t.Error("Info level should be enabled")
	}
	if !core.Enabled(zap.WarnLevel) {
		t.Error("Warn level should be enabled")
	}
	if !core.Enabled(zap.ErrorLevel) {
		t.Error("Error level should be enabled")
	}
}

func testNewZapLogger(t *testing.T) {
	recorder := slogtest.NewLogger()

	// Create zap logger using convenience function
	zapLogger := slogzap.NewZapLogger(recorder)

	// Should default to Info level
	zapLogger.Debug("debug message")
	messages := recorder.GetMessages()
	if len(messages) != 0 {
		t.Error("Debug message should not be logged at Info level")
	}

	// Info should work
	recorder.Clear()
	zapLogger.Info("info message")
	messages = recorder.GetMessages()
	if len(messages) != 1 {
		t.Fatal("Info message should be logged")
	}
	if messages[0].Message != "info message" {
		t.Errorf("Expected 'info message', got %q", messages[0].Message)
	}
}

func testSlogCoreWithCaller(t *testing.T) {
	recorder := slogtest.NewLogger()

	// Create zap logger with caller info
	zapLogger := slogzap.NewZapLogger(recorder, zap.AddCaller())

	zapLogger.Info("test with caller")

	messages := recorder.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(messages))
	}

	msg := messages[0]
	if msg.Fields["caller"] == nil {
		t.Error("Expected caller information in fields")
	}
	callerStr, ok := msg.Fields["caller"].(string)
	if !ok {
		t.Error("Caller should be a string")
	}
	if !strings.Contains(callerStr, "core_test") {
		t.Errorf("Caller should contain test file name, got %s", callerStr)
	}
}

func testSlogCoreCallerUndefined(t *testing.T) {
	recorder := slogtest.NewLogger()
	core := slogzap.NewCore(recorder, zap.DebugLevel)

	// Create entry without caller info
	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Message: "no caller",
		Caller:  zapcore.EntryCaller{}, // Undefined caller
	}

	err := core.Write(entry, nil)
	if err != nil {
		t.Errorf("Write returned error: %v", err)
	}

	messages := recorder.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	// Should not have caller field when undefined
	if _, hasCaller := messages[0].Fields["caller"]; hasCaller {
		t.Error("Should not have caller field when undefined")
	}
}

func testSlogCoreWithStack(t *testing.T) {
	recorder := slogtest.NewLogger()
	core := slogzap.NewCore(recorder, zap.DebugLevel)

	// Create an entry with stack trace
	entry := zapcore.Entry{
		Level:   zapcore.ErrorLevel,
		Message: "error with stack",
		Stack:   "goroutine 1 [running]:\nmain.main()\n\t/tmp/test.go:10 +0x20",
	}

	err := core.Write(entry, nil)
	if err != nil {
		t.Errorf("Write returned error: %v", err)
	}

	messages := recorder.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(messages))
	}

	msg := messages[0]
	if msg.Fields["stacktrace"] == nil {
		t.Error("Expected stacktrace in fields")
	}
	stackStr, ok := msg.Fields["stacktrace"].(string)
	if !ok {
		t.Error("Stacktrace should be a string")
	}
	if !strings.Contains(stackStr, "goroutine 1") {
		t.Errorf("Stacktrace should contain goroutine info, got %s", stackStr)
	}
}

func testSlogCoreComplexFields(t *testing.T) {
	recorder := slogtest.NewLogger()

	zapLogger := slogzap.NewZapLogger(recorder)

	// Test various field types
	zapLogger.Info("complex fields",
		zap.String("string", "value"),
		zap.Int("int", 42),
		zap.Float64("float", 3.14),
		zap.Bool("bool", true),
		zap.Strings("strings", []string{"a", "b", "c"}),
		zap.Ints("ints", []int{1, 2, 3}),
		zap.Duration("duration", 5000000000), // 5 seconds in nanoseconds
		zap.Any("any", map[string]int{"key": 123}),
	)

	messages := recorder.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(messages))
	}

	msg := messages[0]
	// Check basic field types
	assertField(t, msg.Fields, "string", "value")
	assertField(t, msg.Fields, "int", int64(42))
	assertField(t, msg.Fields, "float", 3.14)
	assertField(t, msg.Fields, "bool", true)

	// Check array fields (zap converts to []any)
	assertAnySliceField(t, msg.Fields, "strings", "a", "b", "c")
	assertAnySliceField(t, msg.Fields, "ints", 1, 2, 3)

	// Check duration (zap converts to time.Duration)
	if v, ok := msg.Fields["duration"].(time.Duration); !ok {
		t.Errorf("Duration field should be time.Duration, got %T", msg.Fields["duration"])
	} else if v != 5*time.Second {
		t.Errorf("Duration field mismatch: %v", v)
	}

	// Check complex any field (zap preserves the original type)
	if v, ok := msg.Fields["any"].(map[string]int); !ok {
		t.Errorf("Any field should be map[string]int, got %T", msg.Fields["any"])
	} else if val, exists := v["key"]; !exists || val != 123 {
		t.Errorf("Any field map content mismatch: %v", v)
	}
}

func testSlogCoreNilLogger(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewCore should panic with nil logger")
		}
	}()

	slogzap.NewCore(nil, zap.InfoLevel)
}

func testSlogCoreNilLevel(t *testing.T) {
	recorder := slogtest.NewLogger()

	// Should default to InfoLevel
	core := slogzap.NewCore(recorder, nil)

	// Debug should be disabled
	if core.Enabled(zap.DebugLevel) {
		t.Error("Debug should be disabled with default level")
	}

	// Info should be enabled
	if !core.Enabled(zap.InfoLevel) {
		t.Error("Info should be enabled with default level")
	}
}

func testSlogCoreSync(t *testing.T) {
	recorder := slogtest.NewLogger()

	core := slogzap.NewCore(recorder, zap.InfoLevel)

	// Sync should always succeed (no-op for slog)
	if err := core.Sync(); err != nil {
		t.Errorf("Sync() returned error: %v", err)
	}
}

func testBidirectionalSlogToZap(t *testing.T) {
	// Test slog → zap: Use slog backend with zap API
	baseRecorder := slogtest.NewLogger()
	zapLogger := slogzap.NewZapLogger(baseRecorder)

	// Use zap API
	zapLogger.Info("via zap api",
		zap.String("path", "slog->zap"),
		zap.Int("test_id", 1),
	)

	// Verify slog recorder received the message
	messages := baseRecorder.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.Message != "via zap api" {
		t.Errorf("Expected 'via zap api', got %q", msg.Message)
	}
	assertField(t, msg.Fields, "path", "slog->zap")
	assertField(t, msg.Fields, "test_id", int64(1))
}

func testBidirectionalZapToSlog(t *testing.T) {
	// Test zap → slog: Use zap backend with slog API
	zapConfig := slogzap.NewDefaultConfig()
	slogLogger, err := slogzap.New(zapConfig)
	if err != nil {
		t.Fatalf("Failed to create slog logger: %v", err)
	}

	// Use slog API (this internally uses zap)
	slogLogger.Info().
		WithField("path", "zap->slog").
		WithField("test_id", 2).
		Print("via slog api")

	// This direction is harder to test without capturing zap output,
	// but we can verify it doesn't panic or error
}

func testBidirectionalCompatibility(t *testing.T) {
	// Verify both directions work with the same field types
	baseRecorder := slogtest.NewLogger()
	zapLogger := slogzap.NewZapLogger(baseRecorder)

	// Test complex field types through the adapter
	zapLogger.Info("field compatibility test",
		zap.String("string", "test"),
		zap.Int("int", 42),
		zap.Bool("bool", true),
		zap.Float64("float", 3.14),
		zap.Duration("duration", 1000000000), // 1 second
	)

	messages := baseRecorder.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	// Verify field types are preserved correctly
	fields := messages[0].Fields
	assertField(t, fields, "string", "test")
	assertField(t, fields, "int", int64(42))
	assertField(t, fields, "bool", true)
}

func testSlogCoreConcurrent(t *testing.T) {
	recorder := slogtest.NewLogger()
	zapLogger := slogzap.NewZapLogger(recorder)

	const goroutines = 10
	const msgsPerGoroutine = 10 // Reduced for cleaner test output

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			logger := zapLogger.With(zap.Int("goroutine", id))
			for j := 0; j < msgsPerGoroutine; j++ {
				logger.Info(fmt.Sprintf("msg-%d-%d", id, j),
					zap.Int("index", j),
				)
			}
		}(i)
	}

	wg.Wait()

	messages := recorder.GetMessages()
	expectedCount := goroutines * msgsPerGoroutine
	if len(messages) != expectedCount {
		t.Errorf("Expected %d messages, got %d", expectedCount, len(messages))
	}

	// Verify all goroutines logged their messages
	counts := make(map[int]int)
	for _, msg := range messages {
		if gid, ok := msg.Fields["goroutine"].(int64); ok {
			counts[int(gid)]++
		} else if gid, ok := msg.Fields["goroutine"].(int); ok {
			counts[gid]++
		}
	}

	for i := 0; i < goroutines; i++ {
		if counts[i] != msgsPerGoroutine {
			t.Errorf("Goroutine %d: expected %d messages, got %d", i, msgsPerGoroutine, counts[i])
		}
	}
}

func testSlogCoreCheckDisabled(t *testing.T) {
	recorder := slogtest.NewLogger()

	// Create core with Info level
	core := slogzap.NewCore(recorder, zap.InfoLevel)

	// Create a debug entry (which should be disabled)
	entry := zapcore.Entry{
		Level:   zap.DebugLevel,
		Message: "debug message",
	}

	// Check should return nil for disabled level
	checked := core.Check(entry, nil)
	if checked != nil {
		t.Error("Check should return nil for disabled level")
	}

	// Verify with a CheckedEntry
	ce := &zapcore.CheckedEntry{}
	result := core.Check(entry, ce)
	if result != ce {
		t.Error("Check should return the same CheckedEntry when level is disabled")
	}
}

func testSlogCoreFatalWrite(t *testing.T) {
	recorder := slogtest.NewLogger()
	core := slogzap.NewCore(recorder, zap.DebugLevel)

	// We can't actually test os.Exit, but we can verify the Fatal log is written
	entry := zapcore.Entry{
		Level:   zapcore.FatalLevel,
		Message: "fatal error occurred",
	}

	// Note: This test cannot verify os.Exit behaviour in unit tests.
	// The actual Exit call would need to be tested in integration tests.
	// The Write method will call logger.Fatal() which in our test logger
	// just records a message with Fatal level
	err := core.Write(entry, nil)
	if err != nil {
		t.Errorf("Write returned error: %v", err)
	}

	// Check that we got the original message
	messages := recorder.GetMessages()
	found := false
	for _, msg := range messages {
		if msg.Message == "fatal error occurred" && msg.Level == slog.Fatal {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find fatal message in recorder")
	}

	// Also check for the "zap fatal exit" message
	foundExit := false
	for _, msg := range messages {
		if msg.Message == "zap fatal exit" && msg.Level == slog.Fatal {
			foundExit = true
			break
		}
	}
	if !foundExit {
		t.Error("Expected to find 'zap fatal exit' message")
	}
}

func testSlogCorePanicWrite(t *testing.T) {
	recorder := slogtest.NewLogger()
	core := slogzap.NewCore(recorder, zap.DebugLevel)

	// Test PanicLevel
	t.Run("PanicLevel", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic from PanicLevel")
			} else {
				expected := "zap panic: panic message"
				if r != expected {
					t.Errorf("Expected panic message %q, got %q", expected, r)
				}
			}
		}()

		entry := zapcore.Entry{
			Level:   zapcore.PanicLevel,
			Message: "panic message",
		}

		_ = core.Write(entry, nil)
	})

	// Test DPanicLevel
	t.Run("DPanicLevel", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic from DPanicLevel")
			} else {
				expected := "zap panic: development panic"
				if r != expected {
					t.Errorf("Expected panic message %q, got %q", expected, r)
				}
			}
		}()

		entry := zapcore.Entry{
			Level:   zapcore.DPanicLevel,
			Message: "development panic",
		}

		_ = core.Write(entry, nil)
	})
}

func testConvertFieldsEmpty(t *testing.T) {
	recorder := slogtest.NewLogger()
	core := slogzap.NewCore(recorder, zap.InfoLevel)
	zapLogger := zap.New(core)

	// Test with no fields
	zapLogger.Info("no fields")

	messages := recorder.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.Fields == nil || len(msg.Fields) != 0 {
		t.Errorf("Expected empty fields map, got %v", msg.Fields)
	}

	// Test with empty field slice explicitly
	recorder.Clear()
	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Message: "empty field slice",
	}
	err := core.Write(entry, []zapcore.Field{})
	if err != nil {
		t.Errorf("Write returned error: %v", err)
	}

	messages = recorder.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}
	if messages[0].Fields == nil || len(messages[0].Fields) != 0 {
		t.Errorf("Expected empty fields map for empty field slice, got %v", messages[0].Fields)
	}
}

// mapTestCase represents a test case for zap to slog level mapping
type mapTestCase struct {
	name              string
	zapLevel          zapcore.Level
	expectedSlogLevel slog.LogLevel
	shouldPanic       bool
}

// test runs the zap to slog level mapping test
func (tc mapTestCase) test(t *testing.T) {
	recorder := slogtest.NewLogger()
	core := slogzap.NewCore(recorder, zapcore.DebugLevel)

	// For DPanic, we need to catch the panic
	if tc.shouldPanic {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic from DPanicLevel")
			}
		}()
	}

	entry := zapcore.Entry{
		Level:   tc.zapLevel,
		Message: fmt.Sprintf("test level %v", tc.zapLevel),
	}

	recorder.Clear()
	_ = core.Write(entry, nil)

	// For non-panic levels, check that they were logged at expected level
	if !tc.shouldPanic {
		messages := recorder.GetMessages()
		if len(messages) > 0 {
			msg := messages[0]
			if msg.Level != tc.expectedSlogLevel {
				t.Errorf("Expected slog level %v for zap level %v, got %v",
					tc.expectedSlogLevel, tc.zapLevel, msg.Level)
			}
		}
	}
}

// configTestCase represents a test case for configuration testing
type configTestCase struct {
	name   string
	config zap.Config
}

// test runs the configuration test
func (tc configTestCase) test(t *testing.T) {
	recorder := slogtest.NewLogger()
	// Create zap logger with specific config
	zapLogger, err := tc.config.Build(zap.WrapCore(func(_ zapcore.Core) zapcore.Core {
		return slogzap.NewCore(recorder, zap.InfoLevel)
	}))
	if err != nil {
		t.Fatalf("Failed to build zap logger: %v", err)
	}

	zapLogger.Info("test message")

	messages := recorder.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}
