package testing_test

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"darvaza.org/core"
	slogtest "darvaza.org/slog/internal/testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = runInSubprocessTestCase{}

// runInSubprocessTestCase runs a function in a child process and checks
// how the child ended.
type runInSubprocessTestCase struct {
	fn         func()
	name       string
	wantOutput string
	wantStatus int
}

func (tc runInSubprocessTestCase) Name() string {
	return tc.name
}

func (tc runInSubprocessTestCase) Test(t *testing.T) {
	status, output := slogtest.RunInSubprocess(t, tc.fn)

	core.AssertEqual(t, tc.wantStatus, status, "exit status")
	core.AssertContains(t, output, tc.wantOutput, "output")
}

func newRunInSubprocessTestCase(name string, fn func(), wantStatus int,
	wantOutput string) runInSubprocessTestCase {
	return runInSubprocessTestCase{
		name:       name,
		fn:         fn,
		wantStatus: wantStatus,
		wantOutput: wantOutput,
	}
}

func runInSubprocessTestCases() []runInSubprocessTestCase {
	return []runInSubprocessTestCase{
		newRunInSubprocessTestCase("exits with 1", exitWithText("exit one", 1), 1, "exit one"),
		newRunInSubprocessTestCase("exits with 0", exitWithText("exit zero", 0), 0, "exit zero"),
		newRunInSubprocessTestCase("panics", panicWithText("panic text"), 2, "panic text"),
		newRunInSubprocessTestCase("returns", returnWithText("returned"), 3, "returned"),
		newRunInSubprocessTestCase("returns from a.b(c)|[d]", returnWithText("quoted"), 3, "quoted"),
	}
}

func TestRunInSubprocess(t *testing.T) {
	core.RunTestCases(t, runInSubprocessTestCases())
}

func TestAssertFatalExit(t *testing.T) {
	ok := slogtest.AssertFatalExit(t, exitWithText("fatal text", 1), "fatal text")
	core.AssertTrue(t, ok, "result")
}

// TestRunInSubprocessTimeout checks that the child has a timeout of its
// own when the test has a deadline, so it ends even if the parent dies
// first.
func TestRunInSubprocessTimeout(t *testing.T) {
	status, output := slogtest.RunInSubprocess(t, printHasTimeout)

	_, hasDeadline := t.Deadline()
	core.AssertEqual(t, 3, status, "exit status")
	core.AssertContains(t, output, fmt.Sprintf("has timeout: %t", hasDeadline), "output")
}

// exitWithText returns a function that writes text to standard error,
// then exits with status.
func exitWithText(text string, status int) func() {
	return func() {
		log.New(os.Stderr, "", 0).Print(text)
		// the child process ends here, with the row's status.
		// revive:disable-next-line:deep-exit
		os.Exit(status)
	}
}

// panicWithText returns a function that panics with text.
func panicWithText(text string) func() {
	return func() {
		panic(text)
	}
}

// returnWithText returns a function that writes text to standard output,
// then returns.
func returnWithText(text string) func() {
	return func() {
		log.New(os.Stdout, "", 0).Print(text)
	}
}

// printHasTimeout writes to standard output whether the test binary has
// a timeout, then returns.
func printHasTimeout() {
	// the flag's value is a Duration, and its String parses back.
	timeout := core.Must(time.ParseDuration(flag.Lookup("test.timeout").Value.String()))
	log.New(os.Stdout, "", 0).Printf("has timeout: %t", timeout > 0)
}
