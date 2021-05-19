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

package local

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testMsg struct {
	Val string
}

func (t *testMsg) Marshall() ([]byte, error) {
	return []byte(t.Val), nil
}

func (t *testMsg) Unmarshall(bytes []byte) error {
	t.Val = string(bytes)
	return nil
}

func TestLocal_Subscribe(t *testing.T) {
	l := New(0)

	assert.NoError(t, l.Subscribe("foo", (*testMsg)(nil)))
	assert.IsType(t, ErrAlreadySubscribed, l.Subscribe("foo", (*testMsg)(nil)))
}

func TestLocal_Unsubscribe(t *testing.T) {
	l := New(0)

	assert.NoError(t, l.Subscribe("foo", (*testMsg)(nil)))
	assert.NoError(t, l.Unsubscribe("foo"))
	assert.IsType(t, ErrNotSubscribed, l.Unsubscribe("foo"))
}

func TestLocal_Broadcast(t *testing.T) {
	l := New(1)

	// Valid message:
	vm := &testMsg{Val: "bar"}
	assert.IsType(t, ErrNotSubscribed, l.Broadcast("foo", vm))
	assert.NoError(t, l.Subscribe("foo", (*testMsg)(nil)))
	assert.NoError(t, l.Broadcast("foo", vm))
}

func TestLocal_WaitFor(t *testing.T) {
	l := New(1)

	// Should return nil for unsubscribed topic:
	assert.Nil(t, l.WaitFor("foo"))

	// Valid message:
	assert.NoError(t, l.Subscribe("foo", (*testMsg)(nil)))
	assert.NoError(t, l.Broadcast("foo", &testMsg{Val: "bar"}))
	assert.Equal(t, &testMsg{Val: "bar"}, (<-l.WaitFor("foo")).Message)
}
