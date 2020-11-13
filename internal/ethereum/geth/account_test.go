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

package geth

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var accountAddress = common.HexToAddress("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")

func TestAccount_ValidAddress(t *testing.T) {
	account, err := NewAccount("./testdata/keystore", "test123", accountAddress)
	assert.NoError(t, err)

	assert.Equal(t, accountAddress, account.Address())
	assert.Equal(t, "test123", account.Passphrase())
	assert.Equal(t, accountAddress, account.account.Address)
	assert.NotNil(t, account.wallet)
}

func TestAccount_InvalidAddress(t *testing.T) {
	account, err := NewAccount("./testdata/keystore", "test123", common.HexToAddress(""))
	assert.Error(t, err)
	assert.Nil(t, account)
}
