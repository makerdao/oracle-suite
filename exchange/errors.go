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

package exchange

import "fmt"

var errNoPotentialPricePoint = fmt.Errorf("failed to make request to nil PotentialPricePoint")

//nolint:deadcode,varcheck,unused
var errWrongPotentialPricePoint = fmt.Errorf("failed to make request using wrong PotentialPricePoint")

//nolint:deadcode,varcheck,unused
var errNoExchangeInPotentialPricePoint = fmt.Errorf("failed to make request for nil exchange in given PotentialPricePoint")

var errUnknownExchange = fmt.Errorf("unknown exchange name given")

var errNoPoolPassed = fmt.Errorf("no query worker pool passed")

var errEmptyExchangeResponse = fmt.Errorf("empty exchange response received")
