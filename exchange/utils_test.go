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

package exchange

import (
	"makerdao/gofer/model"
	"makerdao/gofer/query"
)

// mockWorkerPool mock worker pool implementation for tests
type mockWorkerPool struct {
	resp *query.HTTPResponse
}

func newMockWorkerPool(resp *query.HTTPResponse) *mockWorkerPool {
	return &mockWorkerPool{
		resp: resp,
	}
}

func (mwp *mockWorkerPool) Start() {}

func (mwp *mockWorkerPool) Stop() error {
	return nil
}

func (mwp *mockWorkerPool) Query(req *query.HTTPRequest) *query.HTTPResponse {
	return mwp.resp
}

func newPotentialPricePoint(exchangeName, base, quote string) *model.PotentialPricePoint {
	p := &model.Pair{
		Base:  base,
		Quote: quote,
	}
	return &model.PotentialPricePoint{
		Exchange: &model.Exchange{
			Name: exchangeName,
		},
		Pair: p,
	}
}