package logrus_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"darvaza.org/slog"
	slogrus "darvaza.org/slog/handlers/logrus"
)

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

	// Test level transitions
	testLevels := []struct {
		name    string
		method  func() slog.Logger
		level   slog.LogLevel
		enabled bool
		logMsg  string
	}{
		{"Debug", logger.Debug, slog.Debug, true, "level=debug msg=\"test debug\"\n"},
		{"Info", logger.Info, slog.Info, true, "level=info msg=\"test info\"\n"},
		{"Warn", logger.Warn, slog.Warn, true, "level=warning msg=\"test warn\"\n"},
		{"Error", logger.Error, slog.Error, true, "level=error msg=\"test error\"\n"},
	}

	for _, tt := range testLevels {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			l := tt.method()
			if l == nil {
				t.Fatal("logger method returned nil")
			}

			// Check if enabled state matches expected
			if got := l.Enabled(); got != tt.enabled {
				t.Errorf("Enabled() = %v, want %v", got, tt.enabled)
			}

			// Test logging
			l.Printf("test %s", strings.ToLower(tt.name))
			if got := buf.String(); got != tt.logMsg {
				t.Errorf("Log output = %q, want %q", got, tt.logMsg)
			}
		})
	}
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
	if !strings.Contains(output, "key1=value1") {
		t.Errorf("Expected field key1=value1 in output: %s", output)
	}

	// Test WithFields
	buf.Reset()
	fields := map[string]any{
		"key2": "value2",
		"key3": 123,
	}
	l2 := logger.Info().WithFields(fields)
	l2.Print("test message")
	output = buf.String()
	if !strings.Contains(output, "key2=value2") || !strings.Contains(output, "key3=123") {
		t.Errorf("Expected fields in output: %s", output)
	}
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
	if !strings.Contains(output, "base=value") {
		t.Error("Missing base field from parent logger")
	}
	if !strings.Contains(output, "key1=value1") {
		t.Error("Missing key1 field")
	}
	if !strings.Contains(output, "key2=value2") {
		t.Error("Missing key2 field")
	}
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
	if !strings.Contains(output, "method=") {
		t.Error("Expected method field in output with stack trace")
	}
	if !strings.Contains(output, "call-stack=") {
		t.Error("Expected call-stack field in output")
	}
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
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid log level")
		}
	}()

	logrusLogger := logrus.New()
	logger := slogrus.New(logrusLogger)

	// This should panic
	logger.WithLevel(slog.UndefinedLevel)
}
