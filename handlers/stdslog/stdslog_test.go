package stdslog_test

import (
	"context"
	"errors"
	"io"
	stdslog "log/slog"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/mock"
	slogstdslog "darvaza.org/slog/handlers/stdslog"
	slogtest "darvaza.org/slog/internal/testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = roundTripLevelTestCase{}
var _ core.TestCase = roundTripValueTestCase{}
var _ core.TestCase = fieldCollectionTestCase{}

// newTestLogger creates an adaptor over a stdlib handler discarding all
// output, enabled at every level.
func newTestLogger() slog.Logger {
	h := stdslog.NewTextHandler(io.Discard, &stdslog.HandlerOptions{
		Level: stdslog.LevelDebug,
	})
	return slogstdslog.NewWithHandler(h)
}

// newSandwichLogger builds the full round trip: darvaza slog in front,
// stdlib log/slog in the middle, the given recorder at the back.
func newSandwichLogger(recorder slog.Logger) slog.Logger {
	return slogstdslog.New(stdslog.New(slogstdslog.NewHandler(recorder)))
}

func TestCompliance(t *testing.T) {
	ct := slogtest.ComplianceTest{
		FactoryOptions: slogtest.FactoryOptions{
			NewLogger:             newTestLogger,
			NewLoggerWithRecorder: newSandwichLogger,
		},
	}
	ct.Run(t)
}

// TestComplianceThresholds reruns the compliance suite with the
// recorder filtering at each threshold, mirroring the filter handler's
// threshold compliance runs: the sandwich must stay compliant when the
// backend discards entries.
func TestComplianceThresholds(t *testing.T) {
	thresholds := []struct {
		name      string
		threshold slog.LogLevel
	}{
		{"Info", slog.Info},
		{"Warn", slog.Warn},
		{"Error", slog.Error},
	}

	for _, tc := range thresholds {
		t.Run(tc.name, func(t *testing.T) {
			runTestThresholdCompliance(t, tc.threshold)
		})
	}
}

func runTestThresholdCompliance(t *testing.T, threshold slog.LogLevel) {
	t.Helper()

	ct := slogtest.ComplianceTest{
		FactoryOptions: slogtest.FactoryOptions{
			NewLogger: func() slog.Logger {
				recorder := mock.NewLoggerWithThreshold(threshold)
				return newSandwichLogger(recorder)
			},
			NewLoggerWithRecorder: newSandwichLogger,
		},
	}
	ct.Run(t)
}

func TestStress(t *testing.T) {
	suite := slogtest.StressTestSuite{
		NewLogger:             newTestLogger,
		NewLoggerWithRecorder: newSandwichLogger,
	}
	suite.Run(t)
}

// roundTripLevelTestCase pins level preservation through the stdlib
// sandwich; unlike logr, every level survives, Warn included.
type roundTripLevelTestCase struct {
	name  string
	level slog.LogLevel
}

// Name returns the test case name.
func (tc roundTripLevelTestCase) Name() string {
	return tc.name
}

// Test validates the recorded level and message.
func (tc roundTripLevelTestCase) Test(t *testing.T) {
	t.Helper()

	recorder := mock.NewLogger()
	adapter := newSandwichLogger(recorder)

	adapter.WithLevel(tc.level).Print("round trip")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)
	slogtest.AssertMustMessage(t, messages[0], tc.level, "round trip")
}

// newRoundTripLevelTestCase creates a new round-trip level test case.
func newRoundTripLevelTestCase(name string,
	level slog.LogLevel) roundTripLevelTestCase {
	return roundTripLevelTestCase{
		name:  name,
		level: level,
	}
}

func TestRoundTripLevels(t *testing.T) {
	testCases := []roundTripLevelTestCase{
		newRoundTripLevelTestCase("debug", slog.Debug),
		newRoundTripLevelTestCase("info", slog.Info),
		newRoundTripLevelTestCase("warn", slog.Warn),
		newRoundTripLevelTestCase("error", slog.Error),
	}

	core.RunTestCases(t, testCases)
}

