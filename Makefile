PACKAGE ?= gofer
GOFILES := $(shell find . -name '*.go')

PROTO_SRCS := $(shell find . -name '*.proto')
GEN_SRCS := $(PROTO_SRCS:.proto=.pb.go)
CMD_SRCS := cmd/main.go

OUT_DIR := workdir
CMD_TARGET := $(OUT_DIR)/$(PACKAGE)

GO := go
PROTOC := protoc

all: $(GEN_SRCS) $(CMD_TARGET)
.PHONY: all

clean:
	rm -rf $(OUT_DIR) $(GEN_SRCS)
.PHONY: clean

build: clean all
.PHONY: build

$(GEN_SRCS): $(PROTO_SRCS)
	for x in $^; do $(PROTOC) --go_out=. $$x; done

$(CMD_TARGET): GOOS ?= linux
$(CMD_TARGET): GOARCH ?= amd64
$(CMD_TARGET): CGO_ENABLED ?= 0
$(CMD_TARGET): $(CMD_SRCS)
	mkdir -p $(@D)
	$(GO) build -o $@ $<

vendor:
	$(GO) mod vendor
.PHONY: vendor

test: $(GEN_SRCS)
	$(GO) test ./...
.PHONY: test

bench: $(GEN_SRCS)
	$(GO) test -bench=. ./...
.PHONY: bench

lint:
	golangci-lint run ./...
.PHONY: lint
