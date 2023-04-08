#!/bin/sh

set -eu

: ${GO:=go}

MODULES=$(find * -name go.mod -exec dirname '{}' \;)
GROUPS="x handlers"
BASE="$PWD"
MODULE=$($GO list)

namedir() {
	local d="$1" g= n=

	if [ "." = "$d" ]; then
		echo "root"
		return
	fi

	for g in $GROUPS; do
		n="${d#$g/}"
		if [ "x$n" != "x$d" ]; then
			echo "$n" | tr '/' '-'
			return
		fi
	done

	echo "$d" | tr '/' '-'
}

mod() {
	cd "$1"
	$GO list
}

mod_replace() {
	local d="$1"
	grep "=>" "$d/go.mod" | sed -n -e "s;^.*\($MODULE.*\)[ \t]\+=>.*;\1;p"
}

gen_index() {
	local d= n=

	for d; do
		n=$(namedir "$d")
		m=$(mod "$d")
		echo "$n:$d:$m"
	done
}

INDEX=$(gen_index $MODULES)

echo "$INDEX" | while IFS=: read name dir mod; do
	deps=
	for dep in $(mod_replace "$dir"); do
		depname=$(echo "$INDEX" | grep ":$dep$" | cut -d: -f1)
		if [ -n "$depname" ]; then
			deps="${deps:+$deps,}$depname"
		fi
	done

	echo "$name:$dir:$mod:$deps"
done | sort -V