// TestRoundTripFields pins field preservation through the sandwich.
// Values normalise through stdlib slog.Value on the way: int arrives
// as int64. That is inherent to crossing log/slog, not a defect.
func TestRoundTripFields(t *testing.T) {
	recorder := mock.NewLogger()
	adapter := newSandwichLogger(recorder)

	adapter.Info().
		WithField("string", "value").
		WithField("int", 42).
		WithField("bool", true).
		WithField("float", 3.14).
		Print("fields test")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)

	msg := messages[0]
	slogtest.AssertMustMessage(t, msg, slog.Info, "fields test")
	slogtest.AssertField(t, msg, "string", "value")
	slogtest.AssertField(t, msg, "int", int64(42))
	slogtest.AssertField(t, msg, "bool", true)
	slogtest.AssertField(t, msg, "float", 3.14)
}

// roundTripValueTestCase pins how a field value mutates through the
// sandwich: stdlib slog.Value re-types some Go primitives on the way
// in; everything it has no kind for passes through untouched. These
// rows are the normalisation contract local tests assert instead of
// running the shared bidirectional suite, whose WithFields subtest
// expects the original values.
type roundTripValueTestCase struct {
	value    any
	expected any
	name     string
}

// Name returns the test case name.
func (tc roundTripValueTestCase) Name() string {
	return tc.name
}

// Test validates the recorded value, exact type included.
func (tc roundTripValueTestCase) Test(t *testing.T) {
	t.Helper()

	recorder := mock.NewLogger()
	adapter := newSandwichLogger(recorder)

	adapter.Info().WithField("value", tc.value).Print("normalise")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)

	msg := messages[0]
	slogtest.AssertMustMessage(t, msg, slog.Info, "normalise")
	core.AssertEqual(t, 1, len(msg.Fields), "field count")
	slogtest.AssertField(t, msg, "value", tc.expected)
}

// newRoundTripValueTestCase creates a new value normalisation test case.
func newRoundTripValueTestCase(name string, value,
	expected any) roundTripValueTestCase {
	return roundTripValueTestCase{
		name:     name,
		value:    value,
		expected: expected,
	}
}

// newRoundTripValueTestCaseSame creates a row for a value the sandwich
// passes through untouched.
func newRoundTripValueTestCaseSame(name string,
	value any) roundTripValueTestCase {
	return newRoundTripValueTestCase(name, value, value)
}

func roundTripValueTestCases() []roundTripValueTestCase {
	type custom struct {
		Name  string
		Value int
	}

	err := errors.New("kept as error")
	when := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)

	return []roundTripValueTestCase{
		newRoundTripValueTestCaseSame("string", "value"),
		newRoundTripValueTestCase("int", 42, int64(42)),
		newRoundTripValueTestCase("int8", int8(7), int64(7)),
		newRoundTripValueTestCase("uint", uint(42), uint64(42)),
		newRoundTripValueTestCase("float32", float32(1.5), float64(1.5)),
		newRoundTripValueTestCaseSame("float64", 3.14),
		newRoundTripValueTestCaseSame("bool", true),
		newRoundTripValueTestCaseSame("duration", time.Second),
		newRoundTripValueTestCaseSame("time", when),
		newRoundTripValueTestCaseSame("error", err),
		newRoundTripValueTestCaseSame("struct", custom{Name: "test", Value: 123}),
		newRoundTripValueTestCaseSame("slice", core.S("a", "b", "c")),
		newRoundTripValueTestCaseSame("map", map[string]int{"one": 1, "two": 2}),
		newRoundTripValueTestCaseSame("nil", nil),
	}
}

func TestRoundTripValues(t *testing.T) {
	core.RunTestCases(t, roundTripValueTestCases())
}

// TestRoundTripFieldChaining mirrors the bidirectional suite's
// FieldChaining subtest through the sandwich: branches keep the base
// fields plus their own, and nothing leaks between siblings.
func TestRoundTripFieldChaining(t *testing.T) {
	recorder := mock.NewLogger()
	adapter := newSandwichLogger(recorder)

	base := adapter.WithField("app", "test").WithField("version", "1.0")
	userLogger := base.WithField("component", "user")
	adminLogger := base.WithField("component", "admin")

	userLogger.Info().Print("user action")
	adminLogger.Info().Print("admin action")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 2)

	for i, component := range core.S("user", "admin") {
		msg := messages[i]
		slogtest.AssertMustMessage(t, msg, slog.Info, component+" action")
		core.AssertEqual(t, 3, len(msg.Fields), "field count")
		slogtest.AssertField(t, msg, "app", "test")
		slogtest.AssertField(t, msg, "version", "1.0")
		slogtest.AssertField(t, msg, "component", component)
	}
}

