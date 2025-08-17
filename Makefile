.PHONY: all clean generate fmt tidy check-grammar check-spelling check-shell check-jq
.PHONY: coverage codecov clean-coverage race
.PHONY: FORCE

GO ?= go
GOFMT ?= gofmt
GOFMT_FLAGS = -w -l -s
GOGENERATE_FLAGS = -v
GOUP_FLAGS ?= -v
GOUP_PACKAGES ?= ./...
GOTEST_FLAGS ?=
JQ ?= jq

TOOLSDIR := $(CURDIR)/internal/build
TMPDIR ?= $(CURDIR)/.tmp
OUTDIR ?= $(TMPDIR)
COVERAGE_DIR ?= $(TMPDIR)/coverage

# Dynamic version selection based on Go version
# Format: $(TOOLSDIR)/get_version.sh <go_version> <tool_version1> <tool_version2> ..
GOLANGCI_LINT_VERSION ?= $(shell $(TOOLSDIR)/get_version.sh 1.23 v2.3.0)
REVIVE_VERSION ?= $(shell $(TOOLSDIR)/get_version.sh 1.23 v1.7)

GOLANGCI_LINT_URL ?= github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
GOLANGCI_LINT ?= $(GO) run $(GOLANGCI_LINT_URL)

REVIVE_CONF ?= $(TOOLSDIR)/revive.toml
REVIVE_RUN_ARGS ?= -config $(REVIVE_CONF) -formatter friendly
REVIVE_URL ?= github.com/mgechev/revive@$(REVIVE_VERSION)
REVIVE ?= $(GO) run $(REVIVE_URL)

FIX_WHITESPACE ?= $(TOOLSDIR)/fix_whitespace.sh
# Exclude Go files (handled separately by gofmt)
FIX_WHITESPACE_EXCLUDE_GO ?= -name '*.go'
# Exclude binary and image files
FIX_WHITESPACE_EXCLUDE_BINARY_EXTS ?= exe dll so dylib a o test
FIX_WHITESPACE_EXCLUDE_IMAGE_EXTS ?= png jpg jpeg gif ico pdf
FIX_WHITESPACE_EXCLUDE_ARCHIVE_EXTS ?= zip tar gz bz2 xz 7z
FIX_WHITESPACE_EXCLUDE_OTHER_EXTS ?= bin dat
# Combine all exclusions
FIX_WHITESPACE_EXCLUDE_EXTS ?= \
	$(FIX_WHITESPACE_EXCLUDE_ARCHIVE_EXTS) \
	$(FIX_WHITESPACE_EXCLUDE_BINARY_EXTS) \
	$(FIX_WHITESPACE_EXCLUDE_IMAGE_EXTS) \
	$(FIX_WHITESPACE_EXCLUDE_OTHER_EXTS)
FIX_WHITESPACE_EXCLUDE_PATTERNS ?= $(patsubst %,-o -name '*.%',$(FIX_WHITESPACE_EXCLUDE_EXTS))
FIX_WHITESPACE_EXCLUDE ?= $(FIX_WHITESPACE_EXCLUDE_GO) $(FIX_WHITESPACE_EXCLUDE_PATTERNS)
FIX_WHITESPACE_ARGS ?= . \! \( $(FIX_WHITESPACE_EXCLUDE) \)

PNPX ?= pnpx

FIND_FILES_PRUNE_RULES ?= -name vendor -o -name .git -o -name node_modules
FIND_FILES_PRUNE_ARGS ?= \( $(FIND_FILES_PRUNE_RULES) \) -prune
FIND_FILES_GO_ARGS ?= $(FIND_FILES_PRUNE_ARGS) -o -name '*.go'
FIND_FILES_MARKDOWN_ARGS ?= $(FIND_FILES_PRUNE_ARGS) -o -name '*.md'

ifndef MARKDOWNLINT
ifeq ($(shell $(PNPX) markdownlint-cli --version 2>&1 | grep -q '^[0-9]' && echo yes),yes)
MARKDOWNLINT = $(PNPX) markdownlint-cli
else
MARKDOWNLINT = true
endif
endif
MARKDOWNLINT_FLAGS ?= --fix --config $(TOOLSDIR)/markdownlint.json

ifndef LANGUAGETOOL
ifeq ($(shell $(PNPX) @twilio-labs/languagetool-cli --version 2>&1 | grep -qE '^(unknown|[0-9])' && echo yes),yes)
LANGUAGETOOL = $(PNPX) @twilio-labs/languagetool-cli
else
LANGUAGETOOL = true
endif
endif
LANGUAGETOOL_FLAGS ?= --config $(TOOLSDIR)/languagetool.cfg --custom-dict-file $(TMPDIR)/languagetool-dict.txt

