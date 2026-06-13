package zap_test

// cSpell:words zaptest

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/mock"
	slogzap "darvaza.org/slog/handlers/zap"
	slogtest "darvaza.org/slog/internal/testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = reversedLevelTestCase{}
var _ core.TestCase = reversedEnabledTestCase{}

func TestNewReversed(t *testing.T) {
	t.Run("NilParent", runTestNewReversedNilParent)
	t.Run("TypedNilParent", runTestNewReversedTypedNilParent)
	t.Run("ForeignTypedNilParent", runTestNewReversedForeignTypedNilParent)
	t.Run("Unwrap", runTestNewReversedUnwrap)
	t.Run("UnwrapWithFields", runTestNewReversedUnwrapWithFields)
	t.Run("UnwrapWithOptions", runTestNewReversedUnwrapWithOptions)
	t.Run("Forward", runTestNewReversedForward)
	t.Run("ForwardWithOptions", runTestNewReversedForwardWithOptions)
	t.Run("ForwardWithParentFields", runTestNewReversedForwardWithParentFields)
	t.Run("Levels", runTestNewReversedLevels)
	t.Run("Enabled", runTestNewReversedEnabled)
}

func runTestNewReversedNilParent(t *testing.T) {
	t.Helper()
	zl, err := slogzap.NewReversed(nil)
	core.AssertErrorIs(t, err, core.ErrInvalid, "nil parent")
	core.AssertNil(t, zl, "logger")
}

func runTestNewReversedTypedNilParent(t *testing.T) {
	t.Helper()
	var parent *slogzap.Logger
	zl, err := slogzap.NewReversed(parent)
	core.AssertErrorIs(t, err, core.ErrInvalid, "typed-nil parent")
	core.AssertNil(t, zl, "logger")
}

func runTestNewReversedForeignTypedNilParent(t *testing.T) {
	t.Helper()
	var parent *mock.Logger
	zl, err := slogzap.NewReversed(parent)
	core.AssertErrorIs(t, err, core.ErrInvalid, "foreign typed-nil parent")
	core.AssertNil(t, zl, "logger")
}

func runTestNewReversedUnwrap(t *testing.T) {
	t.Helper()
	slogLogger, err := slogzap.New(slogzap.NewDefaultConfig())
	core.AssertMustNoError(t, err, "create parent")

	parent := core.AssertMustTypeIs[*slogzap.Logger](t, slogLogger, "adaptor type")
	wrapped, _ := parent.Unwrap()

	zl, err := slogzap.NewReversed(parent)
	core.AssertMustNoError(t, err, "NewReversed")
	core.AssertSame(t, wrapped, zl, "unwrapped logger")
}

func runTestNewReversedUnwrapWithFields(t *testing.T) {
	t.Helper()
	obsCore, logs := observer.New(zapcore.DebugLevel)
	slogLogger, err := slogzap.New(slogzap.NewDefaultConfig(),
		zap.WrapCore(func(zapcore.Core) zapcore.Core { return obsCore }))
	core.AssertMustNoError(t, err, "create parent")

	parent := slogLogger.WithField("service", "api")
	zl, err := slogzap.NewReversed(parent)
	core.AssertMustNoError(t, err, "NewReversed")

	zl.Info("hello")

	entries := logs.All()
	core.AssertMustEqual(t, 1, len(entries), "entry count")
	core.AssertEqual(t, "hello", entries[0].Message, "message")
	slogtest.AssertFieldValue(t, entries[0].ContextMap(), "service", "api")
}

// newCountingHook returns a zap hook and a counter it increments on
// every entry, to verify options reach the returned logger.
func newCountingHook() (func(zapcore.Entry) error, *int) {
	count := new(int)
	hook := func(zapcore.Entry) error {
		*count++
		return nil
	}
	return hook, count
}

func runTestNewReversedUnwrapWithOptions(t *testing.T) {
	t.Helper()
	obsCore, logs := observer.New(zapcore.DebugLevel)
	slogLogger, err := slogzap.New(slogzap.NewDefaultConfig(),
		zap.WrapCore(func(zapcore.Core) zapcore.Core { return obsCore }))
	core.AssertMustNoError(t, err, "create parent")

	hook, count := newCountingHook()
	zl, err := slogzap.NewReversed(slogLogger, zap.Hooks(hook))
	core.AssertMustNoError(t, err, "NewReversed")

	zl.Info("hello")
	core.AssertEqual(t, 1, *count, "hook count")
	core.AssertEqual(t, 1, logs.Len(), "entry count")
}

func runTestNewReversedForward(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()
	zl, err := slogzap.NewReversed(recorder)
	core.AssertMustNoError(t, err, "NewReversed")

	zl.Info("forwarded", zap.String("key", "value"))

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)

	msg := messages[0]
	core.AssertEqual(t, slog.Info, msg.Level, "level")
	core.AssertEqual(t, "forwarded", msg.Message, "message")
	slogtest.AssertField(t, msg, "key", "value")
}

