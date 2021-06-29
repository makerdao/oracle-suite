//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package config

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p-core/crypto"

	suite "github.com/makerdao/oracle-suite"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/ethereum/geth"
	"github.com/makerdao/oracle-suite/pkg/ghost"
	"github.com/makerdao/oracle-suite/pkg/gofer"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/transport"
	"github.com/makerdao/oracle-suite/pkg/transport/messages"
	"github.com/makerdao/oracle-suite/pkg/transport/p2p"
	"github.com/makerdao/oracle-suite/pkg/transport/p2p/crypto/ethkey"
)

var ErrFailedToLoadConfiguration = errors.New("failed to load Ghost's configuration")
var ErrFailedToReadPassphraseFile = errors.New("failed to read the ethereum password file")
var ErrFailedToParsePrivKeySeed = errors.New("failed to parse the privKeySeed field")

type Config struct {
	Ethereum Ethereum `json:"ethereum"`
	P2P      P2P      `json:"p2p"`
	Options  Options  `json:"options"`
	Feeds    []string `json:"feeds"`
	Pairs    []string `json:"pairs"`
}

type Ethereum struct {
	From     string `json:"from"`
	Keystore string `json:"keystore"`
	Password string `json:"password"`
}

type P2P struct {
	PrivKeySeed      string   `json:"privKeySeed"`
	ListenAddrs      []string `json:"listenAddrs"`
	BootstrapAddrs   []string `json:"bootstrapAddrs"`
	DirectPeersAddrs []string `json:"directPeersAddrs"`
	BlockedAddrs     []string `json:"blockedAddrs"`
	DisableDiscovery bool     `json:"disableDiscovery"`
}

type Options struct {
	Interval int `json:"interval"`
}

type Dependencies struct {
	Context context.Context
	Gofer   gofer.Gofer
	Logger  log.Logger
}

type Instances struct {
	Signer    ethereum.Signer
	Transport transport.Transport
	Ghost     *ghost.Ghost
}

func (c *Config) Configure(deps Dependencies) (*Instances, error) {
	// Create wallet for given account and keystore:
	acc, err := c.configureAccount()
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrFailedToLoadConfiguration, err)
	}

	// Create new signer instance:
	sig := c.configureSigner(acc)

	// Transport:
	tra, err := c.configureTransport(deps.Context, sig, deps.Logger)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrFailedToLoadConfiguration, err)
	}

	// Create and configure Ghost:
	gho, err := c.configureGhost(deps.Gofer, sig, tra, deps.Logger)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrFailedToLoadConfiguration, err)
	}

	return &Instances{
		Signer:    sig,
		Transport: tra,
		Ghost:     gho,
	}, nil
}

func (c *Config) configureAccount() (*geth.Account, error) {
	passphrase, err := c.readAccountPassphrase(c.Ethereum.Password)
	if err != nil {
		return nil, err
	}

	a, err := geth.NewAccount(
		c.Ethereum.Keystore,
		passphrase,
		ethereum.HexToAddress(c.Ethereum.From),
	)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (c *Config) configureSigner(a *geth.Account) ethereum.Signer {
	return geth.NewSigner(a)
}

func (c *Config) configureTransport(ctx context.Context, s ethereum.Signer, l log.Logger) (transport.Transport, error) {
	peerPrivKey, err := c.generatePrivKey()
	if err != nil {
		return nil, err
	}

	cfg := p2p.Config{
		Context:          ctx,
		PeerPrivKey:      peerPrivKey,
		MessagePrivKey:   ethkey.NewPrivKey(s),
		ListenAddrs:      c.P2P.ListenAddrs,
		BootstrapAddrs:   c.P2P.BootstrapAddrs,
		DirectPeersAddrs: c.P2P.DirectPeersAddrs,
		BlockedAddrs:     c.P2P.BlockedAddrs,
		Discovery:        !c.P2P.DisableDiscovery,
		Signer:           s,
		Logger:           l,
		AppName:          "spire",
		AppVersion:       suite.Version,
	}
	cfg.FeedersAddrs = []ethereum.Address{ethereum.HexToAddress(c.Ethereum.From)}
	for _, feed := range c.Feeds {
		cfg.FeedersAddrs = append(cfg.FeedersAddrs, ethereum.HexToAddress(feed))
	}

	p, err := p2p.New(cfg)
	if err != nil {
		return nil, err
	}

	err = p.Subscribe(messages.PriceMessageName, (*messages.Price)(nil))
	if err != nil {
		_ = p.Close()
		return nil, err
	}

	return p, nil
}

func (c *Config) configureGhost(
	g gofer.Gofer,
	s ethereum.Signer,
	t transport.Transport,
	l log.Logger,
) (*ghost.Ghost, error) {

	cfg := ghost.Config{
		Gofer:     g,
		Signer:    s,
		Transport: t,
		Logger:    l,
		Interval:  time.Second * time.Duration(c.Options.Interval),
		Pairs:     nil,
	}

	for _, name := range c.Pairs {
		cfg.Pairs = append(cfg.Pairs, &ghost.Pair{
			AssetPair: name,
		})
	}

	return ghost.NewGhost(cfg)
}

func (c *Config) generatePrivKey() (crypto.PrivKey, error) {
	seedReader := rand.Reader
	if len(c.P2P.PrivKeySeed) != 0 {
		seed, err := hex.DecodeString(c.P2P.PrivKeySeed)
		if err != nil {
			return nil, fmt.Errorf("%v: %v", ErrFailedToParsePrivKeySeed, err)
		}
		if len(seed) != ed25519.SeedSize {
			return nil, fmt.Errorf("%v: seed must be 32 bytes", ErrFailedToParsePrivKeySeed)
		}
		seedReader = bytes.NewReader(seed)
	}
	privKey, _, err := crypto.GenerateEd25519Key(seedReader)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrFailedToParsePrivKeySeed, err)
	}
	return privKey, nil
}

func (c *Config) readAccountPassphrase(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	passphraseFile, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("%v: %v", ErrFailedToReadPassphraseFile, err)
	}
	return strings.TrimSuffix(string(passphraseFile), "\n"), nil
}
