#!/bin/sh
# shellcheck disable=SC1007,SC3043 # empty assignments, -o and local usage
#
# make_coverage.sh - Execute coverage tests for all modules in a monorepo
#
# This script reads module information from an index file and runs Go tests
# with coverage for each module. It then merges all coverage files into a
# single coverage report.
#
# Usage: make_coverage.sh [index_file] [coverage_dir]
#
# Arguments:
#   index_file   - Path to the index file containing module information
#                  (default: .tmp/index)
#   coverage_dir - Directory to store coverage files
#                  (default: .coverage)
#
# Environment variables:
#   GO           - Go command (default: go)
#   GOTEST_FLAGS - Additional flags for 'go test'
#   COVERAGE_HTML - Set to "true" to generate HTML coverage report

set -eu

INDEX="${1:-.tmp/index}"
COVERAGE_DIR="${2:-.coverage}"

if [ ! -s "$INDEX" ]; then
	echo "Error: Index file not found: $INDEX" >&2
	echo "Run 'make .tmp/index' first" >&2
	exit 1
fi

# Create coverage directory
rm -rf "$COVERAGE_DIR"
mkdir -p "$COVERAGE_DIR"
# and make it absolute, because of `go -C`
COVERAGE_DIR=$(cd "$COVERAGE_DIR" && pwd)

# Count total modules
total=$(grep -c '^[^:]*:[^:]*' "$INDEX" || true)

# Run tests for each module
echo "Running coverage tests..."
n=1
failed=0
while IFS=: read -r name dir _rest; do
	[ -n "$name" ] || continue

	# Show progress
	printf "[$n/$total] Testing module: %-12s " "$name"

	COVERPROFILE="$COVERAGE_DIR/coverage_${n}_${name}.prof"
	COVEROUTPUT="$COVERAGE_DIR/test_${n}_${name}.out"

	# shellcheck disable=SC2086 # GOTEST_FLAGS splitting intended
	set -- ${GOTEST_FLAGS:-} "-coverprofile=$COVERPROFILE" "-coverpkg=./..." ./...

	# Run tests quietly, capturing output to file
	if ${GO:-go} -C "$dir" test "$@" > "$COVEROUTPUT" 2>&1; then
		# Extract coverage percentage from output
		coverage=$(grep -aE 'coverage: [0-9.]+% of statements' "$COVEROUTPUT" | tail -1 | sed 's/.*coverage: //')
		printf "✓ %-20s %s\n" "($dir)" "${coverage:-no coverage}"

		# Generate HTML report for individual module if requested
		if [ "${COVERAGE_HTML:-}" = "true" ]; then
			${GO:-go} -C "$dir" tool cover "-html=$COVERPROFILE" -o "$COVERPROFILE.html" 2>/dev/null || true
		fi
	else
		printf "✗ FAILED\n"
		echo "⚠️  ${name} tests failed:" >&2
		grep -aE '(FAIL|Error:|panic:|^---)|^\s+' "$COVEROUTPUT" | tail -20 >&2
		failed=1
	fi

	n=$((n + 1))
done < "$INDEX"

# Merge coverage files
echo
echo "Generating coverage.out..."
set -- "$COVERAGE_DIR"/coverage_*.prof

if [ ! -f "${1:-}" ]; then
	echo "No coverage files found" >&2
	exit 1
fi

# Simple merge: header from first file, then all data lines
head -1 "$1" > "$COVERAGE_DIR/coverage.out"
for f; do
	tail -n +2 "$f" >> "$COVERAGE_DIR/coverage.out"
done

# Optional: generate HTML report
if [ "${COVERAGE_HTML:-}" = "true" ]; then
	echo
	echo "Generating HTML coverage report..."
	${GO:-go} tool cover -html="$COVERAGE_DIR/coverage.out" -o "$COVERAGE_DIR/coverage.html"
	echo "HTML report saved to $COVERAGE_DIR/coverage.html"
fi

# Exit with failure if any tests failed
if [ "$failed" -ne 0 ]; then
	echo "Some tests failed" >&2
	exit 1
fi
