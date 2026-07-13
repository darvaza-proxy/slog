package stdslog_test

import (
	"bytes"
	"io"
	stdslog "log/slog"
	"os"
	"os/exec"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/mock"
	slogstdslog "darvaza.org/slog/handlers/stdslog"
	slogtest "darvaza.org/slog/internal/testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = loggerEnabledTestCase{}

// newTestLoggerAt creates an adaptor over a stdlib handler discarding
// all output, enabled at the given stdlib level and above.
func newTestLoggerAt(level stdslog.Level) slog.Logger {
	h := stdslog.NewTextHandler(io.Discard, &stdslog.HandlerOptions{
		Level: level,
	})
	return slogstdslog.NewWithHandler(h)
}

func TestLoggerLevel(t *testing.T) {
	var nilLogger *slogstdslog.Logger
	core.AssertEqual(t, slog.UndefinedLevel, nilLogger.Level(), "nil receiver")

	logger := core.AssertMustTypeIs[*slogstdslog.Logger](t, newTestLogger(),
		"logger")
	core.AssertEqual(t, slog.UndefinedLevel, logger.Level(), "default")

	debug := core.AssertMustTypeIs[*slogstdslog.Logger](t, logger.Debug(),
		"debug")
	core.AssertEqual(t, slog.Debug, debug.Level(), "debug level")
}

// TestLoggerEnabled pins the divergence documented on Enabled: a logger
// with no level set reports disabled instead of panicking.
func TestLoggerEnabled(t *testing.T) {
	var nilLogger *slogstdslog.Logger
	core.AssertFalse(t, nilLogger.Enabled(), "nil receiver")

	logger := newTestLogger()
	core.AssertFalse(t, logger.Enabled(), "no level set")

	l1, ok := logger.WithEnabled()
	core.AssertSame(t, logger, l1, "WithEnabled logger")
	core.AssertFalse(t, ok, "WithEnabled disabled")

	info := logger.Info()
	core.AssertTrue(t, info.Enabled(), "info enabled")

	l2, ok := info.WithEnabled()
	core.AssertSame(t, info, l2, "WithEnabled info logger")
	core.AssertTrue(t, ok, "WithEnabled enabled")
}

// loggerEnabledTestCase pins Enabled delegation to a backend handler
// filtering at LevelWarn.
type loggerEnabledTestCase struct {
	name    string
	level   slog.LogLevel
	enabled bool
}

// Name returns the test case name.
func (tc loggerEnabledTestCase) Name() string {
	return tc.name
}

// Test validates the Enabled result against the Warn-filtering backend.
func (tc loggerEnabledTestCase) Test(t *testing.T) {
	t.Helper()

	logger := newTestLoggerAt(stdslog.LevelWarn)
	core.AssertEqual(t, tc.enabled, logger.WithLevel(tc.level).Enabled(),
		"enabled")
}

// newLoggerEnabledTestCase creates a new threshold test case.
func newLoggerEnabledTestCase(name string, level slog.LogLevel,
	enabled bool) loggerEnabledTestCase {
	return loggerEnabledTestCase{
		name:    name,
		level:   level,
		enabled: enabled,
	}
}

func TestLoggerEnabledThreshold(t *testing.T) {
	testCases := []loggerEnabledTestCase{
		newLoggerEnabledTestCase("debug", slog.Debug, false),
		newLoggerEnabledTestCase("info", slog.Info, false),
		newLoggerEnabledTestCase("warn", slog.Warn, true),
		newLoggerEnabledTestCase("error", slog.Error, true),
		newLoggerEnabledTestCase("fatal", slog.Fatal, true),
		newLoggerEnabledTestCase("panic", slog.Panic, true),
	}

	core.RunTestCases(t, testCases)
}

// TestLoggerEnabledTerminal pins the terminal contract on Enabled:
// Fatal and Panic report enabled even when the backend filter rejects
// everything; other levels follow the filter.
func TestLoggerEnabledTerminal(t *testing.T) {
	logger := newTestLoggerAt(stdslog.LevelError + 16)

	core.AssertFalse(t, logger.Error().Enabled(), "error disabled")
	core.AssertTrue(t, logger.Fatal().Enabled(), "fatal enabled")
	core.AssertTrue(t, logger.Panic().Enabled(), "panic enabled")
}

func TestWithLevelSameLevel(t *testing.T) {
	logger := newTestLogger().Info()
	core.AssertSame(t, logger, logger.WithLevel(slog.Info), "same level")
	core.AssertNotSame(t, logger, logger.WithLevel(slog.Debug), "level change")
}

func TestWithLevelInvalid(t *testing.T) {
	logger := newTestLogger()

	core.AssertPanic(t, func() {
		logger.WithLevel(slog.UndefinedLevel)
	}, nil, "undefined level")

	core.AssertPanic(t, func() {
		logger.WithLevel(slog.LogLevel(-1))
	}, nil, "negative level")

	// Out-of-range levels pass WithLevel and trip in Enabled once the
	// entry is about to be delivered.
	core.AssertPanic(t, func() {
		logger.WithLevel(slog.LogLevel(42)).Print("x")
	}, nil, "out of range")

	// A fully-disabled backend cannot swallow the diagnostic: the
	// Panic entry it rides on is terminal, hence always enabled.
	disabled := newTestLoggerAt(stdslog.LevelError + 16)
	core.AssertPanic(t, func() {
		disabled.WithLevel(slog.LogLevel(42)).Print("x")
	}, nil, "out of range, disabled backend")
}

func TestWithFieldIdentity(t *testing.T) {
	logger := newTestLogger()

	core.AssertSame(t, logger, logger.WithField("", "value"), "empty label")
	core.AssertSame(t, logger, logger.WithFields(nil), "nil map")
	core.AssertSame(t, logger, logger.WithFields(map[string]any{}),
		"empty map")
	core.AssertNotSame(t, logger, logger.WithField("k", "v"), "real field")
}

func TestWithFieldsInputUntouched(t *testing.T) {
	recorder := mock.NewLogger()
	adapter := newSandwichLogger(recorder)

	fields := map[string]any{"a": "1", "b": "2"}
	adapter.Info().WithFields(fields).Print("untouched")

	core.AssertEqual(t, 2, len(fields), "input length")
	slogtest.AssertFieldValue(t, fields, "a", "1")
	slogtest.AssertFieldValue(t, fields, "b", "2")
}

func TestPrintFormatting(t *testing.T) {
	recorder := mock.NewLogger()
	adapter := newSandwichLogger(recorder)

	adapter.Info().Print("a", "b")
	adapter.Info().Println("a", "b")
	adapter.Info().Printf("n=%d", 42)

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 3)
	slogtest.AssertMustMessage(t, messages[0], slog.Info, "ab")
	// Println's trailing newline is trimmed before delivery.
	slogtest.AssertMustMessage(t, messages[1], slog.Info, "a b")
	slogtest.AssertMustMessage(t, messages[2], slog.Info, "n=42")
}

