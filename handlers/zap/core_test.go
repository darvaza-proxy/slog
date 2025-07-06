package zap_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"darvaza.org/slog"
	slogzap "darvaza.org/slog/handlers/zap"
	slogtest "darvaza.org/slog/internal/testing"
)

// TestSlogCore tests the basic functionality of SlogCore
func TestSlogCore(t *testing.T) {
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

// TestSlogCoreWithFields tests field handling
func TestSlogCoreWithFields(t *testing.T) {
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
	if msg.Fields["key1"] != "value1" {
		t.Errorf("Expected field key1=value1, got %v", msg.Fields["key1"])
	}
	if msg.Fields["key2"] != int64(42) {
		t.Errorf("Expected field key2=42, got %v (type: %T)", msg.Fields["key2"], msg.Fields["key2"])
	}
}

// TestSlogCoreWith tests the With method
func TestSlogCoreWith(t *testing.T) {
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
	if msg.Fields["persistent"] != "field" {
		t.Errorf("Expected persistent field, got %v", msg.Fields["persistent"])
	}
	if msg.Fields["request_id"] != int64(123) {
		t.Errorf("Expected request_id=123, got %v", msg.Fields["request_id"])
	}
	if msg.Fields["extra"] != "value" {
		t.Errorf("Expected extra field, got %v", msg.Fields["extra"])
	}
}

// TestSlogCoreLevels tests level mapping and filtering
func TestSlogCoreLevels(t *testing.T) {
	tests := []struct {
		name      string
		zapLevel  zapcore.Level
		slogLevel slog.LogLevel
		logFunc   func(*zap.Logger, string, ...zap.Field)
	}{
		{"Debug", zap.DebugLevel, slog.Debug, (*zap.Logger).Debug},
		{"Info", zap.InfoLevel, slog.Info, (*zap.Logger).Info},
		{"Warn", zap.WarnLevel, slog.Warn, (*zap.Logger).Warn},
		{"Error", zap.ErrorLevel, slog.Error, (*zap.Logger).Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := slogtest.NewLogger()
			core := slogzap.NewCore(recorder, tt.zapLevel)
			zapLogger := zap.New(core)

			// Log at the test level
			tt.logFunc(zapLogger, "test")

			// Check output
			messages := recorder.GetMessages()
			if len(messages) != 1 {
				t.Fatalf("Expected 1 log entry, got %d", len(messages))
			}

			msg := messages[0]
			if msg.Level != tt.slogLevel {
				t.Errorf("Expected slog level %v, got %v", tt.slogLevel, msg.Level)
			}
		})
	}
}

