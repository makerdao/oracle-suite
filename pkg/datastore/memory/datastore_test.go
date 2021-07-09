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

package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/makerdao/oracle-suite/pkg/datastore/memory/testutil"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/ethereum/mocks"
	"github.com/makerdao/oracle-suite/pkg/log/null"
	"github.com/makerdao/oracle-suite/pkg/oracle"
	"github.com/makerdao/oracle-suite/pkg/transport"
	"github.com/makerdao/oracle-suite/pkg/transport/local"
	"github.com/makerdao/oracle-suite/pkg/transport/messages"
)

func TestDatastore_Prices(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	sig := &mocks.Signer{}
	tra := local.New(ctx, 0, map[string]transport.Message{messages.PriceMessageName: (*messages.Price)(nil)})

	ds, err:= NewDatastore(Config{
		Context:   ctx,
		Signer:    sig,
		Transport: tra,
		Pairs: map[string]*Pair{
			"AAABBB": {Feeds: []ethereum.Address{testutil.Address1, testutil.Address2}},
			"XXXYYY": {Feeds: []ethereum.Address{testutil.Address1, testutil.Address2}},
		},
		Logger: null.New(),
	})
	require.NoError(t, err)
	require.NoError(t, ds.Start())

	sig.On("Recover", testutil.PriceAAABBB1.Price.Signature(), mock.Anything).Return(&testutil.Address1, nil)
	sig.On("Recover", testutil.PriceAAABBB2.Price.Signature(), mock.Anything).Return(&testutil.Address2, nil)
	sig.On("Recover", testutil.PriceXXXYYY1.Price.Signature(), mock.Anything).Return(&testutil.Address1, nil)
	sig.On("Recover", testutil.PriceXXXYYY2.Price.Signature(), mock.Anything).Return(&testutil.Address2, nil)

	assert.NoError(t, tra.Broadcast(messages.PriceMessageName, testutil.PriceAAABBB1))
	assert.NoError(t, tra.Broadcast(messages.PriceMessageName, testutil.PriceAAABBB2))
	assert.NoError(t, tra.Broadcast(messages.PriceMessageName, testutil.PriceXXXYYY1))
	assert.NoError(t, tra.Broadcast(messages.PriceMessageName, testutil.PriceXXXYYY2))

	// Datastore fetches prices asynchronously, so we wait up to 1 second:
	var aaabbb, xxxyyy []*messages.Price
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		aaabbb = ds.Prices().AssetPair("AAABBB")
		xxxyyy = ds.Prices().AssetPair("XXXYYY")
		if len(aaabbb) == 2 && len(xxxyyy) == 2 {
			break
		}
	}

	assert.Contains(t, toOraclePrices(aaabbb), testutil.PriceAAABBB1.Price)
	assert.Contains(t, toOraclePrices(aaabbb), testutil.PriceAAABBB2.Price)
	assert.Contains(t, toOraclePrices(xxxyyy), testutil.PriceXXXYYY1.Price)
	assert.Contains(t, toOraclePrices(xxxyyy), testutil.PriceXXXYYY2.Price)
}

func toOraclePrices(ps []*messages.Price) []*oracle.Price {
	var r []*oracle.Price
	for _, p := range ps {
		r = append(r, p.Price)
	}
	return r
}
