package stdslog_test

import (
	stdslog "log/slog"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
	slogstdslog "darvaza.org/slog/handlers/stdslog"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = mapFromSLogLevelTestCase{}
var _ core.TestCase = mapToSLogLevelTestCase{}
var _ core.TestCase = appendSLogAttrTestCase{}

// testLogValuer resolves to a fixed string, exercising slog.LogValuer
// resolution in AppendSLogAttr.
type testLogValuer struct{}

func (testLogValuer) LogValue() stdslog.Value {
	return stdslog.StringValue("resolved")
}

// mapFromSLogLevelTestCase tests stdlib log/slog to slog level mapping.
type mapFromSLogLevelTestCase struct {
	name     string
	level    stdslog.Level
	expected slog.LogLevel
}

// Name returns the test case name.
func (tc mapFromSLogLevelTestCase) Name() string {
	return tc.name
}

// Test validates the level mapping.
func (tc mapFromSLogLevelTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.expected, slogstdslog.MapFromSLogLevel(tc.level),
		"level")
}

// newMapFromSLogLevelTestCase creates a new level mapping test case.
func newMapFromSLogLevelTestCase(name string, level stdslog.Level,
	expected slog.LogLevel) mapFromSLogLevelTestCase {
	return mapFromSLogLevelTestCase{
		name:     name,
		level:    level,
		expected: expected,
	}
}

func TestMapFromSLogLevel(t *testing.T) {
	testCases := []mapFromSLogLevelTestCase{
		newMapFromSLogLevelTestCase("debug", stdslog.LevelDebug, slog.Debug),
		newMapFromSLogLevelTestCase("info", stdslog.LevelInfo, slog.Info),
		newMapFromSLogLevelTestCase("warn", stdslog.LevelWarn, slog.Warn),
		newMapFromSLogLevelTestCase("error", stdslog.LevelError, slog.Error),
		newMapFromSLogLevelTestCase("below debug", stdslog.LevelDebug-4,
			slog.Debug),
		newMapFromSLogLevelTestCase("between debug and info",
			stdslog.LevelDebug+1, slog.Debug),
		newMapFromSLogLevelTestCase("between info and warn",
			stdslog.LevelInfo+1, slog.Info),
		newMapFromSLogLevelTestCase("between warn and error",
			stdslog.LevelWarn+1, slog.Warn),
		newMapFromSLogLevelTestCase("above error", stdslog.LevelError+4,
			slog.Error),
	}

	core.RunTestCases(t, testCases)
}

// mapToSLogLevelTestCase tests slog to stdlib log/slog level mapping.
type mapToSLogLevelTestCase struct {
	name     string
	expected stdslog.Level
	level    slog.LogLevel
	wantOK   bool
}

// Name returns the test case name.
func (tc mapToSLogLevelTestCase) Name() string {
	return tc.name
}

// Test validates the level mapping.
func (tc mapToSLogLevelTestCase) Test(t *testing.T) {
	t.Helper()
	got, ok := slogstdslog.MapToSLogLevel(tc.level)
	core.AssertEqual(t, tc.wantOK, ok, "ok")
	if tc.wantOK {
		core.AssertEqual(t, tc.expected, got, "level")
	}
}

// newMapToSLogLevelTestCase creates a valid-level mapping test case.
func newMapToSLogLevelTestCase(name string, level slog.LogLevel,
	expected stdslog.Level) mapToSLogLevelTestCase {
	return mapToSLogLevelTestCase{
		name:     name,
		level:    level,
		expected: expected,
		wantOK:   true,
	}
}

// newMapToSLogLevelTestCaseInvalid creates a rejected-level mapping test
// case; the returned level is unspecified and not asserted.
func newMapToSLogLevelTestCaseInvalid(name string,
	level slog.LogLevel) mapToSLogLevelTestCase {
	return mapToSLogLevelTestCase{
		name:   name,
		level:  level,
		wantOK: false,
	}
}

