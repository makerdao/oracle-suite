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

package ethereum

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/ethereum/geth"
)

//nolint:unlambda
var ethClientFactory = func(rpc string) (geth.EthClient, error) {
	return ethclient.Dial(rpc)
}

type Ethereum struct {
	From     string `json:"from"`
	Keystore string `json:"keystore"`
	Password string `json:"password"`
	RPC      string `json:"rpc"`
}

func (c *Ethereum) ConfigureSigner() (ethereum.Signer, error) {
	account, err := c.configureAccount()
	if err != nil {
		return nil, err
	}
	return geth.NewSigner(account), nil
}

func (c *Ethereum) ConfigureEthereumClient(signer ethereum.Signer) (*geth.Client, error) {
	client, err := ethClientFactory(c.RPC)
	if err != nil {
		return nil, err
	}
	return geth.NewClient(client, signer), nil
}

func (c *Ethereum) configureAccount() (*geth.Account, error) {
	if c.From == "" {
		return nil, nil
	}
	passphrase, err := c.readAccountPassphrase(c.Password)
	if err != nil {
		return nil, err
	}
	account, err := geth.NewAccount(c.Keystore, passphrase, ethereum.HexToAddress(c.From))
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (c *Ethereum) readAccountPassphrase(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	passphrase, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read Ethereum password file: %w", err)
	}
	return strings.TrimSuffix(string(passphrase), "\n"), nil
}
