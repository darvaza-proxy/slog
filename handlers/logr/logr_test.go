package logr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"

	"darvaza.org/core"
	"darvaza.org/slog"
	slogtest "darvaza.org/slog/internal/testing"
)

func TestLevel(t *testing.T) {
	// Test nil receiver
	var nilLogger *Logger
	core.AssertEqual(t, slog.UndefinedLevel, nilLogger.Level(), "nil logger level")

	// Test normal logger
	var buf bytes.Buffer
	logrLogger := funcr.New(func(prefix, args string) {
		_, _ = buf.WriteString(prefix + args + "\n")
	}, funcr.Options{})

	logger := New(logrLogger)
	llLogger := core.AssertMustTypeIs[*Logger](t, logger, "logger type")
	core.AssertEqual(t, slog.UndefinedLevel, llLogger.Level(), "default level")

	// Test level-specific logger
	errorLogger := core.AssertMustTypeIs[*Logger](t, logger.Error(), "error logger type")
	core.AssertEqual(t, slog.Error, errorLogger.Level(), "error level")
}

// TestCase interface validations
var _ core.TestCase = logLevelTestCase{}
var _ core.TestCase = levelMappingConsistencyTestCase{}

// logLevelTestCase tests different log levels functionality.
type logLevelTestCase struct {
	buf   *bytes.Buffer
	fn    func() slog.Logger
	level slog.LogLevel
	name  string
}

// Name returns the test case name.
func (tc logLevelTestCase) Name() string {
	return tc.name
}

// Test executes the log level test.
func (tc logLevelTestCase) Test(t *testing.T) {
	t.Helper()
	tc.buf.Reset()
	l := tc.fn()
	l.Printf("test %s message", tc.name)
	expected := fmt.Sprintf("test %s message", tc.name)
	core.AssertContains(t, tc.buf.String(), expected, "test message")
}

// newLogLevelTestCase creates a new log level test case.
func newLogLevelTestCase(name string, buf *bytes.Buffer, level slog.LogLevel,
	fn func() slog.Logger) logLevelTestCase {
	return logLevelTestCase{
		name:  name,
		buf:   buf,
		level: level,
		fn:    fn,
	}
}

// levelMappingConsistencyTestCase tests level mapping consistency between Logger and Sink.
type levelMappingConsistencyTestCase struct {
	slogLevel slog.LogLevel
	logrLevel int
	expected  slog.LogLevel
	name      string
}

// Name returns the test case name.
func (tc levelMappingConsistencyTestCase) Name() string {
	return tc.name
}

// Test executes the level mapping consistency test.
func (tc levelMappingConsistencyTestCase) Test(t *testing.T) {
	t.Helper()

	// Test slog to logr mapping
	mapped := mapToLogrLevel(tc.slogLevel)
	core.AssertEqual(t, tc.logrLevel, mapped, "slog to logr mapping")

	// Test logr to slog mapping (only for valid levels)
	if tc.logrLevel >= 0 {
		reversed := mapFromLogrLevel(tc.logrLevel)
		core.AssertEqual(t, tc.expected, reversed, "logr to slog mapping")
	}
}

// newLevelMappingConsistencyTestCase creates a new level mapping consistency test case.
func newLevelMappingConsistencyTestCase(name string, slogLevel slog.LogLevel,
	logrLevel int, expected slog.LogLevel) levelMappingConsistencyTestCase {
	return levelMappingConsistencyTestCase{
		name:      name,
		slogLevel: slogLevel,
		logrLevel: logrLevel,
		expected:  expected,
	}
}

// levelString returns the string representation of a LogLevel
func levelString(level slog.LogLevel) string {
	switch level {
	case slog.Debug:
		return "DEBUG"
	case slog.Info:
		return "INFO"
	case slog.Warn:
		return "WARN"
	case slog.Error:
		return "ERROR"
	case slog.Fatal:
		return "FATAL"
	case slog.Panic:
		return "PANIC"
	default:
		return fmt.Sprintf("LEVEL(%d)", level)
	}
}

// createTestLogger creates a funcr logger that writes to a buffer
func createTestLogger(buf *bytes.Buffer) logr.Logger {
	return funcr.New(func(prefix, args string) {
		_, _ = buf.WriteString(prefix)
		if args != "" {
			_, _ = buf.WriteString(" ")
			_, _ = buf.WriteString(args)
		}
		_, _ = buf.WriteString("\n")
	}, funcr.Options{
		Verbosity: 2,
	})
}

