package internal

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// TestLogletAncestorInfo tests the logletAncestorInfo struct and its methods
func TestLogletAncestorInfo(t *testing.T) {
	t.Run("CanDelegate", testCanDelegate)
	t.Run("GetBaseMap", testGetBaseMap)
}

// canDelegateTest represents a test case for the canDelegate method
type canDelegateTest struct {
	// Large fields first (strings, interfaces, slices) - 8+ bytes
	description string
	name        string
	info        logletAncestorInfo

	// Small fields last (booleans) - 1 byte
	expected bool
}

// test validates canDelegate behaviour
func (tc canDelegateTest) test(t *testing.T) {
	t.Helper()

	result := tc.info.canDelegate()
	core.AssertEqual(t, tc.expected, result, tc.description)
}

// canDelegateTestCases returns test cases for the canDelegate method
func canDelegateTestCases() []canDelegateTest {
	return []canDelegateTest{
		{
			name: "NilBaseMap",
			info: logletAncestorInfo{
				pathToRoot:  []*Loglet{{}},
				baseMap:     nil,
				totalFields: 1,
			},
			expected:    false,
			description: "nil baseMap delegation",
		},
		{
			name: "SelfHasCache",
			info: logletAncestorInfo{
				pathToRoot:  []*Loglet{{}}, // length 1 means self has cache
				baseMap:     map[string]any{"key": "value"},
				totalFields: 1,
			},
			expected:    true,
			description: "self cache delegation",
		},
		{
			name: "FieldsMatchBaseMap",
			info: logletAncestorInfo{
				pathToRoot:  []*Loglet{{}, {}}, // length 2 means parent has cache
				baseMap:     map[string]any{"key": "value"},
				totalFields: 1, // matches baseMap length
			},
			expected:    true,
			description: "matching fields delegation",
		},
		{
			name: "FieldsExceedBaseMap",
			info: logletAncestorInfo{
				pathToRoot:  []*Loglet{{}, {}},
				baseMap:     map[string]any{"key": "value"},
				totalFields: 2, // exceeds baseMap length
			},
			expected:    false,
			description: "excess fields no delegation",
		},
		{
			name: "EmptyBaseMap",
			info: logletAncestorInfo{
				pathToRoot:  []*Loglet{{}, {}},
				baseMap:     map[string]any{},
				totalFields: 0,
			},
			expected:    true,
			description: "empty map delegation",
		},
		{
			name: "EdgeCase_PathLength1_DifferentFields",
			info: logletAncestorInfo{
				pathToRoot:  []*Loglet{{}}, // length 1
				baseMap:     map[string]any{"key": "value"},
				totalFields: 5, // different from baseMap length
			},
			expected:    true,
			description: "path length 1 delegation",
		},
		{
			name: "EdgeCase_EmptyPath",
			info: logletAncestorInfo{
				pathToRoot:  []*Loglet{}, // empty path
				baseMap:     map[string]any{"key": "value"},
				totalFields: 1,
			},
			expected:    false,
			description: "empty path no delegation",
		},
	}
}

// testCanDelegate tests the canDelegate method with various scenarios
func testCanDelegate(t *testing.T) {
	for _, tc := range canDelegateTestCases() {
		t.Run(tc.name, tc.test)
	}
}

// testGetBaseMap tests the getBaseMap method delegation logic
func testGetBaseMap(t *testing.T) {
	t.Run("ExistingBaseMap", testGetBaseMapExisting)
	t.Run("NilBaseMapWithFields", testGetBaseMapNilWithFields)
	t.Run("NoFieldsInPath", testGetBaseMapNoFields)
	t.Run("AncestorAlreadyCached", testGetBaseMapAlreadyCached)
}

func testGetBaseMapExisting(t *testing.T) {
	t.Helper()

	existingMap := map[string]any{"existing": "value"}
	info := logletAncestorInfo{
		pathToRoot:  []*Loglet{{}},
		baseMap:     existingMap,
		totalFields: 1,
	}

	result := info.getBaseMap()

	// Use core.AssertMustSame for identity comparison
	core.AssertMustSame(t, existingMap, result, "getBaseMap() should return the same baseMap instance when available")
}

