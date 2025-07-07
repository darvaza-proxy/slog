package testing

import (
	"testing"

	"darvaza.org/slog"
)

// TestBidirectionalOptions verifies that BidirectionalTestOptions works correctly
func TestBidirectionalOptions(t *testing.T) {
	t.Run("NilOptions", testNilOptions)
	t.Run("WithExceptions", testWithExceptions)
	t.Run("EmptyExceptions", testEmptyExceptions)
	t.Run("WithUndefinedLevel", testWithUndefinedLevel)
}

func testNilOptions(t *testing.T) {
	t.Helper()
	var opts *BidirectionalTestOptions

	// All levels should map to themselves with nil options
	levels := []slog.LogLevel{
		slog.Debug, slog.Info, slog.Warn, slog.Error, slog.Fatal, slog.Panic,
	}

	for _, level := range levels {
		// Test nil handling - should return original level
		if got := opts.ExpectedLevel(level); got != level {
			t.Errorf("ExpectedLevel(%v) = %v, want %v", level, got, level)
		}
	}
}

func testWithExceptions(t *testing.T) {
	t.Helper()
	opts := &BidirectionalTestOptions{
		AdapterOptions: AdapterOptions{
			LevelExceptions: map[slog.LogLevel]slog.LogLevel{
				slog.Warn:  slog.Info, // adapter limitation mapping
				slog.Debug: slog.Info, // no debug support
			},
		},
	}

	// Test mapped levels
	if got := opts.ExpectedLevel(slog.Warn); got != slog.Info {
		t.Errorf("ExpectedLevel(Warn) = %v, want Info", got)
	}
	if got := opts.ExpectedLevel(slog.Debug); got != slog.Info {
		t.Errorf("ExpectedLevel(Debug) = %v, want Info", got)
	}

	// Test unmapped levels remain unchanged
	if got := opts.ExpectedLevel(slog.Error); got != slog.Error {
		t.Errorf("ExpectedLevel(Error) = %v, want Error", got)
	}
	if got := opts.ExpectedLevel(slog.Info); got != slog.Info {
		t.Errorf("ExpectedLevel(Info) = %v, want Info", got)
	}
}

func testEmptyExceptions(t *testing.T) {
	t.Helper()
	opts := &BidirectionalTestOptions{
		AdapterOptions: AdapterOptions{
			LevelExceptions: map[slog.LogLevel]slog.LogLevel{},
		},
	}

	// All levels should map to themselves with empty map
	if got := opts.ExpectedLevel(slog.Warn); got != slog.Warn {
		t.Errorf("ExpectedLevel(Warn) = %v, want Warn", got)
	}
}

func testWithUndefinedLevel(t *testing.T) {
	t.Helper()
	opts := &BidirectionalTestOptions{
		AdapterOptions: AdapterOptions{
			LevelExceptions: map[slog.LogLevel]slog.LogLevel{
				slog.Warn:  slog.UndefinedLevel, // Skip Warn messages
				slog.Debug: slog.UndefinedLevel, // Skip Debug messages
			},
		},
	}

	// Test that Warn and Debug map to UndefinedLevel
	if got := opts.ExpectedLevel(slog.Warn); got != slog.UndefinedLevel {
		t.Errorf("ExpectedLevel(Warn) = %v, want UndefinedLevel", got)
	}
	if got := opts.ExpectedLevel(slog.Debug); got != slog.UndefinedLevel {
		t.Errorf("ExpectedLevel(Debug) = %v, want UndefinedLevel", got)
	}

	// Test that other levels remain unchanged
	if got := opts.ExpectedLevel(slog.Info); got != slog.Info {
		t.Errorf("ExpectedLevel(Info) = %v, want Info", got)
	}
	if got := opts.ExpectedLevel(slog.Error); got != slog.Error {
		t.Errorf("ExpectedLevel(Error) = %v, want Error", got)
	}
}

// TestBidirectionalWithOptionsIntegration tests the integration with a mock adapter
func TestBidirectionalWithOptionsIntegration(t *testing.T) {
	// Create a mock adapter that changes Warn to Info
	mockAdapter := func(backend slog.Logger) slog.Logger {
		return &levelMappingLogger{
			backend: backend,
			mapping: map[slog.LogLevel]slog.LogLevel{
				slog.Warn: slog.Info,
			},
		}
	}

	// This should pass with the correct options
	opts := &BidirectionalTestOptions{
		AdapterOptions: AdapterOptions{
			LevelExceptions: map[slog.LogLevel]slog.LogLevel{
				slog.Warn: slog.Info,
			},
		},
	}

	TestBidirectionalWithOptions(t, "MockAdapter", mockAdapter, opts)
}

// levelMappingLogger is a test logger that maps levels
type levelMappingLogger struct {
	backend slog.Logger
	mapping map[slog.LogLevel]slog.LogLevel
}

func (l *levelMappingLogger) mapLevel(level slog.LogLevel) slog.LogLevel {
	if mapped, ok := l.mapping[level]; ok {
		return mapped
	}
	return level
}

func (l *levelMappingLogger) Debug() slog.Logger { return l.WithLevel(slog.Debug) }
func (l *levelMappingLogger) Info() slog.Logger  { return l.WithLevel(slog.Info) }
func (l *levelMappingLogger) Warn() slog.Logger  { return l.WithLevel(slog.Warn) }
func (l *levelMappingLogger) Error() slog.Logger { return l.WithLevel(slog.Error) }
func (l *levelMappingLogger) Fatal() slog.Logger { return l.WithLevel(slog.Fatal) }
func (l *levelMappingLogger) Panic() slog.Logger { return l.WithLevel(slog.Panic) }

func (l *levelMappingLogger) WithLevel(level slog.LogLevel) slog.Logger {
	return l.backend.WithLevel(l.mapLevel(level))
}

func (l *levelMappingLogger) WithStack(skip int) slog.Logger {
	return &levelMappingLogger{
		backend: l.backend.WithStack(skip),
		mapping: l.mapping,
	}
}

func (l *levelMappingLogger) WithField(key string, value any) slog.Logger {
	return &levelMappingLogger{
		backend: l.backend.WithField(key, value),
		mapping: l.mapping,
	}
}

func (l *levelMappingLogger) WithFields(fields map[string]any) slog.Logger {
	return &levelMappingLogger{
		backend: l.backend.WithFields(fields),
		mapping: l.mapping,
	}
}

func (l *levelMappingLogger) Enabled() bool {
	return l.backend.Enabled()
}

func (l *levelMappingLogger) WithEnabled() (slog.Logger, bool) {
	return l, l.Enabled()
}

func (l *levelMappingLogger) Print(args ...any) {
	l.backend.Print(args...)
}

func (l *levelMappingLogger) Println(args ...any) {
	l.backend.Println(args...)
}

func (l *levelMappingLogger) Printf(format string, args ...any) {
	l.backend.Printf(format, args...)
}
