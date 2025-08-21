package discard_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = discardLogletTestCase{}

type discardLogletTestCase struct {
	method  func() slog.Logger
	name    string
	level   slog.LogLevel
	enabled bool
}

func (tc discardLogletTestCase) Name() string {
	return tc.name
}

func (tc discardLogletTestCase) Test(t *testing.T) {
	t.Helper()

	l := tc.method()
	core.AssertMustNotNil(t, l, "logger method returned nil")

	core.AssertEqual(t, tc.enabled, l.Enabled(), "Enabled() for level %s", tc.name)

	wl, enabled := l.WithEnabled()
	core.AssertMustNotNil(t, wl, "WithEnabled returned nil logger")
	core.AssertEqual(t, tc.enabled, enabled, "WithEnabled() enabled for %s", tc.name)
}

func newDiscardLogletTestCase(name string,
	method func() slog.Logger,
	level slog.LogLevel, enabled bool) discardLogletTestCase {
	return discardLogletTestCase{
		name:    name,
		method:  method,
		level:   level,
		enabled: enabled,
	}
}

func discardLogletTestCases() []discardLogletTestCase {
	logger := discard.New()
	return []discardLogletTestCase{
		newDiscardLogletTestCase("Debug", logger.Debug, slog.Debug, false),
		newDiscardLogletTestCase("Info", logger.Info, slog.Info, false),
		newDiscardLogletTestCase("Warn", logger.Warn, slog.Warn, false),
		newDiscardLogletTestCase("Error", logger.Error, slog.Error, false),
		newDiscardLogletTestCase("Fatal", logger.Fatal, slog.Fatal, true),
		newDiscardLogletTestCase("Panic", logger.Panic, slog.Panic, true),
	}
}

func TestDiscardLoglet(t *testing.T) {
	core.RunTestCases(t, discardLogletTestCases())
}

func TestDiscardWithFields(t *testing.T) {
	logger := discard.New()

	l1 := logger.WithField("key1", "value1")
	core.AssertSame(t, logger, l1, "WithField should return same instance")

	l2 := logger.WithField("", "value")
	core.AssertSame(t, logger, l2, "WithField empty key")

	fields := map[string]any{
		"key2": "value2",
		"key3": 123,
	}
	l3 := logger.WithFields(fields)
	core.AssertSame(t, logger, l3, "WithFields should return same instance")

	l4 := logger.WithFields(map[string]any{})
	core.AssertSame(t, logger, l4, "WithFields empty map")

	fieldsWithEmpty := map[string]any{
		"":     "should be removed",
		"key4": "value4",
	}
	l5 := logger.WithFields(fieldsWithEmpty)
	core.AssertSame(t, logger, l5, "WithFields should return same instance")
}

func TestDiscardWithStack(t *testing.T) {
	logger := discard.New()
	l := logger.WithStack(1)
	core.AssertNotNil(t, l, "WithStack returned nil")
}

func TestDiscardChaining(t *testing.T) {
	logger := discard.New()

	// Chain with level change should create new instance
	l1 := logger.
		WithField("key1", "value1").
		WithField("key2", "value2").
		Info()

	core.AssertNotSame(t, logger, l1, "Chain with level change should create new instance")
	core.AssertFalse(t, l1.Enabled(), "Info logger should not be enabled")

	// Branch with more fields should return same instance
	l2 := l1.
		WithField("key3", "value3").
		WithField("key4", "value4")

	core.AssertSame(t, l1, l2, "Chain with only fields should return same instance")
}

func TestDiscardLevelValidation(t *testing.T) {
	core.AssertPanic(t, func() {
		logger := discard.New()
		logger.WithLevel(slog.UndefinedLevel)
	}, nil, "invalid level panic")
}

func TestDiscardPrintMethods(t *testing.T) {
	logger := discard.New()

	// Test that non-fatal print methods don't panic
	core.AssertNoPanic(t, func() {
		logger.Debug().Print("test")
	}, "Debug Print")

	core.AssertNoPanic(t, func() {
		logger.Info().Printf("test %d", 123)
	}, "Info Printf")

	core.AssertNoPanic(t, func() {
		logger.Warn().Println("test")
	}, "Warn Println")

	core.AssertNoPanic(t, func() {
		logger.Error().Print("test")
	}, "Error Print")
}
