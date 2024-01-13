#!/bin/sh

set -eu

INDEX="$1"

PROJECTS="$(cut -d':' -f1 "$INDEX")"
COMMANDS="tidy get build test up"

expand() {
	local prefix="$1" suffix="$2"
	local x= out=
	shift 2

	for x; do
		out="${out:+$out }${prefix}$x${suffix}"
	done

	echo "$out"
}

prefixed() {
	local prefix="${1:+$1-}"
	shift
	expand "$prefix" "" "$@"
}

suffixed() {
	local suffix="${1:+-$1}"
	shift
	expand "" "$suffix" "$@"
}

# packed remove excess whitespace from lines of commands
packed() {
	sed -e 's/^[ \t]\+//' -e 's/[ \t]\+$//' -e '/^$/d;' -e '/^#/d';
}

# packet_oneline converts a multiline script into packed single-line equivalent
packed_oneline() {
	packed | tr '\n' ';' | sed -e 's|;$||' -e 's|then;|then |g' -e 's|;[ \t]*|; |g'
}

gen_install_tools() {
	cat <<EOT
for url in \$(GO_INSTALL_URLS); do \$(GO) install -v \$\$url; done
EOT
}

gen_revive_exclude() {
	local self="$1"
	local dirs= d=

	dirs="$(cut -d: -f2 "$INDEX" | grep -v '^.$' || true)"
	if [ "." != "$self" ]; then
		dirs=$(echo "$dirs" | sed -n -e "s;^$self/\(.*\)$;\1;p")
	fi

	for d in $dirs; do
		printf -- "-exclude ./$d/... "
	done
}

for cmd in $COMMANDS; do
	all="$(prefixed $cmd $PROJECTS)"
	depsx=

	cat <<EOT
.PHONY: $cmd $all
$cmd: $all

EOT

	# default calls
	case "$cmd" in
	tidy)
		call="$(cat <<-EOT | packed
		\$(GO) mod tidy

		# go vet and revive only if there are .go files
		#
		$(cat <<-EOL | packed_oneline
			set -e
			FILES="\$\$(\$(GO) list -f '{{len .GoFiles}}')"
			if [ -n "\$\$FILES" ]; then
				\$(GO) vet ./...
				\$(REVIVE) \$(REVIVE_RUN_ARGS) ./...
			fi
			EOL
			)
		EOT
		)"
		depsx="fmt \$(REVIVE)"
		;;
	up)
		call="\$(GO) get -u -v ./...
\$(GO) mod tidy"
		;;
	test)
		call="\$(GO) $cmd ./..."
		;;
	*)
		call="\$(GO) $cmd -v ./..."
		;;
	esac

	case "$cmd" in
	build|test)
		sequential=true ;;
	*)
		sequential=false ;;
	esac

	while IFS=: read name dir mod deps; do

		deps=$(echo "$deps" | tr ',' ' ')

		# cd $dir
		if [ "." = "$dir" ]; then
			# root
			cd=
		else
			cd="cd '$dir'; "
		fi

		callx="$call"
		if [ "$name" = root ]; then
			# special case
			case "$cmd" in
			get)
				cmdx="get -tags tools"
				;;
			up)
				cmdx="get -tags tools -u"
				;;
			*)
				cmdx=
				;;
			esac

			[ -z "$cmdx" ] || cmdx="\$(GO) $cmdx -v ./..."

			if [ "up" = "$cmd" ]; then
				callx="$cmdx
\$(GO) mod tidy
$(gen_install_tools)"
			elif [ "get" = "$cmd" ]; then
				callx="$cmdx
$(gen_install_tools)"
			elif [ -n "$cmdx" ]; then
				classx="$cmdx"
			fi

		fi

		if [ "build" = "$cmd" ]; then
			# special build flags for cmd/*
			#
			callx="$(cat <<-EOL | packed_oneline
			set -e
			MOD="\$\$(\$(GO) list -f '{{.ImportPath}}' ./...)"
			if echo "\$\$MOD" | grep -q -e '.*/cmd/[^/]\+\$\$'; then
				\$(GO_BUILD_CMD) ./...
			elif [ -n "\$\$MOD" ]; then
				\$(GO_BUILD) ./...
			fi
			EOL
			)"
		fi

		if [ "tidy" = "$cmd" ]; then
			# exclude submodules when running revive
			#
			exclude=$(gen_revive_exclude "$dir")
			if [ -n "$exclude" ]; then
				callx=$(echo "$callx" | sed -e "s;\(REVIVE)\);\1 $exclude;")
			fi
		fi

		if ! $sequential; then
			deps=
		fi

		cat <<EOT
$cmd-$name:${deps:+ $(prefixed $cmd $deps)}${depsx:+ | $depsx} ; \$(info \$(M) $cmd: $name)
$(echo "$callx" | sed -e "/^$/d;" -e "s|^|\t\$(Q) $cd|")

EOT
	done < "$INDEX"
done

for x in $PROJECTS; do
	cat <<EOT
$x: $(suffixed $x get build tidy)
EOT
done
