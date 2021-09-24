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

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecover_WithCallback(t *testing.T) {
	var panicVal interface{}
	defer func() {
		assert.Nil(t, recover())
		assert.Equal(t, panicVal, "panic")
	}()
	h := (&Recover{Recover: func(err interface{}) {
		panicVal = err
	}}).Handle(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		panic("panic")
	}))
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
}

func TestRecover_WithoutCallback(t *testing.T) {
	defer func() {
		assert.Nil(t, recover())
	}()
	h := (&Recover{}).Handle(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		panic("panic")
	}))
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
}
