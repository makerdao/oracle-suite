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
	"errors"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/makerdao/gofer/internal/marshal"
)

func StartServer(addr string) error {
	log.Printf("[WEB] starting server at %s", addr)
	return http.ListenAndServe(addr, nil)
}

func internalServerError(w http.ResponseWriter, srvErr error) {
	log.Printf("[WEB] 500: %s", srvErr.Error())
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
			log.Printf("[WEB] %s", err.Error())
		}
	}()
	return func() {
		wg.Wait()
	}
}

func recoverHandler() func() {
	return func() {
		if r := recover(); r != nil {
			var err error
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("unknown panic")
			}

			if err != nil {
				log.Printf("[WEB] recovered panic: %s", err.Error())
			}
		}
	}
}

func marshallerHandler(
	handler func(m marshal.Marshaller, r *http.Request) error,
) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		defer recoverHandler()

		m, err := marshal.NewMarshal(marshal.JSON)
		if err != nil {
			internalServerError(w, err)
			return
		}

		asJSON(w)

		wait := asyncCopy(w, m)
		defer func() {
			if err := m.Close(); err != nil {
				internalServerError(w, err)
				return
			}

			wait()
		}()

		if err := handler(m, r); err != nil {
			internalServerError(w, err)
			return
		}
	}
}
