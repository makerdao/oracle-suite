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

package marshal

import (
	"io/ioutil"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMarshaller(t *testing.T) {
	expectedMap := map[FormatType]interface{}{
		Plain:  (*plain)(nil),
		JSON:   (*json)(nil),
		NDJSON: (*json)(nil),
		Trace:  (*trace)(nil),
	}
	formatMap := map[FormatType]string{
		Plain:  "plain",
		JSON:   "json",
		NDJSON: "ndjson",
		Trace:  "trace",
	}
	for ct, st := range formatMap {
		t.Run(st, func(t *testing.T) {
			m, err := NewMarshal(ct)

			assert.Nil(t, err)
			assert.Implements(t, (*Marshaller)(nil), m)
			assert.IsType(t, m, &Marshal{})
			assert.IsType(t, m.marshaller, expectedMap[ct])
		})
	}
}

func TestBufferedMarshaller_RW(t *testing.T) {
	for _, live := range []bool{false, true} {
		t.Run("live:"+strconv.FormatBool(live), func(t *testing.T) {
			bm := newBufferedMarshaller(live, func(i interface{}) ([]marshalledItem, error) {
				if s, ok := i.([]marshalledItem); ok {
					// this code should be called only when the live var is set to false
					s = append(s, []byte("notlive"))
					return s, nil
				}

				return []marshalledItem{marshalledItem(i.(string))}, nil
			})

			assert.Nil(t, bm.Write("foo"))
			assert.Nil(t, bm.Write("bar"))
			assert.Nil(t, bm.Close())

			r, err := ioutil.ReadAll(bm)

			if live {
				assert.Equal(t, "foobar", string(r))
			} else {
				assert.Equal(t, "foobarnotlive", string(r))
			}

			assert.Nil(t, err)
		})
	}
}

func TestBufferedMarshaller_RW_Async(t *testing.T) {
	for _, live := range []bool{false, true} {
		t.Run("live:"+strconv.FormatBool(live), func(t *testing.T) {
			bm := newBufferedMarshaller(live, func(i interface{}) ([]marshalledItem, error) {
				if s, ok := i.([]marshalledItem); ok {
					// this code should be called only when the live var is set to false
					s = append(s, []byte("notlive"))
					return s, nil
				}

				return []marshalledItem{marshalledItem(i.(string))}, nil
			})

			var r []byte
			var err error

			wg := sync.WaitGroup{}

			wg.Add(1)
			go func() {
				r, err = ioutil.ReadAll(bm)
				wg.Done()
			}()

			wg.Add(1)
			go func() {
				assert.Nil(t, bm.Write("foo"))
				assert.Nil(t, bm.Write("bar"))
				assert.Nil(t, bm.Close())
				wg.Done()
			}()

			wg.Wait()

			if live {
				assert.Equal(t, "foobar", string(r))
			} else {
				assert.Equal(t, "foobarnotlive", string(r))
			}

			assert.Nil(t, err)
		})
	}
}
