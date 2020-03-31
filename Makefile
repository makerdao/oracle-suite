PACKAGE ?= gofer
GOFILES = $(shell find . -name '*.go')

vendor:
	go mod vendor
.PHONY: vendor

workdir:
	@mkdir -p workdir
.PHONY: workdir

build: $(GOFILES)
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o workdir/gofer ./cmd/main.go
.PHONY: build

test:
	go test ./...
.PHONY: test

bench:
	go test -bench=. ./...
.PHONY: bench

lint:
	golangci-lint run ./...
.PHONY: lint

proto:
	mkdir -p model
	protoc --go_out=model *.proto
.PHONY: proto
	
build: proto
	go build
.PHONY: build
