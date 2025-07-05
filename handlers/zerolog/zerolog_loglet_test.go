package zerolog_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"darvaza.org/slog"
	slogzerolog "darvaza.org/slog/handlers/zerolog"
)

func TestZerologLoglet(t *testing.T) {
	// Create a zerolog logger with buffer
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.DebugLevel)

	// Create slog adapter
	logger := slogzerolog.New(&zl)

	// Test level transitions
	testLevels := []struct {
		name     string
		method   func() slog.Logger
		level    slog.LogLevel
		enabled  bool
		logLevel string
	}{
		{"Debug", logger.Debug, slog.Debug, true, "debug"},
		{"Info", logger.Info, slog.Info, true, "info"},
		{"Warn", logger.Warn, slog.Warn, true, "warn"},
		{"Error", logger.Error, slog.Error, true, "error"},
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

			var result map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
				t.Fatalf("Failed to parse log output: %v", err)
			}

			if result["level"] != tt.logLevel {
				t.Errorf("Log level = %v, want %v", result["level"], tt.logLevel)
			}
			expectedMsg := "test " + strings.ToLower(tt.name)
			if result["message"] != expectedMsg {
				t.Errorf("Log message = %v, want %v", result["message"], expectedMsg)
			}
		})
	}
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
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	if result["key1"] != "value1" {
		t.Errorf("Expected field key1=value1 in output: %v", result)
	}

	// Test WithFields
	buf.Reset()
	fields := map[string]any{
		"key2": "value2",
		"key3": 123,
	}
	l2 := logger.Info().WithFields(fields)
	l2.Print("test message")

	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	if result["key2"] != "value2" {
		t.Errorf("Expected field key2=value2 in output: %v", result)
	}
	if result["key3"] != float64(123) { // JSON unmarshals numbers as float64
		t.Errorf("Expected field key3=123 in output: %v", result)
	}
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
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Check all fields are present
	if result["base"] != "value" {
		t.Error("Missing base field from parent logger")
	}
	if result["key1"] != "value1" {
		t.Error("Missing key1 field")
	}
	if result["key2"] != "value2" {
		t.Error("Missing key2 field")
	}
}

func TestZerologWithStack(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.DebugLevel)
	logger := slogzerolog.New(&zl)

	// Test WithStack - verify it doesn't crash
	l := logger.Info().WithStack(0)
	if l == nil {
		t.Fatal("WithStack returned nil")
	}

	// WithStack in Loglet stores the stack but zerolog needs special config to output it
	// Just verify the logger works correctly
	buf.Reset()
	l.Print("test with stack")

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Verify basic log output works
	if result["message"] != "test with stack" {
		t.Errorf("Expected message in output: %v", result)
	}
}

func TestZerologDisabledLevels(t *testing.T) {
	var buf bytes.Buffer
	// Create logger that only logs Info and above
	zl := zerolog.New(&buf).Level(zerolog.InfoLevel)
	logger := slogzerolog.New(&zl)

	// Debug should be disabled
	if logger.Debug().Enabled() {
		t.Error("Debug should be disabled when zerolog level is Info")
	}

	// Info should be enabled
	if !logger.Info().Enabled() {
		t.Error("Info should be enabled when zerolog level is Info")
	}
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
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Zerolog puts errors in the "error" field
	if result["error"] != "test error" {
		t.Errorf("Expected error field in output: %v", result)
	}
}

func TestZerologLevelValidation(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid log level")
		}
	}()

	zl := zerolog.New(bytes.NewBuffer(nil))
	logger := slogzerolog.New(&zl)

	// This should panic
	logger.WithLevel(slog.UndefinedLevel)
}
