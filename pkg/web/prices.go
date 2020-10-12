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

package web

import (
	"net/http"
	"net/url"

	"github.com/makerdao/gofer/internal/marshal"
	"github.com/makerdao/gofer/pkg/cli"
	"github.com/makerdao/gofer/pkg/graph"
)

func PricesHandler(l graph.PriceModels) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m, err := marshal.NewMarshal(marshal.JSON)
		if err != nil {
			badRequest(w, err)
			return
		}
		asJSON(w)
		defer closeAndFinish(m, w, asyncCopy(w, m))

		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			badRequest(w, err)
			return
		}
		if err := cli.Prices(values["pair"], l, m); err != nil {
			internalServerError(w, err)
			return
		}
	}
}
