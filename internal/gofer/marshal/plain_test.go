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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/internal/gofer/marshal/testutil"
	"github.com/makerdao/gofer/pkg/gofer"
)

func TestPlain_Nodes(t *testing.T) {
	var err error
	m := newPlain()

	ab := gofer.Pair{Base: "A", Quote: "B"}
	cd := gofer.Pair{Base: "C", Quote: "D"}
	ns := testutil.Models(ab, cd)

	err = m.Write(ns[ab])
	assert.NoError(t, err)

	err = m.Write(ns[cd])
	assert.NoError(t, err)

	b, err := m.Bytes()
	assert.NoError(t, err)

	expected := `
A/B
C/D
`[1:]

	assert.Equal(t, expected, string(b))
}

func TestPlain_Prices(t *testing.T) {
	var err error
	m := newPlain()

	ab := gofer.Pair{Base: "A", Quote: "B"}
	cd := gofer.Pair{Base: "C", Quote: "D"}
	ns := testutil.Prices(ab, cd)

	err = m.Write(ns[ab])
	assert.NoError(t, err)

	cdt := ns[cd]
	cdt.Error = "something"
	err = m.Write(ns[cd])
	assert.NoError(t, err)

	b, err := m.Bytes()
	assert.NoError(t, err)

	expected := `
A/B 10.000000
C/D - something
`[1:]

	assert.Equal(t, expected, string(b))
}
