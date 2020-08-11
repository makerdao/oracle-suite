PACKAGE ?= gofer
GOFILES := $(shell { git ls-files; } | grep ".go$$") # git ls-files -o --exclude-standard;

OUT_DIR := workdir
COVER_FILE := $(OUT_DIR)/cover.out
TEST_FLAGS ?= all

GO := go

clean:
	rm -rf $(OUT_DIR)
.PHONY: clean

vendor:
	$(GO) mod vendor
.PHONY: vendor

test:
	$(GO) test -tags $(TEST_FLAGS) ./...
.PHONY: test

test-e2e:
	$(GO) test -tags $(TEST_FLAGS) ./e2ehelper
.PHONY: e2etest

bench:
	$(GO) test -tags $(TEST_FLAGS) -bench=. ./...
.PHONY: bench

lint:
	golangci-lint run ./...
.PHONY: lint

cover:
	@mkdir -p $(dir $(COVER_FILE))
	$(GO) test -tags $(TEST_FLAGS) -coverprofile=$(COVER_FILE) ./...
	go tool cover -func=$(COVER_FILE)
.PHONY: cover

add-license: $(GOFILES)
	for x in $^; do tmp=$$(cat LICENSE_HEADER; sed -n '/^package \|^\/\/ *+build /,$$p' $$x); echo "$$tmp" > $$x; done
.PHONY: add-license

test-license: $(GOFILES)
	@grep -vlz "$$(tr '\n' . < LICENSE_HEADER)" $^ && exit 1 || exit 0
.PHONY: test-license

test-all: test lint test-license
.PHONY: test-all
