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
	"fmt"
	"strings"

	"github.com/makerdao/gofer/pkg/gofer/graph"
)

type plain struct {
	bufferedMarshaller *bufferedMarshaller
}

func newPlain() *plain {
	return &plain{newBufferedMarshaller(false, func(item interface{}) ([]marshalledItem, error) {
		if i, ok := item.([]marshalledItem); ok {
			strs := make([]string, len(i))
			for n, s := range i {
				strs[n] = string(s)
			}

			return []marshalledItem{[]byte(strings.Join(strs, "\n"))}, nil
		}

		var err error
		var ret []marshalledItem

		switch i := item.(type) {
		case graph.AggregatorTick:
			plainHandleTick(&ret, i)
		case graph.Aggregator:
			plainHandleGraph(&ret, i)
		case map[graph.Pair][]string:
			plainHandleOrigins(&ret, i)
		default:
			return nil, fmt.Errorf("unsupported data type")
		}

		return ret, err
	})}
}

// Read implements the Marshaller interface.
func (j *plain) Read(p []byte) (int, error) {
	return j.bufferedMarshaller.Read(p)
}

// Write implements the Marshaller interface.
func (j *plain) Write(item interface{}) error {
	return j.bufferedMarshaller.Write(item)
}

// Close implements the Marshaller interface.
func (j *plain) Close() error {
	return j.bufferedMarshaller.Close()
}

func plainHandleTick(ret *[]marshalledItem, tick graph.AggregatorTick) {
	if tick.Error != nil {
		*ret = append(*ret, []byte(fmt.Sprintf("%s - %s", tick.Pair.String(), strings.TrimSpace(tick.Error.Error()))))
	} else {
		*ret = append(*ret, []byte(fmt.Sprintf("%s %f", tick.Pair.String(), tick.Price)))
	}
}

func plainHandleGraph(ret *[]marshalledItem, graph graph.Aggregator) {
	*ret = append(*ret, []byte(graph.Pair().String()))
}

func plainHandleOrigins(ret *[]marshalledItem, origins map[graph.Pair][]string) {
	for p, os := range origins {
		*ret = append(*ret, []byte(p.String()+":"))
		for _, o := range os {
			*ret = append(*ret, []byte(o))
		}
		*ret = append(*ret, []byte{})
	}
}
