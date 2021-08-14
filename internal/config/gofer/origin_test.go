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

package gofer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsingOriginParamsFailParsing(t *testing.T) {
	parsed, err := parseOriginParams(nil)
	assert.Nil(t, parsed)
	assert.Error(t, err)

	parsed, err = parseOriginParams([]byte(""))
	assert.Nil(t, parsed)
	assert.Error(t, err)
}

func TestParsingOriginParams(t *testing.T) {
	parsed, err := parseOriginParams([]byte(`{}`))
	assert.NoError(t, err)
	assert.NotNil(t, parsed)

	// Parsing only apiKey
	parsed, err = parseOriginParams([]byte(`{"apiKey":"test"}`))
	assert.NoError(t, err)
	assert.NotNil(t, parsed)
	assert.Equal(t, "test", parsed.APIKey)

	// Parsing contracts
	parsed, err = parseOriginParams([]byte(`{"contracts":{"BTC/ETH":"0x00000"}}`))
	assert.NoError(t, err)
	assert.NotNil(t, parsed)
	assert.Equal(t, "0x00000", parsed.Contracts["BTC/ETH"])

	// Parsing symbol aliases
	parsed, err = parseOriginParams([]byte(`{"symbolAliases":{"ETH":"WETH"}}`))
	assert.NoError(t, err)
	assert.NotNil(t, parsed)
	assert.Equal(t, "WETH", parsed.SymbolAliases["ETH"])
}
