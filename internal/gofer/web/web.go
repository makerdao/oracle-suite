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
	"net/http"
	"sync"

	"github.com/makerdao/gofer/internal/gofer/marshal"
	"github.com/makerdao/gofer/pkg/log"
)

func StartServer(addr string, l log.Logger) error {
	l.WithField("addr", addr).Infof("Starting server")
	return http.ListenAndServe(addr, nil)
}

func internalServerError(w http.ResponseWriter, l log.Logger, err error) {
	l.WithError(err).Error("Internal server error")
	w.WriteHeader(http.StatusInternalServerError)
}

func asJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func asyncCopy(dst io.Writer, src io.Reader, l log.Logger) func() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_, err := io.Copy(dst, src)
		wg.Done()
		if err == io.EOF {
			return
		}
		if err != nil {
			l.WithError(err).Error("Error during serving data")
		}
	}()
	return func() {
		wg.Wait()
	}
}

func recoverHandler(l log.Logger) func() {
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
				l.WithError(err).Error("Recovered panic error")
			}
		}
	}
}

func marshallerHandler(
	handler func(m marshal.Marshaller, r *http.Request) error,
	l log.Logger,
) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		defer recoverHandler(l)

		m, err := marshal.NewMarshal(marshal.JSON)
		if err != nil {
			internalServerError(w, l, err)
			return
		}

		asJSON(w)

		wait := asyncCopy(w, m, l)
		defer func() {
			if err := m.Close(); err != nil {
				internalServerError(w, l, err)
				return
			}

			wait()
		}()

		if err := handler(m, r); err != nil {
			internalServerError(w, l, err)
			return
		}
	}
}