func testGetBaseMapNilWithFields(t *testing.T) {
	t.Helper()

	var base Loglet
	parent := base.WithField("key", "value")
	child := parent.WithLevel(slog.Info) // No fields, should trigger caching

	info := logletAncestorInfo{
		pathToRoot:  []*Loglet{&child, &parent},
		baseMap:     nil,
		totalFields: 1,
	}

	result := info.getBaseMap()
	core.AssertNotNil(t, result, "getBaseMap() should cache ancestor and return non-nil map")

	// Verify the parent got cached
	core.AssertNotNil(t, parent.fieldsMap, "getBaseMap() should have cached the parent's fieldsMap")

	// Verify the info was updated
	core.AssertNotNil(t, info.baseMap, "getBaseMap() should have updated info.baseMap")
	core.AssertEqual(t, len(result), info.totalFields, "getBaseMap() should have updated info.totalFields")
}

func testGetBaseMapNoFields(t *testing.T) {
	t.Helper()

	var base Loglet
	child := base.WithLevel(slog.Info) // No fields

	info := logletAncestorInfo{
		pathToRoot:  []*Loglet{&child},
		baseMap:     nil,
		totalFields: 0,
	}

	result := info.getBaseMap()
	core.AssertNil(t, result, "getBaseMap() should return nil when no fields in path")
}

func testGetBaseMapAlreadyCached(t *testing.T) {
	t.Helper()

	var base Loglet
	parent := base.WithField("key", "value")
	child := parent.WithLevel(slog.Info)

	// Pre-cache the parent
	parent.fieldsMap = map[string]any{"key": "value"}

	info := logletAncestorInfo{
		pathToRoot:  []*Loglet{&child, &parent},
		baseMap:     nil,
		totalFields: 1,
	}

	result := info.getBaseMap()
	core.AssertNil(t, result, "getBaseMap() should return nil when ancestor already cached but no delegation")
}

// TestFindCachedAncestor tests the findCachedAncestor method traversal
func TestFindCachedAncestor(t *testing.T) {
	t.Run("NoCachedAncestor", testFindCachedAncestorNone)
	t.Run("CachedParent", testFindCachedAncestorParent)
	t.Run("CachedSelf", testFindCachedAncestorSelf)
	t.Run("MultipleLevels", testFindCachedAncestorMultiple)
	t.Run("NilLoglet", testFindCachedAncestorNil)
	t.Run("EmptyLoglet", testFindCachedAncestorEmpty)
}

func testFindCachedAncestorNone(t *testing.T) {
	t.Helper()

	var base Loglet
	parent := base.WithField("key", "value")
	child := parent.WithField("child", "value")

	info := child.findCachedAncestor()

	core.AssertNil(t, info.baseMap, "findCachedAncestor() should return nil baseMap when no cached ancestor")
	core.AssertEqual(t, 2, len(info.pathToRoot), "findCachedAncestor() pathToRoot length")
	core.AssertEqual(t, 2, info.totalFields, "findCachedAncestor() totalFields")
}

func testFindCachedAncestorParent(t *testing.T) {
	t.Helper()

	var base Loglet
	parent := base.WithField("key", "value")
	child := parent.WithField("child", "value")

	// Cache the parent
	cachedMap := map[string]any{"key": "value"}
	parent.fieldsMap = cachedMap

	info := child.findCachedAncestor()

	core.AssertNotNil(t, info.baseMap, "findCachedAncestor() should return cached parent's map")
	core.AssertEqual(t, len(cachedMap), len(info.baseMap), "findCachedAncestor() baseMap length")
	core.AssertEqual(t, 1, len(info.pathToRoot), "findCachedAncestor() pathToRoot length")
	// totalFields should be: cached map size (1) + child fields (1) = 2
	expectedFields := len(cachedMap) + 1 // child's field
	core.AssertEqual(t, expectedFields, info.totalFields, "findCachedAncestor() totalFields")
}

