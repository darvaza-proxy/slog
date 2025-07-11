#!/bin/sh
# shellcheck disable=SC1007,SC3043 # empty assignments and local usage

set -eu

: "${GO:=go}"

# list of directories with go.mod
MODULES=$(find ./* -name go.mod | sed -e 's;^\./;;' | tr '\n' '\0' | xargs -n1 -0r dirname)
# shellcheck disable=2178 # space delimited list of grouping prefixes
GROUPS="handlers"

mod() {
	local d="${1:-.}"
	grep ^module "$d/go.mod" | cut -d' ' -f2
}

namedir() {
	local d="$1" g= n=

	if [ "." = "$d" ]; then
		echo "root"
		return
	fi

	# shellcheck disable=2086,2128 # word splitting of $GROUPS intended, not array.
	for g in $GROUPS; do
		n="${d#"$g/"}"
		if [ "$n" != "$d" ]; then
			echo "$n" | tr '/' '-'
			return
		fi
	done

	echo "$d" | tr '/' '-'
}

mod_replace() {
	local d="$1"
	grep "=>" "$d/go.mod" | sed -n -e "s;^.*\($ROOT_MODULE.*\)[ \t]\+=>.*;\1;p"
}

gen_index() {
	local d= n=

	for d; do
		n=$(namedir "$d")
		m=$(mod "$d")
		echo "$n:$d:$m"
	done
}

ROOT_MODULE=$(mod)
# shellcheck disable=SC2086 # word splitting intended
INDEX=$(gen_index $MODULES)

echo "$INDEX" | while IFS=: read -r name dir mod; do
	deps=
	for dep in $(mod_replace "$dir"); do
		depname=$(echo "$INDEX" | grep ":$dep$" | cut -d: -f1 | tr '\n' ',' | sed -e 's|,\+$||g')
		if [ -n "$depname" ]; then
			deps="${deps:+$deps,}$depname"
		fi
	done

	echo "$name:$dir:$mod:$deps"
done | sort -V
