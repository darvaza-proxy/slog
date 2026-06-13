package zap_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/mock"
	slogzap "darvaza.org/slog/handlers/zap"
	slogtest "darvaza.org/slog/internal/testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = levelTestCase{}
var _ core.TestCase = mapTestCase{}
var _ core.TestCase = configTestCase{}

// assertAnySliceField is a test helper for []any fields (interface slices)
func assertAnySliceField(t *testing.T, fields map[string]any, key string, expected ...any) {
	t.Helper()
	actual, ok := fields[key].([]any)
	core.AssertMustTrue(t, ok, "field []any type")
	core.AssertMustEqual(t, len(expected), len(actual), "field length")
	for i, v := range expected {
		core.AssertEqual(t, v, actual[i], "field element")
	}
}

func TestSlogCore(t *testing.T) {
	t.Run("BasicFunctionality", runTestSlogCoreBasic)
	t.Run("WithFields", runTestSlogCoreWithFields)
	t.Run("With", runTestSlogCoreWith)
	t.Run("WithEmpty", runTestSlogCoreWithEmpty)
	t.Run("Levels", runTestSlogCoreLevels)
	t.Run("Enabled", runTestSlogCoreEnabled)
}

func TestNewZapLogger(t *testing.T) {
	t.Run("ConvenienceConstructor", runTestNewZapLogger)
}

func TestSlogCoreCaller(t *testing.T) {
	t.Run("WithCaller", runTestSlogCoreWithCaller)
	t.Run("CallerUndefined", runTestSlogCoreCallerUndefined)
}

func TestSlogCoreStack(t *testing.T) {
	t.Run("WithStack", runTestSlogCoreWithStack)
}

func TestSlogCoreFields(t *testing.T) {
	t.Run("ComplexFields", runTestSlogCoreComplexFields)
	t.Run("ConvertFieldsEmpty", runTestConvertFieldsEmpty)
}

func TestSlogCoreErrorCases(t *testing.T) {
	t.Run("NilLogger", runTestSlogCoreNilLogger)
	t.Run("NilLevel", runTestSlogCoreNilLevel)
	t.Run("InvalidLevel", runTestZapLoggerInvalidLevel)
}

func TestSlogCoreSync(t *testing.T) {
	t.Run("Sync", runTestSlogCoreSync)
}

func TestBidirectionalIntegration(t *testing.T) {
	t.Run("slog_to_zap_direction", runTestBidirectionalSlogToZap)
	t.Run("zap_to_slog_direction", runTestBidirectionalZapToSlog)
	t.Run("both_directions_compatibility", runTestBidirectionalCompatibility)
}

func TestSlogCoreConcurrent(t *testing.T) {
	t.Run("ConcurrentAccess", runTestSlogCoreConcurrent)
}

func TestSlogCoreCheck(t *testing.T) {
	t.Run("CheckDisabled", runTestSlogCoreCheckDisabled)
}

func TestSlogCoreWrite(t *testing.T) {
	t.Run("FatalWrite", runTestSlogCoreFatalWrite)
	t.Run("FatalHook", runTestSlogCoreFatalHook)
	t.Run("PanicCoreWrite", runTestSlogCorePanicCoreWrite)
	t.Run("PanicWrite", runTestSlogCorePanicWrite)
	t.Run("DPanicProduction", runTestSlogCoreDPanicProduction)
	t.Run("DPanicDevelopment", runTestSlogCoreDPanicDevelopment)
}

func newMapTestCase(
	name string, zapLevel zapcore.Level, expectedSlogLevel slog.LogLevel,
) mapTestCase {
	return mapTestCase{
		name:              name,
		zapLevel:          zapLevel,
		expectedSlogLevel: expectedSlogLevel,
	}
}

