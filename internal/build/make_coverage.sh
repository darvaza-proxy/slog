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

# Extract package name from test output line
extract_package_name() {
	local line="$1"
	# Handle both formats: "ok <package>" and "<package>"
	echo "$line" | sed -n 's/^ok[[:space:]]\+\([^[:space:]]\+\)[[:space:]].*/\1/p; s/^[[:space:]]*\([^[:space:]]\+\)[[:space:]].*/\1/p' | head -1
}

# Extract coverage percentage from test output line
extract_coverage_percentage() {
	local line="$1"
	echo "$line" | sed -n 's/.*coverage: \([0-9.]\+%\).*/\1/p'
}

# Convert full package path to module-relative subpackage path
convert_to_subpackage_path() {
	local package="$1"
	local module_path="$2"
	local module_name="$3"
	local path="$package"

	if [ -n "$module_path" ]; then
		sub_package="${package#"$module_path/"}"

		if [ "$package" = "$module_path" ]; then
			# root
			path="$module_name"
		elif [ "$package" != "$sub_package" ]; then
			# sub package
			path="$module_name/$sub_package"
		fi
	fi

	echo "$path"
}

# Helper function to format coverage output
format_coverage_output() {
	local stdout_file="$1"
	local module_name="$2"
	local module_dir="$3"

	# Read module path from go.mod
	local module_path=""
	if [ -f "$module_dir/go.mod" ]; then
		module_path=$(awk '/^module / {print $2}' "$module_dir/go.mod")
	fi

	if [ -s "$stdout_file" ]; then
		grep -E 'coverage: [0-9.]+%' "$stdout_file" | while read -r line; do
			package=$(extract_package_name "$line")
			coverage=$(extract_coverage_percentage "$line")
			subpackage=$(convert_to_subpackage_path "$package" "$module_path" "$module_name")

			printf "coverage: %-30s %s\n" "$subpackage" "$coverage"
		done
	else
		echo "coverage: $module_name: no test output"
	fi
}

# Use absolute path for coverage directory
mkdir -p "$COVERAGE_DIR"
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
	format_coverage_output "$COVERSTDOUT" "$MODULE_NAME" "$MODULE_DIR"
else
	exit_code=$?
	echo "Tests failed for $MODULE_NAME" >&2
	# Show failed test output
	if [ -s "$COVERSTDOUT" ]; then
		grep -aE '(FAIL|Error:|panic:|^---)|^[[:space:]]+' "$COVERSTDOUT" | tail -20 >&2
	fi
	exit $exit_code
fi
