# Gofer
> As in a [tool](https://en.wikipedia.org/wiki/Gofer) that specializes in the delivery of special items. 

A command line interface for the [Gofer Go library](https://github.com/makerdao/gofer) to enable getting reliable crypto-asset prices.

## Usage

### Getting prices

Single price:

```sh
gofer price BTC/USD
```

Multiple prices:

```sh
gofer price BAT/USD MKR/USD
```

Price with full aggregation trace in JSON format:

```sh
gofer price -format json ETH/BTC
```

### Listing supported price pairs

```sh
gofer pairs
```

### Listing exchanges

Listing all supported exchanges:

```sh
gofer exchanges
```

List exchanges that will be queried to get the price a specific pair: 

```sh
gofer exchanges BTC/USD
```

### Getting help

Show general help screen with available commands

```sh
gofer -h
```
```text
Usage:  gofer [OPTION...] COMMAND

A CLI interface to the Gofer Go library that provides reliable data for blockchain tools.

Options:
  -h, -help     Display this help and exit
  -config PATH  path to config file (default: ./gofer.json)

Commands:
  exchanges
  pairs
  price

For help on individual commands, run: gofer COMMAND -h
```

To get help for a specific command use

```sh
gofer COMMAND -h
```

## Installation

To install the `gofer` command into your `${GOPATH}/bin` directory:

```sh 
go get -u github.com/makerdao/gofer/cli/cmd/gofer
```

Should you wish to compile the binary from code, try:

```sh 
git clone https://github.com/makerdao/gofer/cli.git
cd gofer-cli
make
```

## License
[The GNU Affero General Public License](../../LICENSE)