func mapTestCases() []mapTestCase {
	return []mapTestCase{
		newMapTestCase("PanicLevel", zapcore.PanicLevel, slog.Panic),
		newMapTestCase("DPanicLevel", zapcore.DPanicLevel, slog.Error),
		newMapTestCase("UnknownLevel", zapcore.Level(99), slog.Info),
		newMapTestCase("InvalidLevel", zapcore.InvalidLevel, slog.Info),
	}
}

func TestMapZapToSlogLevel(t *testing.T) {
	core.RunTestCases(t, mapTestCases())
}

func newConfigTestCase(name string, config zap.Config) configTestCase {
	return configTestCase{
		name:   name,
		config: config,
	}
}

func configTestCases() []configTestCase {
	return []configTestCase{
		newConfigTestCase("Development", zap.NewDevelopmentConfig()),
		newConfigTestCase("Production", zap.NewProductionConfig()),
	}
}

func TestSlogCoreWithConfigurations(t *testing.T) {
	core.RunTestCases(t, configTestCases())
}

// Test functions

func runTestSlogCoreBasic(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()

	// Create a zap logger using our SlogCore
	zapCore := slogzap.NewCore(recorder, zap.DebugLevel)
	zapLogger := zap.New(zapCore)

	// Test basic logging
	zapLogger.Info("test message")

	// Check the output
	messages := recorder.GetMessages()
	core.AssertMustEqual(t, 1, len(messages), "log entry count")

	msg := messages[0]
	core.AssertEqual(t, slog.Info, msg.Level, "level")
	core.AssertEqual(t, "test message", msg.Message, "message")
}

func runTestSlogCoreWithFields(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()

	zapCore := slogzap.NewCore(recorder, zap.DebugLevel)
	zapLogger := zap.New(zapCore)

	// Log with fields
	zapLogger.Info("test with fields",
		zap.String("key1", "value1"),
		zap.Int("key2", 42),
	)

	// Check the output
	messages := recorder.GetMessages()
	core.AssertMustEqual(t, 1, len(messages), "log entry count")

	msg := messages[0]
	slogtest.AssertField(t, msg, "key1", "value1")
	slogtest.AssertField(t, msg, "key2", int64(42))
}

func runTestSlogCoreWith(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()

	zapCore := slogzap.NewCore(recorder, zap.DebugLevel)
	zapLogger := zap.New(zapCore)

	// Create a logger with persistent fields
	childLogger := zapLogger.With(
		zap.String("persistent", "field"),
		zap.Int("request_id", 123),
	)

	// Log with additional fields
	childLogger.Info("child message", zap.String("extra", "value"))

	// Check the output
	messages := recorder.GetMessages()
	core.AssertMustEqual(t, 1, len(messages), "log entry count")

	msg := messages[0]
	slogtest.AssertField(t, msg, "persistent", "field")
	slogtest.AssertField(t, msg, "request_id", int64(123))
	slogtest.AssertField(t, msg, "extra", "value")
}

func runTestSlogCoreWithEmpty(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()
	zapCore := slogzap.NewCore(recorder, zap.DebugLevel)

	// With() should return the same core when no fields are provided
	sameCore := zapCore.With([]zapcore.Field{})
	core.AssertSame(t, zapCore, sameCore, "With() no fields")
}

// levelTestCase represents a test case for level mapping
type levelTestCase struct {
	logFunc   func(*zap.Logger, string, ...zap.Field)
	name      string
	zapLevel  zapcore.Level
	slogLevel slog.LogLevel
}

func (tc levelTestCase) Name() string {
	return tc.name
}

func (tc levelTestCase) Test(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()
	zapCore := slogzap.NewCore(recorder, tc.zapLevel)
	zapLogger := zap.New(zapCore)

	// Log at the test level
	tc.logFunc(zapLogger, "test")

	// Check output
	messages := recorder.GetMessages()
	core.AssertMustEqual(t, 1, len(messages), "log entry count")

	msg := messages[0]
	core.AssertEqual(t, tc.slogLevel, msg.Level, "level")
}