func TestMapToSLogLevel(t *testing.T) {
	testCases := []mapToSLogLevelTestCase{
		newMapToSLogLevelTestCase("debug", slog.Debug, stdslog.LevelDebug),
		newMapToSLogLevelTestCase("info", slog.Info, stdslog.LevelInfo),
		newMapToSLogLevelTestCase("warn", slog.Warn, stdslog.LevelWarn),
		newMapToSLogLevelTestCase("error", slog.Error, stdslog.LevelError),
		newMapToSLogLevelTestCase("fatal", slog.Fatal, stdslog.LevelError+4),
		newMapToSLogLevelTestCase("panic", slog.Panic, stdslog.LevelError+8),
		newMapToSLogLevelTestCaseInvalid("undefined", slog.UndefinedLevel),
		newMapToSLogLevelTestCaseInvalid("below range", slog.LogLevel(-1)),
		newMapToSLogLevelTestCaseInvalid("above range", slog.LogLevel(42)),
	}

	core.RunTestCases(t, testCases)
}

// appendSLogAttrTestCase tests attribute flattening into key/value pairs.
type appendSLogAttrTestCase struct {
	attr     stdslog.Attr
	prefix   string
	name     string
	expected []any
}

// Name returns the test case name.
func (tc appendSLogAttrTestCase) Name() string {
	return tc.name
}

// Test validates the flattened key/value pairs.
func (tc appendSLogAttrTestCase) Test(t *testing.T) {
	t.Helper()
	got := slogstdslog.AppendSLogAttr(nil, tc.prefix, tc.attr)
	core.AssertSliceEqual(t, tc.expected, got, "pairs")
}

// newAppendSLogAttrTestCase creates a new attribute flattening test case.
func newAppendSLogAttrTestCase(name, prefix string, attr stdslog.Attr,
	expected []any) appendSLogAttrTestCase {
	return appendSLogAttrTestCase{
		name:     name,
		prefix:   prefix,
		attr:     attr,
		expected: expected,
	}
}

func appendSLogAttrTestCases() []appendSLogAttrTestCase {
	return []appendSLogAttrTestCase{
		newAppendSLogAttrTestCase("string attr", "",
			stdslog.String("k", "v"),
			core.S[any]("k", "v")),
		newAppendSLogAttrTestCase("prefixed attr", "req.",
			stdslog.String("k", "v"),
			core.S[any]("req.k", "v")),
		newAppendSLogAttrTestCase("int attr", "",
			stdslog.Int("n", 42),
			core.S[any]("n", int64(42))),
		newAppendSLogAttrTestCase("empty key dropped", "",
			stdslog.String("", "v"),
			nil),
		newAppendSLogAttrTestCase("group", "",
			stdslog.Group("g", stdslog.String("k", "v"), stdslog.Int("n", 1)),
			core.S[any]("g.k", "v", "g.n", int64(1))),
		newAppendSLogAttrTestCase("nested group", "",
			stdslog.Group("a", stdslog.Group("b", stdslog.String("k", "v"))),
			core.S[any]("a.b.k", "v")),
		newAppendSLogAttrTestCase("inline group", "",
			stdslog.Group("", stdslog.String("k", "v")),
			core.S[any]("k", "v")),
		newAppendSLogAttrTestCase("prefixed group", "p.",
			stdslog.Group("g", stdslog.String("k", "v")),
			core.S[any]("p.g.k", "v")),
		newAppendSLogAttrTestCase("empty group dropped", "",
			stdslog.Group("g"),
			nil),
		newAppendSLogAttrTestCase("log valuer resolved", "",
			stdslog.Any("lv", testLogValuer{}),
			core.S[any]("lv", "resolved")),
	}
}

func TestAppendSLogAttr(t *testing.T) {
	core.RunTestCases(t, appendSLogAttrTestCases())
}

func TestSLogRecordAttrs(t *testing.T) {
	record := stdslog.NewRecord(time.Time{}, stdslog.LevelInfo, "msg", 0)
	record.AddAttrs(stdslog.String("a", "1"),
		stdslog.Group("g", stdslog.Int("n", 2)))

	got := slogstdslog.SLogRecordAttrs(record, "p.")
	core.AssertSliceEqual(t, core.S[any]("p.a", "1", "p.g.n", int64(2)), got,
		"pairs")
}
