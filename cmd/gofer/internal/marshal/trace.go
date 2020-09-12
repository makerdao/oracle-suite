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
	str := printTree(func(node interface{}) (string, []interface{}) {
		var c []interface{}
		var s string

		switch typedTick := node.(type) {
		case graph.IndirectTick:
			s = fmt.Sprintf(
				"Aggregator(%s)=%f",
				typedTick.Pair,
				typedTick.Price,
			)

			if typedTick.Error != nil {
				s += fmt.Sprintf(
					"\nError: %s",
					strings.TrimSpace(typedTick.Error.Error()),
				)
			}

			for _, t := range typedTick.OriginTicks {
				c = append(c, t)
			}
			for _, t := range typedTick.IndirectTick {
				c = append(c, t)
			}
		case graph.OriginTick:
			s = fmt.Sprintf(
				"Origin(%s, %s)=%f",
				typedTick.Pair,
				typedTick.Origin,
				typedTick.Price,
			)

			if typedTick.Error != nil {
				s += fmt.Sprintf(
					"\nError: %s",
					strings.TrimSpace(typedTick.Error.Error()),
				)
			}
		}

		return s, c
	}, []interface{}{t})

	*ret = append(*ret, []byte(fmt.Sprintf("Price for %s:", t.Pair)))
	*ret = append(*ret, []byte(str))
}

func traceHandleGraph(ret *[]marshalledItem, g graph.Aggregator) {
	str := printTree(func(node interface{}) (string, []interface{}) {
		var c []interface{}
		var s string

		switch typedNode := node.(type) {
		case graph.Aggregator:
			s = fmt.Sprintf(
				"%s(%s)",
				reflect.TypeOf(node).Elem().String(),
				typedNode.Pair(),
			)

			for _, n := range typedNode.Children() {
				c = append(c, n)
			}
		case graph.Origin:
			s = fmt.Sprintf(
				"%s(%s, %s)",
				reflect.TypeOf(node).Elem().String(),
				typedNode.OriginPair().Pair,
				typedNode.OriginPair().Origin,
			)
		default:
			s = fmt.Sprintf(
				"%s",
				reflect.TypeOf(node).Elem().String(),
			)
		}

		return s, c
	}, []interface{}{g})

	*ret = append(*ret, []byte(fmt.Sprintf("Graph for %s:", g.Pair())))
	*ret = append(*ret, []byte(str))
}

func traceHandleString(ret *[]marshalledItem, s string) {
	*ret = append(*ret, []byte(s))
}

func printTree(printer func(interface{}) (string, []interface{}), nodes []interface{}) string {
	const middle = "├──"
	const last = "└──"
	const level = "│  "
	const empty = "   "

	prefixLines := func(str string, first, rest string) string {
		return first + strings.ReplaceAll(strings.TrimRight(str, "\n"), "\n", "\n"+rest)
	}

	str := strings.Builder{}
	for i, node := range nodes {
		nodeStr, nodeChildren := printer(node)
		isLast := i == len(nodes) - 1

		if isLast {
			str.WriteString(prefixLines(nodeStr, last, empty+level))
		} else {
			str.WriteString(prefixLines(nodeStr, middle, level+level))
		}

		str.WriteString("\n")

		if len(nodeChildren) > 0 {
			subTree := printTree(printer, nodeChildren)

			if isLast {
				subTree = prefixLines(subTree, empty, empty)
			} else {
				subTree = prefixLines(subTree, level, level)
			}

			str.WriteString(subTree)
			str.WriteString("\n")
		}
	}

	return str.String()
}
