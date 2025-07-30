#!/bin/sh
# shellcheck disable=SC1007,SC3043 # empty assignments and local usage
#
# make_coverage.sh - Execute coverage tests for a single module
#
# This script runs Go tests with coverage for a single module and generates
# coverage profile (.prof), function coverage (.func), and HTML report (.html).
#
# Usage: make_coverage.sh <module_name> <module_dir> <coverage_dir>
#
# Arguments:
#   module_name  - Name of the module (e.g., "root", "cmp", "config")
#   module_dir   - Directory containing the module (e.g., ".", "cmp", "config")
#   coverage_dir - Directory to store coverage files
#
# Environment variables:
#   GO           - Go command (default: go)
#   GOTEST_FLAGS - Additional flags for 'go test'

set -eu

MODULE_NAME="${1:?Module name required}"
MODULE_DIR="${2:?Module directory required}"
COVERAGE_DIR="${3:?Coverage directory required}"

# Helper function to format coverage output
format_coverage_output() {
	local stdout_file="$1"
	local module_name="$2"

	if [ -s "$stdout_file" ]; then
		# Extract coverage percentage and format with module name
		grep -E 'coverage: [0-9.]+%' "$stdout_file" | tail -1 | \
			sed "s|coverage: \([0-9.]\+%\) of statements in \./\.\.\.|Coverage: \1 of statements in $module_name|" || \
			echo "Coverage: no coverage data"
	else
		echo "Coverage: no test output"
	fi
}

# Use absolute path for coverage directory
COVERAGE_DIR=$(cd "$COVERAGE_DIR" && pwd)

# Output files
COVERPROFILE="$COVERAGE_DIR/coverage_${MODULE_NAME}.prof"
COVERFUNC="$COVERAGE_DIR/coverage_${MODULE_NAME}.func"
COVERHTML="$COVERAGE_DIR/coverage_${MODULE_NAME}.html"
COVERSTDOUT="$COVERAGE_DIR/coverage_${MODULE_NAME}.stdout"

# shellcheck disable=SC2086 # GOTEST_FLAGS splitting intended
set -- ${GOTEST_FLAGS:-} \
	"-covermode=atomic" \
	"-coverprofile=$COVERPROFILE" \
	"-coverpkg=./..." \
	./...

# Run tests with coverage
# Note: The makefile already cd's into the module directory before calling this script
if ${GO:-go} test "$@" > "$COVERSTDOUT" 2>&1; then
	# Generate function coverage report
	${GO:-go} -C "$MODULE_DIR" tool cover -func="$COVERPROFILE" > "$COVERFUNC" 2>/dev/null || true

	# Generate HTML coverage report
	${GO:-go} -C "$MODULE_DIR" tool cover -html="$COVERPROFILE" -o "$COVERHTML" 2>/dev/null || true

	# Display formatted coverage output
	format_coverage_output "$COVERSTDOUT" "$MODULE_NAME"
else
	exit_code=$?
	echo "Tests failed for $MODULE_NAME" >&2
	# Show failed test output
	if [ -s "$COVERSTDOUT" ]; then
		grep -aE '(FAIL|Error:|panic:|^---)|^[[:space:]]+' "$COVERSTDOUT" | tail -20 >&2
	fi
	exit $exit_code
fi
