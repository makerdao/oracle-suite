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

package internal

import (
	"math/big"
	"math/rand"
	"time"

	"github.com/makerdao/oracle-suite/pkg/oracle"
	"github.com/makerdao/oracle-suite/pkg/transport/messages"
)

type MessageGenerator struct {
	pairs []string
}

func NewMessageGenerator(pairs []string) *MessageGenerator {
	return &MessageGenerator{
		pairs: pairs,
	}
}

func (m *MessageGenerator) pair() string {
	idx := rand.Intn(len(m.pairs))
	return m.pairs[idx]
}

func (m *MessageGenerator) ValidPriceMessage() *messages.Price {
	return &messages.Price{
		Price: &oracle.Price{
			Wat: m.pair(),
			Val: big.NewInt(int64(10 + rand.Intn(100))),
			Age: time.Now(),
		},
		Trace: nil,
	}
}