ifndef CSPELL
ifeq ($(shell $(PNPX) cspell --version 2>&1 | grep -q '^[0-9]' && echo yes),yes)
CSPELL = $(PNPX) cspell
else
CSPELL = true
endif
endif
CSPELL_FLAGS ?= --no-progress --dot --config $(TOOLSDIR)/cspell.json

ifndef SHELLCHECK
ifeq ($(shell $(PNPX) shellcheck --version 2>&1 | grep -q '^ShellCheck' && echo yes),yes)
SHELLCHECK = $(PNPX) shellcheck
else
SHELLCHECK = true
endif
endif
SHELLCHECK_FLAGS ?=

V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell if [ "$$(tput colors 2> /dev/null || echo 0)" -ge 8 ]; then printf "\033[34;1m▶\033[0m"; else printf "▶"; fi)

GO_BUILD = $(GO) build -v
GO_BUILD_CMD = $(GO_BUILD) -o "$(OUTDIR)"

all: get generate tidy build

clean: ; $(info $(M) cleaning…)
	rm -rf $(TMPDIR)

$(TMPDIR)/index: $(TOOLSDIR)/gen_index.sh Makefile FORCE ; $(info $(M) generating index…)
	$Q mkdir -p $(@D)
	$Q $< > $@~
	$Q if cmp $@ $@~ 2> /dev/null >&2; then rm $@~; else mv $@~ $@; fi

$(TMPDIR)/gen.mk: $(TOOLSDIR)/gen_mk.sh $(TMPDIR)/index Makefile ; $(info $(M) generating subproject rules…)
	$Q mkdir -p $(@D)
	$Q $< $(TMPDIR)/index > $@~
	$Q if cmp $@ $@~ 2> /dev/null >&2; then rm $@~; else mv $@~ $@; fi

$(TMPDIR)/languagetool-dict.txt: $(TOOLSDIR)/cspell.json | check-jq ; $(info $(M) generating languagetool dictionary…)
	$Q mkdir -p $(@D)
	$Q $(JQ) -r '.words[]' $< | sort > $@

include $(TMPDIR)/gen.mk

fmt: ; $(info $(M) reformatting sources…)
	$Q find . $(FIND_FILES_GO_ARGS) -print0 | xargs -0 -r $(GOFMT) $(GOFMT_FLAGS)
	$Q $(FIX_WHITESPACE) $(FIX_WHITESPACE_ARGS)
ifneq ($(MARKDOWNLINT),true)
	$Q find . $(FIND_FILES_MARKDOWN_ARGS) -print0 | xargs -0 -r $(MARKDOWNLINT) $(MARKDOWNLINT_FLAGS)
endif

ifneq ($(LANGUAGETOOL),true)
check-grammar: $(TMPDIR)/languagetool-dict.txt FORCE ; $(info $(M) checking grammar…)
	$Q find . $(FIND_FILES_MARKDOWN_ARGS) -print0 | xargs -0 -r $(LANGUAGETOOL) $(LANGUAGETOOL_FLAGS)
else
check-grammar: FORCE ; $(info $(M) grammar checks disabled)
endif

ifneq ($(CSPELL),true)
TIDY_SPELLING = check-spelling
check-spelling: FORCE ; $(info $(M) checking spelling…)
	$Q $(CSPELL) $(CSPELL_FLAGS) "**/*.{go,md}"
else
TIDY_SPELLING =
check-spelling: FORCE ; $(info $(M) spell checking disabled)
endif

ifneq ($(SHELLCHECK),true)
TIDY_SHELL = check-shell
check-shell: FORCE ; $(info $(M) checking shell scripts…)
	$Q find . $(FIND_FILES_PRUNE_ARGS) -o -name '*.sh' -print0 | xargs -0 -r $(SHELLCHECK) $(SHELLCHECK_FLAGS)
else
TIDY_SHELL =
check-shell: FORCE ; $(info $(M) shell checks disabled)
endif

tidy: fmt $(TIDY_SPELLING) $(TIDY_SHELL)

generate: ; $(info $(M) running go:generate…)
	$Q git grep -l '^//go:generate' | sort -uV | xargs -r -n1 $(GO) generate $(GOGENERATE_FLAGS)


# Generate Codecov upload script
# This target prepares codecov.sh script for uploading coverage
# data to Codecov with proper module flags
codecov: $(COVERAGE_DIR)/coverage.out ; $(info $(M) preparing codecov data)
	$Q $(TOOLSDIR)/make_codecov.sh $(TMPDIR)/index $(COVERAGE_DIR)

check-jq: FORCE
	$Q $(JQ) --version >/dev/null 2>&1 || { \
		echo "Warning: jq is required to import cspell's custom dictionary but was not found" >&2; \
		echo "  Install jq or set JQ variable to override" >&2; \
		false; \
	}
