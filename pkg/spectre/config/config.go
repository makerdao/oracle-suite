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

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/libp2p/go-libp2p-core/crypto"

	suite "github.com/makerdao/oracle-suite"
	"github.com/makerdao/oracle-suite/pkg/datastore"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	ethereumGeth "github.com/makerdao/oracle-suite/pkg/ethereum/geth"
	"github.com/makerdao/oracle-suite/pkg/log"
	oracleGeth "github.com/makerdao/oracle-suite/pkg/oracle/geth"
	"github.com/makerdao/oracle-suite/pkg/spectre"
	"github.com/makerdao/oracle-suite/pkg/transport"
	"github.com/makerdao/oracle-suite/pkg/transport/messages"
	"github.com/makerdao/oracle-suite/pkg/transport/p2p"
)

var ErrFailedToLoadConfiguration = errors.New("failed to load Spectre's configuration")
var ErrFailedToReadPassphraseFile = errors.New("failed to read the ethereum password file")
var ErrFailedToParsePrivKeySeed = errors.New("failed to parse the privKeySeed field")

type Config struct {
	Ethereum Ethereum        `json:"ethereum"`
	P2P      P2P             `json:"p2p"`
	Options  Options         `json:"options"`
	Feeds    []string        `json:"feeds"`
	Pairs    map[string]Pair `json:"pairs"`
}

type Ethereum struct {
	From     string `json:"from"`
	Keystore string `json:"keystore"`
	Password string `json:"password"`
	RPC      string `json:"rpc"`
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

type Pair struct {
	Oracle           string  `json:"oracle"`
	OracleSpread     float64 `json:"oracleSpread"`
	OracleExpiration int64   `json:"oracleExpiration"`
	MsgExpiration    int64   `json:"msgExpiration"`
}

type Dependencies struct {
	Context context.Context
	Logger  log.Logger
}

type Instances struct {
	Ethereum  ethereum.Client
	Signer    ethereum.Signer
	Transport transport.Transport
	Spectre   *spectre.Spectre
}

func (c *Config) Configure(deps Dependencies) (*Instances, error) {
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

	// Create Ethereum client:
	eth, err := c.configureEthClient(sig)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrFailedToLoadConfiguration, err)
	}

	// Datastore:
	dat := c.configureDatastore(sig, tra, deps.Logger)

	// Create and configure Spectre:
	spe := c.configureSpectre(sig, dat, deps.Logger, eth)

	return &Instances{
		Ethereum:  eth,
		Signer:    sig,
		Transport: tra,
		Spectre:   spe,
	}, nil
}

func (c *Config) configureAccount() (*ethereumGeth.Account, error) {
	passphrase, err := c.readAccountPassphrase(c.Ethereum.Password)
	if err != nil {
		return nil, err
	}

	a, err := ethereumGeth.NewAccount(
		c.Ethereum.Keystore,
		passphrase,
		ethereum.HexToAddress(c.Ethereum.From),
	)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (c *Config) configureSigner(a *ethereumGeth.Account) ethereum.Signer {
	return ethereumGeth.NewSigner(a)
}

func (c *Config) configureTransport(ctx context.Context, s ethereum.Signer, l log.Logger) (transport.Transport, error) {
	peerPrivKey, err := c.generatePrivKey()
	if err != nil {
		return nil, err
	}

	// Spectre doesn't create any messages so there is no need to
	// configure the MessagePrivKey field.
	cfg := p2p.Config{
		Context:          ctx,
		PeerPrivKey:      peerPrivKey,
		ListenAddrs:      c.P2P.ListenAddrs,
		BootstrapAddrs:   c.P2P.BootstrapAddrs,
		DirectPeersAddrs: c.P2P.DirectPeersAddrs,
		BlockedAddrs:     c.P2P.BlockedAddrs,
		Discovery:        !c.P2P.DisableDiscovery,
		Signer:           s,
		Logger:           l,
		AppName:          "spectre",
		AppVersion:       suite.Version,
	}
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

func (c *Config) configureEthClient(s ethereum.Signer) (*ethereumGeth.Client, error) {
	client, err := ethclient.Dial(c.Ethereum.RPC)
	if err != nil {
		return nil, err
	}

	return ethereumGeth.NewClient(client, s), nil
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

	for name := range c.Pairs {
		cfg.Pairs[name] = &datastore.Pair{Feeds: feeds}
	}

	return datastore.NewDatastore(cfg)
}

func (c *Config) configureSpectre(
	s ethereum.Signer,
	d spectre.Datastore,
	l log.Logger,
	e ethereum.Client,
) *spectre.Spectre {

	cfg := spectre.Config{
		Signer:    s,
		Interval:  time.Second * time.Duration(c.Options.Interval),
		Datastore: d,
		Logger:    l,
		Pairs:     nil,
	}

	for name, pair := range c.Pairs {
		cfg.Pairs = append(cfg.Pairs, &spectre.Pair{
			AssetPair:        name,
			OracleSpread:     pair.OracleSpread,
			OracleExpiration: time.Second * time.Duration(pair.OracleExpiration),
			PriceExpiration:  time.Second * time.Duration(pair.MsgExpiration),
			Median:           oracleGeth.NewMedian(e, ethereum.HexToAddress(pair.Oracle)),
		})
	}

	return spectre.NewSpectre(cfg)
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
	passphraseFile, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("%v: %v", ErrFailedToReadPassphraseFile, err)
	}
	return strings.TrimSuffix(string(passphraseFile), "\n"), nil
}
