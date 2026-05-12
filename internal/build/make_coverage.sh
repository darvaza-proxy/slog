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

mkdir -p "$COVERAGE_DIR"

COVERAGE_DIR=$(cd "$COVERAGE_DIR" && pwd)
COVERAGE_BASE="${COVERAGE_DIR}/coverage_${MODULE_NAME}"
COVERAGE_TMP_BASE="${COVERAGE_DIR}/.coverage_${MODULE_NAME}"

# Clean-up
rm -rf "${COVERAGE_BASE}"*
rm -rf "${COVERAGE_TMP_BASE}"*

# Extract package name from test output line
extract_package_name() {
	local line="$1"
	# Handle both formats: "ok <package>" and "<package>"
	echo "$line" | sed -n 's/^ok[[:space:]]\+\([^[:space:]]\+\)[[:space:]].*/\1/p; s/^[[:space:]]*\([^[:space:]]\+\)[[:space:]].*/\1/p' | head -1
}

# Extract coverage percentage from test output line
extract_coverage_percentage() {
	local line="$1"
	echo "$line" | sed -n 's/.*coverage: \([0-9.]\+\)%.*/\1/p'
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

# Helper function to format dual coverage output
format_dual_coverage_output() {
	local module_stdout="$1"
	local self_stdout="$2"
	local module_name="$3"
	local module_dir="$4"

	# Read module path from go.mod
	local module_path=""
	if [ -f "$module_dir/go.mod" ]; then
		module_path=$(awk '/^module / {print $2}' "$module_dir/go.mod")
	fi

	if [ -s "$module_stdout" ]; then
		local self_cov_file="${COVERAGE_TMP_BASE}_self_coverage"
		local processed_packages="${COVERAGE_TMP_BASE}_packages"

		# Extract self-coverage percentages by package (take only the last occurrence)
		if [ -s "$self_stdout" ]; then
			grep -E 'coverage: [0-9.]+%' "$self_stdout" | while read -r line; do
				package=$(extract_package_name "$line")
				coverage=$(extract_coverage_percentage "$line")
				subpackage=$(convert_to_subpackage_path "$package" "$module_path" "$module_name")

				# Remove previous entry for this package and add new one
				grep -v "^$subpackage:" "$self_cov_file" 2>/dev/null > "${self_cov_file}~" || true
				echo "$subpackage:$coverage" >> "${self_cov_file}~"
				mv "${self_cov_file}~" "$self_cov_file"
			done
		fi

		# Process module coverage and combine with self coverage
		# Sort packages by directory structure (depth first, then alphabetically)
		grep -E 'coverage: [0-9.]+%' "$module_stdout" | while read -r line; do
			package=$(extract_package_name "$line")
			module_cov=$(extract_coverage_percentage "$line")
			subpackage=$(convert_to_subpackage_path "$package" "$module_path" "$module_name")

			# Skip if already processed
			if grep -q "^$subpackage$" "$processed_packages" 2>/dev/null; then
				continue
			fi
			echo "$subpackage" >> "$processed_packages"

			# Look up self coverage for this package
			self_cov=""
			if [ -f "$self_cov_file" ]; then
				self_cov=$(grep "^$subpackage:" "$self_cov_file" 2>/dev/null | cut -d: -f2 || echo "")
			fi

			# Display format based on whether we have self coverage
			if [ -n "$self_cov" ] && [ "$self_cov" != "0.0" ]; then
				# Both self and module coverage (only show dual when self-coverage exists)
				printf "coverage: %-30.30s %5.1f%% (%4.1f%%)\n" "$subpackage" "$self_cov" "$module_cov"
			else
				# Only module coverage (for packages with no Go files or 0% self coverage)
				printf "coverage: %-30.30s %5.1f%%\n" "$subpackage" "$module_cov"
			fi
		done | sort -k2

		# Cleanup temp files
		rm -f "${self_cov_file}~" "${processed_packages}~"
		rm -f "${self_cov_file}" "${processed_packages}"
	else
		echo "coverage: $module_name: no test output"
	fi
}

# Output files
COVERPROFILE="$COVERAGE_BASE.prof"
COVERFUNC="$COVERAGE_BASE.func"
COVERHTML="$COVERAGE_BASE.html"
COVERSTDOUT="$COVERAGE_BASE.stdout"

COVERPROFILE_SELF="${COVERAGE_BASE}_self.prof"
COVERFUNC_SELF="${COVERAGE_BASE}_self.func"
COVERSTDOUT_SELF="${COVERAGE_BASE}_self.stdout"

# shellcheck disable=SC2086 # GOTEST_FLAGS splitting intended
set -- ${GOTEST_FLAGS:-} "-covermode=atomic"

# Run tests with coverage
if ${GO:-go} -C "$MODULE_DIR" test "$@" "-coverprofile=$COVERPROFILE" "-coverpkg=./..." ./... > "$COVERSTDOUT" 2>&1; then
	# Run per-package self-coverage
	# Use go list to find all packages with Go files (excluding test-only packages)
	${GO:-go} -C "$MODULE_DIR" list -f '{{if .GoFiles}}{{.ImportPath}}{{end}}' ./... | while read -r pkg; do
		if [ -n "$pkg" ]; then
			pkg_prof_file="${COVERAGE_BASE}/$pkg.prof"
			mkdir -p "${pkg_prof_file%/*}"

			${GO:-go} -C "$MODULE_DIR" test "$@" "-coverprofile=$pkg_prof_file" \
				"$pkg" >> "$COVERSTDOUT_SELF" 2>&1 || true
		fi
	done

	# merge self-coverage profiles
	find "${COVERAGE_BASE}" -name "*.prof" -type f -print0 \
		| xargs -r0 "$(dirname "$0")/merge_coverage.sh" > "$COVERPROFILE_SELF"

	# Generate function coverage reports
	${GO:-go} -C "$MODULE_DIR" tool cover -func="$COVERPROFILE" -o "$COVERFUNC" 2>/dev/null || true
	${GO:-go} -C "$MODULE_DIR" tool cover -func="$COVERPROFILE_SELF" -o "$COVERFUNC_SELF" 2>/dev/null || true

	# Generate HTML coverage report using self-coverage for accurate development view
	${GO:-go} -C "$MODULE_DIR" tool cover -html="$COVERPROFILE_SELF" -o "$COVERHTML" 2>/dev/null || true

	# Display dual coverage output
	format_dual_coverage_output "$COVERSTDOUT" "$COVERSTDOUT_SELF" "$MODULE_NAME" "$MODULE_DIR"
else
	exit_code=$?
	echo "Tests failed for $MODULE_NAME" >&2
	# Show failed test output
	if [ -s "$COVERSTDOUT" ]; then
		grep -aE '(FAIL|Error:|panic:|^---)|^[[:space:]]+' "$COVERSTDOUT" | tail -20 >&2
	fi
	exit $exit_code
fi
