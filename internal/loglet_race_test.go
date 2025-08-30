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

	// Create a deeper chain to increase chance of race
	// The key is having intermediate nodes with fields that will be cached
	var root internal.Loglet
	level1 := root.WithField("f1", "v1")
	level2 := level1.WithField("f2", "v2")
	level3 := level2.WithField("f3", "v3")

	// Create multiple leaf nodes that will all traverse the same ancestors
	// This increases the chance of concurrent access to ancestor fieldsMap
	leafNodes := make([]internal.Loglet, 10)
	for i := range leafNodes {
		leafNodes[i] = level3.WithField("leaf", i)
	}

	// Use WaitGroup to coordinate goroutines
	var wg sync.WaitGroup

	// Launch goroutines that access FieldsMap on different nodes simultaneously
	// The race happens when:
	// 1. Leaf nodes start building their fieldsMap
	// 2. They call findCachedAncestor() which reads ancestor.fieldsMap
	// 3. Meanwhile, ancestors are also building their fieldsMap (writing)
	for i := range leafNodes {
		wg.Add(1)
		go func(id int, node *internal.Loglet) {
			defer wg.Done()

			// Force all goroutines to start at roughly the same time
			// This increases the chance of the race

			// Access FieldsMap which triggers buildFieldsMap
			fields := node.FieldsMap()

			// Verify fields are correct
			if fields == nil {
				t.Errorf("worker %d: fields is nil", id)
				return
			}

			// Should have all ancestor fields
			if v, ok := fields["f1"]; !ok || v != "v1" {
				t.Errorf("worker %d: missing or wrong f1", id)
			}
			if v, ok := fields["f2"]; !ok || v != "v2" {
				t.Errorf("worker %d: missing or wrong f2", id)
			}
		}(i, &leafNodes[i])
	}

	// Also trigger computation on intermediate nodes concurrently
	// This ensures ancestors are being written while leaves are reading
	wg.Add(3)
	go func() {
		defer wg.Done()
		_ = level1.FieldsMap()
	}()
	go func() {
		defer wg.Done()
		_ = level2.FieldsMap()
	}()
	go func() {
		defer wg.Done()
		_ = level3.FieldsMap()
	}()

	// Wait for all goroutines to complete
	wg.Wait()
}

// runTestNestedHierarchyRace tests a more complex scenario with nested
// field map computations happening simultaneously at different levels
func runTestNestedHierarchyRace(t *testing.T) {
	t.Helper()

	// Create a deeper hierarchy
	var root internal.Loglet
	root = root.WithField("app", "test")

	level1 := root.WithField("level", "1")
	level2 := level1.WithField("level", "2")
	level3 := level2.WithField("level", "3")
	level4 := level3.WithField("level", "4")

	// Create multiple branches at different levels
	branch2a := level2.WithField("branch", "2a")
	branch2b := level2.WithField("branch", "2b")
	branch3a := level3.WithField("branch", "3a")
	branch3b := level3.WithField("branch", "3b")

	var wg sync.WaitGroup

	// Access different levels concurrently
	// This tests the conditional locking logic where:
	// - A node doesn't lock itself (already holds lock)
	// - But does lock ancestor nodes

	loglets := []*internal.Loglet{
		&level4, &branch2a, &branch2b, &branch3a, &branch3b,
	}

	for _, ll := range loglets {
		wg.Add(1)
		go func(loglet *internal.Loglet) {
			defer wg.Done()

			// Multiple accesses to trigger caching at different times
			for i := 0; i < 5; i++ {
				fields := loglet.FieldsMap()

				// Verify app field is always present (from root)
				if app, ok := fields["app"]; !ok || app != "test" {
					t.Errorf("missing or incorrect 'app' field")
				}
			}
		}(ll)
	}

	wg.Wait()

	// Verify all loglets have correct final state
	core.AssertEqual(t, "test", level4.FieldsMap()["app"], "level4 app")
	core.AssertEqual(t, "4", level4.FieldsMap()["level"], "level4 level")
	core.AssertEqual(t, "2a", branch2a.FieldsMap()["branch"], "branch2a")
	core.AssertEqual(t, "2b", branch2b.FieldsMap()["branch"], "branch2b")
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
	for i := 0; i < 20; i++ {
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