// TestWithStack pins the stack crossing the sandwich as a plain "stack"
// field, and its absence without WithStack.
func TestWithStack(t *testing.T) {
	recorder := mock.NewLogger()
	adapter := newSandwichLogger(recorder)

	adapter.Info().WithStack(0).Print("with stack")
	adapter.Info().Print("without stack")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 2)

	stack, ok := messages[0].Fields["stack"].(string)
	core.AssertTrue(t, ok, "stack field string")
	core.AssertTrue(t, stack != "", "stack field non-empty")
	slogtest.AssertNoField(t, messages[1], "stack")
}

// TestWithStackIdentity pins WithStack's collection short-circuit: a
// nil receiver and a disabled leveled logger return the receiver
// unchanged instead of allocating a stack-carrying copy. The enabled
// leg is covered by TestWithStack.
func TestWithStackIdentity(t *testing.T) {
	var nilLogger *slogstdslog.Logger
	core.AssertNil(t, nilLogger.WithStack(0), "nil receiver")

	disabled := newTestLoggerAt(stdslog.LevelWarn).Info()
	core.AssertSame(t, disabled, disabled.WithStack(0), "disabled skip")
}

func TestPanicPrint(t *testing.T) {
	logger := newTestLogger()

	core.AssertPanic(t, func() {
		logger.Panic().Print("boom")
	}, "boom", "panic print")

	// An empty message still panics, with a nil payload.
	core.AssertPanic(t, func() {
		logger.Panic().Print()
	}, nil, "panic empty message")
}

