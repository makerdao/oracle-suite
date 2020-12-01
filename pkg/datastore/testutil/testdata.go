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

package testutil

import (
	"math/big"
	"time"

	"github.com/makerdao/gofer/pkg/ethereum"
	"github.com/makerdao/gofer/pkg/oracle"
	"github.com/makerdao/gofer/pkg/transport/messages"
)

var (
	Address1     = ethereum.HexToAddress("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")
	Address2     = ethereum.HexToAddress("0x8eb3daaf5cb4138f5f96711c09c0cfd0288a36e9")
	PriceAAABBB1 = &messages.Price{
		Price: &oracle.Price{
			Wat: "AAABBB",
			Val: big.NewInt(10),
			Age: time.Unix(100, 0),
			V:   1,
		},
		Trace: nil,
	}
	PriceAAABBB2 = &messages.Price{
		Price: &oracle.Price{
			Wat: "AAABBB",
			Val: big.NewInt(20),
			Age: time.Unix(200, 0),
			V:   2,
		},
		Trace: nil,
	}
	PriceAAABBB3 = &messages.Price{
		Price: &oracle.Price{
			Wat: "AAABBB",
			Val: big.NewInt(30),
			Age: time.Unix(300, 0),
			V:   3,
		},
		Trace: nil,
	}
	PriceAAABBB4 = &messages.Price{
		Price: &oracle.Price{
			Wat: "AAABBB",
			Val: big.NewInt(30),
			Age: time.Unix(400, 0),
			V:   4,
		},
		Trace: nil,
	}
	PriceXXXYYY1 = &messages.Price{
		Price: &oracle.Price{
			Wat: "XXXYYY",
			Val: big.NewInt(10),
			Age: time.Unix(100, 0),
			V:   5,
		},
		Trace: nil,
	}
	PriceXXXYYY2 = &messages.Price{
		Price: &oracle.Price{
			Wat: "XXXYYY",
			Val: big.NewInt(20),
			Age: time.Unix(200, 0),
			V:   6,
		},
		Trace: nil,
	}
)
