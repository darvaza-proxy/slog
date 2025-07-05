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

	"darvaza.org/slog"
	slogtest "darvaza.org/slog/internal/testing"
)

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

// assertContains checks if the output contains the expected string
func assertContains(t *testing.T, output, expected string) {
	t.Helper()
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got: %s", expected, output)
	}
}

// testBasicLogging tests basic logging functionality
func testBasicLogging(t *testing.T) {
	t.Helper()
	var buf bytes.Buffer
	logrLogger := createTestLogger(&buf)
	logger := New(logrLogger)

	// Test basic logging
	logger.Info().Print("test info message")
	assertContains(t, buf.String(), "test info message")
}

// testLogLevels tests different log levels
func testLogLevels(t *testing.T) {
	t.Helper()
	var buf bytes.Buffer
	logrLogger := createTestLogger(&buf)
	logger := New(logrLogger)

	tests := []struct {
		name  string
		level slog.LogLevel
		fn    func() slog.Logger
	}{
		{"Debug", slog.Debug, logger.Debug},
		{"Info", slog.Info, logger.Info},
		{"Warn", slog.Warn, logger.Warn},
		{"Error", slog.Error, logger.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			l := tt.fn()
			l.Printf("test %s message", tt.name)

			assertContains(t, buf.String(), fmt.Sprintf("test %s message", tt.name))
		})
	}
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
	assertContains(t, output, "key1")
	assertContains(t, output, "value1")

	// Test WithFields
	buf.Reset()
	fields := map[string]any{
		"key2": "value2",
		"key3": 123,
	}
	logger.Info().WithFields(fields).Print("with fields")
	output = buf.String()
	assertContains(t, output, "key2")
	assertContains(t, output, "key3")
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
	var buf bytes.Buffer
	logrLogger := createTestLogger(&buf)
	logger := New(logrLogger)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		} else {
			// core.Panic wraps the message in a PanicError
			errStr := fmt.Sprintf("%v", r)
			if !strings.Contains(errStr, "panic message") {
				t.Errorf("unexpected panic value: %v", r)
			}
		}
	}()

	logger.Panic().Print("panic message")
}

// TestSink tests the Sink adapter (slog.Logger as logr)
func TestSink(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		var buf bytes.Buffer
		slogLogger := newTestLogger(&buf)

		logrLogger := NewLogr(slogLogger)

		logrLogger.Info("test message")
		assertContains(t, buf.String(), "test message")
	})

	t.Run("V-Levels", func(t *testing.T) {
		var buf bytes.Buffer
		slogLogger := newTestLogger(&buf)

		logrLogger := NewLogr(slogLogger)

		// V(0) should map to Info (not Warn any more)
		logrLogger.V(0).Info("info level")
		assertContains(t, buf.String(), "info level")

		buf.Reset()
		// V(1) should map to Debug
		logrLogger.V(1).Info("debug level")
		assertContains(t, buf.String(), "debug level")

		buf.Reset()
		// V(2) should also map to Debug
		logrLogger.V(2).Info("debug level 2")
		assertContains(t, buf.String(), "debug level 2")
	})

	t.Run("Error", func(t *testing.T) {
		var buf bytes.Buffer
		slogLogger := newTestLogger(&buf)

		logrLogger := NewLogr(slogLogger)

		err := errors.New("test error")
		logrLogger.Error(err, "error occurred")

		output := buf.String()
		assertContains(t, output, "error occurred")
		assertContains(t, output, "test error")
	})

	t.Run("WithValues", func(t *testing.T) {
		var buf bytes.Buffer
		slogLogger := newTestLogger(&buf)

		logrLogger := NewLogr(slogLogger)

		logrLogger.WithValues("key1", "value1", "key2", 42).Info("with values")

		output := buf.String()
		assertContains(t, output, "key1")
		assertContains(t, output, "value1")
		assertContains(t, output, "key2")
		assertContains(t, output, "42")
	})

	t.Run("WithName", func(t *testing.T) {
		var buf bytes.Buffer
		slogLogger := newTestLogger(&buf)

		logrLogger := NewLogr(slogLogger)

		logrLogger.WithName("component").WithName("subcomponent").Info("named logger")

		assertContains(t, buf.String(), "component.subcomponent")
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
	assertContains(t, output, "test message")
	assertContains(t, output, "round")
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
	tests := []struct {
		name      string
		slogLevel slog.LogLevel
		logrLevel int
		expected  slog.LogLevel
	}{
		{"Debug", slog.Debug, 1, slog.Debug},
		{"Info", slog.Info, 0, slog.Info},
		{"Warn", slog.Warn, 0, slog.Info}, // logr has no warn, maps to info
		{"Error", slog.Error, -1, slog.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSlogToLogrMapping(t, tt.slogLevel, tt.logrLevel)
			testLogrToSlogMapping(t, tt.logrLevel, tt.expected)
		})
	}
}

func testSlogToLogrMapping(t *testing.T, slogLevel slog.LogLevel, expected int) {
	t.Helper()
	mapped := mapToLogrLevel(slogLevel)
	if mapped != expected {
		t.Errorf("mapToLogrLevel(%v) = %d, want %d", slogLevel, mapped, expected)
	}
}

func testLogrToSlogMapping(t *testing.T, logrLevel int, expected slog.LogLevel) {
	t.Helper()
	if logrLevel >= 0 {
		backMapped := mapFromLogrLevel(logrLevel)
		if backMapped != expected {
			t.Errorf("mapFromLogrLevel(%d) = %v, want %v", logrLevel, backMapped, expected)
		}
	}
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
	slogtest.AssertMessageCount(t, messages, 1)
	slogtest.AssertMessage(t, messages[0], slog.Info, "test message")

	// Test with odd number of key-value pairs
	recorder.Clear()
	sink.Info(0, "test message", "key1", "value1", "incomplete")
	messages = recorder.GetMessages()
	slogtest.AssertMessageCount(t, messages, 1)
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
	var buf bytes.Buffer
	logrLogger := createTestLogger(&buf)
	logger := New(logrLogger)

	// Test that invalid level triggers panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid log level")
		}
	}()

	logger.WithLevel(slog.UndefinedLevel)
}