func runTestNewReversedForwardWithOptions(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()
	hook, count := newCountingHook()
	zl, err := slogzap.NewReversed(recorder, zap.Hooks(hook))
	core.AssertMustNoError(t, err, "NewReversed")

	zl.Info("forwarded")
	core.AssertEqual(t, 1, *count, "hook count")
	slogtest.AssertMustMessageCount(t, recorder.GetMessages(), 1)
}

func runTestNewReversedForwardWithParentFields(t *testing.T) {
	t.Helper()
	recorder := mock.NewLogger()
	parent := recorder.WithField("service", "api")
	zl, err := slogzap.NewReversed(parent)
	core.AssertMustNoError(t, err, "NewReversed")

	zl.Info("forwarded")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)
	slogtest.AssertField(t, messages[0], "service", "api")
}

// reversedLevelTestCase exercises level delegation from the reversed
// zap logger to a threshold-filtered parent.
type reversedLevelTestCase struct {
	logFunc   func(*zap.Logger, string, ...zap.Field)
	name      string
	wantCount int
	threshold slog.LogLevel
	wantLevel slog.LogLevel
}

func (tc reversedLevelTestCase) Name() string {
	return tc.name
}

func (tc reversedLevelTestCase) Test(t *testing.T) {
	t.Helper()
	recorder := mock.NewLoggerWithThreshold(tc.threshold)
	zl, err := slogzap.NewReversed(recorder)
	core.AssertMustNoError(t, err, "NewReversed")

	tc.logFunc(zl, "test message")

	messages := recorder.GetMessages()
	core.AssertMustEqual(t, tc.wantCount, len(messages), "message count")
	if tc.wantCount > 0 {
		core.AssertEqual(t, tc.wantLevel, messages[0].Level, "level")
	}
}

func newReversedLevelTestCase(
	name string, threshold slog.LogLevel,
	logFunc func(*zap.Logger, string, ...zap.Field),
	wantCount int, wantLevel slog.LogLevel,
) reversedLevelTestCase {
	return reversedLevelTestCase{
		name:      name,
		threshold: threshold,
		logFunc:   logFunc,
		wantCount: wantCount,
		wantLevel: wantLevel,
	}
}

func reversedLevelTestCases() []reversedLevelTestCase {
	return []reversedLevelTestCase{
		newReversedLevelTestCase("DebugBelowInfoThreshold", slog.Info,
			(*zap.Logger).Debug, 0, slog.UndefinedLevel),
		newReversedLevelTestCase("InfoAtInfoThreshold", slog.Info,
			(*zap.Logger).Info, 1, slog.Info),
		newReversedLevelTestCase("WarnAboveInfoThreshold", slog.Info,
			(*zap.Logger).Warn, 1, slog.Warn),
		newReversedLevelTestCase("InfoBelowErrorThreshold", slog.Error,
			(*zap.Logger).Info, 0, slog.UndefinedLevel),
		newReversedLevelTestCase("ErrorAtErrorThreshold", slog.Error,
			(*zap.Logger).Error, 1, slog.Error),
		newReversedLevelTestCase("DebugAtDebugThreshold", slog.Debug,
			(*zap.Logger).Debug, 1, slog.Debug),
	}
}

func runTestNewReversedLevels(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, reversedLevelTestCases())
}

// reversedEnabledTestCase probes the reversed logger's level enabler
// directly, including zap levels with no slog equivalent.
type reversedEnabledTestCase struct {
	name      string
	threshold slog.LogLevel
	zapLevel  zapcore.Level
	want      bool
}

func (tc reversedEnabledTestCase) Name() string {
	return tc.name
}

func (tc reversedEnabledTestCase) Test(t *testing.T) {
	t.Helper()
	recorder := mock.NewLoggerWithThreshold(tc.threshold)
	zl, err := slogzap.NewReversed(recorder)
	core.AssertMustNoError(t, err, "NewReversed")

	core.AssertEqual(t, tc.want, zl.Core().Enabled(tc.zapLevel), "enabled")
}

func newReversedEnabledTestCase(
	name string, threshold slog.LogLevel, zapLevel zapcore.Level, want bool,
) reversedEnabledTestCase {
	return reversedEnabledTestCase{
		name:      name,
		threshold: threshold,
		zapLevel:  zapLevel,
		want:      want,
	}
}

func reversedEnabledTestCases() []reversedEnabledTestCase {
	return []reversedEnabledTestCase{
		newReversedEnabledTestCase("InvalidLevel", slog.Debug,
			zapcore.InvalidLevel, false),
		newReversedEnabledTestCase("OutOfRangeLevel", slog.Debug,
			zapcore.Level(99), false),
		newReversedEnabledTestCase("BelowRangeLevel", slog.Debug,
			zapcore.Level(-2), false),
		newReversedEnabledTestCase("DPanicAtErrorThreshold", slog.Error,
			zapcore.DPanicLevel, true),
		newReversedEnabledTestCase("DPanicBelowFatalThreshold", slog.Fatal,
			zapcore.DPanicLevel, false),
		newReversedEnabledTestCase("DebugBelowInfoThreshold", slog.Info,
			zapcore.DebugLevel, false),
		newReversedEnabledTestCase("InfoAtInfoThreshold", slog.Info,
			zapcore.InfoLevel, true),
	}
}

func runTestNewReversedEnabled(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, reversedEnabledTestCases())
}
