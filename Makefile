.PHONY: all clean generate fmt tidy

GO ?= go
GOFMT ?= gofmt
GOFMT_FLAGS = -w -l -s
GOGENERATE_FLAGS = -v

GOPATH ?= $(shell $(GO) env GOPATH)
GOBIN ?= $(GOPATH)/bin

REVIVE ?= $(GOBIN)/revive
REVIVE_RUN_ARGS ?= -formatter friendly -set_exit_status
REVIVE_INSTALL_URL ?= github.com/mgechev/revive

V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell if [ "$$(tput colors 2> /dev/null || echo 0)" -ge 8 ]; then printf "\033[34;1m▶\033[0m"; else printf "▶"; fi)

HANDLERS = cblog filtered logrus noop zap zerolog

TMPDIR ?= .tmp

all: get generate tidy build

clean: ; $(info $(M) cleaning…)
	rm -rf $(TMPDIR)

$(TMPDIR)/gen.mk: internal/tools/gen_mk.sh Makefile ; $(info $(M) generating subproject rules)
	$Q mkdir -p $(@D)
	$Q $< $(HANDLERS) > $@~
	$Q if cmp $@ $@~ 2> /dev/null >&2; then rm $@~; else mv $@~ $@; fi

include $(TMPDIR)/gen.mk

fmt: ; $(info $(M) reformatting sources…)
	$Q find . -name '*.go' | xargs -r $(GOFMT) $(GOFMT_FLAGS)

tidy: fmt

generate: ; $(info $(M) running go:generate…)
	$Q git grep -l '^//go:generate' | sort -uV | xargs -r -n1 $(GO) generate $(GOGENERATE_FLAGS)

$(REVIVE):
	$Q $(GO) install -v $(REVIVE_INSTALL_URL)
