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
	if [ "0a" != "$last_byte" ]; then
		printf '\n' >> "$file"
	elif [ "$(wc -c < "$file")" -eq 1 ]; then
		# File only contains a newline, truncate it
		: > "$file"
	fi
}

# Helper function to run find with auto-pruning
run_find() {
	local paths= quoted=

	# Collect paths until we hit a find option (starts with '-')
	while [ $# -gt 0 ]; do
		case "$1" in
		-*)
			# Found a find option, stop collecting paths
			break
			;;
		*)
			# Add path with proper escaping for spaces and special chars
			quoted=$(printf '%s' "$1" | sed -e "s/'/'\\\\''/g" -e "s/^/'/" -e "s/$/'/")
			paths="${paths:+$paths }$quoted"
			shift
			;;
		esac
	done

	# Wrap user conditions in parentheses if they exist
	[ $# -eq 0 ] || set -- \( "$@" \)
	# combine auto-pruning and user conditions
	set -- \( -name .git -o -name node_modules \) -prune -o "$@" -type f
	# combine escaped paths with find options
	eval "set -- ${paths:-.} \"\$@\""

	find "$@" -print0 | xargs -0 -r "$0" --
}

if [ "${1:-}" = "--" ]; then
	# Explicit file mode
	shift
	for file; do
		fix_file "$file"
	done
else
	# Find mode with arguments
	run_find "$@"
fi
