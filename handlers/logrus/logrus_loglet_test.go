package logrus_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"darvaza.org/core"
	"darvaza.org/slog"
	slogrus "darvaza.org/slog/handlers/logrus"
)

func TestLevel(t *testing.T) {
	// Test nil receiver
	var nilLogger *slogrus.Logger
	core.AssertEqual(t, slog.UndefinedLevel, nilLogger.Level(), "nil logger level")

	// Test normal logger
	logrusLogger := logrus.New()
	logger := slogrus.New(logrusLogger)
	rlLogger := core.AssertMustTypeIs[*slogrus.Logger](t, logger, "logger type")
	core.AssertEqual(t, slog.UndefinedLevel, rlLogger.Level(), "default level")

	// Test level-specific logger
	warnLogger := core.AssertMustTypeIs[*slogrus.Logger](t, logger.Warn(), "warn logger type")
	core.AssertEqual(t, slog.Warn, warnLogger.Level(), "warn level")
}

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = logrusLevelTestCase{}

func TestLogrusLoglet(t *testing.T) {
	// Create a logrus logger with buffer
	var buf bytes.Buffer
	logrusLogger := logrus.New()
	logrusLogger.Out = &buf
	logrusLogger.SetLevel(logrus.DebugLevel)
	logrusLogger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
	})

	// Create slog adapter
	logger := slogrus.New(logrusLogger)

	// Test level transitions using TestCase pattern
	core.RunTestCases(t, logrusLevelTestCases(logger, &buf))
}

func TestLogrusWithFields(t *testing.T) {
	var buf bytes.Buffer
	logrusLogger := logrus.New()
	logrusLogger.Out = &buf
	logrusLogger.SetLevel(logrus.DebugLevel)
	logrusLogger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
		SortingFunc: func(keys []string) {
			// Sort keys for consistent output
			for i := 0; i < len(keys); i++ {
				for j := i + 1; j < len(keys); j++ {
					if keys[i] > keys[j] {
						keys[i], keys[j] = keys[j], keys[i]
					}
				}
			}
		},
	})

	logger := slogrus.New(logrusLogger)

	// Test WithField
	buf.Reset()
	l1 := logger.Info().WithField("key1", "value1")
	l1.Print("test message")
	output := buf.String()
	core.AssertContains(t, output, "key1=value1", "key1 field")

	// Test WithFields
	buf.Reset()
	fields := map[string]any{
		"key2": "value2",
		"key3": 123,
	}
	l2 := logger.Info().WithFields(fields)
	l2.Print("test message")
	output = buf.String()
	core.AssertContains(t, output, "key2=value2", "key2 field")
	core.AssertContains(t, output, "key3=123", "key3 field")
}

func TestLogrusChaining(t *testing.T) {
	var buf bytes.Buffer
	logrusLogger := logrus.New()
	logrusLogger.Out = &buf
	logrusLogger.SetLevel(logrus.DebugLevel)
	logrusLogger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
	})

	logger := slogrus.New(logrusLogger)

	// Test method chaining preserves fields and level
	buf.Reset()
	l := logger.
		WithField("base", "value").
		Info().
		WithField("key1", "value1").
		WithField("key2", "value2")

	l.Print("chained message")
	output := buf.String()

	// Check all fields are present
	core.AssertContains(t, output, "base=value", "base field")
	core.AssertContains(t, output, "key1=value1", "key1 field")
	core.AssertContains(t, output, "key2=value2", "key2 field")
}

func TestLogrusWithStack(t *testing.T) {
	var buf bytes.Buffer
	logrusLogger := logrus.New()
	logrusLogger.Out = &buf
	logrusLogger.SetLevel(logrus.DebugLevel)
	logrusLogger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
	})

	logger := slogrus.New(logrusLogger)

	// Test WithStack
	buf.Reset()
	l := logger.Info().WithStack(0)
	l.Print("test with stack")

	output := buf.String()
	// Check for stack trace fields
	core.AssertContains(t, output, "method=", "method field")
	core.AssertContains(t, output, "call-stack=", "call-stack field")
}

func TestLogrusDisabledLevels(t *testing.T) {
	var buf bytes.Buffer
	logrusLogger := logrus.New()
	logrusLogger.Out = &buf
	logrusLogger.SetLevel(logrus.InfoLevel) // Only Info and above

	logger := slogrus.New(logrusLogger)

	// Debug should be disabled
	if logger.Debug().Enabled() {
		t.Error("Debug should be disabled when logrus level is Info")
	}

	// Info should be enabled
	if !logger.Info().Enabled() {
		t.Error("Info should be enabled when logrus level is Info")
	}
}

func TestLogrusLevelValidation(t *testing.T) {
	core.AssertPanic(t, func() {
		logrusLogger := logrus.New()
		logger := slogrus.New(logrusLogger)
		logger.WithLevel(slog.UndefinedLevel)
	}, nil, "invalid level panic")
}

type logrusLevelTestCase struct {
	name    string
	method  func() slog.Logger
	level   slog.LogLevel
	enabled bool
	logMsg  string
	buffer  *bytes.Buffer
}

func (tc logrusLevelTestCase) Name() string {
	return tc.name
}

func (tc logrusLevelTestCase) Test(t *testing.T) {
	t.Helper()
	tc.buffer.Reset()
	l := tc.method()
	core.AssertMustNotNil(t, l, "logger method")

	// Check if enabled state matches expected
	core.AssertEqual(t, tc.enabled, l.Enabled(), "enabled")

	// Test logging
	l.Printf("test %s", strings.ToLower(tc.name))
	core.AssertEqual(t, tc.logMsg, tc.buffer.String(), "log output")
}

func newLogrusLevelTestCase(
	name string, method func() slog.Logger, level slog.LogLevel,
	logMsg string, buffer *bytes.Buffer,
) logrusLevelTestCase {
	return logrusLevelTestCase{
		name:    name,
		method:  method,
		level:   level,
		enabled: true,
		logMsg:  logMsg,
		buffer:  buffer,
	}
}

func logrusLevelTestCases(logger slog.Logger, buffer *bytes.Buffer) []logrusLevelTestCase {
	return []logrusLevelTestCase{
		newLogrusLevelTestCase("Debug", logger.Debug, slog.Debug, "level=debug msg=\"test debug\"\n", buffer),
		newLogrusLevelTestCase("Info", logger.Info, slog.Info, "level=info msg=\"test info\"\n", buffer),
		newLogrusLevelTestCase("Warn", logger.Warn, slog.Warn, "level=warning msg=\"test warn\"\n", buffer),
		newLogrusLevelTestCase("Error", logger.Error, slog.Error, "level=error msg=\"test error\"\n", buffer),
	}
}
