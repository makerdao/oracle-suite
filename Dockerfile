FROM golang:alpine as builder

RUN apk update && apk add git make pkgconf gcc libc-dev openssl

WORKDIR $GOPATH/src/github.com/makerdao/gofer

COPY main.go ./
COPY lib lib
COPY go.mod ./
COPY go.sum ./

ENV GO111MODULE on
ENV GOSUMDB off
ENV GOPROXY direct
ENV GOBIN /usr/local/go/bin

ENTRYPOINT ash

RUN go get ./...
RUN go build -o gofer -i main.go

FROM alpine:latest

RUN apk update && apk add ca-certificates tzdata pkgconf openssl

WORKDIR /opt/gofer

COPY --from=builder /go/src/github.com/makerdao/gofer/gofer bin/gofer
COPY config/ config

ENV SERVER_EXTRA_ARGS=

ENTRYPOINT /opt/gofer/bin/gofer \
	$SERVER_EXTRA_ARGS