// TestPanicDisabledBackend pins the terminal contract on Print: a
// Panic entry still delivers its record — bypassing the backend
// filter — and panics, even when the backend rejects everything.
func TestPanicDisabledBackend(t *testing.T) {
	var buf bytes.Buffer
	h := stdslog.NewTextHandler(&buf, &stdslog.HandlerOptions{
		Level: stdslog.LevelError + 16,
	})
	logger := slogstdslog.NewWithHandler(h)

	core.AssertPanic(t, func() {
		logger.Panic().Print("boom")
	}, "boom", "panic print")
	core.AssertContains(t, buf.String(), "boom", "record delivered")
}

// TestPanicPrintTrimmedPayload pins that the Panic payload carries the
// same trimmed message as the delivered record: a padded message must
// not reach the panic value untrimmed. AssertPanic's substring match
// would miss this, so it inspects the PanicError payload exactly. A
// whitespace-only message trims to empty and takes the nil-payload
// branch.
func TestPanicPrintTrimmedPayload(t *testing.T) {
	padded := panicPrintPayload(t, "  boom  ")
	core.AssertEqual(t, "boom", errorString(padded.Recovered()),
		"padded payload trimmed")

	blank := panicPrintPayload(t, "   ")
	core.AssertNil(t, blank.Recovered(), "whitespace-only payload nil")
}

// panicPrintPayload runs a Panic-level Print and returns the recovered
// PanicError.
func panicPrintPayload(t *testing.T, msg string) (pe *core.PanicError) {
	t.Helper()
	logger := newTestLogger()
	defer func() {
		pe = core.AssertMustTypeIs[*core.PanicError](t, recover(),
			"panic error")
	}()
	logger.Panic().Print(msg)
	return nil
}

// errorString renders a PanicError payload as a string, or "" when the
// payload is not an error.
func errorString(payload any) string {
	if err, ok := payload.(error); ok {
		return err.Error()
	}
	return ""
}

// fatalExitEnv marks the re-executed child in TestFatalExit.
const fatalExitEnv = "SLOG_STDSLOG_TEST_FATAL_EXIT"

// TestFatalExit re-executes the test binary so the child process can
// really exit: Fatal must deliver the record — bypassing the backend
// filter — and then exit with status 1.
func TestFatalExit(t *testing.T) {
	if os.Getenv(fatalExitEnv) == "1" {
		runFatalExitChild()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestFatalExit$")
	cmd.Env = append(os.Environ(), fatalExitEnv+"=1")
	out, err := cmd.Output()

	exitErr := core.AssertMustTypeIs[*exec.ExitError](t, err, "exit error")
	core.AssertEqual(t, 1, exitErr.ExitCode(), "exit status")
	core.AssertContains(t, string(out), "fatal boom", "record delivered")
}

// runFatalExitChild exercises Fatal in the re-executed child. The
// backend filter rejects everything; delivery must happen anyway,
// then exit(1). The fallback exit code marks a broken contract.
func runFatalExitChild() {
	h := stdslog.NewTextHandler(os.Stdout, &stdslog.HandlerOptions{
		Level: stdslog.LevelError + 16,
	})
	logger := slogstdslog.NewWithHandler(h)
	logger.Fatal().Print("fatal boom")

	// unreachable while Fatal honours the terminal contract
	// revive:disable:deep-exit
	os.Exit(3)
	// revive:enable:deep-exit
}

// TestAttrsSortedOrder pins field emission in sorted key order,
// regardless of attachment order, observed through a real TextHandler.
func TestAttrsSortedOrder(t *testing.T) {
	var buf bytes.Buffer
	logger := slogstdslog.NewWithHandler(stdslog.NewTextHandler(&buf, nil))

	logger.Info().
		WithField("zebra", 1).
		WithField("alpha", 2).
		WithField("mid", 3).
		Print("sorted")

	core.AssertContains(t, buf.String(), "alpha=2 mid=3 zebra=1",
		"sorted fields")
}

func TestLoggerUnwrap(t *testing.T) {
	h := stdslog.NewTextHandler(io.Discard, nil)
	logger := core.AssertMustTypeIs[*slogstdslog.Logger](t,
		slogstdslog.NewWithHandler(h), "logger")
	core.AssertSame(t, h, logger.Unwrap(), "unwrap")
}

// TestNewDefault pins New(nil) falling back to the stdlib default
// logger, which enables Info and above.
func TestNewDefault(t *testing.T) {
	logger := slogstdslog.New(nil)
	core.AssertMustNotNil(t, logger, "New(nil)")
	core.AssertTrue(t, logger.Info().Enabled(), "info enabled")
	core.AssertFalse(t, logger.Debug().Enabled(), "debug disabled")
}