// testBasicLogging tests basic logging functionality
func testBasicLogging(t *testing.T) {
	t.Helper()
	var buf bytes.Buffer
	logrLogger := createTestLogger(&buf)
	logger := New(logrLogger)

	// Test basic logging
	logger.Info().Print("test info message")
	core.AssertContains(t, buf.String(), "test info message", "info message")
}

// testLogLevels tests different log levels
func testLogLevels(t *testing.T) {
	t.Helper()
	var buf bytes.Buffer
	logrLogger := createTestLogger(&buf)
	logger := New(logrLogger)

	tests := []logLevelTestCase{
		newLogLevelTestCase("Debug", &buf, slog.Debug, logger.Debug),
		newLogLevelTestCase("Info", &buf, slog.Info, logger.Info),
		newLogLevelTestCase("Warn", &buf, slog.Warn, logger.Warn),
		newLogLevelTestCase("Error", &buf, slog.Error, logger.Error),
	}

	core.RunTestCases(t, tests)
}

// TestLogger tests the Logger adapter (logr as slog.Logger)
func TestLogger(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		testBasicLogging(t)
	})

	t.Run("Levels", func(t *testing.T) {
		testLogLevels(t)
	})

	t.Run("WithFields", func(t *testing.T) {
		testWithFields(t)
	})

	t.Run("Enabled", func(t *testing.T) {
		testEnabled(t)
	})

	t.Run("Panic", func(t *testing.T) {
		testPanic(t)
	})

	t.Run("DisabledLogger", func(t *testing.T) {
		testDisabledLogger(t)
	})

	t.Run("NilLogger", func(t *testing.T) {
		testNilLogger(t)
	})
}

// testWithFields tests field attachment functionality
func testWithFields(t *testing.T) {
	t.Helper()
	var buf bytes.Buffer
	logrLogger := createTestLogger(&buf)
	logger := New(logrLogger)

	// Test WithField
	logger.Info().WithField("key1", "value1").Print("with field")
	output := buf.String()
	core.AssertContains(t, output, "key1", "key1")
	core.AssertContains(t, output, "value1", "value1")

	// Test WithFields
	buf.Reset()
	fields := map[string]any{
		"key2": "value2",
		"key3": 123,
	}
	logger.Info().WithFields(fields).Print("with fields")
	output = buf.String()
	core.AssertContains(t, output, "key2", "key2")
	core.AssertContains(t, output, "key3", "key3")
}

// testEnabled tests the Enabled() method
func testEnabled(t *testing.T) {
	t.Helper()
	// With Verbosity: 0, only V(0) is enabled
	// Our mapping: Warn/Info -> V(0), Debug -> V(1)
	logrLogger := funcr.New(func(_, _ string) {}, funcr.Options{
		Verbosity: 0, // Only V(0) enabled
	})

	logger := New(logrLogger)

	// Debug (V(1)) should be disabled
	if logger.Debug().Enabled() {
		t.Error("Debug should be disabled with verbosity 0")
	}

	// Info (V(0)) should be enabled
	if !logger.Info().Enabled() {
		t.Error("Info should be enabled with verbosity 0")
	}

	// Warn (V(0)) should be enabled
	if !logger.Warn().Enabled() {
		t.Error("Warn should be enabled with verbosity 0")
	}
}

// testPanic tests panic functionality
func testPanic(t *testing.T) {
	t.Helper()

	core.AssertPanic(t, func() {
		var buf bytes.Buffer
		logrLogger := createTestLogger(&buf)
		logger := New(logrLogger)
		logger.Panic().Print("panic message")
	}, nil, "panic")
}

