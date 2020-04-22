//  Copyright (C) 2020  Maker Foundation
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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ModelSuite struct {
	suite.Suite
}

func (suite *ModelSuite) TestPriceEqual() {
	p := &Pair{
		Base:  "BTC",
		Quote: "ETH",
	}

	same := &Pair{
		Base:  "BTC",
		Quote: "ETH",
	}
	other := &Pair{
		Base:  "BTC",
		Quote: "USDC",
	}

	assert.True(suite.T(), p.Equal(same))
	assert.False(suite.T(), p.Equal(other))
}

func (suite *ModelSuite) TestValidateExchange() {
	assert.Error(suite.T(), ValidateExchange(nil))
	assert.Error(suite.T(), ValidateExchange(&Exchange{}))

	assert.NoError(suite.T(), ValidateExchange(&Exchange{Name: "test"}))
}

func (suite *ModelSuite) TestValidatePair() {
	assert.Error(suite.T(), ValidatePair(nil))
	assert.Error(suite.T(), ValidatePair(&Pair{}))
	assert.Error(suite.T(), ValidatePair(&Pair{Base: "BTC"}))
	assert.Error(suite.T(), ValidatePair(&Pair{Quote: "BTC"}))

	assert.NoError(suite.T(), ValidatePair(&Pair{Base: "ETH", Quote: "BTC"}))
}

func (suite *ModelSuite) TestValidatePotentialPricePoint() {
	p := &Pair{Base: "BTC", Quote: "ETH"}
	ex := &Exchange{Name: "test"}
	pp := &PotentialPricePoint{Pair: p, Exchange: ex}

	assert.Error(suite.T(), ValidatePotentialPricePoint(nil))
	assert.Error(suite.T(), ValidatePotentialPricePoint(&PotentialPricePoint{}))

	assert.Error(suite.T(), ValidatePotentialPricePoint(&PotentialPricePoint{Pair: p}))
	assert.Error(suite.T(), ValidatePotentialPricePoint(&PotentialPricePoint{Pair: &Pair{}}))
	assert.Error(suite.T(), ValidatePotentialPricePoint(&PotentialPricePoint{Pair: &Pair{Base: "BTC"}}))
	assert.Error(suite.T(), ValidatePotentialPricePoint(&PotentialPricePoint{Pair: &Pair{Quote: "BTC"}}))

	assert.Error(suite.T(), ValidatePotentialPricePoint(&PotentialPricePoint{Exchange: ex}))
	assert.Error(suite.T(), ValidatePotentialPricePoint(&PotentialPricePoint{Pair: p, Exchange: &Exchange{}}))

	assert.NoError(suite.T(), ValidatePotentialPricePoint(pp))
}

func (suite *ModelSuite) TestPriceToAndFromFloat() {
	p := 0.0234561
	assert.NotEqual(suite.T(), p, PriceFromFloat(p))
	assert.Equal(suite.T(), p, PriceToFloat(PriceFromFloat(p)))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestModelSuite(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}