func newLevelTestCase(
	name string, zapLevel zapcore.Level, slogLevel slog.LogLevel,
	logFunc func(*zap.Logger, string, ...zap.Field),
) levelTestCase {
	return levelTestCase{
		name:      name,
		zapLevel:  zapLevel,
		slogLevel: slogLevel,
		logFunc:   logFunc,
	}
}

func levelTestCases() []levelTestCase {
	return []levelTestCase{
		newLevelTestCase("Debug", zap.DebugLevel, slog.Debug, (*zap.Logger).Debug),
		newLevelTestCase("Info", zap.InfoLevel, slog.Info, (*zap.Logger).Info),
		newLevelTestCase("Warn", zap.WarnLevel, slog.Warn, (*zap.Logger).Warn),
		newLevelTestCase("Error", zap.ErrorLevel, slog.Error, (*zap.Logger).Error),
	}
}

func runTestSlogCoreLevels(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, levelTestCases())
}

func runTestSlogCoreEnabled(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()

	// Create core with Info level
	zapCore := slogzap.NewCore(recorder, zap.InfoLevel)

	// Debug should be disabled
	core.AssertFalse(t, zapCore.Enabled(zap.DebugLevel), "debug disabled")

	// Info and above should be enabled
	core.AssertTrue(t, zapCore.Enabled(zap.InfoLevel), "info enabled")
	core.AssertTrue(t, zapCore.Enabled(zap.WarnLevel), "warn enabled")
	core.AssertTrue(t, zapCore.Enabled(zap.ErrorLevel), "error enabled")
}

func runTestNewZapLogger(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()

	// Create zap logger using convenience function
	zapLogger := slogzap.NewZapLogger(recorder)

	// Should default to Info level
	zapLogger.Debug("debug message")
	messages := recorder.GetMessages()
	core.AssertEqual(t, 0, len(messages), "debug count")

	// Info should work
	recorder.Clear()
	zapLogger.Info("info message")
	messages = recorder.GetMessages()
	core.AssertMustEqual(t, 1, len(messages), "info message count")
	core.AssertEqual(t, "info message", messages[0].Message, "message")
}

func runTestSlogCoreWithCaller(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()

	// Create zap logger with caller info
	zapLogger := slogzap.NewZapLogger(recorder, zap.AddCaller())

	zapLogger.Info("test with caller")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)

	msg := messages[0]
	core.AssertNotNil(t, msg.Fields["caller"], "caller")
	callerStr, ok := msg.Fields["caller"].(string)
	core.AssertTrue(t, ok, "caller type")
	core.AssertContains(t, callerStr, "core_test", "caller filename")
}

func runTestSlogCoreCallerUndefined(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()
	zapCore := slogzap.NewCore(recorder, zap.DebugLevel)

	// Create entry without caller info
	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Message: "no caller",
		Caller:  zapcore.EntryCaller{}, // Undefined caller
	}

	err := zapCore.Write(entry, nil)
	core.AssertNil(t, err, "write error")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)

	// Should not have caller field when undefined
	_, hasCaller := messages[0].Fields["caller"]
	core.AssertFalse(t, hasCaller, "Should not have caller field when undefined")
}

func runTestSlogCoreWithStack(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()
	zapCore := slogzap.NewCore(recorder, zap.DebugLevel)

	// Create an entry with stack trace
	entry := zapcore.Entry{
		Level:   zapcore.ErrorLevel,
		Message: "error with stack",
		Stack:   "goroutine 1 [running]:\nmain.main()\n\t/tmp/test.go:10 +0x20",
	}

	err := zapCore.Write(entry, nil)
	core.AssertNil(t, err, "write error")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)

	msg := messages[0]
	core.AssertNotNil(t, msg.Fields["stacktrace"], "stacktrace field")
	stackStr, ok := msg.Fields["stacktrace"].(string)
	core.AssertTrue(t, ok, "Stacktrace field type")
	core.AssertContains(t, stackStr, "goroutine 1", "goroutine info")
}

