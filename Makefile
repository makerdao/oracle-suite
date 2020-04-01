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