// fieldCollectionTestCase pins the field-collection short-circuit:
// once a level is set, collection is bound to Enabled(). A disabled
// leveled entry skips collection, so WithField returns the same
// logger; an unleveled entry collects speculatively; the
// always-enabled Fatal and Panic levels collect regardless of the
// backend threshold. The backend filters at Warn throughout.
type fieldCollectionTestCase struct {
	name    string
	level   slog.LogLevel
	collect bool
}

// Name returns the test case name.
func (tc fieldCollectionTestCase) Name() string {
	return tc.name
}

// Test asserts whether WithField collected (a new instance) or
// short-circuited (the same instance).
func (tc fieldCollectionTestCase) Test(t *testing.T) {
	t.Helper()

	h := stdslog.NewTextHandler(io.Discard,
		&stdslog.HandlerOptions{Level: stdslog.LevelWarn})
	logger := slogstdslog.NewWithHandler(h)
	if tc.level != slog.UndefinedLevel {
		logger = logger.WithLevel(tc.level)
	}

	got := logger.WithField("k", "v")
	if tc.collect {
		core.AssertNotSame(t, logger, got, "collected")
	} else {
		core.AssertSame(t, logger, got, "skipped")
	}
}

// newFieldCollectionTestCase creates a leveled field-collection case.
func newFieldCollectionTestCase(name string, level slog.LogLevel,
	collect bool) fieldCollectionTestCase {
	return fieldCollectionTestCase{
		name:    name,
		level:   level,
		collect: collect,
	}
}

// newFieldCollectionTestCaseUnleveled creates an unleveled case; an
// entry with no level set always collects speculatively.
func newFieldCollectionTestCaseUnleveled(name string) fieldCollectionTestCase {
	return fieldCollectionTestCase{
		name:    name,
		level:   slog.UndefinedLevel,
		collect: true,
	}
}

func fieldCollectionTestCases() []fieldCollectionTestCase {
	return []fieldCollectionTestCase{
		newFieldCollectionTestCaseUnleveled("unleveled collects speculatively"),
		newFieldCollectionTestCase("debug disabled skips", slog.Debug, false),
		newFieldCollectionTestCase("info disabled skips", slog.Info, false),
		newFieldCollectionTestCase("warn enabled collects", slog.Warn, true),
		newFieldCollectionTestCase("error enabled collects", slog.Error, true),
		newFieldCollectionTestCase("fatal always collects", slog.Fatal, true),
		newFieldCollectionTestCase("panic always collects", slog.Panic, true),
	}
}

func TestFieldCollection(t *testing.T) {
	core.RunTestCases(t, fieldCollectionTestCases())
}

// TestNewSLoggerGroups exercises the injection artefact end to end:
// stdlib groups become dot-separated key prefixes.
func TestNewSLoggerGroups(t *testing.T) {
	recorder := mock.NewLogger()
	logger := slogstdslog.NewSLogger(recorder)
	core.AssertMustNotNil(t, logger, "NewSLogger")

	logger.WithGroup("req").Info("hello", "id", 42)

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)
	slogtest.AssertMustMessage(t, messages[0], slog.Info, "hello")
	slogtest.AssertField(t, messages[0], "req.id", int64(42))
}

// TestHandlerEnabled pins the Enabled delegation against a
// level-filtering backend: the threshold is the backend's decision.
func TestHandlerEnabled(t *testing.T) {
	recorder := mock.NewLoggerWithThreshold(slog.Warn)
	h := slogstdslog.NewHandler(recorder)
	ctx := context.Background()

	core.AssertTrue(t, h.Enabled(ctx, stdslog.LevelError), "error enabled")
	core.AssertTrue(t, h.Enabled(ctx, stdslog.LevelWarn), "warn enabled")
	core.AssertFalse(t, h.Enabled(ctx, stdslog.LevelInfo), "info disabled")
	core.AssertFalse(t, h.Enabled(ctx, stdslog.LevelDebug), "debug disabled")
}

// TestNewHandlerNil pins the nil constructors.
func TestNewHandlerNil(t *testing.T) {
	core.AssertNil(t, slogstdslog.NewHandler(nil), "NewHandler nil")
	core.AssertNil(t, slogstdslog.NewSLogger(nil), "NewSLogger nil")
	core.AssertNil(t, slogstdslog.NewWithHandler(nil), "NewWithHandler nil")
}