// TestSink tests the Sink adapter (slog.Logger as logr)
func TestSink(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		var buf bytes.Buffer
		slogLogger := newTestLogger(&buf)

		logrLogger := NewLogr(slogLogger)

		logrLogger.Info("test message")
		core.AssertContains(t, buf.String(), "test message", "test message")
	})

	t.Run("V-Levels", func(t *testing.T) {
		var buf bytes.Buffer
		slogLogger := newTestLogger(&buf)

		logrLogger := NewLogr(slogLogger)

		// V(0) should map to Info (not Warn any more)
		logrLogger.V(0).Info("info level")
		core.AssertContains(t, buf.String(), "info level", "info level")

		buf.Reset()
		// V(1) should map to Debug
		logrLogger.V(1).Info("debug level")
		core.AssertContains(t, buf.String(), "debug level", "debug level")

		buf.Reset()
		// V(2) should also map to Debug
		logrLogger.V(2).Info("debug level 2")
		core.AssertContains(t, buf.String(), "debug level 2", "debug level 2")
	})

	t.Run("Error", func(t *testing.T) {
		var buf bytes.Buffer
		slogLogger := newTestLogger(&buf)

		logrLogger := NewLogr(slogLogger)

		err := errors.New("test error")
		logrLogger.Error(err, "error occurred")

		output := buf.String()
		core.AssertContains(t, output, "error occurred", "error occurred")
		core.AssertContains(t, output, "test error", "test error")
	})

	t.Run("WithValues", func(t *testing.T) {
		var buf bytes.Buffer
		slogLogger := newTestLogger(&buf)

		logrLogger := NewLogr(slogLogger)

		logrLogger.WithValues("key1", "value1", "key2", 42).Info("with values")

		output := buf.String()
		core.AssertContains(t, output, "key1", "key1")
		core.AssertContains(t, output, "value1", "value1")
		core.AssertContains(t, output, "key2", "key2")
		core.AssertContains(t, output, "42", "42")
	})

	t.Run("WithName", func(t *testing.T) {
		var buf bytes.Buffer
		slogLogger := newTestLogger(&buf)

		logrLogger := NewLogr(slogLogger)

		logrLogger.WithName("component").WithName("subcomponent").Info("named logger")

		output := buf.String()
		core.AssertContains(t, output, "component.subcomponent", "component.subcomponent")
	})

	t.Run("Enabled", func(t *testing.T) {
		var buf bytes.Buffer
		slogLogger := newTestLogger(&buf)

		// Create a Sink with the logger
		sink := NewSink(slogLogger)

		// Test that V-levels are properly mapped
		if !sink.Enabled(0) { // Info/Warn
			t.Error("V(0) should be enabled")
		}
		if !sink.Enabled(1) { // Debug
			t.Error("V(1) should be enabled")
		}
		if !sink.Enabled(2) { // Debug (higher verbosity)
			t.Error("V(2) should be enabled")
		}
	})
}

// TestSinkDisabled tests Sink behaviour with disabled and nil loggers
func TestSinkDisabled(t *testing.T) {
	t.Run("DisabledSink", func(t *testing.T) {
		// Create a disabled slog logger for testing
		disabledLogger := &disabledTestLogger{}
		sink := NewSink(disabledLogger)

		// Test that all V-levels report as disabled
		testSinkDisabledLevels(t, sink)

		// Test that Info and Error don't panic when logger is disabled
		sink.Info(0, "should not panic")
		sink.Error(errors.New("test error"), "should not panic")
	})

	t.Run("NilSink", func(t *testing.T) {
		// Test nil logger in sink
		sink := &Sink{logger: nil}

		if sink.Enabled(0) {
			t.Error("Sink with nil logger should not be enabled")
		}

		// These should not panic
		sink.Info(0, "test")
		sink.Error(nil, "test")
	})
}

// testSinkDisabledLevels verifies all V-levels are disabled
func testSinkDisabledLevels(t *testing.T, sink logr.LogSink) {
	t.Helper()
	for i := 0; i <= 2; i++ {
		if sink.Enabled(i) {
			t.Errorf("V(%d) should be disabled for disabled logger", i)
		}
	}
}

// TestRoundTrip tests that we can round-trip between slog and logr
// testDisabledLogger tests behaviour when logger is disabled
func testDisabledLogger(t *testing.T) {
	t.Helper()
	var buf bytes.Buffer
	logger := createDisabledLogger(&buf)

	// Test each level
	testDisabledLevel(t, logger.Debug(), &buf, "Debug")
	testDisabledLevel(t, logger.Info(), &buf, "Info")
	testDisabledLevel(t, logger.Warn(), &buf, "Warn")

	// Error level should always be enabled in logr
	testEnabledError(t, logger.Error(), &buf)

	// Test WithEnabled for efficiency
	testWithEnabledDisabled(t, logger.Debug())
}

