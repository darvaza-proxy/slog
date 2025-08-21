package filter_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
)

func TestLevel(t *testing.T) {
	// Test nil receiver for LogEntry.Level()
	var nilEntry *filter.LogEntry
	core.AssertEqual(t, slog.UndefinedLevel, nilEntry.Level(), "nil log entry level")

	// Test normal functionality
	parent := &mockLogger{enabled: true}
	logger := filter.New(parent, slog.Info)

	infoEntry := logger.Info()
	filterEntry := core.AssertMustTypeIs[*filter.LogEntry](t, infoEntry, "info entry type")
	core.AssertEqual(t, slog.Info, filterEntry.Level(), "info level")

	errorEntry := logger.Error()
	filterError := core.AssertMustTypeIs[*filter.LogEntry](t, errorEntry, "error entry type")
	core.AssertEqual(t, slog.Error, filterError.Level(), "error level")
}

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = filterLogletTestCase{}

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

type filterLogletTestCase struct {
	method  func() slog.Logger
	name    string
	level   slog.LogLevel
	enabled bool
}

func (tc filterLogletTestCase) Name() string {
	return tc.name
}

func (tc filterLogletTestCase) Test(t *testing.T) {
	t.Helper()

	l := tc.method()
	core.AssertMustNotNil(t, l, "logger method returned nil")

	core.AssertEqual(t, tc.enabled, l.Enabled(), "Enabled() for level %s", tc.name)
}

func newFilterLogletTestCase(name string,
	method func() slog.Logger,
	level slog.LogLevel, enabled bool) filterLogletTestCase {
	return filterLogletTestCase{
		name:    name,
		method:  method,
		level:   level,
		enabled: enabled,
	}
}

func filterLogletTestCases() []filterLogletTestCase {
	base := &mockLogger{enabled: true}
	logger := filter.New(base, slog.Info)
	return []filterLogletTestCase{
		newFilterLogletTestCase("Debug", logger.Debug, slog.Debug, false),
		newFilterLogletTestCase("Info", logger.Info, slog.Info, true),
		newFilterLogletTestCase("Warn", logger.Warn, slog.Warn, true),
		newFilterLogletTestCase("Error", logger.Error, slog.Error, true),
		newFilterLogletTestCase("Fatal", logger.Fatal, slog.Fatal, true),
		newFilterLogletTestCase("Panic", logger.Panic, slog.Panic, true),
	}
}

func TestFilterLoglet(t *testing.T) {
	core.RunTestCases(t, filterLogletTestCases())
}

func TestFilterWithFields(t *testing.T) {
	base := &mockLogger{enabled: true}
	logger := filter.New(base, slog.Info)

	l1 := logger.WithField("root", "value")
	core.AssertMustNotNil(t, l1, "WithField returned nil")

	l2 := logger.Info().WithField("key1", "value1")
	core.AssertMustNotNil(t, l2, "WithField on enabled logger returned nil")

	fields := map[string]any{
		"key2": "value2",
		"key3": 123,
	}
	l3 := logger.Info().WithFields(fields)
	core.AssertNotNil(t, l3, "WithFields returned nil")
}

func TestFilterWithStack(t *testing.T) {
	base := &mockLogger{enabled: true}
	logger := filter.New(base, slog.Info)

	l1 := logger.WithStack(1)
	core.AssertMustNotNil(t, l1, "WithStack on root returned nil")

	l2 := logger.Info().WithStack(1)
	core.AssertNotNil(t, l2, "WithStack on enabled logger returned nil")
}

func TestFilterChaining(t *testing.T) {
	base := &mockLogger{enabled: true}
	logger := filter.New(base, slog.Info)

	l := logger.
		WithField("key1", "value1").
		WithField("key2", "value2").
		Info().
		WithField("key3", "value3")

	core.AssertMustNotNil(t, l, "Chained logger is nil")

	core.AssertTrue(t, l.Enabled(), "Info logger enabled")
}

func TestFilterFieldTransformation(t *testing.T) {
	base := &mockLogger{enabled: true}

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

	core.AssertTrue(t, transformed, "FieldFilter was not called")
}

func TestFilterMessageFilter(t *testing.T) {
	base := &mockLogger{enabled: true}

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

	core.AssertTrue(t, filtered, "MessageFilter was not called")
}

func TestFilterParentless(t *testing.T) {
	logger := filter.NewNoop()

	logger.Debug().Print("test")
	logger.Info().Print("test")
	logger.Error().Print("test")

	core.AssertTrue(t, logger.Fatal().Enabled(), "Fatal parentless enabled")
	core.AssertTrue(t, logger.Panic().Enabled(), "Panic parentless enabled")
}
