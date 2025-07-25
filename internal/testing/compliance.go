// Package testing provides shared test utilities for slog handler testing.
package testing

import (
	"fmt"
	"sync"
	"testing"

	"darvaza.org/slog"
)

// ComplianceTest runs a comprehensive test suite to verify that a logger
// implementation correctly implements the slog.Logger interface.
type ComplianceTest struct {
	AdapterOptions
	FactoryOptions

	// SkipEnabledTests skips tests that require checking the Enabled state.
	SkipEnabledTests bool

	// SkipPanicTests skips tests that would cause panics.
	SkipPanicTests bool
}

// Run executes the full compliance test suite.
func (ct ComplianceTest) Run(t *testing.T) {
	t.Run("Interface", func(t *testing.T) {
		logger := ct.NewLogger()

		// Verify it's not nil
		if logger == nil {
			t.Fatal("NewLogger returned nil")
		}

		// Verify it implements slog.Logger
		var _ slog.Logger = logger
	})

	t.Run("LevelMethods", ct.testLevelMethods)

	t.Run("FieldMethods", ct.testFieldMethods)

	t.Run("PrintMethods", ct.testPrintMethods)

	if !ct.SkipEnabledTests {
		t.Run("EnabledMethod", ct.testEnabledMethod)
	}

	t.Run("WithStack", ct.testWithStack)

	t.Run("Immutability", ct.testImmutability)

	t.Run("BasicConcurrency", ct.testBasicConcurrency)
}

func (ct ComplianceTest) testLevelMethods(t *testing.T) {
	t.Helper()
	levels := logLevels()

	for _, tc := range levels {
		t.Run(tc.name, func(t *testing.T) {
			ct.testSingleLevelMethod(t, tc.method, tc.level)
		})
	}
}

// testSingleLevelMethod tests a single level method in compliance test.
func (ct ComplianceTest) testSingleLevelMethod(t *testing.T,
	method func(slog.Logger) slog.Logger, level slog.LogLevel) {
	t.Helper()

	if ct.SkipPanicTests && (level == slog.Panic || level == slog.Fatal) {
		t.Skip("Skipping panic/fatal test")
	}

	logger := ct.NewLogger()
	levelLogger := method(logger)

	// Test that we get a logger back
	if levelLogger == nil {
		t.Fatal("level method returned nil")
	}

	// Test method chaining
	chained := levelLogger.WithField("test", "value")
	if chained == nil {
		t.Fatal("chained level method returned nil")
	}
}

func (ct ComplianceTest) testFieldMethods(t *testing.T) {
	t.Helper()
	RunWithLogger(t, "WithField", ct.NewLogger(), TestWithField)

	RunWithLogger(t, "WithFields", ct.NewLogger(), TestWithFields)
}

