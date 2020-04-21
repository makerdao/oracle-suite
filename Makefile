PACKAGE ?= gofer
GOFILES := $(shell find . -name '*.go')

OUT_DIR := workdir
CMD_TARGET := $(OUT_DIR)/$(PACKAGE)
COVER_FILE := $(OUT_DIR)/cover.out

GO := go

clean:
	rm -rf $(OUT_DIR)
.PHONY: clean

build: clean
.PHONY: build

$(CMD_TARGET): GOOS ?= linux
$(CMD_TARGET): GOARCH ?= amd64
$(CMD_TARGET): CGO_ENABLED ?= 0
$(CMD_TARGET): $(CMD_SRCS)
	mkdir -p $(@D)
	$(GO) build -o $@ $<

vendor:
	$(GO) mod vendor
.PHONY: vendor

test:
	$(GO) test ./...
.PHONY: test

bench:
	$(GO) test -bench=. ./...
.PHONY: bench

lint:
	golangci-lint run ./...
.PHONY: lint

cover:
	@mkdir -p $(dir $(COVER_FILE))
	$(GO) test -coverprofile=$(COVER_FILE) ./...
	go tool cover -func=$(COVER_FILE)
.PHONY: cover
