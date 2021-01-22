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
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/pkg/transport"
)

type testMessageHandler struct {
	Topic   string
	Raw     []byte
	Message transport.Message
}

func (t *testMessageHandler) Published(topic string, raw []byte, message transport.Message) {
	t.Topic = topic
	t.Raw = raw
	t.Message = message
}

func (t *testMessageHandler) Received(topic string, raw *pubsub.Message, message transport.Message) {
	t.Topic = topic
	t.Raw = raw.Data
	t.Message = message
}

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

func TestMessageHandlerSet_Published(t *testing.T) {
	mhs := NewMessageHandlerSet()

	msg := &testMsg{Val: "abc"}
	mh1 := &testMessageHandler{}
	mh2 := &testMessageHandler{}

	mhs.Add(mh1, mh2)
	mhs.Published("foo", []byte("bar"), msg)

	// All message handlers should be invoked:
	assert.Equal(t, "foo", mh1.Topic)
	assert.Equal(t, "foo", mh2.Topic)
	assert.Equal(t, []byte("bar"), mh1.Raw)
	assert.Equal(t, []byte("bar"), mh2.Raw)
	assert.Equal(t, msg, mh1.Message)
	assert.Equal(t, msg, mh2.Message)
}

func TestMessageHandlerSet_Received(t *testing.T) {
	mhs := NewMessageHandlerSet()

	msg := &testMsg{Val: "abc"}
	mh1 := &testMessageHandler{}
	mh2 := &testMessageHandler{}

	mhs.Add(mh1, mh2)
	mhs.Received("foo", &pubsub.Message{
		Message: &pb.Message{
			Data: []byte("bar"),
		},
	}, msg)

	// All message handlers should be invoked:
	assert.Equal(t, "foo", mh1.Topic)
	assert.Equal(t, "foo", mh2.Topic)
	assert.Equal(t, []byte("bar"), mh1.Raw)
	assert.Equal(t, []byte("bar"), mh2.Raw)
	assert.Equal(t, msg, mh1.Message)
	assert.Equal(t, msg, mh2.Message)
}