func runTestSlogCoreComplexFields(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()

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
	slogtest.AssertMustMessageCount(t, messages, 1)

	msg := messages[0]
	// Check basic field types
	slogtest.AssertField(t, msg, "string", "value")
	slogtest.AssertField(t, msg, "int", int64(42))
	slogtest.AssertField(t, msg, "float", 3.14)
	slogtest.AssertField(t, msg, "bool", true)

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

func runTestSlogCoreNilLogger(t *testing.T) {
	t.Helper()
	core.AssertPanic(t, func() {
		slogzap.NewCore(nil, zap.InfoLevel)
	}, nil, "NewCore nil logger panic")
}

func runTestSlogCoreNilLevel(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()

	// Should default to InfoLevel
	zapCore := slogzap.NewCore(recorder, nil)

	// Debug should be disabled
	core.AssertFalse(t, zapCore.Enabled(zap.DebugLevel), "Debug default level")

	// Info should be enabled
	core.AssertTrue(t, zapCore.Enabled(zap.InfoLevel), "Info default level")
}

func runTestSlogCoreSync(t *testing.T) {
	recorder := mock.NewLogger()

	zapCore := slogzap.NewCore(recorder, zap.InfoLevel)

	// Sync should always succeed (no-op for slog)
	core.AssertNil(t, zapCore.Sync(), "Sync() returned error")
}

func runTestBidirectionalSlogToZap(t *testing.T) {
	t.Helper()
	// Test slog → zap: Use slog backend with zap API
	baseRecorder := mock.NewLogger()
	zapLogger := slogzap.NewZapLogger(baseRecorder)

	// Use zap API
	zapLogger.Info("via zap api",
		zap.String("path", "slog->zap"),
		zap.Int("test_id", 1),
	)

	// Verify slog recorder received the message
	messages := baseRecorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)

	msg := messages[0]
	core.AssertEqual(t, "via zap api", msg.Message, "message text")
	slogtest.AssertField(t, msg, "path", "slog->zap")
	slogtest.AssertField(t, msg, "test_id", int64(1))
}

func runTestBidirectionalZapToSlog(t *testing.T) {
	t.Helper()
	// Test zap → slog: Use zap backend with slog API
	zapConfig := slogzap.NewDefaultConfig()
	slogLogger, err := slogzap.New(zapConfig)
	core.AssertMustNil(t, err, "create slog logger")

	// Use slog API (this internally uses zap)
	slogLogger.Info().
		WithField("path", "zap->slog").
		WithField("test_id", 2).
		Print("via slog api")

	// This direction is harder to test without capturing zap output,
	// but we can verify it doesn't panic or error
}

func runTestBidirectionalCompatibility(t *testing.T) {
	t.Helper()
	// Verify both directions work with the same field types
	baseRecorder := mock.NewLogger()
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
	slogtest.AssertMustMessageCount(t, messages, 1)

	// Verify field types are preserved correctly
	msg := messages[0]
	slogtest.AssertField(t, msg, "string", "test")
	slogtest.AssertField(t, msg, "int", int64(42))
	slogtest.AssertField(t, msg, "bool", true)
}

func runTestSlogCoreConcurrent(t *testing.T) {
	recorder := mock.NewLogger()
	zapLogger := slogzap.NewZapLogger(recorder)

	const goroutines = 10
	const msgsPerGoroutine = 10 // Reduced for cleaner test output

	runZapConcurrentLoggers(zapLogger, goroutines, msgsPerGoroutine)

	messages := recorder.GetMessages()
	expectedCount := goroutines * msgsPerGoroutine
	core.AssertEqual(t, expectedCount, len(messages), "message count")

	counts := tallyZapGoroutineCounts(messages)
	for i := range goroutines {
		core.AssertEqual(t, msgsPerGoroutine, counts[i], "Goroutine %d message count", i)
	}
}

