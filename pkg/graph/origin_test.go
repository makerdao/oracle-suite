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

package graph

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOriginNode_OriginPair(t *testing.T) {
	op := OriginPair{
		Origin: "foo",
		Pair:   Pair{Base: "A", Quote: "B"},
	}

	o := NewOriginNode(op)
	assert.Equal(t, op, o.OriginPair())
}

func TestOriginNode_Ingest_Valid(t *testing.T) {
	op := OriginPair{
		Origin: "foo",
		Pair:   Pair{Base: "A", Quote: "B"},
	}

	ot := OriginTick{
		Tick: Tick{
			Pair:      Pair{Base: "A", Quote: "B"},
			Price:     10,
			Bid:       10,
			Ask:       10,
			Volume24h: 10,
			Timestamp: time.Unix(0, 0),
		},
		Origin: "foo",
		Error:  nil,
	}

	o := NewOriginNode(op)
	err := o.Ingest(ot)

	assert.Equal(t, op, o.OriginPair())
	assert.Equal(t, ot, o.Tick())
	assert.Nil(t, err)
}

func TestOriginNode_Ingest_IncompatiblePair(t *testing.T) {
	op := OriginPair{
		Origin: "foo",
		Pair:   Pair{Base: "A", Quote: "B"},
	}

	ot := OriginTick{
		Tick: Tick{
			Pair:      Pair{Base: "A", Quote: "C"},
			Price:     10,
			Bid:       10,
			Ask:       10,
			Volume24h: 10,
			Timestamp: time.Unix(0, 0),
		},
		Origin: "foo",
		Error:  nil,
	}

	o := NewOriginNode(op)
	err := o.Ingest(ot)

	assert.True(t, errors.As(err, &IngestedIncompatiblePairErr{}))
	assert.Equal(t, err, o.tick.Error)
}

func TestOriginNode_Ingest_IncompatibleOrigin(t *testing.T) {
	op := OriginPair{
		Origin: "foo",
		Pair:   Pair{Base: "A", Quote: "B"},
	}

	ot := OriginTick{
		Tick: Tick{
			Pair:      Pair{Base: "A", Quote: "B"},
			Price:     10,
			Bid:       10,
			Ask:       10,
			Volume24h: 10,
			Timestamp: time.Unix(0, 0),
		},
		Origin: "bar",
		Error:  nil,
	}

	o := NewOriginNode(op)
	err := o.Ingest(ot)

	assert.True(t, errors.As(err, &IngestedIncompatibleOriginErr{}))
	assert.Equal(t, err, o.tick.Error)
}

func TestOriginNode_Ingest_IncompatibleEverything(t *testing.T) {
	op := OriginPair{
		Origin: "foo",
		Pair:   Pair{Base: "A", Quote: "B"},
	}

	ot := OriginTick{
		Tick: Tick{
			Pair:      Pair{Base: "A", Quote: "C"},
			Price:     10,
			Bid:       10,
			Ask:       10,
			Volume24h: 10,
			Timestamp: time.Unix(0, 0),
		},
		Origin: "bar",
		Error:  nil,
	}

	o := NewOriginNode(op)
	err := o.Ingest(ot)

	assert.True(t, errors.As(err, &IngestedIncompatibleOriginErr{}))
	assert.True(t, errors.As(err, &IngestedIncompatiblePairErr{}))
	assert.Equal(t, err, o.tick.Error)
}

func TestOriginNode_Ingest_TickWithError(t *testing.T) {
	err := errors.New("something")

	op := OriginPair{
		Origin: "foo",
		Pair:   Pair{Base: "A", Quote: "B"},
	}

	ot := OriginTick{
		Tick: Tick{
			Pair:      Pair{Base: "A", Quote: "B"},
			Price:     10,
			Bid:       10,
			Ask:       10,
			Volume24h: 10,
			Timestamp: time.Unix(0, 0),
		},
		Origin: "foo",
		Error:  err,
	}

	o := NewOriginNode(op)
	err2 := o.Ingest(ot)

	assert.Nil(t, err2)
	assert.Equal(t, err, o.tick.Error)
}
