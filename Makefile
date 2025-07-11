.PHONY: all clean generate fmt tidy check-grammar
.PHONY: FORCE

GO ?= go
GOFMT ?= gofmt
GOFMT_FLAGS = -w -l -s
GOGENERATE_FLAGS = -v
GOUP_FLAGS ?= -v
GOUP_PACKAGES ?= ./...
GOTEST_FLAGS ?=

TOOLSDIR := $(CURDIR)/internal/build
TMPDIR ?= $(CURDIR)/.tmp
OUTDIR ?= $(TMPDIR)

# Dynamic version selection based on Go version
# Format: $(TOOLSDIR)/get_version.sh <go_version> <tool_version1> <tool_version2> ..
GOLANGCI_LINT_VERSION ?= $(shell $(TOOLSDIR)/get_version.sh 1.23 v1.64)
REVIVE_VERSION ?= $(shell $(TOOLSDIR)/get_version.sh 1.23 v1.7)

GOLANGCI_LINT_URL ?= github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
GOLANGCI_LINT ?= $(GO) run $(GOLANGCI_LINT_URL)

REVIVE_CONF ?= $(TOOLSDIR)/revive.toml
REVIVE_RUN_ARGS ?= -config $(REVIVE_CONF) -formatter friendly
REVIVE_URL ?= github.com/mgechev/revive@$(REVIVE_VERSION)
REVIVE ?= $(GO) run $(REVIVE_URL)

PNPX ?= pnpx

ifndef MARKDOWNLINT
ifeq ($(shell $(PNPX) markdownlint-cli --version 2>&1 | grep -q '^[0-9]' && echo yes),yes)
MARKDOWNLINT = $(PNPX) markdownlint-cli
else
MARKDOWNLINT = true
endif
endif
MARKDOWNLINT_FLAGS ?= --fix --config $(TOOLSDIR)/markdownlint.json

ifndef LANGUAGETOOL
ifeq ($(shell $(PNPX) @twilio-labs/languagetool-cli --version 2>&1 | grep -q '^[0-9]' && echo yes),yes)
LANGUAGETOOL = $(PNPX) @twilio-labs/languagetool-cli
else
LANGUAGETOOL = true
endif
endif
LANGUAGETOOL_FLAGS ?= --config $(TOOLSDIR)/languagetool.cfg

V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell if [ "$$(tput colors 2> /dev/null || echo 0)" -ge 8 ]; then printf "\033[34;1m▶\033[0m"; else printf "▶"; fi)

# Find all markdown files
MARKDOWN_FILES = $(shell find . \( -name vendor -o -name .git \) -prune -o -name '*.md' -print)

# Find all go files
GO_FILES = $(shell find . \( -name vendor -o -name .git \) -prune -o -name '*.go' -print)

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

include $(TMPDIR)/gen.mk

fmt: ; $(info $(M) reformatting sources…)
	$Q echo "$(GO_FILES)" | tr ' ' '\n' | xargs -r $(GOFMT) $(GOFMT_FLAGS)
	$Q echo "$(MARKDOWN_FILES)" | tr ' ' '\n' | xargs -r sed -i 's/[ \t]*$$//'
ifneq ($(MARKDOWNLINT),true)
	$Q echo "$(MARKDOWN_FILES)" | tr ' ' '\n' | xargs -r $(MARKDOWNLINT) $(MARKDOWNLINT_FLAGS)
endif

check-grammar: ; $(info $(M) checking grammar with LanguageTool…)
ifneq ($(LANGUAGETOOL),true)
	$Q echo "$(MARKDOWN_FILES) $(GO_FILES)" | tr ' ' '\n' | xargs -r $(LANGUAGETOOL) $(LANGUAGETOOL_FLAGS)
endif

tidy: fmt check-grammar

generate: ; $(info $(M) running go:generate…)
	$Q git grep -l '^//go:generate' | sort -uV | xargs -r -n1 $(GO) generate $(GOGENERATE_FLAGS)