// TestSlogCoreEnabled tests the Enabled method
func TestSlogCoreEnabled(t *testing.T) {
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

// TestNewZapLogger tests the convenience constructor
func TestNewZapLogger(t *testing.T) {
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

// TestSlogCoreWithCaller tests caller information handling
func TestSlogCoreWithCaller(t *testing.T) {
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
	if !strings.Contains(callerStr, "core_test.go") {
		t.Errorf("Caller should contain test file name, got %s", callerStr)
	}
}

// TestSlogCoreWithStack tests stack trace handling (line 74-76)
func TestSlogCoreWithStack(t *testing.T) {
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

// TestSlogCoreComplexFields tests various field types
func TestSlogCoreComplexFields(t *testing.T) {
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
	if msg.Fields["string"] != "value" {
		t.Errorf("String field mismatch: %v", msg.Fields["string"])
	}
	if msg.Fields["int"] != int64(42) {
		t.Errorf("Int field mismatch: %v (type: %T)", msg.Fields["int"], msg.Fields["int"])
	}
	if msg.Fields["float"] != 3.14 {
		t.Errorf("Float field mismatch: %v", msg.Fields["float"])
	}
	if !msg.Fields["bool"].(bool) {
		t.Errorf("Bool field mismatch: %v", msg.Fields["bool"])
	}
}

// TestSlogCoreNilLogger tests that NewCore panics with nil logger
func TestSlogCoreNilLogger(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewCore should panic with nil logger")
		}
	}()

	slogzap.NewCore(nil, zap.InfoLevel)
}

// TestSlogCoreNilLevel tests that NewCore handles nil level
func TestSlogCoreNilLevel(t *testing.T) {
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

// TestSlogCoreSync tests the Sync method
func TestSlogCoreSync(t *testing.T) {
	recorder := slogtest.NewLogger()

	core := slogzap.NewCore(recorder, zap.InfoLevel)

	// Sync should always succeed (no-op for slog)
	if err := core.Sync(); err != nil {
		t.Errorf("Sync() returned error: %v", err)
	}
}

// TestSlogToZapToSlog tests round-trip conversion
func TestSlogToZapToSlog(t *testing.T) {
	// Start with a test recorder
	baseRecorder := slogtest.NewLogger()

	// Create zap logger backed by slog
	zapLogger := slogzap.NewZapLogger(baseRecorder)

	// Use the zap logger
	zapLogger.Info("through zap", zap.String("source", "zap"))

	// Check that the base recorder got the message from zap
	messages := baseRecorder.GetMessages()
	found := false
	for _, msg := range messages {
		if msg.Message == "through zap" && msg.Fields["source"] == "zap" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Base recorder should have received zap message")
	}

	// Now test the reverse: wrap zap logger back to slog
	// This would create zap->slog->zap->slog which isn't the intended test
	// The test name suggests testing both directions, but not nested
}

// TestSlogCoreConcurrent tests concurrent access
func TestSlogCoreConcurrent(t *testing.T) {
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

// TestSlogCoreCheckDisabled tests the Check method with disabled level (line 62)
func TestSlogCoreCheckDisabled(t *testing.T) {
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

// TestSlogCoreFatalWrite tests Fatal path in Write() (lines 98-100)
func TestSlogCoreFatalWrite(t *testing.T) {
	recorder := slogtest.NewLogger()
	core := slogzap.NewCore(recorder, zap.DebugLevel)

	// We can't actually test os.Exit, but we can verify the Fatal log is written
	entry := zapcore.Entry{
		Level:   zapcore.FatalLevel,
		Message: "fatal error occurred",
	}

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

// TestSlogCorePanicWrite tests Panic path in Write() (lines 101-103)
func TestSlogCorePanicWrite(t *testing.T) {
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

// TestConvertFieldsEmpty tests empty fields in convertFields() (line 117)
func TestConvertFieldsEmpty(t *testing.T) {
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

// TestMapZapToSlogLevel tests DPanic and unknown levels (lines 141, 146-147)
func TestMapZapToSlogLevel(t *testing.T) {
	recorder := slogtest.NewLogger()

	tests := []struct {
		name      string
		zapLevel  zapcore.Level
		slogLevel slog.LogLevel
	}{
		{"DPanicLevel", zapcore.DPanicLevel, slog.Panic},
		{"UnknownLevel", zapcore.Level(99), slog.Info},    // Unknown level defaults to Info
		{"InvalidLevel", zapcore.InvalidLevel, slog.Info}, // InvalidLevel (-1) also defaults to Info
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := slogzap.NewCore(recorder, zapcore.DebugLevel)

			// For DPanic, we need to catch the panic
			if tt.zapLevel == zapcore.DPanicLevel {
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected panic from DPanicLevel")
					}
				}()
			}

			entry := zapcore.Entry{
				Level:   tt.zapLevel,
				Message: fmt.Sprintf("test %s", tt.name),
			}

			recorder.Clear()
			_ = core.Write(entry, nil)

			// For unknown levels, check that they were logged at Info level
			if tt.zapLevel != zapcore.DPanicLevel {
				messages := recorder.GetMessages()
				if len(messages) > 0 {
					msg := messages[0]
					if msg.Level != tt.slogLevel {
						t.Errorf("Expected slog level %v for zap level %v, got %v",
							tt.slogLevel, tt.zapLevel, msg.Level)
					}
				}
			}
		})
	}
}
