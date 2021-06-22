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

	"github.com/libp2p/go-libp2p-core/crypto"

	suite "github.com/makerdao/oracle-suite"
	"github.com/makerdao/oracle-suite/pkg/datastore"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/ethereum/geth"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/spire"
	"github.com/makerdao/oracle-suite/pkg/transport"
	"github.com/makerdao/oracle-suite/pkg/transport/messages"
	"github.com/makerdao/oracle-suite/pkg/transport/p2p"
	"github.com/makerdao/oracle-suite/pkg/transport/p2p/crypto/ethkey"
)

var ErrFailedToLoadConfiguration = errors.New("failed to load Spire's configuration")
var ErrFailedToReadPassphraseFile = errors.New("failed to read the ethereum password file")
var ErrFailedToParsePrivKeySeed = errors.New("failed to parse the privKeySeed field")

type Config struct {
	Ethereum Ethereum `json:"ethereum"`
	P2P      P2P      `json:"p2p"`
	RPC      RPC      `json:"rpc"`
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

type RPC struct {
	Disable bool   `json:"disable"`
	Address string `json:"address"`
}

type Dependencies struct {
	Context context.Context
	Logger  log.Logger
}

func (c *Config) ConfigureAgent(deps Dependencies) (*spire.Agent, error) {
	// Ethereum account:
	acc, err := c.configureAccount()
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrFailedToLoadConfiguration, err)
	}

	// Signer:
	sig := c.configureSigner(acc)

	// Transport:
	tra, err := c.configureTransport(deps.Context, sig, deps.Logger)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrFailedToLoadConfiguration, err)
	}

	// Datastore:
	dat := c.configureDatastore(sig, tra, deps.Logger)

	// Spire's RPC Agent:
	srv, err := spire.NewAgent(spire.AgentConfig{
		Datastore: dat,
		Transport: tra,
		Signer:    sig,
		Network:   "tcp",
		Address:   c.RPC.Address,
		Logger:    deps.Logger,
		SkipRPC:   c.RPC.Disable,
	})
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrFailedToLoadConfiguration, err)
	}

	return srv, nil
}

func (c *Config) ConfigureSpire(deps Dependencies) (*spire.Spire, error) {
	// Ethereum account:
	acc, err := c.configureAccount()
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrFailedToLoadConfiguration, err)
	}

	// Signer:
	sig := c.configureSigner(acc)

	// Spire:
	return spire.NewSpire(spire.Config{
		Signer:  sig,
		Network: "tcp",
		Address: c.RPC.Address,
	}), nil
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
		AssetPairs:       c.Pairs,
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

func (c *Config) configureDatastore(s ethereum.Signer, t transport.Transport, l log.Logger) *datastore.Datastore {
	cfg := datastore.Config{
		Signer:    s,
		Transport: t,
		Pairs:     make(map[string]*datastore.Pair),
		Logger:    l,
	}

	var feeds []ethereum.Address
	for _, feed := range c.Feeds {
		feeds = append(feeds, ethereum.HexToAddress(feed))
	}

	for _, pair := range c.Pairs {
		cfg.Pairs[pair] = &datastore.Pair{Feeds: feeds}
	}

	return datastore.NewDatastore(cfg)
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
