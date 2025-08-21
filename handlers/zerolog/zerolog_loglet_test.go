package zerolog_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"darvaza.org/core"
	"darvaza.org/slog"
	slogzerolog "darvaza.org/slog/handlers/zerolog"
)

func TestLevel(t *testing.T) {
	zl := zerolog.New(bytes.NewBuffer(nil))
	logger := slogzerolog.New(&zl)
	zlLogger := core.AssertMustTypeIs[*slogzerolog.Logger](t, logger, "logger type")
	core.AssertEqual(t, slog.UndefinedLevel, zlLogger.Level(), "default level")

	debugLogger := core.AssertMustTypeIs[*slogzerolog.Logger](t, logger.Debug(), "debug logger type")
	core.AssertEqual(t, slog.Debug, debugLogger.Level(), "debug level")
}

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = zerologLogletTestCase{}

type zerologLogletTestCase struct {
	name     string
	level    slog.LogLevel
	enabled  bool
	logLevel string
}

func (tc zerologLogletTestCase) Name() string {
	return tc.name
}

func (tc zerologLogletTestCase) Test(t *testing.T) {
	t.Helper()

	// Create a zerolog logger with buffer for this specific test case
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.DebugLevel)
	logger := slogzerolog.New(&zl)

	buf.Reset()

	// Get the level method based on the test case level
	var l slog.Logger
	switch tc.level {
	case slog.Debug:
		l = logger.Debug()
	case slog.Info:
		l = logger.Info()
	case slog.Warn:
		l = logger.Warn()
	case slog.Error:
		l = logger.Error()
	default:
		t.Fatalf("Unknown level: %v", tc.level)
	}

	core.AssertMustNotNil(t, l, "logger method")

	// Check if enabled state matches expected
	core.AssertEqual(t, tc.enabled, l.Enabled(), "Enabled() for %s", tc.name)

	// Test logging
	l.Printf("test %s", strings.ToLower(tc.name))

	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	core.AssertMustNil(t, err, "parse log output")

	level, ok := result["level"].(string)
	core.AssertMustTrue(t, ok, "level is string")
	core.AssertEqual(t, tc.logLevel, level, "Log level")

	expectedMsg := "test " + strings.ToLower(tc.name)
	message, ok := result["message"].(string)
	core.AssertMustTrue(t, ok, "message is string")
	core.AssertEqual(t, expectedMsg, message, "Log message")
}

func newZerologLogletTestCase(name string, level slog.LogLevel, logLevel string) zerologLogletTestCase {
	return zerologLogletTestCase{
		name:     name,
		level:    level,
		enabled:  true,
		logLevel: logLevel,
	}
}

func zerologLogletTestCases() []zerologLogletTestCase {
	return []zerologLogletTestCase{
		newZerologLogletTestCase("Debug", slog.Debug, "debug"),
		newZerologLogletTestCase("Info", slog.Info, "info"),
		newZerologLogletTestCase("Warn", slog.Warn, "warn"),
		newZerologLogletTestCase("Error", slog.Error, "error"),
	}
}

func TestZerologLoglet(t *testing.T) {
	core.RunTestCases(t, zerologLogletTestCases())
}

func TestZerologWithFields(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.DebugLevel)
	logger := slogzerolog.New(&zl)

	// Test WithField
	buf.Reset()
	l1 := logger.Info().WithField("key1", "value1")
	l1.Print("test message")

	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	core.AssertMustNil(t, err, "parse log output")
	core.AssertEqual(t, "value1", result["key1"], "field key1")

	// Test WithFields
	buf.Reset()
	fields := map[string]any{
		"key2": "value2",
		"key3": 123,
	}
	l2 := logger.Info().WithFields(fields)
	l2.Print("test message")

	err = json.Unmarshal(buf.Bytes(), &result)
	core.AssertMustNil(t, err, "parse log output")
	core.AssertEqual(t, "value2", result["key2"], "field key2")
	key3, ok := result["key3"].(float64)
	core.AssertMustTrue(t, ok, "key3 is float64")
	core.AssertEqual(t, float64(123), key3, "field key3")
}

func TestZerologChaining(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.DebugLevel)
	logger := slogzerolog.New(&zl)

	// Test method chaining preserves fields and level
	buf.Reset()
	l := logger.
		WithField("base", "value").
		Info().
		WithField("key1", "value1").
		WithField("key2", "value2")

	l.Print("chained message")

	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	core.AssertMustNil(t, err, "parse log output")

	// Check all fields are present
	core.AssertEqual(t, "value", result["base"], "base field")
	core.AssertEqual(t, "value1", result["key1"], "key1 field")
	core.AssertEqual(t, "value2", result["key2"], "key2 field")
}

func TestZerologWithStack(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.DebugLevel)
	logger := slogzerolog.New(&zl)

	// Test WithStack - verify it doesn't crash
	l := logger.Info().WithStack(0)
	core.AssertNotNil(t, l, "WithStack returned nil")

	// WithStack in Loglet stores the stack but zerolog needs special config to output it
	// Just verify the logger works correctly
	buf.Reset()
	l.Print("test with stack")

	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	core.AssertMustNil(t, err, "parse log output")

	// Verify basic log output works
	core.AssertEqual(t, "test with stack", result["message"], "message content")
}

func TestZerologDisabledLevels(t *testing.T) {
	var buf bytes.Buffer
	// Create logger that only logs Info and above
	zl := zerolog.New(&buf).Level(zerolog.InfoLevel)
	logger := slogzerolog.New(&zl)

	// Debug should be disabled
	core.AssertFalse(t, logger.Debug().Enabled(), "Debug enabled at Info level")

	// Info should be enabled
	core.AssertTrue(t, logger.Info().Enabled(), "Info enabled at Info level")
}

func TestZerologErrorField(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.DebugLevel)
	logger := slogzerolog.New(&zl)

	// Test error field handling
	buf.Reset()
	err := fmt.Errorf("test error")
	l := logger.Error().WithField(slog.ErrorFieldName, err)
	l.Print("error occurred")

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	core.AssertMustNil(t, err, "parse log output")

	// Zerolog puts errors in the "error" field
	core.AssertEqual(t, "test error", result["error"], "error field")
}

func TestZerologLevelValidation(t *testing.T) {
	core.AssertPanic(t, func() {
		zl := zerolog.New(bytes.NewBuffer(nil))
		logger := slogzerolog.New(&zl)
		logger.WithLevel(slog.UndefinedLevel)
	}, nil, "invalid level panic")
}
