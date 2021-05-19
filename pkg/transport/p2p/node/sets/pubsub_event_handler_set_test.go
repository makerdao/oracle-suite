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

package sets

import (
	"testing"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
)

func TestPubSubEventHandlerSet_Handle(t *testing.T) {
	ehs := NewPubSubEventHandlerSet()

	pe := pubsub.PeerEvent{
		Type: 1,
		Peer: "a",
	}

	// All event handlers should be invoked:
	calls := 0
	ehs.Add(PubSubEventHandlerFunc(func(topic string, event pubsub.PeerEvent) {
		assert.Equal(t, "foo", topic)
		assert.Equal(t, pe, event)
		calls++
	}))
	ehs.Add(PubSubEventHandlerFunc(func(topic string, event pubsub.PeerEvent) {
		assert.Equal(t, "foo", topic)
		assert.Equal(t, pe, event)
		calls++
	}))
	ehs.Handle("foo", pe)
	assert.Equal(t, 2, calls)
}
