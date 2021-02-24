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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/pkg/gofer/graph"

	"github.com/makerdao/gofer/internal/gofer/marshal/testutil"
)

func TestPlain_Graph(t *testing.T) {
	var err error
	j := newPlain()

	err = j.Write(testutil.Graph(graph.Pair{Base: "A", Quote: "B"}))
	assert.NoError(t, err)

	err = j.Write(testutil.Graph(graph.Pair{Base: "C", Quote: "D"}))
	assert.NoError(t, err)

	b, err := j.Bytes()
	assert.NoError(t, err)

	expected := `
A/B
C/D
`[1:]

	assert.Equal(t, expected, string(b))
}

func TestPlain_Ticks(t *testing.T) {
	var err error

	ab := testutil.Graph(graph.Pair{Base: "A", Quote: "B"})
	cd := testutil.Graph(graph.Pair{Base: "C", Quote: "D"})
	j := newPlain()

	err = j.Write(ab.Tick())
	assert.NoError(t, err)

	cdt := cd.Tick()
	cdt.Error = errors.New("something")
	err = j.Write(cdt)
	assert.NoError(t, err)

	b, err := j.Bytes()
	assert.NoError(t, err)

	expected := `
A/B 10.000000
C/D - something
`[1:]

	assert.Equal(t, expected, string(b))
}

func TestPlain_Origins(t *testing.T) {
	p := graph.Pair{Base: "A", Quote: "B"}
	j := newPlain()

	err := j.Write(map[graph.Pair][]string{
		p: {"a", "b", "c"},
	})

	assert.NoError(t, err)

	b, _ := j.Bytes()

	expected := `
A/B:
a
b
c
`[1:]

	assert.Equal(t, expected, string(b))
}
