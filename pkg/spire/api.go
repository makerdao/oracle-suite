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

package spire

import (
	"github.com/makerdao/gofer/pkg/datastore"
	"github.com/makerdao/gofer/pkg/transport"
	"github.com/makerdao/gofer/pkg/transport/messages"
)

type NoArgument = struct{}

type Datastore interface {
	Prices() *datastore.PriceStore
	Start() error
	Stop() error
}

type API struct {
	transport transport.Transport
	datastore Datastore
}

func (n *API) BroadcastPrice(price *messages.Price, _ *NoArgument) error {
	return n.transport.Broadcast(messages.PriceMessageName, price)
}

func (n *API) GetPrices(assetPair *string, prices *[]*messages.Price) error {
	*prices = n.datastore.Prices().AssetPair(*assetPair).Messages()
	return nil
}