// createDisabledLogger creates a logger with nothing enabled
func createDisabledLogger(buf *bytes.Buffer) slog.Logger {
	logrLogger := funcr.New(func(prefix, args string) {
		_, _ = buf.WriteString(prefix)
		if args != "" {
			_, _ = buf.WriteString(" ")
			_, _ = buf.WriteString(args)
		}
		_, _ = buf.WriteString("\n")
	}, funcr.Options{
		Verbosity: -1, // Nothing enabled
	})
	return New(logrLogger)
}

// testDisabledLevel verifies a disabled logger level produces no output
func testDisabledLevel(t *testing.T, logger slog.Logger, buf *bytes.Buffer, level string) {
	t.Helper()
	if logger.Enabled() {
		t.Errorf("%s should be disabled with verbosity -1", level)
	}
	logger.Print("should not appear")
	if buf.Len() > 0 {
		t.Errorf("Disabled %s logger should not produce output", level)
	}
}

// testEnabledError verifies error level is always enabled
func testEnabledError(t *testing.T, errorLogger slog.Logger, buf *bytes.Buffer) {
	t.Helper()
	if !errorLogger.Enabled() {
		t.Error("Error should always be enabled in logr")
	}
	errorLogger.Print("error message")
	if !strings.Contains(buf.String(), "error message") {
		t.Error("Error logger should produce output even with verbosity -1")
	}
}

// testWithEnabledDisabled verifies WithEnabled behaviour for disabled loggers
func testWithEnabledDisabled(t *testing.T, logger slog.Logger) {
	t.Helper()
	l, enabled := logger.WithEnabled()
	if enabled {
		t.Error("WithEnabled should return false for disabled logger")
	}
	if l == nil {
		t.Error("WithEnabled should still return a valid logger")
	}
}

// testNilLogger tests behaviour with nil loggers
func testNilLogger(t *testing.T) {
	t.Helper()

	// Test nil receiver
	var logger *Logger
	if logger.Enabled() {
		t.Error("nil logger should not be enabled")
	}

	// Test logger with nil sink
	logrLogger := logr.New(nil)
	wrappedLogger := New(logrLogger)
	if wrappedLogger.Enabled() {
		t.Error("logger with nil sink should not be enabled")
	}

	// Verify methods don't panic on nil
	wrappedLogger.Print("test")
	wrappedLogger.Printf("test %s", "format")
	wrappedLogger.Println("test")

	// Verify WithEnabled works correctly
	l, enabled := wrappedLogger.WithEnabled()
	if enabled {
		t.Error("WithEnabled should return false for logger with nil sink")
	}
	if l != wrappedLogger {
		t.Error("WithEnabled should return the same logger instance")
	}
}

// TestRoundTrip tests that we can round-trip between slog and logr
func TestRoundTrip(t *testing.T) {
	var buf bytes.Buffer

	// Start with a test slog logger
	originalSlog := newTestLogger(&buf)

	// Convert to logr
	asLogr := NewLogr(originalSlog)

	// Convert back to slog
	backToSlog := New(asLogr)

	// Test that it still works
	backToSlog.Info().WithField("round", "trip").Print("test message")

	output := buf.String()
	core.AssertContains(t, output, "test message", "test message")
	core.AssertContains(t, output, "round", "round")
}

// testLogger is a simple slog.Logger implementation for testing
type testLogger struct {
	buf    *bytes.Buffer
	level  slog.LogLevel
	fields map[string]any
}

func newTestLogger(buf *bytes.Buffer) slog.Logger {
	return &testLogger{
		buf:    buf,
		level:  slog.Info,
		fields: make(map[string]any),
	}
}

func (*testLogger) Enabled() bool {
	return true
}

func (tl *testLogger) WithEnabled() (slog.Logger, bool) {
	return tl, true
}

func (tl *testLogger) Print(args ...any) {
	tl.write(fmt.Sprint(args...))
}

func (tl *testLogger) Println(args ...any) {
	tl.write(fmt.Sprintln(args...))
}

func (tl *testLogger) Printf(format string, args ...any) {
	tl.write(fmt.Sprintf(format, args...))
}

func (tl *testLogger) write(msg string) {
	entry := maps.Clone(tl.fields)
	entry["level"] = levelString(tl.level)
	entry["msg"] = strings.TrimSpace(msg)

	data, _ := json.Marshal(entry)
	_, _ = tl.buf.Write(data)
	_, _ = tl.buf.WriteString("\n")
}

