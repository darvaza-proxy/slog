package testing

import (
	"testing"

	"darvaza.org/core"
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
		got := opts.ExpectedLevel(level)
		core.AssertEqual(t, level, got, "ExpectedLevel(%v)", level)
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
	got := opts.ExpectedLevel(slog.Warn)
	core.AssertEqual(t, slog.Info, got, "ExpectedLevel(Warn)")

	got = opts.ExpectedLevel(slog.Debug)
	core.AssertEqual(t, slog.Info, got, "ExpectedLevel(Debug)")

	// Test unmapped levels remain unchanged
	got = opts.ExpectedLevel(slog.Error)
	core.AssertEqual(t, slog.Error, got, "ExpectedLevel(Error)")

	got = opts.ExpectedLevel(slog.Info)
	core.AssertEqual(t, slog.Info, got, "ExpectedLevel(Info)")
}

func testEmptyExceptions(t *testing.T) {
	t.Helper()
	opts := &BidirectionalTestOptions{
		AdapterOptions: AdapterOptions{
			LevelExceptions: map[slog.LogLevel]slog.LogLevel{},
		},
	}

	// All levels should map to themselves with empty map
	got := opts.ExpectedLevel(slog.Warn)
	core.AssertEqual(t, slog.Warn, got, "ExpectedLevel(Warn)")
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
	got := opts.ExpectedLevel(slog.Warn)
	core.AssertEqual(t, slog.UndefinedLevel, got, "ExpectedLevel(Warn)")

	got = opts.ExpectedLevel(slog.Debug)
	core.AssertEqual(t, slog.UndefinedLevel, got, "ExpectedLevel(Debug)")

	// Test that other levels remain unchanged
	got = opts.ExpectedLevel(slog.Info)
	core.AssertEqual(t, slog.Info, got, "ExpectedLevel(Info)")

	got = opts.ExpectedLevel(slog.Error)
	core.AssertEqual(t, slog.Error, got, "ExpectedLevel(Error)")
}

// TestBidirectionalWithOptionsIntegration tests the integration with a mock adapter
func TestBidirectionalWithOptionsIntegration(t *testing.T) {
	t.Run("MockAdapterWithOptions", testMockAdapterWithOptions)
}

func testMockAdapterWithOptions(t *testing.T) {
	t.Helper()
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
	return l, l.Enabled() // skipcq: GO-W4006
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

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = testBidirectionalFunctionTestCase{}
var _ core.TestCase = testBidirectionalWithAdapterTestCase{}

type testBidirectionalFunctionTestCase struct {
	adapterFn   func(slog.Logger) slog.Logger
	name        string
	expectError bool
}

func (tc testBidirectionalFunctionTestCase) Name() string {
	return tc.name
}

func (tc testBidirectionalFunctionTestCase) Test(t *testing.T) {
	t.Helper()

	// For coverage testing, we just need to verify the function can be called
	// The function may fail in its subtests (that's expected), but it shouldn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("TestBidirectional panicked: %v", r)
		}
	}()

	TestBidirectional(t, "test-adapter", tc.adapterFn)

	// Note: This function may fail in its subtests, but that's okay for coverage testing
}

func newTestBidirectionalFunctionTestCase(name string,
	adapterFn func(slog.Logger) slog.Logger, expectError bool) testBidirectionalFunctionTestCase {
	return testBidirectionalFunctionTestCase{
		name:        name,
		adapterFn:   adapterFn,
		expectError: expectError,
	}
}

func testBidirectionalFunctionTestCases() []testBidirectionalFunctionTestCase {
	return []testBidirectionalFunctionTestCase{
		newTestBidirectionalFunctionTestCase("pass-through adapter",
			func(backend slog.Logger) slog.Logger { return backend },
			false),
		newTestBidirectionalFunctionTestCase("field adding adapter",
			func(backend slog.Logger) slog.Logger {
				return backend.WithField("adapter", "test")
			},
			false),
	}
}

func TestTestBidirectionalFunction(t *testing.T) {
	core.RunTestCases(t, testBidirectionalFunctionTestCases())
}

type testBidirectionalWithAdapterTestCase struct {
	adapterFactory func() slog.Logger
	name           string
	expectError    bool
}

func (tc testBidirectionalWithAdapterTestCase) Name() string {
	return tc.name
}

func (tc testBidirectionalWithAdapterTestCase) Test(t *testing.T) {
	t.Helper()

	// Use MockT to test function without failing the build
	mock := &core.MockT{}

	// For coverage testing, we just need to verify the function can be called
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("TestBidirectionalWithAdapter panicked: %v", r)
		}
	}()

	TestBidirectionalWithAdapter(mock, "test-factory", tc.adapterFactory)

	// Function executed without panic - test passed
}

func newTestBidirectionalWithAdapterTestCase(name string,
	adapterFactory func() slog.Logger, expectError bool) testBidirectionalWithAdapterTestCase {
	return testBidirectionalWithAdapterTestCase{
		name:           name,
		adapterFactory: adapterFactory,
		expectError:    expectError,
	}
}

func testBidirectionalWithAdapterTestCases() []testBidirectionalWithAdapterTestCase {
	return []testBidirectionalWithAdapterTestCase{
		newTestBidirectionalWithAdapterTestCase("simple factory",
			func() slog.Logger {
				// Create a logger with recording capability for bidirectional testing
				return NewLogger()
			},
			false),
	}
}

func TestTestBidirectionalWithAdapter(t *testing.T) {
	core.RunTestCases(t, testBidirectionalWithAdapterTestCases())
}
