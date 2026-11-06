package testing

import (
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ core.TestCase = subprocessPatternTestCase{}
	_ core.TestCase = checkFatalExitTestCase{}
)

// subprocessPatternTestCase tests the -test.run pattern built from a
// test name.
type subprocessPatternTestCase struct {
	name     string
	testName string
	want     string
}

func (tc subprocessPatternTestCase) Name() string {
	return tc.name
}

func (tc subprocessPatternTestCase) Test(t *testing.T) {
	core.AssertEqual(t, tc.want, subprocessPattern(tc.testName), "pattern")
}

func newSubprocessPatternTestCase(name, testName, want string) subprocessPatternTestCase {
	return subprocessPatternTestCase{
		name:     name,
		testName: testName,
		want:     want,
	}
}

func subprocessPatternTestCases() []subprocessPatternTestCase {
	return []subprocessPatternTestCase{
		newSubprocessPatternTestCase("test", "TestA", `^TestA$`),
		newSubprocessPatternTestCase("subtest", "TestA/sub_test", `^TestA$/^sub_test$`),
		newSubprocessPatternTestCase("nested subtest", "TestA/b/c", `^TestA$/^b$/^c$`),
		newSubprocessPatternTestCase("regular expression syntax", `TestA/a.b(c)|[d]`,
			`^TestA$/^a\.b\(c\)\|\[d\]$`),
	}
}

func TestSubprocessPattern(t *testing.T) {
	core.RunTestCases(t, subprocessPatternTestCases())
}

// checkFatalExitTestCase checks how a child ended against what Fatal
// does.
type checkFatalExitTestCase struct {
	name   string
	output string
	text   string
	status int
	want   bool
}

func (tc checkFatalExitTestCase) Name() string {
	return tc.name
}

func (tc checkFatalExitTestCase) Test(t *testing.T) {
	mock := &core.MockT{}
	ok := checkFatalExit(mock, tc.status, tc.output, tc.text)

	core.AssertEqual(t, tc.want, ok, "result")
	core.AssertEqual(t, !ok, mock.HasErrors(), "errors reported on failure")
	if !ok {
		last, logged := mock.LastLog()
		core.AssertMustTrue(t, logged, "log on failure")
		core.AssertContains(t, last, tc.output, "output logged on failure")
	}
}

func newCheckFatalExitTestCase(name string, status int, output, text string,
	want bool) checkFatalExitTestCase {
	return checkFatalExitTestCase{
		name:   name,
		status: status,
		output: output,
		text:   text,
		want:   want,
	}
}

// newCheckFatalExitPassCase builds a row that checkFatalExit accepts.
func newCheckFatalExitPassCase(name string, status int, output, text string) checkFatalExitTestCase {
	return newCheckFatalExitTestCase(name, status, output, text, true)
}

// newCheckFatalExitFailCase builds a row that checkFatalExit rejects.
func newCheckFatalExitFailCase(name string, status int, output, text string) checkFatalExitTestCase {
	return newCheckFatalExitTestCase(name, status, output, text, false)
}

func checkFatalExitTestCases() []checkFatalExitTestCase {
	return []checkFatalExitTestCase{
		newCheckFatalExitPassCase("status 1 with the text", 1, "fatal text\n", "fatal text"),
		newCheckFatalExitFailCase("status 1 without the text", 1, "other text\n", "fatal text"),
		newCheckFatalExitFailCase("status 1 with an empty text", 1, "fatal text\n", ""),
		newCheckFatalExitFailCase("status 2 with the text", 2, "fatal text\n", "fatal text"),
		newCheckFatalExitFailCase("status 3 with the text", 3, "fatal text\n", "fatal text"),
		newCheckFatalExitFailCase("status 0 with the text", 0, "fatal text\n", "fatal text"),
	}
}

func TestCheckFatalExit(t *testing.T) {
	core.RunTestCases(t, checkFatalExitTestCases())
}
