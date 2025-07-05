package filter_test

import (
	"testing"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
)

// mockLogger is a simple logger for testing
type mockLogger struct {
	enabled bool
	level   slog.LogLevel
}

func (m *mockLogger) Enabled() bool                    { return m.enabled }
func (m *mockLogger) WithEnabled() (slog.Logger, bool) { return m, m.enabled }
func (m *mockLogger) Print(_ ...any)                   {}
func (m *mockLogger) Println(_ ...any)                 {}
func (m *mockLogger) Printf(_ string, _ ...any)        {}
func (m *mockLogger) Debug() slog.Logger               { return m.WithLevel(slog.Debug) }
func (m *mockLogger) Info() slog.Logger                { return m.WithLevel(slog.Info) }
func (m *mockLogger) Warn() slog.Logger                { return m.WithLevel(slog.Warn) }
func (m *mockLogger) Error() slog.Logger               { return m.WithLevel(slog.Error) }
func (m *mockLogger) Fatal() slog.Logger               { return m.WithLevel(slog.Fatal) }
func (m *mockLogger) Panic() slog.Logger               { return m.WithLevel(slog.Panic) }
func (m *mockLogger) WithLevel(level slog.LogLevel) slog.Logger {
	return &mockLogger{enabled: true, level: level}
}
func (m *mockLogger) WithStack(_ int) slog.Logger             { return m }
func (m *mockLogger) WithField(_ string, _ any) slog.Logger   { return m }
func (m *mockLogger) WithFields(_ map[string]any) slog.Logger { return m }

func TestFilterLoglet(t *testing.T) {
	// Create a base logger
	base := &mockLogger{enabled: true}

	// Create a filter logger with Info threshold
	logger := filter.New(base, slog.Info)

	// Test level transitions
	testLevels := []struct {
		name    string
		method  func() slog.Logger
		level   slog.LogLevel
		enabled bool
	}{
		{"Debug", logger.Debug, slog.Debug, false},
		{"Info", logger.Info, slog.Info, true},
		{"Warn", logger.Warn, slog.Warn, true},
		{"Error", logger.Error, slog.Error, true},
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
		})
	}
}

func TestFilterWithFields(t *testing.T) {
	base := &mockLogger{enabled: true}
	logger := filter.New(base, slog.Info)

	// Test WithField on root logger (should store in Loglet)
	l1 := logger.WithField("root", "value")
	if l1 == nil {
		t.Fatal("WithField returned nil")
	}

	// Test WithField on enabled logger
	l2 := logger.Info().WithField("key1", "value1")
	if l2 == nil {
		t.Fatal("WithField on enabled logger returned nil")
	}

	// Test WithFields
	fields := map[string]any{
		"key2": "value2",
		"key3": 123,
	}
	l3 := logger.Info().WithFields(fields)
	if l3 == nil {
		t.Fatal("WithFields returned nil")
	}
}

func TestFilterWithStack(t *testing.T) {
	base := &mockLogger{enabled: true}
	logger := filter.New(base, slog.Info)

	// Test WithStack on root logger
	l1 := logger.WithStack(1)
	if l1 == nil {
		t.Fatal("WithStack on root returned nil")
	}

	// Test WithStack on enabled logger
	l2 := logger.Info().WithStack(1)
	if l2 == nil {
		t.Fatal("WithStack on enabled logger returned nil")
	}
}

func TestFilterChaining(t *testing.T) {
	base := &mockLogger{enabled: true}
	logger := filter.New(base, slog.Info)

	// Test method chaining preserves fields and level
	l := logger.
		WithField("key1", "value1").
		WithField("key2", "value2").
		Info().
		WithField("key3", "value3")

	if l == nil {
		t.Fatal("Chained logger is nil")
	}

	// Should be enabled (Info level)
	if !l.Enabled() {
		t.Error("Info logger should be enabled")
	}
}

func TestFilterFieldTransformation(t *testing.T) {
	base := &mockLogger{enabled: true}

	// Test with field filter
	transformed := false
	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Info,
		FieldFilter: func(key string, val any) (string, any, bool) {
			transformed = true
			if key == "password" {
				return key, "[REDACTED]", true
			}
			return key, val, true
		},
	}

	l := logger.Info().WithField("password", "secret123")
	l.Print("test")

	if !transformed {
		t.Error("FieldFilter was not called")
	}
}

func TestFilterMessageFilter(t *testing.T) {
	base := &mockLogger{enabled: true}

	// Test with message filter
	filtered := false
	logger := &filter.Logger{
		Parent:    base,
		Threshold: slog.Info,
		MessageFilter: func(msg string) (string, bool) {
			filtered = true
			return "[FILTERED] " + msg, true
		},
	}

	l := logger.Info()
	l.Print("test message")

	if !filtered {
		t.Error("MessageFilter was not called")
	}
}

func TestFilterParentless(t *testing.T) {
	// Test parentless logger (only Fatal/Panic should work)
	logger := filter.NewNoop()

	// These should not panic
	logger.Debug().Print("test")
	logger.Info().Print("test")
	logger.Error().Print("test")

	// Fatal and Panic levels are enabled
	if !logger.Fatal().Enabled() {
		t.Error("Fatal should be enabled for parentless logger")
	}
	if !logger.Panic().Enabled() {
		t.Error("Panic should be enabled for parentless logger")
	}
}
