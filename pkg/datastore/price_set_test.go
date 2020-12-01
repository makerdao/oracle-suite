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

package datastore

import (
	"math"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/pkg/datastore/testutil"
	"github.com/makerdao/gofer/pkg/transport/messages"
)

func TestPriceSet_Len(t *testing.T) {
	ps := NewPriceSet([]*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	})

	assert.Equal(t, 4, ps.Len())
}

func TestPriceSet_Messages(t *testing.T) {
	ps := NewPriceSet([]*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	})

	assert.Len(t, ps.Messages(), 4)
	assert.Contains(t, ps.Messages(), testutil.PriceAAABBB1)
	assert.Contains(t, ps.Messages(), testutil.PriceAAABBB2)
	assert.Contains(t, ps.Messages(), testutil.PriceAAABBB3)
	assert.Contains(t, ps.Messages(), testutil.PriceAAABBB4)
}

func TestPriceSet_OraclePrices(t *testing.T) {
	ps := NewPriceSet([]*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	})

	assert.Len(t, ps.OraclePrices(), 4)
	assert.Contains(t, ps.OraclePrices(), testutil.PriceAAABBB1.Price)
	assert.Contains(t, ps.OraclePrices(), testutil.PriceAAABBB2.Price)
	assert.Contains(t, ps.OraclePrices(), testutil.PriceAAABBB3.Price)
	assert.Contains(t, ps.OraclePrices(), testutil.PriceAAABBB4.Price)
}

func TestPriceSet_Truncate(t *testing.T) {
	msgs := []*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	}

	ps1 := NewPriceSet(msgs)
	ps1.Truncate(5)
	assert.Len(t, ps1.Messages(), 4)

	ps2 := NewPriceSet(msgs)
	ps2.Truncate(4)
	assert.Len(t, ps2.Messages(), 4)

	ps3 := NewPriceSet(msgs)
	ps3.Truncate(3)
	assert.Len(t, ps3.Messages(), 3)
}

func TestPriceSet_Median_Even(t *testing.T) {
	ps := NewPriceSet([]*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	})

	assert.Equal(t, big.NewInt(25), ps.Median())
}

func TestPriceSet_Median_Odd(t *testing.T) {
	ps := NewPriceSet([]*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
	})

	assert.Equal(t, big.NewInt(20), ps.Median())
}

func TestPriceSet_Median_Empty(t *testing.T) {
	ps := NewPriceSet([]*messages.Price{})

	assert.Equal(t, big.NewInt(0), ps.Median())
}

func TestPriceSet_Spread(t *testing.T) {
	ps := NewPriceSet([]*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	})

	tests := []struct {
		price int64
		want  float64
	}{
		{
			price: 0,
			want:  math.Inf(1),
		},
		{
			price: 20,
			want:  25,
		},
		{
			price: 25,
			want:  0,
		},
		{
			price: 50,
			want:  50,
		},
	}
	for n, tt := range tests {
		t.Run("Case:"+strconv.Itoa(n+1), func(t *testing.T) {
			assert.Equal(t, tt.want, ps.Spread(big.NewInt(tt.price)))
		})
	}
}

func TestPriceSet_ClearOlderThan(t *testing.T) {
	ps := NewPriceSet([]*messages.Price{
		testutil.PriceAAABBB1,
		testutil.PriceAAABBB2,
		testutil.PriceAAABBB3,
		testutil.PriceAAABBB4,
	})

	ps.ClearOlderThan(time.Unix(300, 0))

	assert.Len(t, ps.OraclePrices(), 2)
	assert.Contains(t, ps.OraclePrices(), testutil.PriceAAABBB3.Price)
	assert.Contains(t, ps.OraclePrices(), testutil.PriceAAABBB4.Price)
}