func (tl *testLogger) Debug() slog.Logger {
	return tl.WithLevel(slog.Debug)
}

func (tl *testLogger) Info() slog.Logger {
	return tl.WithLevel(slog.Info)
}

func (tl *testLogger) Warn() slog.Logger {
	return tl.WithLevel(slog.Warn)
}

func (tl *testLogger) Error() slog.Logger {
	return tl.WithLevel(slog.Error)
}

func (tl *testLogger) Fatal() slog.Logger {
	return tl.WithLevel(slog.Fatal)
}

func (tl *testLogger) Panic() slog.Logger {
	return tl.WithLevel(slog.Panic)
}

func (tl *testLogger) WithLevel(level slog.LogLevel) slog.Logger {
	return &testLogger{
		buf:    tl.buf,
		level:  level,
		fields: tl.fields,
	}
}

func (tl *testLogger) WithStack(skip int) slog.Logger {
	newFields := maps.Clone(tl.fields)
	newFields["stack"] = fmt.Sprintf("stack(skip=%d)", skip)
	return &testLogger{
		buf:    tl.buf,
		level:  tl.level,
		fields: newFields,
	}
}

func (tl *testLogger) WithField(label string, value any) slog.Logger {
	if label == "" {
		return tl
	}
	newFields := maps.Clone(tl.fields)
	newFields[label] = value
	return &testLogger{
		buf:    tl.buf,
		level:  tl.level,
		fields: newFields,
	}
}

func (tl *testLogger) WithFields(fields map[string]any) slog.Logger {
	newFields := maps.Clone(tl.fields)
	for k, v := range fields {
		if k != "" {
			newFields[k] = v
		}
	}
	return &testLogger{
		buf:    tl.buf,
		level:  tl.level,
		fields: newFields,
	}
}

// disabledTestLogger is a test logger that is always disabled
type disabledTestLogger struct{}

func (*disabledTestLogger) Enabled() bool                           { return false }
func (d *disabledTestLogger) WithEnabled() (slog.Logger, bool)      { return d, false }
func (*disabledTestLogger) Print(...any)                            {}
func (*disabledTestLogger) Println(...any)                          {}
func (*disabledTestLogger) Printf(string, ...any)                   {}
func (d *disabledTestLogger) Debug() slog.Logger                    { return d }
func (d *disabledTestLogger) Info() slog.Logger                     { return d }
func (d *disabledTestLogger) Warn() slog.Logger                     { return d }
func (d *disabledTestLogger) Error() slog.Logger                    { return d }
func (d *disabledTestLogger) Fatal() slog.Logger                    { return d }
func (d *disabledTestLogger) Panic() slog.Logger                    { return d }
func (d *disabledTestLogger) WithLevel(slog.LogLevel) slog.Logger   { return d }
func (d *disabledTestLogger) WithStack(int) slog.Logger             { return d }
func (d *disabledTestLogger) WithField(string, any) slog.Logger     { return d }
func (d *disabledTestLogger) WithFields(map[string]any) slog.Logger { return d }

// TestCompliance runs the shared compliance test suite
func TestCompliance(t *testing.T) {
	newLogger := func() slog.Logger {
		logrLogger := funcr.New(func(_, _ string) {}, funcr.Options{})
		return New(logrLogger)
	}

	newWithRecorder := func(recorder slog.Logger) slog.Logger {
		logrLogger := logr.New(NewSink(recorder))
		return New(logrLogger)
	}

	compliance := slogtest.ComplianceTest{
		FactoryOptions: slogtest.FactoryOptions{
			NewLogger:             newLogger,
			NewLoggerWithRecorder: newWithRecorder,
		},
		AdapterOptions: slogtest.AdapterOptions{
			LevelExceptions: map[slog.LogLevel]slog.LogLevel{
				slog.Warn: slog.Info, // logr maps Warn to Info (V(0))
			},
		},
	}

	compliance.Run(t)
}

// TestLevelMethods tests level methods using shared test utilities
func TestLevelMethods(t *testing.T) {
	slogtest.TestLevelMethods(t, func() slog.Logger {
		logrLogger := funcr.New(func(_, _ string) {}, funcr.Options{})
		return New(logrLogger)
	})
}

