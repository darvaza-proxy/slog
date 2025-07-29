#!/bin/sh
# shellcheck disable=SC1007,SC3043 # empty assignments and local usage
#
# merge_coverage.sh - Merge multiple Go coverage profiles into one
#
# This script merges multiple coverage profile files into a single file.
# It takes the header from the first file and appends all data lines from
# all files.
#
# Usage: merge_coverage.sh <input_files...>
#
# Example: merge_coverage.sh coverage_*.prof > coverage.out

set -eu

# Check if we have any input files
if [ $# -eq 0 ]; then
	echo "Error: No input files provided" >&2
	exit 1
fi

# Get header from first existing file
if [ -s "$1" ]; then
	head -1 "$1"
else
	echo "Error: First input file is empty or does not exist: $1" >&2
	exit 1
fi

# Append data lines from all files
for f; do
	if [ -s "$f" ]; then
		tail -n +2 "$f"
	fi
done
