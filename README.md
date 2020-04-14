# gofer

[![CircleCI](https://circleci.com/gh/makerdao/gofer.svg?style=svg&circle-token=a7007c0430edac55d1625526a2ad7c0151bbc8c6)](https://circleci.com/gh/makerdao/gofer)

## Tech

  - `levigo` LevelDB interface used for persistance

## Building and testing

Build binaries to `workdir/`

```sh
make build
```

Run tests

```sh
make test
```

Run benchmarks

```sh
make bench
```

## Structure

  - `cmd/` CLI entrypoints
  - `app/` run-time entrypoint
  - `store/` persistance layer
  - `store/leveldb/` LevelDB store implementation
  - `reducer/` business logic
  - `config/` run-time configuration using JSON files
