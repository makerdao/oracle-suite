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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/makerdao/gofer/pkg/datastore/testutil"
	"github.com/makerdao/gofer/pkg/ethereum"
	"github.com/makerdao/gofer/pkg/ethereum/mocks"
	"github.com/makerdao/gofer/pkg/log/null"
	"github.com/makerdao/gofer/pkg/oracle"
	"github.com/makerdao/gofer/pkg/transport/local"
	"github.com/makerdao/gofer/pkg/transport/messages"
)

func TestDatastore_Prices(t *testing.T) {
	sig := &mocks.Signer{}
	tra := local.New(0)

	ds := NewDatastore(Config{
		Signer:    sig,
		Transport: tra,
		Pairs: map[string]*Pair{
			"AAABBB": {Feeds: []ethereum.Address{testutil.Address1, testutil.Address2}},
			"XXXYYY": {Feeds: []ethereum.Address{testutil.Address1, testutil.Address2}},
		},
		Logger: null.New(),
	})

	assert.NoError(t, ds.Start())
	defer ds.Stop()

	sig.On("Recover", testutil.PriceAAABBB1.Price.Signature(), mock.Anything).Return(&testutil.Address1, nil)
	sig.On("Recover", testutil.PriceAAABBB2.Price.Signature(), mock.Anything).Return(&testutil.Address2, nil)
	sig.On("Recover", testutil.PriceXXXYYY1.Price.Signature(), mock.Anything).Return(&testutil.Address1, nil)
	sig.On("Recover", testutil.PriceXXXYYY2.Price.Signature(), mock.Anything).Return(&testutil.Address2, nil)

	assert.NoError(t, tra.Broadcast(messages.PriceMessageName, testutil.PriceAAABBB1))
	assert.NoError(t, tra.Broadcast(messages.PriceMessageName, testutil.PriceAAABBB2))
	assert.NoError(t, tra.Broadcast(messages.PriceMessageName, testutil.PriceXXXYYY1))
	assert.NoError(t, tra.Broadcast(messages.PriceMessageName, testutil.PriceXXXYYY2))

	// Datastore fetches prices asynchronously, so we wait up to 1 second:
	var aaabbb, xxxyyy []*oracle.Price
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		aaabbb = ds.Prices().AssetPair("AAABBB").OraclePrices()
		xxxyyy = ds.Prices().AssetPair("XXXYYY").OraclePrices()
		if len(aaabbb) == 2 && len(xxxyyy) == 2 {
			break
		}
	}

	assert.Contains(t, aaabbb, testutil.PriceAAABBB1.Price)
	assert.Contains(t, aaabbb, testutil.PriceAAABBB2.Price)
	assert.Contains(t, xxxyyy, testutil.PriceXXXYYY1.Price)
	assert.Contains(t, xxxyyy, testutil.PriceXXXYYY2.Price)
}
