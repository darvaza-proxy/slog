#!/bin/sh

set -eu

PROJECTS="$*"
COMMANDS="get tidy build test up"

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

for cmd in $COMMANDS; do
	all="$(prefixed $cmd root $PROJECTS)"
	depsx=

	cat <<EOT
.PHONY: $cmd $all
$cmd: $all

EOT

	case "$cmd" in
	tidy)
		call="$(cat <<EOT | sed -e '/^$/d;'
\$(GO) mod tidy
\$(GO) vet ./...
\$(REVIVE) \$(REVIVE_RUN_ARGS) ./...
EOT
)"
		depsx="\$(REVIVE)"
		;;
	up)
		call="\$(GO) get -u -v ./...
\$(GO) mod tidy"
		;;
	*)
		call="\$(GO) $cmd -v ./..."
		;;
	esac

	# tidy up call

	case "$cmd" in
	build|test)
		sequential=true ;;
	*)
		sequential=false ;;
	esac

	for x in . $PROJECTS; do
		if [ "$x" = . ]; then
			k="root"
			cd=
		else
			k="$x"
			cd="cd 'handlers/$x' \&\& "
		fi

		callx="$call"
		if [ "$k" = root ]; then
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
\$(GO) install -v \$(REVIVE_INSTALL_URL)"
			elif [ "get" = "$cmd" ]; then
				callx="$cmdx
\$(GO) install -v \$(REVIVE_INSTALL_URL)"
			elif [ "tidy" = "$cmd" ]; then
				exclude=
				for x in $PROJECTS; do
					exclude="${exclude:+$exclude }-exclude ./handlers/$x/..."
				done
				callx=$(echo "$call" | sed -e "s;\(REVIVE)\);\1 $exclude;")
			elif [ -n "$cmdx" ]; then
				classx="$cmdx"
			fi
		fi

		deps=
		if $sequential; then
			[ "$k" = root ] || deps=root
		fi

		cat <<EOT
$cmd-$k:${deps:+ $(prefixed $cmd $deps)}${depsx:+ | $depsx} ; \$(info \$(M) $cmd: $k)
$(echo "$callx" | sed -e "/^$/d;" -e "s|^|\t\$(Q) $cd|")

EOT
	done
done

for x in $PROJECTS; do
	cat <<EOT
$x: $(suffixed $x get build tidy)
EOT
done
