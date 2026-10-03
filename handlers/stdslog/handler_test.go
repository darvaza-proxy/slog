package stdslog_test

import (
	"context"
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
var _ core.TestCase = handlerLevelCapTestCase{}
var _ core.TestCase = handlerAttrsTestCase{}

// newTestRecord creates a stdlib record without timestamp or PC.
func newTestRecord(level stdslog.Level, msg string) stdslog.Record {
	return stdslog.NewRecord(time.Time{}, level, msg, 0)
}

func TestHandlerNilReceiver(t *testing.T) {
	var h *slogstdslog.Handler
	ctx := context.Background()
	record := newTestRecord(stdslog.LevelInfo, "msg")

	core.AssertFalse(t, h.Enabled(ctx, stdslog.LevelInfo), "enabled")
	core.AssertNil(t, h.Handle(ctx, record), "handle error")
	core.AssertNil(t, h.WithAttrs(core.S(stdslog.String("k", "v"))),
		"with attrs")
	core.AssertNil(t, h.WithGroup("g"), "with group")

	// A zero-value Handler carries no backend logger and behaves the same.
	zero := &slogstdslog.Handler{}
	core.AssertFalse(t, zero.Enabled(ctx, stdslog.LevelError), "zero enabled")
	core.AssertNil(t, zero.Handle(ctx, record), "zero handle error")
	core.AssertSame(t, zero,
		zero.WithAttrs(core.S(stdslog.String("k", "v"))), "zero with attrs")
	core.AssertSame(t, zero, zero.WithGroup("g"), "zero with group")
}

// handlerLevelCapTestCase pins inbound levels capping at slog.Error:
// records above LevelError must never trigger Fatal or Panic terminal
// behaviour in the backend.
type handlerLevelCapTestCase struct {
	name     string
	level    stdslog.Level
	expected slog.LogLevel
}

// Name returns the test case name.
func (tc handlerLevelCapTestCase) Name() string {
	return tc.name
}

// Test validates the record is delivered, capped, without panicking.
func (tc handlerLevelCapTestCase) Test(t *testing.T) {
	t.Helper()

	recorder := mock.NewLogger()
	h := slogstdslog.NewHandler(recorder)

	core.AssertNoPanic(t, func() {
		_ = h.Handle(context.Background(), newTestRecord(tc.level, "caps"))
	}, "handle")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)
	slogtest.AssertMustMessage(t, messages[0], tc.expected, "caps")
}

// newHandlerLevelCapTestCase creates a new level capping test case.
func newHandlerLevelCapTestCase(name string, level stdslog.Level,
	expected slog.LogLevel) handlerLevelCapTestCase {
	return handlerLevelCapTestCase{
		name:     name,
		level:    level,
		expected: expected,
	}
}

func TestHandlerLevelCap(t *testing.T) {
	testCases := []handlerLevelCapTestCase{
		newHandlerLevelCapTestCase("error", stdslog.LevelError, slog.Error),
		newHandlerLevelCapTestCase("fatal equivalent",
			stdslog.LevelError+4, slog.Error),
		newHandlerLevelCapTestCase("panic equivalent",
			stdslog.LevelError+8, slog.Error),
	}

	core.RunTestCases(t, testCases)
}

func TestHandlerIdentity(t *testing.T) {
	h := slogstdslog.NewHandler(mock.NewLogger())

	core.AssertSame(t, h, h.WithAttrs(nil), "nil attrs")
	core.AssertSame(t, h, h.WithAttrs(core.S[stdslog.Attr]()), "empty attrs")
	core.AssertSame(t, h, h.WithGroup(""), "empty group")
	// Attrs that flatten to nothing leave the handler untouched.
	core.AssertSame(t, h,
		h.WithAttrs(core.S(stdslog.String("", "v"))), "dropped attrs")
	core.AssertNotSame(t, h,
		h.WithAttrs(core.S(stdslog.String("k", "v"))), "real attrs")
}

