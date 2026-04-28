package internal_test

import (
	"sync"
	"testing"

	"darvaza.org/core"

	"darvaza.org/slog/internal"
)

// TestLogletFieldsMapRace tests for race conditions when multiple goroutines
// access the fields map concurrently
func TestLogletFieldsMapRace(t *testing.T) {
	t.Run("concurrent FieldsMap access", runTestConcurrentFieldsMapAccess)
	t.Run("nested hierarchy race", runTestNestedHierarchyRace)
	t.Run("peek with concurrency", runTestPeekFieldsMapConcurrency)
}

// runTestConcurrentFieldsMapAccess tests the original race condition where:
// - One goroutine is computing fieldsMap (writing)
// - Another goroutine is checking cached ancestors (reading)
func runTestConcurrentFieldsMapAccess(t *testing.T) {
	t.Helper()

	// Build a chain whose intermediate nodes will end up with cached
	// fieldsMaps, then fan out to leaves that all traverse those
	// ancestors so leaf reads race against ancestor writes.
	var root internal.Loglet
	level1 := root.WithField("f1", "v1")
	level2 := level1.WithField("f2", "v2")
	level3 := level2.WithField("f3", "v3")

	leafNodes := make([]internal.Loglet, 10)
	for i := range leafNodes {
		leafNodes[i] = level3.WithField("leaf", i)
	}

	var wg sync.WaitGroup
	runLeafFieldsMapWorkers(t, leafNodes, &wg)
	runIntermediateFieldsMapTouches(&wg, &level1, &level2, &level3)
	wg.Wait()
}

// runLeafFieldsMapWorkers fans the leaf set out across goroutines so
// every leaf builds its fieldsMap concurrently with the others.
func runLeafFieldsMapWorkers(t *testing.T, leaves []internal.Loglet, wg *sync.WaitGroup) {
	t.Helper()
	for i := range leaves {
		wg.Add(1)
		go func(id int, node *internal.Loglet) {
			defer wg.Done()
			assertLeafAncestorFields(t, id, node.FieldsMap())
		}(i, &leaves[i])
	}
}

// assertLeafAncestorFields checks that a leaf's resolved fieldsMap
// contains the ancestor entries that should have been inherited.
func assertLeafAncestorFields(t *testing.T, id int, fields map[string]any) {
	t.Helper()
	if fields == nil {
		t.Errorf("worker %d: fields is nil", id)
		return
	}
	if v, ok := fields["f1"]; !ok || v != "v1" {
		t.Errorf("worker %d: missing or wrong f1", id)
	}
	if v, ok := fields["f2"]; !ok || v != "v2" {
		t.Errorf("worker %d: missing or wrong f2", id)
	}
}

// runIntermediateFieldsMapTouches triggers computation on intermediate
// nodes concurrently so ancestors are being written while leaves read.
func runIntermediateFieldsMapTouches(wg *sync.WaitGroup, loglets ...*internal.Loglet) {
	wg.Add(len(loglets))
	for _, ll := range loglets {
		go func(l *internal.Loglet) {
			defer wg.Done()
			_ = l.FieldsMap()
		}(ll)
	}
}

// runTestNestedHierarchyRace tests a more complex scenario with nested
// field map computations happening simultaneously at different levels
func runTestNestedHierarchyRace(t *testing.T) {
	t.Helper()

	var root internal.Loglet
	root = root.WithField("app", "test")

	level1 := root.WithField("level", "1")
	level2 := level1.WithField("level", "2")
	level3 := level2.WithField("level", "3")
	level4 := level3.WithField("level", "4")

	branch2a := level2.WithField("branch", "2a")
	branch2b := level2.WithField("branch", "2b")
	branch3a := level3.WithField("branch", "3a")
	branch3b := level3.WithField("branch", "3b")

	// Exercise the conditional-locking path: a node skips locking
	// itself (it already holds the lock) but does lock ancestors.
	loglets := []*internal.Loglet{
		&level4, &branch2a, &branch2b, &branch3a, &branch3b,
	}
	runHierarchyAppFieldWorkers(t, loglets)

	core.AssertEqual(t, "test", level4.FieldsMap()["app"], "level4 app")
	core.AssertEqual(t, "4", level4.FieldsMap()["level"], "level4 level")
	core.AssertEqual(t, "2a", branch2a.FieldsMap()["branch"], "branch2a")
	core.AssertEqual(t, "2b", branch2b.FieldsMap()["branch"], "branch2b")
}

// runHierarchyAppFieldWorkers fans out one goroutine per loglet, each
// re-resolving the fieldsMap several times to broaden the race window.
func runHierarchyAppFieldWorkers(t *testing.T, loglets []*internal.Loglet) {
	t.Helper()
	var wg sync.WaitGroup
	for _, ll := range loglets {
		wg.Add(1)
		go func(loglet *internal.Loglet) {
			defer wg.Done()
			assertAppFieldRepeatedly(t, loglet)
		}(ll)
	}
	wg.Wait()
}

// assertAppFieldRepeatedly resolves the loglet's fieldsMap a few times
// in a row, asserting the ancestor "app" field is always present.
func assertAppFieldRepeatedly(t *testing.T, loglet *internal.Loglet) {
	t.Helper()
	for range 5 {
		fields := loglet.FieldsMap()
		if app, ok := fields["app"]; !ok || app != "test" {
			t.Error("missing or incorrect 'app' field")
		}
	}
}

// runTestPeekFieldsMapConcurrency verifies that peekFieldsMap with
// conditional locking works correctly
func runTestPeekFieldsMapConcurrency(t *testing.T) {
	t.Helper()

	var base internal.Loglet
	parent := base.WithField("parent", "value")
	child := parent.WithField("child", "value")

	// Pre-cache the parent to test peeking existing cache
	_ = parent.FieldsMap()

	var wg sync.WaitGroup

	// Multiple goroutines trying to build child's map
	// They will all peek at parent's cached map
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fields := child.FieldsMap()
			if fields["parent"] != "value" {
				t.Error("incorrect parent field")
			}
		}()
	}

	wg.Wait()
}