// runZapConcurrentLoggers fans out N goroutines, each emitting M
// messages tagged with its goroutine id, and waits for them all to
// finish before returning.
func runZapConcurrentLoggers(zapLogger *zap.Logger, goroutines, msgsPerGoroutine int) {
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func(id int) {
			defer wg.Done()
			logger := zapLogger.With(zap.Int("goroutine", id))
			for j := range msgsPerGoroutine {
				logger.Info(fmt.Sprintf("msg-%d-%d", id, j),
					zap.Int("index", j),
				)
			}
		}(i)
	}
	wg.Wait()
}

// tallyZapGoroutineCounts walks the recorded messages and tallies how
// many came from each goroutine. The "goroutine" field may surface as
// either int or int64 depending on field-encoding paths.
func tallyZapGoroutineCounts(messages []mock.Message) map[int]int {
	counts := make(map[int]int)
	for _, msg := range messages {
		gid, ok := readGoroutineID(msg.Fields)
		if ok {
			counts[gid]++
		}
	}
	return counts
}

// readGoroutineID extracts the per-goroutine id from a message's
// fields, accepting either int or int64 to absorb encoding drift.
func readGoroutineID(fields map[string]any) (int, bool) {
	if gid, ok := fields["goroutine"].(int64); ok {
		return int(gid), true
	}
	if gid, ok := fields["goroutine"].(int); ok {
		return gid, true
	}
	return 0, false
}

func runTestSlogCoreCheckDisabled(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()

	// Create core with Info level
	zapCore := slogzap.NewCore(recorder, zap.InfoLevel)

	// Create a debug entry (which should be disabled)
	entry := zapcore.Entry{
		Level:   zap.DebugLevel,
		Message: "debug message",
	}

	// Check should return nil for disabled level
	checked := zapCore.Check(entry, nil)
	if checked != nil {
		t.Error("Check should return nil for disabled level")
	}

	// Verify with a CheckedEntry
	ce := &zapcore.CheckedEntry{}
	result := zapCore.Check(entry, ce)
	if result != ce {
		t.Error("Check should return the same CheckedEntry when level is disabled")
	}
}

func runTestSlogCoreFatalWrite(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()
	zapCore := slogzap.NewCore(recorder, zap.DebugLevel)

	// Write only records the entry; the exit belongs to the slog
	// backend or to zap's CheckedEntry machinery, not to the Core.
	entry := zapcore.Entry{
		Level:   zapcore.FatalLevel,
		Message: "fatal error occurred",
	}
	core.AssertNoError(t, zapCore.Write(entry, nil), "write")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)
	core.AssertEqual(t, slog.Fatal, messages[0].Level, "level")
	core.AssertEqual(t, "fatal error occurred", messages[0].Message, "message")
}

// runTestSlogCoreFatalHook verifies the zap layer owns Fatal terminal
// behaviour; WriteThenPanic substitutes the untestable os.Exit.
func runTestSlogCoreFatalHook(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()
	zapLogger := zap.New(slogzap.NewCore(recorder, zap.DebugLevel),
		zap.WithFatalHook(zapcore.WriteThenPanic))

	core.AssertPanic(t, func() {
		zapLogger.Fatal("fatal error occurred")
	}, "fatal error occurred", "fatal hook")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)
	core.AssertEqual(t, slog.Fatal, messages[0].Level, "level")
}

func runTestSlogCorePanicCoreWrite(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()
	zapCore := slogzap.NewCore(recorder, zap.DebugLevel)

	// A bare Core.Write never panics; like every zapcore.Core it
	// only writes, and CheckedEntry adds the terminal action.
	entry := zapcore.Entry{
		Level:   zapcore.PanicLevel,
		Message: "panic message",
	}
	core.AssertNoPanic(t, func() {
		_ = zapCore.Write(entry, nil)
	}, "core write")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)
	core.AssertEqual(t, slog.Panic, messages[0].Level, "level")
}