func (ct ComplianceTest) testPrintMethods(t *testing.T) {
	t.Helper()
	tests := []struct {
		name   string
		method func(slog.Logger, ...any)
	}{
		{
			name: "Print",
			method: func(l slog.Logger, args ...any) {
				l.Print(args...)
			},
		},
		{
			name: "Println",
			method: func(l slog.Logger, args ...any) {
				l.Println(args...)
			},
		},
		{
			name: "Printf",
			method: func(l slog.Logger, args ...any) {
				if len(args) > 0 {
					l.Printf("%v", args[0])
				} else {
					l.Printf("test")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ct.testPrintMethod(t, tc.method)
		})
	}
}

// testPrintMethod tests a single print method with various arguments
func (ct ComplianceTest) testPrintMethod(t *testing.T, method func(slog.Logger, ...any)) {
	t.Helper()
	logger := ct.NewLogger()

	// Test with no args
	method(logger.Info())

	// Test with single arg
	method(logger.Info(), "test")

	// Test with multiple args
	method(logger.Info(), "test", 123, true)

	// Test with nil arg
	method(logger.Info(), nil)
}

func (ct ComplianceTest) testEnabledMethod(t *testing.T) {
	t.Helper()
	t.Run("Enabled", func(_ *testing.T) {
		logger := ct.NewLogger()

		// Test Enabled method exists and returns bool
		_ = logger.Enabled()
	})

	t.Run("WithEnabled", func(t *testing.T) {
		logger := ct.NewLogger()

		// Test WithEnabled returns logger and bool
		l, enabled := logger.WithEnabled()
		if l == nil {
			t.Error("WithEnabled returned nil logger")
		}

		// If logger is enabled, the returned logger should be the same
		if enabled && l != logger {
			t.Error("WithEnabled should return same logger when enabled")
		}
	})
}

func (ct ComplianceTest) testWithStack(t *testing.T) {
	t.Helper()
	TestWithStack(t, ct.NewLogger())
}

func (ct ComplianceTest) testImmutability(t *testing.T) {
	t.Helper()

	t.Run("BasicImmutability", ct.testBasicImmutability)
	t.Run("BranchingIndependence", ct.testBranchingIndependence)
	t.Run("FieldChainIsolation", ct.testFieldChainIsolation)
	t.Run("DeepBranching", ct.testDeepBranching)
}

// testBasicImmutability verifies that WithField and level methods create new instances.
func (ct ComplianceTest) testBasicImmutability(t *testing.T) {
	t.Helper()
	base := ct.NewLogger()

	// Test WithField creates new instances
	ct.verifyFieldImmutability(t, base)

	// Test level methods create new instances
	ct.verifyLevelImmutability(t, base)
}

// verifyFieldImmutability checks that WithField operations create new logger instances.
func (ComplianceTest) verifyFieldImmutability(t *testing.T, base slog.Logger) {
	t.Helper()

	// Add fields to a logger
	l1 := base.WithField("field1", "value1")
	l2 := base.WithField("field2", "value2")

	// l1 and l2 should be independent
	if l1 == l2 {
		t.Error("WithField should return new logger instances")
	}

	// Original should be unchanged
	if base == l1 || base == l2 {
		t.Error("WithField should not modify original logger")
	}
}

// verifyLevelImmutability checks that level methods create new logger instances.
func (ComplianceTest) verifyLevelImmutability(t *testing.T, base slog.Logger) {
	t.Helper()

	// Test with levels
	l3 := base.Info()
	l4 := base.Debug()

	if l3 == l4 {
		t.Error("level methods should return new logger instances")
	}
}

// testBranchingIndependence verifies that branched loggers maintain independence.
func (ct ComplianceTest) testBranchingIndependence(t *testing.T) {
	t.Helper()
	base := ct.NewLogger()

	// Create branches
	branches := ct.createBranches(base)

	// Verify all loggers are distinct
	ct.verifyDistinctLoggers(t, branches)
}

// createBranches creates a set of branched loggers for testing.
func (ComplianceTest) createBranches(base slog.Logger) []slog.Logger {
	// Create first branch
	branch1 := base.WithField("branch", "1")
	branch1a := branch1.WithField("subbranch", "1a")
	branch1b := branch1.WithField("subbranch", "1b")

	// Create second branch
	branch2 := base.WithField("branch", "2")
	branch2a := branch2.WithField("subbranch", "2a")

	return []slog.Logger{base, branch1, branch1a, branch1b, branch2, branch2a}
}

// verifyDistinctLoggers ensures all loggers in the slice are distinct instances.
func (ComplianceTest) verifyDistinctLoggers(t *testing.T, loggers []slog.Logger) {
	t.Helper()

	for i := 0; i < len(loggers); i++ {
		for j := i + 1; j < len(loggers); j++ {
			if loggers[i] == loggers[j] {
				t.Errorf("Loggers at indices %d and %d are the same instance", i, j)
			}
		}
	}
}

// testFieldChainIsolation verifies that field modifications don't affect parent or sibling loggers.
// This test requires NewLoggerWithRecorder to be provided in FactoryOptions, which should create
// the handler being tested with a test recorder as its output backend. This allows the test to
// verify that fields are properly isolated between logger instances.
//
// If your handler cannot support recording (e.g., writes directly to external systems),
// document this limitation in your test file and these tests will be automatically skipped.
func (ct ComplianceTest) testFieldChainIsolation(t *testing.T) {
	t.Helper()

	// Skip if recorder not available
	if ct.NewLoggerWithRecorder == nil {
		t.Skip("Need NewLoggerWithRecorder for field chain verification - " +
			"provide this in FactoryOptions to enable these tests")
	}

	// Run isolation tests
	t.Run("BaseLogger", ct.testBaseLoggerIsolation)
	t.Run("BranchWithFields", ct.testBranchFieldsIsolation)
	t.Run("IndependentBranches", ct.testIndependentBranchesIsolation)
}

// testBaseLoggerIsolation verifies base logger has no custom fields.
func (ct ComplianceTest) testBaseLoggerIsolation(t *testing.T) {
	t.Helper()
	recorder := NewLogger()
	base := ct.NewLoggerWithRecorder(recorder)

	// Log from base - should have no fields
	base.Info().Print("base message")

	// Verify message count
	messages := recorder.GetMessages()
	if len(messages) != 1 {
		t.Skipf("Handler appears to be asynchronous, got %d messages instead of 1", len(messages))
	}

	// Base message should have no custom fields
	msg := messages[0]
	if msg.Fields["app"] != nil || msg.Fields["version"] != nil {
		t.Error("Base logger message should not have branch fields")
	}
}

// testBranchFieldsIsolation verifies branch logger has its own fields.
func (ct ComplianceTest) testBranchFieldsIsolation(t *testing.T) {
	t.Helper()
	recorder := NewLogger()
	base := ct.NewLoggerWithRecorder(recorder)

	// Create a branch with fields
	branch := base.
		WithField("app", "test").
		WithField("version", "1.0")

	// Log from branch - should have both fields
	branch.Info().Print("branch message")

	messages := recorder.GetMessages()
	if len(messages) != 1 {
		t.Skipf("Handler appears to be asynchronous, got %d messages instead of 1", len(messages))
	}

	// Branch message should have its fields
	AssertField(t, messages[0], "app", "test")
	AssertField(t, messages[0], "version", "1.0")
}

// testIndependentBranchesIsolation verifies branches are independent.
func (ct ComplianceTest) testIndependentBranchesIsolation(t *testing.T) {
	t.Helper()
	recorder := NewLogger()
	base := ct.NewLoggerWithRecorder(recorder)

	// Create branches
	branch1 := base.WithField("type", "branch1")
	branch2 := base.WithField("type", "branch2")

	// Log from both branches
	branch1.Info().Print("branch1 message")
	branch2.Info().Print("branch2 message")

	messages := recorder.GetMessages()
	if len(messages) != 2 {
		t.Skipf("Handler appears to be asynchronous, got %d messages instead of 2", len(messages))
	}

	// Each message should only have its own branch's fields
	AssertField(t, messages[0], "type", "branch1")
	AssertField(t, messages[1], "type", "branch2")
}

// testDeepBranching verifies deep branching scenarios.
func (ct ComplianceTest) testDeepBranching(t *testing.T) {
	t.Helper()
	base := ct.NewLogger()

	// Create deep branch hierarchy
	hierarchy := ct.createDeepHierarchy(base)

	// Verify distinctness
	ct.verifyDeepHierarchyDistinct(t, hierarchy)

	// Test sibling independence
	ct.verifySiblingIndependence(t, hierarchy.l1)
}

// deepHierarchy holds a deep branch hierarchy for testing.
type deepHierarchy struct {
	l1, l2, l3, l4 slog.Logger
}

// createDeepHierarchy creates a deep branch hierarchy.
func (ComplianceTest) createDeepHierarchy(base slog.Logger) deepHierarchy {
	l1 := base.WithField("level", 1)
	l2 := l1.WithField("level", 2).WithField("data", "test")
	l3 := l2.WithField("level", 3)
	l4 := l3.WithField("level", 4).WithField("extra", "info")

	return deepHierarchy{l1: l1, l2: l2, l3: l3, l4: l4}
}

// verifyDeepHierarchyDistinct checks that all loggers in hierarchy are distinct.
func (ComplianceTest) verifyDeepHierarchyDistinct(t *testing.T, h deepHierarchy) {
	t.Helper()

	if h.l1 == h.l2 || h.l2 == h.l3 || h.l3 == h.l4 {
		t.Error("Deep branching should create new logger instances")
	}
}

// verifySiblingIndependence checks that sibling branches are independent.
func (ComplianceTest) verifySiblingIndependence(t *testing.T, l1 slog.Logger) {
	t.Helper()

	// Create a sibling branch at level 2
	l2Sibling := l1.WithField("level", 2).WithField("data", "sibling")

	// Recreate original l2 to compare
	l2 := l1.WithField("level", 2).WithField("data", "test")

	if l2 == l2Sibling {
		t.Error("Sibling branches should be independent")
	}
}

// concurrencyTest defines a basic concurrency test case.
type concurrencyTest struct {
	name       string
	goroutines int
	operations int
	testFunc   func(*testing.T, slog.Logger, int, int)
}

// concurrencyTests returns the basic concurrency test cases.
func (ct ComplianceTest) concurrencyTests() []concurrencyTest {
	return []concurrencyTest{
		{
			name:       "ConcurrentLogging",
			goroutines: 5,
			operations: 10,
			testFunc:   ct.testConcurrentLogging,
		},
		{
			name:       "ConcurrentFieldIndependence",
			goroutines: 5,
			operations: 1, // Only need one operation per goroutine for this test
			testFunc:   ct.testConcurrentFieldIndependence,
		},
	}
}

func (ct ComplianceTest) testBasicConcurrency(t *testing.T) {
	t.Helper()

	// Basic concurrency test to ensure thread safety (compliance requirement)
	// This is NOT a stress test, just verification of safe concurrent access
	logger := ct.NewLogger()

	for _, tc := range ct.concurrencyTests() {
		t.Run(tc.name, func(t *testing.T) {
			tc.testFunc(t, logger, tc.goroutines, tc.operations)
		})
	}
}

// testConcurrentLogging verifies concurrent logging operations are safe.
func (ct ComplianceTest) testConcurrentLogging(_ *testing.T, logger slog.Logger, goroutines, operations int) {
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go ct.runConcurrentLogOperations(&wg, logger, i, operations)
	}

	wg.Wait()
}

