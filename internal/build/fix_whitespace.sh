#!/bin/sh
# shellcheck disable=SC1007,SC3043 # empty assignments and local usage
# fix_whitespace.sh - Find files and fix trailing whitespace and EOF newlines
#
# Usage: fix_whitespace.sh [find arguments]
#        fix_whitespace.sh -- file1 file2 ...
#
# Automatically prunes .git and node_modules directories
#
# Environment Variables:
#   SED - sed command to use (default: sed)
#         Set to "gsed" on macOS or "sed -i ''" for BSD compatibility
#
# Examples:
#   fix_whitespace.sh . -name '*.go' -o -name '*.md'
#   fix_whitespace.sh src/ -name '*.js'
#   fix_whitespace.sh -- README.md LICENCE.txt
#   SED="gsed" fix_whitespace.sh . -name '*.md'

set -eu

# Function to fix a single file
fix_file() {
	local file="$1" last_byte

	# Skip if not a regular file
	[ -f "$file" ] || return 0

	# Remove trailing whitespace
	${SED:-sed} -i 's/[[:space:]]*$//' "$file"

	# Leave empty files alone
	[ -s "$file" ] || return 0


	# Check last byte to see if file ends with newline
	# Use od to get hexadecimal representation of last byte
	last_byte=$(tail -c 1 "$file" | od -An -tx1 | tr -d ' \t')

	# If last byte is not newline (0x0a), add one
	if [ "x$last_byte" != "x0a" ]; then
		printf '\n' >> "$file"
	elif [ "$(wc -c < "$file")" -eq 1 ]; then
		# File only contains a newline, truncate it
		: > "$file"
	fi
}

# Helper function to run find with auto-pruning
run_find() {
	local paths=

	# Collect paths until we hit a find option (starts with -)
	while [ $# -gt 0 ] && [ "${1#-}" = "$1" ]; do
		paths="$paths $1"
		shift
	done

	# Default to current directory if no paths specified
	if [ -z "$paths" ]; then
		paths="."
	fi

	# Wrap user conditions in parentheses if they exist
	[ $# -eq 0 ] || set -- \( "$@" \)

	# Execute find with auto-pruning and user conditions
	# shellcheck disable=SC2086 # intentional word splitting for paths
	find $paths \( -name .git -o -name node_modules \) -prune -o "$@" -type f -print0 | xargs -0 -r "$0" --
}

# Handle different argument patterns
if [ $# -eq 0 ]; then
	# No arguments - search current directory
	run_find .
elif [ "${1:-}" = "--" ]; then
	# Explicit file mode
	shift
	for file; do
		fix_file "$file"
	done
else
	# Find mode with arguments
	run_find "$@"
fi
