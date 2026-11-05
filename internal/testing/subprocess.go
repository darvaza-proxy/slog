package testing

// cSpell:words GOCOVERDIR

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"darvaza.org/core"
)

// subprocessEnv names the environment variable that marks the child
// process of [RunInSubprocess], set to the name of the test that runs fn.
const subprocessEnv = "SLOG_TEST_SUBPROCESS"

// subprocessReturned is the exit status of a child whose fn returned.
const subprocessReturned = 3

// RunInSubprocess runs fn in a child process, and returns the child's exit
// status and combined output, so a test can see fn end the program as
// Fatal does. The child runs the calling test again, marked by an
// environment variable set to its name, and reaches the call, where it
// runs fn and exits with status 3 if fn returns.
//
// The child repeats whatever the test does before the call, so the call
// comes first in its test, once. The child gets the parent's environment,
// where `go test -cover` sets GOCOVERDIR, so a child that ends through
// os.Exit adds its coverage. It is stopped at the test's deadline, when
// the test has one, and its own -test.timeout ends there too, so it
// doesn't outlive a parent that dies first.
func RunInSubprocess(t *testing.T, fn func()) (status int, output string) {
	t.Helper()

	if os.Getenv(subprocessEnv) == t.Name() {
		fn()
		// the child ends here, with the status that says fn returned.
		// revive:disable-next-line:deep-exit
		os.Exit(subprocessReturned)
	}

	return runSubprocess(t)
}

// runSubprocess runs the test binary again for the test t, marked as the
// child, and waits for it to end.
func runSubprocess(t *testing.T) (status int, output string) {
	t.Helper()

	exe, err := os.Executable()
	core.AssertMustNoError(t, err, "test binary")

	ctx := t.Context()
	args := []string{"-test.run=" + subprocessPattern(t.Name())}
	if deadline, ok := t.Deadline(); ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()

		// the child times out on its own too, so it ends even if the
		// parent dies before it can stop the child.
		args = append(args, "-test.timeout="+time.Until(deadline).String())
	}

	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Env = append(os.Environ(), subprocessEnv+"="+t.Name())

	out, err := cmd.CombinedOutput()
	if err != nil {
		exitErr := core.AssertMustTypeIs[*exec.ExitError](t, err, "child exit")
		return exitErr.ExitCode(), string(out)
	}
	return 0, string(out)
}

// subprocessPattern returns the -test.run pattern that selects the test
// named name, with each level of the name quoted and anchored.
func subprocessPattern(name string) string {
	levels := strings.Split(name, "/")
	for i, level := range levels {
		levels[i] = "^" + regexp.QuoteMeta(level) + "$"
	}
	return strings.Join(levels, "/")
}

// AssertFatalExit runs fn in a child process through [RunInSubprocess],
// and verifies that it ends the program as Fatal does: exit status 1, with
// text in the output. A test that fails in the child also exits with
// status 1, so text is what tells the two apart. As with
// [RunInSubprocess], the call comes first in its test.
// Returns true if the assertion passed, false otherwise.
func AssertFatalExit(t *testing.T, fn func(), text string) bool {
	t.Helper()
	status, output := RunInSubprocess(t, fn)
	return checkFatalExit(t, status, output, text)
}

// checkFatalExit verifies a child's exit status and output for
// [AssertFatalExit], and logs the output when either is wrong.
func checkFatalExit(t core.T, status int, output, text string) bool {
	t.Helper()
	ok := core.AssertEqual(t, 1, status, "exit status")
	if !core.AssertContains(t, output, text, "output") {
		ok = false
	}
	if !ok {
		t.Logf("output:\n%s", output)
	}
	return ok
}
