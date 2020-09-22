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
		case graph.AggregatorTick:
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

func traceHandleTick(ret *[]marshalledItem, t graph.AggregatorTick) {
	str := renderTree(func(node interface{}) ([]byte, []interface{}) {
		var c []interface{}
		var s []byte

		switch typedTick := node.(type) {
		case graph.AggregatorTick:
			s = renderNode(
				"AggregatorTick",
				mergeKVMap(
					[]param{
						{key: "pair", value: typedTick.Pair.String()},
						{key: "price", value: typedTick.Price},
						{key: "timestamp", value: typedTick.Timestamp.Format(time.RFC3339Nano)},
					},
					typedTick.Parameters,
				),
				typedTick.Error,
			)

			for _, t := range typedTick.OriginTicks {
				c = append(c, t)
			}
			for _, t := range typedTick.AggregatorTicks {
				c = append(c, t)
			}
		case graph.OriginTick:
			s = renderNode(
				"OriginTick",
				[]param{
					{key: "pair", value: typedTick.Pair.String()},
					{key: "origin", value: typedTick.Origin},
					{key: "price", value: typedTick.Price},
					{key: "timestamp", value: typedTick.Timestamp.Format(time.RFC3339Nano)},
				},
				typedTick.Error,
			)
		}

		return s, c
	}, []interface{}{t}, 0)

	*ret = append(*ret, []byte(fmt.Sprintf("Price for %s:", t.Pair)))
	*ret = append(*ret, str)
}

func traceHandleGraph(ret *[]marshalledItem, g graph.Aggregator) {
	str := renderTree(func(node interface{}) ([]byte, []interface{}) {
		var c []interface{}
		var s []byte

		switch typedNode := node.(type) {
		case graph.Aggregator:
			s = renderNode(
				reflect.TypeOf(node).Elem().String(),
				[]param{
					{key: "pair", value: typedNode.Pair().String()},
				},
				nil,
			)

			for _, n := range typedNode.Children() {
				c = append(c, n)
			}
		case graph.Origin:
			s = renderNode(
				reflect.TypeOf(node).Elem().String(),
				[]param{
					{key: "pair", value: typedNode.OriginPair().Pair},
					{key: "origin", value: typedNode.OriginPair().Origin},
				},
				nil,
			)
		}

		return s, c
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

type param struct {
	key   string
	value interface{}
}

func renderNode(typ string, params []param, err error) []byte {
	str := bytes.Buffer{}
	if err != nil {
		str.WriteString(color("[ERROR] ", red))
	}

	str.WriteString(typ)
	str.WriteString("(")
	for i, p := range params {
		str.WriteString(color(p.key, green))
		str.WriteString(":")
		str.WriteString(fmt.Sprintf("%v", p.value))
		if i != len(params)-1 {
			str.WriteString(", ")
		}
	}
	str.WriteString(")")
	if err != nil {
		str.WriteString("\n")
		str.WriteString(color("Error: " + strings.TrimSpace(err.Error()), red))
	}

	return str.Bytes()
}

//nolint:gocyclo
func renderTree(printer func(interface{}) ([]byte, []interface{}), nodes []interface{}, level int) []byte {
	const (
		first  = string(green + "┌──" + reset)
		middle = string(green + "├──" + reset)
		last   = string(green + "└──" + reset)
		vline  = string(green + "│  " + reset)
		hline  = string(green + "───" + reset)
		empty  = string(green + "   " + reset)
	)

	s := bytes.Buffer{}
	for i, node := range nodes {
		nodeStr, nodeChildren := printer(node)
		isFirst := i == 0
		isLast := i == len(nodes)-1
		hasChild := len(nodeChildren) > 0
		firstLinePrefix := string(reset)
		restLinesPrefix := string(reset)

		switch {
		case level == 0 && isFirst && isLast:
			firstLinePrefix += hline
		case level == 0 && isFirst:
			firstLinePrefix += first
		case isLast:
			firstLinePrefix += last
		default:
			firstLinePrefix += middle
		}

		switch {
		case isLast && hasChild:
			restLinesPrefix += empty + vline
		case !isLast && hasChild:
			restLinesPrefix += vline + vline
		case isLast && !hasChild:
			restLinesPrefix += empty + empty
		case !isLast && !hasChild:
			restLinesPrefix += vline + empty
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

func mergeKVMap(target []param, kv map[string]string) []param {
	for _, k := range sortKeys(kv) {
		target = append(target, param{key: k, value: kv[k]})
	}
	return target
}

func sortKeys(kv map[string]string) []string {
	var ks []string
	for k := range kv {
		ks = append(ks, k)
	}

	sort.Strings(ks)

	return ks
}
