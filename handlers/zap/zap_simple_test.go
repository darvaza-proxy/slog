package zap_test

import (
	"testing"

	"darvaza.org/slog"
	slogzap "darvaza.org/slog/handlers/zap"
)

func TestZapLogletSimple(t *testing.T) {
	// Create using the default config
	logger, err := slogzap.New(nil)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if logger == nil {
		t.Fatal("New() returned nil")
	}

	// Test level transitions
	testLevels := []struct {
		name   string
		method func() slog.Logger
		level  slog.LogLevel
	}{
		{"Debug", logger.Debug, slog.Debug},
		{"Info", logger.Info, slog.Info},
		{"Warn", logger.Warn, slog.Warn},
		{"Error", logger.Error, slog.Error},
		{"Fatal", logger.Fatal, slog.Fatal},
		{"Panic", logger.Panic, slog.Panic},
	}

	for _, tt := range testLevels {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.method()
			if l == nil {
				t.Fatal("logger method returned nil")
			}

			// Test WithEnabled
			wl, _ := l.WithEnabled()
			if wl == nil {
				t.Fatal("WithEnabled returned nil logger")
			}
		})
	}
}

func TestZapWithFieldsSimple(t *testing.T) {
	logger, err := slogzap.New(nil)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Test WithField
	l1 := logger.WithField("key1", "value1")
	if l1 == nil {
		t.Fatal("WithField returned nil")
	}
	if l1 == logger {
		t.Error("WithField should return a new logger")
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
	if l3 == logger {
		t.Error("WithFields should return a new logger")
	}

	// Test WithFields with empty map
	l4 := logger.WithFields(map[string]any{})
	if l4 != logger {
		t.Error("WithFields with empty map should return same logger")
	}
}

func TestZapWithStackSimple(t *testing.T) {
	logger, err := slogzap.New(nil)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Test WithStack
	l := logger.WithStack(1)
	if l == nil {
		t.Fatal("WithStack returned nil")
	}
	if l == logger {
		t.Error("WithStack should return a new logger")
	}
}

func TestZapChainingSimple(t *testing.T) {
	logger, err := slogzap.New(nil)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Test method chaining preserves immutability
	base := logger.WithField("base", "value")
	l1 := base.Info()
	l2 := base.Error()

	// l1 and l2 should be different instances
	if l1 == l2 {
		t.Error("Different log levels should create different instances")
	}

	// Adding fields to l1 shouldn't affect l2
	l1WithField := l1.WithField("key1", "value1")
	if l1WithField == l1 {
		t.Error("WithField should create a new instance")
	}
}

func TestZapLevelValidationSimple(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid log level")
		}
	}()

	logger, err := slogzap.New(nil)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	// This should panic
	logger.WithLevel(slog.UndefinedLevel)
}

func TestZapNoOp(t *testing.T) {
	logger := slogzap.NewNoop()
	if logger == nil {
		t.Fatal("NewNoop returned nil")
	}

	// Should not panic
	logger.Debug().Print("test")
	logger.Info().Printf("test %d", 123)
	logger.Warn().Println("test")
}
