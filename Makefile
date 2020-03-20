PACKAGE ?= gofer

vendor:
	go mod vendor
.PHONY: vendor

test:
	go test ./...
.PHONY: test