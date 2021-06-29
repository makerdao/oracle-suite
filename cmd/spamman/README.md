# Spamman CLI Readme

Spamman allows testing Spire network by generating different type of messages and spamming it with them.

- Table of contents

## Installation

To install it, you'll first need Go installed on your machine. Then you can use
standard Go command:

```bash
go get -u github.com/makerdao/oracle-suite/cmd/spamman
```

Alternatively, you can build Spire using `Makefile` directly from the repository:

```bash
git clone https://github.com/makerdao/oracle-suite.git
cd oracle-suite
make
```

This will build the binary in `bin` directory.

## Quick Start

Spamman consists of two main parts:

1. P2P node (agent) responsible for connecting to other peers.
2. Spamman CLI, that allows for generating, sending and receiving messages.

## Configuration

Spamman uses `spire` configuration format due to it needs to be connected to same network.

### Minimal steps

- Create a new ethereum key.
- Edit the `spire.json` configuration file to point to the created file

```bash
{
  "ethereum": {
    "from": "0xYourEthereumAddress",
    "keystore": "/path/to/the/keystore/directory",
    "password": "/path/to/a/text/file/with/polain/text/password"
  }, 
  ...
}
```

- Run spamman.

```bash
spamman -c ./spire.json -r 60 --gen.valid.msg run
```

- Now it will start generating and sending messages to configured network.

## Available messages type

For now Spamman could generate 3 message types:

 - `gen.valid.msg` (with allowed feed address configured) - Fully valid, signed `Price` messages.
 - `gen.valid.msg` (without allowed feed address configured) - Fully valid, signed `Price` messages, but they will be rejected by network due to signer restrictions.
 - `gen.invalid.signature` - Valid message with invalid signature, messages should be rejected by network.