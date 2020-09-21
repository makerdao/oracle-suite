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

package cli

import (
	"io"
	"log"
	"net/http"

	"github.com/makerdao/gofer/internal/marshal"
	"github.com/makerdao/gofer/pkg/graph"
)

func Server(args []string, l pricer) error {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		var err error

		srvErr := func(srvErr error) {
			log.Println(srvErr)
			writer.WriteHeader(http.StatusBadRequest)
		}

		pairStr := request.URL.Query().Get("pair")
		pair, err := graph.NewPair(pairStr)
		if err != nil {
			srvErr(err)
			return
		}

		ticks, err := l.Ticks(pair)
		if err != nil {
			srvErr(err)
			return
		}

		m, _ := marshal.NewMarshal(marshal.JSON)
		for _, t := range ticks {
			err = m.Write(t, nil)
			if err != nil {
				srvErr(err)
				return
			}
		}

		err = m.Close()
		if err != nil {
			srvErr(err)
			return
		}

		writer.Header().Set("Content-Type", "application/json")

		_, err = io.Copy(writer, m)
		if err != nil {
			srvErr(err)
			return
		}
	})

	return http.ListenAndServe(":8080", nil)
}
