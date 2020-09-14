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
	"sort"
	"strings"
	"time"

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

			return []marshalledItem{[]byte(strings.Join(strs, "\n"))}, nil
		}

		var err error
		var ret []marshalledItem

		switch i := item.(type) {
		case graph.IndirectTick:
			traceHandleTick(&ret, i)
		case graph.Aggregator:
			traceHandleGraph(&ret, i)
		case map[graph.Pair][]string:
			traceHandleOrigins(&ret, i)
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
	str := renderTree(func(node interface{}) ([]byte, []interface{}) {
		var c []interface{}
		var s string

		switch typedTick := node.(type) {
		case graph.IndirectTick:
			s = fmt.Sprintf(
				"IndirectTick(%s, %f, %s)",
				typedTick.Pair,
				typedTick.Price,
				typedTick.Timestamp.Format(time.RFC3339),
			)

			if typedTick.Error != nil {
				s = "[IGNORED] " + s + fmt.Sprintf(
					"\nError: %s",
					strings.TrimSpace(typedTick.Error.Error()),
				)
			}

			for _, t := range typedTick.OriginTicks {
				c = append(c, t)
			}
			for _, t := range typedTick.IndirectTicks {
				c = append(c, t)
			}
		case graph.OriginTick:
			s = fmt.Sprintf(
				"OriginTick(%s, %s, %f, %s)",
				typedTick.Pair,
				typedTick.Origin,
				typedTick.Price,
				typedTick.Timestamp.Format(time.RFC3339),
			)

			if typedTick.Error != nil {
				s = "[IGNORED] " + s + fmt.Sprintf(
					"\nError: %s",
					strings.TrimSpace(typedTick.Error.Error()),
				)
			}
		}

		return []byte(s), c
	}, []interface{}{t}, 0)

	*ret = append(*ret, []byte(fmt.Sprintf("Price for %s:", t.Pair)))
	*ret = append(*ret, str)
}

func traceHandleGraph(ret *[]marshalledItem, g graph.Aggregator) {
	str := renderTree(func(node interface{}) ([]byte, []interface{}) {
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
			s = reflect.TypeOf(node).Elem().String()
		}

		return []byte(s), c
	}, []interface{}{g}, 0)

	*ret = append(*ret, []byte(fmt.Sprintf("Graph for %s:", g.Pair())))
	*ret = append(*ret, str)
}

func traceHandleOrigins(ret *[]marshalledItem, origins map[graph.Pair][]string) {
	type originPair struct {
		pair    graph.Pair
		origins []string
	}

	var s []interface{}
	for p, o := range origins {
		s = append(s, originPair{pair: p, origins: o})
	}

	sort.Slice(s, func(i, j int) bool {
		return s[i].(originPair).pair.String() > s[j].(originPair).pair.String()
	})

	str := renderTree(func(node interface{}) ([]byte, []interface{}) {
		var c []interface{}
		var s string

		switch typedNode := node.(type) {
		case originPair:
			s = typedNode.pair.String()
			for _, o := range typedNode.origins {
				c = append(c, o)
			}
		case string:
			s = typedNode
		}

		return []byte(s), c
	}, s, 0)

	*ret = append(*ret, str)
}

//nolint:gocyclo
func renderTree(printer func(interface{}) ([]byte, []interface{}), nodes []interface{}, level int) []byte {
	const (
		first  = "┌──"
		middle = "├──"
		last   = "└──"
		vline  = "│  "
		hline  = "───"
		empty  = "   "
	)

	s := bytes.Buffer{}
	for i, node := range nodes {
		nodeStr, nodeChildren := printer(node)
		isFirst := i == 0
		isLast := i == len(nodes)-1
		hasChild := len(nodeChildren) > 0
		firstLinePrefix := ""
		restLinesPrefix := ""

		switch {
		case level == 0 && isFirst && isLast:
			firstLinePrefix = hline
		case level == 0 && isFirst:
			firstLinePrefix = first
		case isLast:
			firstLinePrefix = last
		default:
			firstLinePrefix = middle
		}

		switch {
		case isLast && hasChild:
			restLinesPrefix = empty + vline
		case !isLast && hasChild:
			restLinesPrefix = vline + vline
		case isLast && !hasChild:
			restLinesPrefix = empty + empty
		case !isLast && !hasChild:
			restLinesPrefix = vline + empty
		}

		s.Write(prependLines(nodeStr, firstLinePrefix, restLinesPrefix))
		s.WriteByte('\n')

		if len(nodeChildren) > 0 {
			subTree := renderTree(printer, nodeChildren, level+1)

			if isLast {
				subTree = prependLines(subTree, empty, empty)
			} else {
				subTree = prependLines(subTree, vline, vline)
			}

			s.Write(subTree)
			s.WriteByte('\n')
		}
	}

	return s.Bytes()
}

func prependLines(s []byte, first, rest string) []byte {
	bts := bytes.Buffer{}
	bts.WriteString(first)
	bts.Write(bytes.ReplaceAll(bytes.TrimRight(s, "\n"), []byte{'\n'}, append([]byte{'\n'}, rest...)))
	return bts.Bytes()
}
