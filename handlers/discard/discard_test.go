package discard_test

import (
	"testing"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
)

func TestDiscardLoglet(t *testing.T) {
	logger := discard.New()

	// Test level transitions
	testLevels := []struct {
		name    string
		method  func() slog.Logger
		level   slog.LogLevel
		enabled bool
	}{
		{"Debug", logger.Debug, slog.Debug, false},
		{"Info", logger.Info, slog.Info, false},
		{"Warn", logger.Warn, slog.Warn, false},
		{"Error", logger.Error, slog.Error, false},
		{"Fatal", logger.Fatal, slog.Fatal, true},
		{"Panic", logger.Panic, slog.Panic, true},
	}

	for _, tt := range testLevels {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.method()
			if l == nil {
				t.Fatal("logger method returned nil")
			}

			// Check if enabled state matches expected
			if got := l.Enabled(); got != tt.enabled {
				t.Errorf("Enabled() = %v, want %v", got, tt.enabled)
			}

			// Test WithEnabled
			wl, enabled := l.WithEnabled()
			if wl == nil {
				t.Fatal("WithEnabled returned nil logger")
			}
			if enabled != tt.enabled {
				t.Errorf("WithEnabled() enabled = %v, want %v", enabled, tt.enabled)
			}
		})
	}
}

func TestDiscardWithFields(t *testing.T) {
	logger := discard.New()

	// Test WithField
	l1 := logger.WithField("key1", "value1")
	if l1 == nil {
		t.Fatal("WithField returned nil")
	}

	// Test WithField with empty key (should return same logger)
	l2 := logger.WithField("", "value")
	if l2 != logger {
		t.Error("WithField with empty key should return same logger")
	}

	// Test WithFields
	fields := map[string]any{
		"key2": "value2",
		"key3": 123,
	}
	l3 := logger.WithFields(fields)
	if l3 == nil {
		t.Fatal("WithFields returned nil")
	}

	// Test WithFields with empty map
	l4 := logger.WithFields(map[string]any{})
	if l4 != logger {
		t.Error("WithFields with empty map should return same logger")
	}

	// Test WithFields removes empty keys
	fieldsWithEmpty := map[string]any{
		"":     "should be removed",
		"key4": "value4",
	}
	l5 := logger.WithFields(fieldsWithEmpty)
	if l5 == nil {
		t.Fatal("WithFields returned nil")
	}
}

func TestDiscardWithStack(t *testing.T) {
	logger := discard.New()

	// Test WithStack
	l := logger.WithStack(1)
	if l == nil {
		t.Fatal("WithStack returned nil")
	}
}

func TestDiscardChaining(t *testing.T) {
	// Test method chaining preserves fields and level
	logger := discard.New()

	l := logger.
		WithField("key1", "value1").
		WithField("key2", "value2").
		Fatal().
		WithField("key3", "value3")

	if l == nil {
		t.Fatal("Chained logger is nil")
	}

	// Should be enabled (Fatal level)
	if !l.Enabled() {
		t.Error("Fatal logger should be enabled")
	}
}

func TestDiscardLevelValidation(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid log level")
		}
	}()

	logger := discard.New()
	// This should panic
	logger.WithLevel(slog.UndefinedLevel)
}

func TestDiscardPrint(_ *testing.T) {
	logger := discard.New()

	// Non-fatal levels should not panic
	logger.Debug().Print("test")
	logger.Info().Printf("test %d", 123)
	logger.Warn().Println("test")

	// Can't test Fatal/Panic Print methods as they exit/panic
}