func testFindCachedAncestorSelf(t *testing.T) {
	t.Helper()

	var base Loglet
	loglet := base.WithField("key", "value")

	// Cache self
	cachedMap := map[string]any{"key": "value"}
	loglet.fieldsMap = cachedMap

	info := loglet.findCachedAncestor()

	core.AssertNotNil(t, info.baseMap, "findCachedAncestor() should return self's cached map")
	core.AssertEqual(t, len(cachedMap), len(info.baseMap), "findCachedAncestor() baseMap length")
	core.AssertEqual(t, 0, len(info.pathToRoot), "findCachedAncestor() pathToRoot length")
	core.AssertEqual(t, len(cachedMap), info.totalFields, "findCachedAncestor() totalFields")
}

func testFindCachedAncestorMultiple(t *testing.T) {
	t.Helper()

	var base Loglet
	grandparent := base.WithField("gp", "value")
	parent := grandparent.WithField("p", "value")
	child := parent.WithField("c", "value")

	// Cache the grandparent (furthest ancestor)
	cachedMap := map[string]any{"gp": "value"}
	grandparent.fieldsMap = cachedMap

	info := child.findCachedAncestor()

	core.AssertNotNil(t, info.baseMap, "findCachedAncestor() should return furthest cached ancestor's map")
	core.AssertEqual(t, len(cachedMap), len(info.baseMap), "findCachedAncestor() baseMap length")
	core.AssertEqual(t, 2, len(info.pathToRoot), "findCachedAncestor() pathToRoot length")
	// totalFields should be: cached map size (1) + parent fields (1) + child fields (1) = 3
	expectedFields := len(cachedMap) + 2 // parent + child fields
	core.AssertEqual(t, expectedFields, info.totalFields, "findCachedAncestor() totalFields")
}

func testFindCachedAncestorNil(t *testing.T) {
	t.Helper()

	var loglet *Loglet
	info := loglet.findCachedAncestor()

	core.AssertNil(t, info.baseMap, "findCachedAncestor() should return nil baseMap for nil loglet")
	core.AssertEqual(t, 0, len(info.pathToRoot), "findCachedAncestor() pathToRoot length")
	core.AssertEqual(t, 0, info.totalFields, "findCachedAncestor() totalFields")
}

func testFindCachedAncestorEmpty(t *testing.T) {
	t.Helper()

	var base Loglet
	info := base.findCachedAncestor()

	core.AssertNil(t, info.baseMap, "findCachedAncestor() should return nil baseMap for empty loglet")
	core.AssertEqual(t, 1, len(info.pathToRoot), "findCachedAncestor() pathToRoot length")
	core.AssertEqual(t, 0, info.totalFields, "findCachedAncestor() totalFields")
}

// TestAncestorIntegration tests the integration between ancestor methods
func TestAncestorIntegration(t *testing.T) {
	t.Run("DelegationWorkflow", testAncestorIntegrationWorkflow)
}

func testAncestorIntegrationWorkflow(t *testing.T) {
	t.Helper()

	var base Loglet
	parent := base.WithField("key", "value")
	child := parent.WithLevel(slog.Info) // No fields, should delegate

	// First call should establish caching
	info := child.findCachedAncestor()
	baseMap := info.getBaseMap()

	core.AssertNotNil(t, baseMap, "Integration: getBaseMap() should establish caching and return map")

	// Verify delegation is possible
	core.AssertTrue(t, info.canDelegate(), "Integration: should be able to delegate after caching")

	// Verify parent was cached
	core.AssertNotNil(t, parent.fieldsMap, "Integration: parent should have been cached")

	// Second call should find existing cache
	info2 := child.findCachedAncestor()
	core.AssertNotNil(t, info2.baseMap, "Integration: second call should find existing cached ancestor")
}