// TestHandlerAttrsPersist pins the eager WithAttrs design: attached
// attributes reach the backend on every subsequent record.
func TestHandlerAttrsPersist(t *testing.T) {
	recorder := mock.NewLogger()
	h := slogstdslog.NewHandler(recorder).
		WithAttrs(core.S(stdslog.String("app", "x")))
	ctx := context.Background()

	core.AssertNil(t, h.Handle(ctx, newTestRecord(stdslog.LevelInfo, "one")),
		"first handle")
	core.AssertNil(t, h.Handle(ctx, newTestRecord(stdslog.LevelInfo, "two")),
		"second handle")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 2)
	slogtest.AssertField(t, messages[0], "app", "x")
	slogtest.AssertField(t, messages[1], "app", "x")
}

// handlerAttrsTestCase pins attribute flattening through Handle: group
// prefixes, handler-attached and call-site attributes, and exact field
// counts so stray fields fail.
type handlerAttrsTestCase struct {
	expected     map[string]any
	name         string
	group        string
	handlerAttrs []stdslog.Attr
	recordAttrs  []stdslog.Attr
}

// Name returns the test case name.
func (tc handlerAttrsTestCase) Name() string {
	return tc.name
}

// Test validates the recorded fields, exactly.
func (tc handlerAttrsTestCase) Test(t *testing.T) {
	t.Helper()

	recorder := mock.NewLogger()
	h := slogstdslog.NewHandler(recorder)
	if tc.group != "" {
		h = h.WithGroup(tc.group)
	}
	if len(tc.handlerAttrs) > 0 {
		h = h.WithAttrs(tc.handlerAttrs)
	}

	record := newTestRecord(stdslog.LevelInfo, "attrs")
	record.AddAttrs(tc.recordAttrs...)
	core.AssertNil(t, h.Handle(context.Background(), record), "handle error")

	messages := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, messages, 1)
	slogtest.AssertMustMessage(t, messages[0], slog.Info, "attrs")
	for key, value := range tc.expected {
		slogtest.AssertField(t, messages[0], key, value)
	}
	core.AssertEqual(t, len(tc.expected), len(messages[0].Fields),
		"field count")
}

// newHandlerAttrsTestCase creates a new attribute flattening test case.
func newHandlerAttrsTestCase(name, group string,
	handlerAttrs, recordAttrs []stdslog.Attr,
	expected map[string]any) handlerAttrsTestCase {
	return handlerAttrsTestCase{
		name:         name,
		group:        group,
		handlerAttrs: handlerAttrs,
		recordAttrs:  recordAttrs,
		expected:     expected,
	}
}

func handlerAttrsTestCases() []handlerAttrsTestCase {
	return []handlerAttrsTestCase{
		newHandlerAttrsTestCase("handler attrs", "",
			core.S(stdslog.String("k", "v")), nil,
			map[string]any{"k": "v"}),
		newHandlerAttrsTestCase("record attrs", "",
			nil, core.S(stdslog.Int("n", 1)),
			map[string]any{"n": int64(1)}),
		newHandlerAttrsTestCase("handler and record attrs", "",
			core.S(stdslog.String("a", "1")),
			core.S(stdslog.String("b", "2")),
			map[string]any{"a": "1", "b": "2"}),
		newHandlerAttrsTestCase("nested group value", "",
			nil, core.S(stdslog.Group("a",
				stdslog.Group("b", stdslog.String("k", "v")))),
			map[string]any{"a.b.k": "v"}),
		newHandlerAttrsTestCase("group prefix on record", "req",
			nil, core.S(stdslog.String("id", "x")),
			map[string]any{"req.id": "x"}),
		newHandlerAttrsTestCase("group prefix on handler attrs", "g",
			core.S(stdslog.String("k", "v")), nil,
			map[string]any{"g.k": "v"}),
		newHandlerAttrsTestCase("inline group", "",
			nil, core.S(stdslog.Group("",
				stdslog.String("k", "v"))),
			map[string]any{"k": "v"}),
		newHandlerAttrsTestCase("empty key dropped", "",
			nil, core.S(stdslog.String("", "v")),
			map[string]any{}),
		newHandlerAttrsTestCase("duplicate keys last wins", "",
			nil, core.S(stdslog.String("k", "one"),
				stdslog.String("k", "two")),
			map[string]any{"k": "two"}),
	}
}

func TestHandlerAttrs(t *testing.T) {
	core.RunTestCases(t, handlerAttrsTestCases())
}
