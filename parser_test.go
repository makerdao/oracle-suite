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

	"github.com/stretchr/testify/suite"
)

type ConfigParserSuite struct {
	suite.Suite
}

func (suite *ConfigParserSuite) TestGoferLibPrices() {
	cfg, err := ReadFile("non-existing.json")
	suite.Error(err)
	suite.Nil(cfg)

	cfg, err = ReadFile("./config.sample.json")
	suite.NoError(err)
	suite.NotNil(cfg)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestConfigParserSuite(t *testing.T) {
	suite.Run(t, &ConfigParserSuite{})
}