func runTestSlogCorePanicWrite(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()
	zapLogger := zap.New(slogzap.NewCore(recorder, zap.DebugLevel))

	core.AssertPanic(t, func() {
		zapLogger.Panic("panic message")
	}, "panic message", "zap layer panic")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)
	core.AssertEqual(t, slog.Panic, messages[0].Level, "level")
}

func runTestSlogCoreDPanicProduction(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()
	zapLogger := zap.New(slogzap.NewCore(recorder, zap.DebugLevel))

	core.AssertNoPanic(t, func() {
		zapLogger.DPanic("development panic")
	}, "production DPanic")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)
	core.AssertEqual(t, slog.Error, messages[0].Level, "level")
}

func runTestSlogCoreDPanicDevelopment(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()
	zapLogger := zap.New(slogzap.NewCore(recorder, zap.DebugLevel),
		zap.Development())

	core.AssertPanic(t, func() {
		zapLogger.DPanic("development panic")
	}, "development panic", "development DPanic")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)
	core.AssertEqual(t, slog.Error, messages[0].Level, "level")
}

func runTestConvertFieldsEmpty(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()
	zapCore := slogzap.NewCore(recorder, zap.InfoLevel)
	zapLogger := zap.New(zapCore)

	// Test with no fields
	zapLogger.Info("no fields")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)

	msg := messages[0]
	core.AssertTrue(t, len(msg.Fields) == 0, "empty fields map, got %v", msg.Fields)

	// Test with empty field slice explicitly
	recorder.Clear()
	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Message: "empty field slice",
	}
	err := zapCore.Write(entry, []zapcore.Field{})
	core.AssertNil(t, err, "write error")

	messages = recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)
	core.AssertTrue(t, len(messages[0].Fields) == 0,
		"Expected empty fields map for empty field slice, got %v", messages[0].Fields)
}

// runTestZapLoggerInvalidLevel tests invalid level handling in zap logger
func runTestZapLoggerInvalidLevel(t *testing.T) {
	t.Helper()
	zapConfig := slogzap.NewDefaultConfig()
	logger, err := slogzap.New(zapConfig)
	core.AssertMustNil(t, err, "create logger")

	// Test invalid level panic - use a value above normal range but within int8
	core.AssertPanic(t, func() {
		logger.WithLevel(slog.LogLevel(100)).Print("invalid level")
	}, nil, "invalid level panic")
}

// mapTestCase represents a test case for zap to slog level mapping
// through SlogCore.Write.
type mapTestCase struct {
	name              string
	expectedSlogLevel slog.LogLevel
	zapLevel          zapcore.Level
}

func (tc mapTestCase) Name() string {
	return tc.name
}

func (tc mapTestCase) Test(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()
	zapCore := slogzap.NewCore(recorder, zapcore.DebugLevel)

	entry := zapcore.Entry{
		Level:   tc.zapLevel,
		Message: fmt.Sprintf("test level %v", tc.zapLevel),
	}
	core.AssertNoError(t, zapCore.Write(entry, nil), "write")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)
	core.AssertEqual(t, tc.expectedSlogLevel, messages[0].Level,
		"slog level for zap level %v", tc.zapLevel)
}

// configTestCase represents a test case for configuration testing
type configTestCase struct {
	config zap.Config
	name   string
}

func (tc configTestCase) Name() string {
	return tc.name
}

func (tc configTestCase) Test(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()
	// Create zap logger with specific config
	zapLogger, err := tc.config.Build(zap.WrapCore(func(_ zapcore.Core) zapcore.Core {
		return slogzap.NewCore(recorder, zap.InfoLevel)
	}))
	if err != nil {
		t.Fatalf("Failed to build zap logger: %v", err)
	}

	zapLogger.Info("test message")

	messages := recorder.GetMessages()
	core.AssertEqual(t, 1, len(messages), "message count")
}
