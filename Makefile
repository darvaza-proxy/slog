.PHONY: all clean generate fmt tidy

GO ?= go
GOFMT ?= gofmt
GOFMT_FLAGS = -w -l -s
GOGENERATE_FLAGS = -v

GOPATH ?= $(shell $(GO) env GOPATH)
GOBIN ?= $(GOPATH)/bin

REVIVE ?= $(GOBIN)/revive
REVIVE_FLAGS ?= -formatter friendly
REVIVE_INSTALL_URL ?= github.com/mgechev/revive

PROJECTS = logrus zap zerolog cblog

TMPDIR ?= .tmp

all: get generate fmt tidy build

clean:
	rm -rf $(TMPDIR)

$(TMPDIR)/gen.mk: tools/gen_mk.sh Makefile
	@echo "$< $(PROJECTS) > $@" >&2
	@mkdir -p $(@D)
	@$< $(PROJECTS) > $@~
	@if cmp $@ $@~ 2> /dev/null >&2; then rm $@~; else mv $@~ $@; fi

include $(TMPDIR)/gen.mk

fmt:
	@find . -name '*.go' | xargs -r $(GOFMT) $(GOFMT_FLAGS)

tidy: fmt

generate:
	@git grep -l '^//go:generate' | sed -n -e 's|\(.*\)/[^/]\+\.go$$|\1|p' | sort -u | while read d; do \
		git grep -l '^//go:generate' "$$d"/*.go | xargs -r $(GO) generate $(GOGENERATE_FLAGS); \
	done

$(REVIVE):
	$(GO) install -v $(REVIVE_INSTALL_URL)
