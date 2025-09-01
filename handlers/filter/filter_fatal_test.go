package filter

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// TestParentlessFatalExit tests Fatal() with parentless logger which calls os.Exit(1)
// This covers entry.go:122 - the only remaining uncovered line
func TestParentlessFatalExit(t *testing.T) {
	// Check if we're in the subprocess
	if val, _ := os.LookupEnv("TEST_FATAL_SUBPROCESS"); val == "1" {
		// We're in the subprocess - run the code that will exit
		logger := NewNoop() // Creates parentless logger with Fatal threshold
		entry := logger.Fatal()
		entry.Print("fatal message") // This will call os.Exit(1)
		return                       // Should never reach here
	}

	// We're in the main test process - execute test binary as subprocess
	cmd := exec.Command(os.Args[0], "-test.run=^TestParentlessFatalExit$")
	cmd.Env = append(os.Environ(), "TEST_FATAL_SUBPROCESS=1")

	// Capture output
	output, err := cmd.CombinedOutput()

	// Check that the process exited with error
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected process to exit with error, got: %v", err)
	}

	// Check exit code is 1 (default for os.Exit(1))
	exitCode := exitErr.ExitCode()
	core.AssertEqual(t, 1, exitCode, "exit code")

	// Check that the fatal message was logged to stderr
	outputStr := string(output)
	core.AssertTrue(t, strings.Contains(outputStr, "fatal message"), "output contains message")
}

// TestParentlessFatalWithFields tests Fatal() with fields attached
func TestParentlessFatalWithFields(t *testing.T) {
	// Check if we're in the subprocess
	if val, _ := os.LookupEnv("TEST_FATAL_FIELDS_SUBPROCESS"); val == "1" {
		// We're in the subprocess - run the code that will exit
		logger := New(nil, slog.Fatal) // Parentless logger
		entry := logger.Fatal().
			WithField("code", 500).
			WithField("error", "critical failure")
		entry.Print("system fatal error") // This will call os.Exit(1)
		return                            // Should never reach here
	}

	// We're in the main test process - execute test binary as subprocess
	cmd := exec.Command(os.Args[0], "-test.run=^TestParentlessFatalWithFields$")
	cmd.Env = append(os.Environ(), "TEST_FATAL_FIELDS_SUBPROCESS=1")

	// Capture output
	output, err := cmd.CombinedOutput()

	// Check that the process exited with error
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected process to exit with error, got: %v", err)
	}

	// Check exit code is 1
	exitCode := exitErr.ExitCode()
	core.AssertEqual(t, 1, exitCode, "exit code")

	// Check that the message was logged
	outputStr := string(output)
	core.AssertTrue(t, strings.Contains(outputStr, "system fatal error"), "output contains message")
}

// TestFatalBehaviourVerification verifies Fatal level behaviour without actually exiting
func TestFatalBehaviourVerification(t *testing.T) {
	// Test that Fatal level is always enabled for parentless logger
	logger := NewNoop()
	entry := logger.Fatal()
	core.AssertTrue(t, entry.Enabled(), "Fatal always enabled for parentless")

	// Test that WithEnabled returns true for Fatal
	enabledEntry, ok := entry.WithEnabled()
	core.AssertTrue(t, ok, "Fatal WithEnabled returns true")
	core.AssertNotNil(t, enabledEntry, "Fatal WithEnabled returns entry")

	// Test that Fatal level passes threshold check
	logger2 := New(nil, slog.Error) // Sets to Fatal for nil parent
	entry2 := logger2.Fatal()
	core.AssertTrue(t, entry2.Enabled(), "Fatal enabled even with Error threshold")
}