// TestFieldMethods tests field methods using shared test utilities
func TestFieldMethods(t *testing.T) {
	slogtest.TestFieldMethods(t, func() slog.Logger {
		logrLogger := funcr.New(func(_, _ string) {}, funcr.Options{})
		return New(logrLogger)
	})
}

// TestWithStack tests stack functionality using shared test utilities
func TestWithStack(t *testing.T) {
	logrLogger := funcr.New(func(_, _ string) {}, funcr.Options{})
	logger := New(logrLogger)
	slogtest.TestWithStack(t, logger)
}

// TestStress tests the logger under stress conditions
func TestStress(t *testing.T) {
	newLogger := func() slog.Logger {
		logrLogger := funcr.New(func(_, _ string) {}, funcr.Options{})
		return New(logrLogger)
	}

	factory := func(recorder slog.Logger) slog.Logger {
		logrLogger := logr.New(NewSink(recorder))
		return New(logrLogger)
	}

	suite := slogtest.StressTestSuite{
		NewLogger:             newLogger,
		NewLoggerWithRecorder: factory,
	}

	suite.Run(t)
}

// TestBidirectionalAdapter tests using both adapters together
func TestBidirectionalAdapter(t *testing.T) {
	// Define expected level mappings for logr
	opts := &slogtest.BidirectionalTestOptions{
		AdapterOptions: slogtest.AdapterOptions{
			LevelExceptions: map[slog.LogLevel]slog.LogLevel{
				slog.Warn: slog.Info, // logr has no native Warn level
			},
		},
	}

	// Test slog -> logr -> slog round trip
	slogtest.TestBidirectionalWithOptions(t, "SlogToLogr", func(backend slog.Logger) slog.Logger {
		// Create a logr logger that outputs to the backend using NewSink
		logrLogger := logr.New(NewSink(backend))
		// Wrap it back as slog
		return New(logrLogger)
	}, opts)
}

// TestLevelMappingConsistency tests that level mappings are consistent between Logger and Sink
func TestLevelMappingConsistency(t *testing.T) {
	tests := []levelMappingConsistencyTestCase{
		newLevelMappingConsistencyTestCase("Debug", slog.Debug, 1, slog.Debug),
		newLevelMappingConsistencyTestCase("Info", slog.Info, 0, slog.Info),
		newLevelMappingConsistencyTestCase("Warn", slog.Warn, 0, slog.Info), // logr has no warn, maps to info
		newLevelMappingConsistencyTestCase("Error", slog.Error, -1, slog.Error),
	}

	core.RunTestCases(t, tests)
}

// Type assertion helpers for better testability
func asCallDepthLogSink(sink logr.LogSink) (logr.CallDepthLogSink, bool) {
	cdSink, ok := sink.(logr.CallDepthLogSink)
	return cdSink, ok
}

func asCallStackHelperLogSink(sink logr.LogSink) (logr.CallStackHelperLogSink, bool) {
	helperSink, ok := sink.(logr.CallStackHelperLogSink)
	return helperSink, ok
}

// TestSinkInterfaceHelpers tests the interface assertion helpers
func TestSinkInterfaceHelpers(t *testing.T) {
	recorder := slogtest.NewLogger()
	sink := NewSink(recorder)

	t.Run("CallDepthLogSink", func(t *testing.T) {
		testCallDepthInterface(t, sink)
	})

	t.Run("CallStackHelperLogSink", func(t *testing.T) {
		testCallStackHelperInterface(t, sink)
	})
}

func testCallDepthInterface(t *testing.T, sink logr.LogSink) {
	t.Helper()
	cdSink, ok := asCallDepthLogSink(sink)
	if !ok {
		t.Error("Sink should implement CallDepthLogSink")
		return
	}

	newSink := cdSink.WithCallDepth(5)
	if sinkImpl, ok := newSink.(*Sink); ok {
		if sinkImpl.callDepth != 5 {
			t.Errorf("WithCallDepth(5) should set callDepth to 5, got %d", sinkImpl.callDepth)
		}
	}
}

func testCallStackHelperInterface(t *testing.T, sink logr.LogSink) {
	t.Helper()
	helperSink, ok := asCallStackHelperLogSink(sink)
	if !ok {
		t.Error("Sink should implement CallStackHelperLogSink")
		return
	}

	helper := helperSink.GetCallStackHelper()
	if helper == nil {
		t.Error("GetCallStackHelper returned nil")
		return
	}

	// Test that calling the helper doesn't panic
	helper()
}

