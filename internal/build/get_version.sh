#!/bin/sh
# shellcheck disable=SC1007,SC3043 # empty assignments and local usage

set -eu

xy_ver() {
	local ver="${1:-}"
	local x y

	if echo "v${ver}" | grep -qe '^v[0-9]\+\.[0-9]\+\(\..*\|$\)'; then
		x="$(echo "$ver" | cut -d. -f1)"
		y="$(echo "$ver" | cut -d. -f2)"
		echo $((x * 1000 + y))
	else
		echo "invalid version format: $ver" >&2
		return 1
	fi
}

GO_VERSION=$(${GO:-go} version | sed -ne 's|.* go\([0-9][^ ]\+\)[ $].*|\1|p')

if [ $# -eq 0 ]; then
	# no arguments, go version
	echo "$GO_VERSION"
	exit 0
fi

if [ $# -gt 1 ]; then
	# base go version and at least one target
	GO_VER=$(xy_ver "$GO_VERSION")
	[ -n "$GO_VER" ] || exit 1

	BASE_VER=$(xy_ver "$1")
	[ -n "$BASE_VER" ] || exit 1

	shift
	if [ "$GO_VER" -ge "$BASE_VER" ]; then
		VER="$BASE_VER"

		for value; do
			[ "$GO_VER" -gt "$VER" ] || break

			: $(( VER = VER + 1 ))
		done

		# match or last
		echo "$value"
		exit 0
	fi
fi

echo "unknown"
exit 1
