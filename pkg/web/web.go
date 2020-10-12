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
	"io"
	"log"
	"net/http"
	"sync"
)

func badRequest(w http.ResponseWriter, srvErr ...error) {
	log.Println(srvErr)
	w.WriteHeader(http.StatusBadRequest)
}
func internalServerError(w http.ResponseWriter, srvErr ...error) {
	log.Println(srvErr)
	w.WriteHeader(http.StatusInternalServerError)
}

func asJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func asyncCopy(dst io.Writer, src io.Reader) func() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_, err := io.Copy(dst, src)
		wg.Done()
		if err == io.EOF {
			return
		}
		if err != nil {
			panic(err.Error())
		}
	}()
	return func() {
		wg.Wait()
	}
}

func closeAndFinish(c io.Closer, w http.ResponseWriter, done func()) {
	if err := c.Close(); err != nil {
		badRequest(w, err)
		return
	}
	done()
}