// TestSinkWithEmptyValues tests Sink behaviour with empty key-value pairs
func TestSinkWithEmptyValues(t *testing.T) {
	recorder := slogtest.NewLogger()
	sink := NewSink(recorder)

	// Test with empty key-value pairs
	sink.Info(0, "test message")
	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)

	slogtest.AssertMessage(t, messages[0], slog.Info, "test message")

	// Test with odd number of key-value pairs
	recorder.Clear()
	sink.Info(0, "test message", "key1", "value1", "incomplete")
	messages = recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)

	slogtest.AssertMessage(t, messages[0], slog.Info, "test message")
	slogtest.AssertField(t, messages[0], "key1", "value1")
	slogtest.AssertNoField(t, messages[0], "incomplete")
}

// TestSinkCallDepthHandling tests proper call depth handling
func TestSinkCallDepthHandling(t *testing.T) {
	recorder := slogtest.NewLogger()
	sink := NewSink(recorder)

	depthSink, ok := asCallDepthLogSink(sink)
	if !ok {
		t.Error("Sink should implement CallDepthLogSink")
		return
	}

	t.Run("SingleCallDepth", func(t *testing.T) {
		testSingleCallDepth(t, depthSink)
	})

	t.Run("ChainedCallDepth", func(t *testing.T) {
		testChainedCallDepth(t, depthSink)
	})
}

func testSingleCallDepth(t *testing.T, depthSink logr.CallDepthLogSink) {
	t.Helper()
	sink2 := depthSink.WithCallDepth(5)
	if sinkImpl, ok := sink2.(*Sink); ok {
		if sinkImpl.callDepth != 5 {
			t.Errorf("WithCallDepth(5) should set callDepth to 5, got %d", sinkImpl.callDepth)
		}
	}
}

func testChainedCallDepth(t *testing.T, depthSink logr.CallDepthLogSink) {
	t.Helper()
	sink2 := depthSink.WithCallDepth(5)

	depthSink2, ok := asCallDepthLogSink(sink2)
	if !ok {
		t.Error("Chained sink should still implement CallDepthLogSink")
		return
	}

	sink3 := depthSink2.WithCallDepth(3)
	if sinkImpl, ok := sink3.(*Sink); ok {
		if sinkImpl.callDepth != 8 {
			t.Errorf("Chained WithCallDepth should add depths, expected 8, got %d", sinkImpl.callDepth)
		}
	}
}

// TestLoggerUnwrap tests the Unwrap method
func TestLoggerUnwrap(t *testing.T) {
	originalLogr := funcr.New(func(_, _ string) {}, funcr.Options{})
	logger := New(originalLogr)

	unwrapped := logger.(*Logger).Unwrap()
	if unwrapped != originalLogr {
		t.Error("Unwrap() should return the original logr.Logger")
	}
}

// TestInvalidLogLevel tests behaviour with invalid log levels
func TestInvalidLogLevel(t *testing.T) {
	t.Run("UndefinedLevel", func(t *testing.T) {
		core.AssertPanic(t, func() {
			var buf bytes.Buffer
			logrLogger := createTestLogger(&buf)
			logger := New(logrLogger)
			logger.WithLevel(slog.UndefinedLevel)
		}, nil, "undefined level panic")
	})

	t.Run("NegativeLevel", func(t *testing.T) {
		core.AssertPanic(t, func() {
			var buf bytes.Buffer
			logrLogger := createTestLogger(&buf)
			logger := New(logrLogger)
			logger.WithLevel(slog.LogLevel(-1))
		}, nil, "negative level panic")
	})
}

// TestWithLevelSameLevel tests WithLevel when called with the same level
func TestWithLevelSameLevel(t *testing.T) {
	var buf bytes.Buffer
	logrLogger := createTestLogger(&buf)
	logger := New(logrLogger)

	// Create an Info level logger
	infoLogger := logger.Info()

	// Call WithLevel with the same level (Info)
	sameLogger := infoLogger.WithLevel(slog.Info)

	// Should return the same logger instance
	core.AssertSame(t, infoLogger, sameLogger, "same level logger")
}