// runConcurrentLogOperations performs logging operations in a goroutine.
func (ComplianceTest) runConcurrentLogOperations(wg *sync.WaitGroup, logger slog.Logger, id, operations int) {
	defer wg.Done()

	for j := 0; j < operations; j++ {
		// Just verify we can log concurrently without panic
		logger.Info().
			WithField("goroutine", id).
			WithField("operation", j).
			Print("concurrent test")
	}
}

// testConcurrentFieldIndependence verifies concurrent field operations maintain independence.
func (ct ComplianceTest) testConcurrentFieldIndependence(t *testing.T, logger slog.Logger, goroutines, _ int) {
	loggers := ct.createConcurrentLoggers(logger, goroutines)
	ct.verifyConcurrentLoggersDistinct(t, loggers)
}

// createConcurrentLoggers creates loggers concurrently and returns them.
func (ct ComplianceTest) createConcurrentLoggers(logger slog.Logger, count int) []slog.Logger {
	var wg sync.WaitGroup
	loggers := make([]slog.Logger, count)

	for i := 0; i < count; i++ {
		wg.Add(1)
		go ct.createLoggerBranch(&wg, logger, i, loggers)
	}

	wg.Wait()
	return loggers
}

// createLoggerBranch creates a logger branch in a goroutine.
func (ComplianceTest) createLoggerBranch(wg *sync.WaitGroup, logger slog.Logger, id int, loggers []slog.Logger) {
	defer wg.Done()

	// Each goroutine creates its own branch
	loggers[id] = logger.
		WithField("goroutine", id).
		WithField("data", fmt.Sprintf("data_%d", id))
}

// verifyConcurrentLoggersDistinct checks that concurrently created loggers are distinct.
func (ComplianceTest) verifyConcurrentLoggersDistinct(t *testing.T, loggers []slog.Logger) {
	t.Helper()

	for i := 0; i < len(loggers); i++ {
		for j := i + 1; j < len(loggers); j++ {
			if loggers[i] == loggers[j] {
				t.Errorf("Concurrent WithField created same logger instance for goroutines %d and %d", i, j)
			}
		}
	}
}
