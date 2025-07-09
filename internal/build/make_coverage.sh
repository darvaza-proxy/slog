#!/bin/sh
# make_coverage.sh - Execute coverage tests directly from .tmp/index
#
# Usage: make_coverage.sh [index_file] [coverage_dir]

set -eu

INDEX="${1:-.tmp/index}"
COVERAGE_DIR="${2:-.coverage}"

if [ ! -f "$INDEX" ]; then
	echo "Error: Index file not found: $INDEX" >&2
	echo "Run 'make .tmp/index' first" >&2
	exit 1
fi

# Create coverage directory
mkdir -p "$COVERAGE_DIR"

# Run tests for each module using make targets
echo "Running coverage tests..."
n=1
failed=0
# shellcheck disable=SC2013
for name in $(cut -d: -f1 "$INDEX"); do
	[ -z "$name" ] && continue

	COVERPROFILE="$COVERAGE_DIR/coverage_${n}_${name}.out"
	if ! make "test-${name}" \
		"GOTEST_FLAGS=${GOTEST_FLAGS:+$GOTEST_FLAGS }-coverprofile=$COVERPROFILE"; then
		echo "⚠️  ${name} tests failed" >&2
		failed=1
	fi

	n=$((n + 1))
done

# Merge coverage files
echo
echo "Merging coverage files..."
set -- "$COVERAGE_DIR"/coverage_*.out

if [ $# -eq 0 ]; then
	echo "No coverage files found" >&2
	exit 1
fi

if command -v gocovmerge >/dev/null 2>&1; then
	echo "Using gocovmerge..."
	gocovmerge "$@" > "$COVERAGE_DIR/coverage.out"
else
	echo "Manual merge (no gocovmerge found)..."
	# Manual merge
	head -1 "$1" > "$COVERAGE_DIR/coverage.out"
	for f; do
		tail -n +2 "$f" >> "$COVERAGE_DIR/coverage.out"
	done
fi

# Show summary
echo
echo "Coverage summary:"
${GO:-go} tool cover -func="$COVERAGE_DIR/coverage.out" | tail -1

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
