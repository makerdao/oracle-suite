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
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/makerdao/gofer/pkg/graph"
)

type trace struct {
	bufferedMarshaller *bufferedMarshaller
}

func newTrace() *trace {
	return &trace{newBufferedMarshaller(false, func(item interface{}, ierr error) ([]marshalledItem, error) {
		if i, ok := item.([]marshalledItem); ok {
			strs := make([]string, len(i))
			for n, s := range i {
				strs[n] = string(s)
			}

			return []marshalledItem{[]byte(strings.Join(strs, "\n") + "\n")}, nil
		}

		var err error
		var ret []marshalledItem

		switch i := item.(type) {
		case graph.IndirectTick:
			traceHandleTick(&ret, i)
		case graph.Aggregator:
			traceHandleGraph(&ret, i)
		case string:
			traceHandleString(&ret, i)
		default:
			return nil, fmt.Errorf("unsupported data type")
		}

		return ret, err
	})}
}

// Read implements the Marshaller interface.
func (j *trace) Read(p []byte) (int, error) {
	return j.bufferedMarshaller.Read(p)
}

// Write implements the Marshaller interface.
func (j *trace) Write(item interface{}, err error) error {
	return j.bufferedMarshaller.Write(item, err)
}

// Close implements the Marshaller interface.
func (j *trace) Close() error {
	return j.bufferedMarshaller.Close()
}

func traceHandleTick(ret *[]marshalledItem, t graph.IndirectTick) {
	buf := &bytes.Buffer{}

	var recur func(interface{}, int)
	recur = func(tick interface{}, level int) {
		prefix := strings.Repeat(" ", level)
		switch typedTick := tick.(type) {
		case graph.IndirectTick:
			fmt.Fprintf(
				buf,
				"%sAggregator(%s)=%f\n",
				prefix,
				typedTick.Pair,
				typedTick.Price,
			)

			if typedTick.Error != nil {
				fmt.Fprintf(
					buf,
					"%s error: %s\n",
					prefix,
					strings.TrimSpace(typedTick.Error.Error()),
				)
			}

			for _, t := range typedTick.ExchangeTicks {
				recur(t, level+1)
			}
			for _, t := range typedTick.IndirectTick {
				recur(t, level+1)
			}
		case graph.ExchangeTick:
			fmt.Fprintf(
				buf,
				"%sExchange(%s, %s)=%f\n",
				prefix,
				typedTick.Pair,
				typedTick.Exchange,
				typedTick.Price,
			)

			if typedTick.Error != nil {
				fmt.Fprintf(
					buf,
					"%s error: %s\n",
					prefix,
					strings.TrimSpace(typedTick.Error.Error()),
				)
			}
		}
	}

	fmt.Fprintf(buf, "Trace for %s pair:\n", t.Pair)
	recur(t, 0)

	*ret = append(*ret, buf.Bytes())
}

func traceHandleGraph(ret *[]marshalledItem, g graph.Aggregator) {
	buf := &bytes.Buffer{}

	var recur func(graph.Node, int)
	recur = func(node graph.Node, level int) {
		prefix := strings.Repeat(" ", level)

		switch typedNode := node.(type) {
		case graph.Aggregator:
			fmt.Fprintf(
				buf,
				"%s%s(%s)\n",
				prefix,
				reflect.TypeOf(node).Elem().String(),
				typedNode.Pair(),

			)
		case graph.Exchange:
			fmt.Fprintf(
				buf,
				"%s%s(%s, %s)\n",
				prefix,
				reflect.TypeOf(node).Elem().String(),
				typedNode.ExchangePair().Pair,
				typedNode.ExchangePair().Exchange,
			)
		default:
			fmt.Fprintf(
				buf,
				"%s%s\n",
				prefix,
				reflect.TypeOf(node).Elem().String(),
			)
		}

		for _, c := range node.Children() {
			recur(c, level+1)
		}
	}

	fmt.Fprintf(buf, "Graph for %s pair:\n", g.Pair())
	recur(g, 0)

	*ret = append(*ret, buf.Bytes())
}

func traceHandleString(ret *[]marshalledItem, s string) {
	*ret = append(*ret, []byte(s))
}
