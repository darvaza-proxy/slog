.PHONY: all clean generate fmt tidy
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
GOLANGCI_LINT_VERSION ?= $(shell $(TOOLSDIR)/get_version.sh 1.22 v1.63.4)
REVIVE_VERSION ?= $(shell $(TOOLSDIR)/get_version.sh 1.22 v1.4)

GOLANGCI_LINT_URL ?= github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
GOLANGCI_LINT ?= $(GO) run $(GOLANGCI_LINT_URL)

REVIVE_CONF ?= $(TOOLSDIR)/revive.toml
REVIVE_RUN_ARGS ?= -config $(REVIVE_CONF) -formatter friendly
REVIVE_URL ?= github.com/mgechev/revive@$(REVIVE_VERSION)
REVIVE ?= $(GO) run $(REVIVE_URL)

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

include $(TMPDIR)/gen.mk

fmt: ; $(info $(M) reformatting sources…)
	$Q find . -name '*.go' | xargs -r $(GOFMT) $(GOFMT_FLAGS)

tidy: fmt

generate: ; $(info $(M) running go:generate…)
	$Q git grep -l '^//go:generate' | sort -uV | xargs -r -n1 $(GO) generate $(GOGENERATE_FLAGS)
