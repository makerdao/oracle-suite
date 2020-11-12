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
	"github.com/ethereum/go-ethereum/accounts"
)

// TODO: Merge Signer to Account. This will allow us to hide go-ethereum types
//       like accounts.Wallet and *accounts.Account.

// Account represent single Ethereum account.
type Account interface {
	// Address returns a address of this account.
	Address() Address
	// Passphrase returns a password of this account.
	Passphrase() string
	// Wallet returns the go-ethereum wallet to which this account belongs to.
	Wallet() accounts.Wallet
	// Account returns the go-ethereum representation of this account.
	Account() *accounts.Account
}
