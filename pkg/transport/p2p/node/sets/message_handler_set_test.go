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

	"github.com/makerdao/oracle-suite/pkg/transport"
)

type testMessageHandler struct {
	topic  string
	raw    []byte
	psMsg  *pubsub.Message
	msg    transport.Message
	result pubsub.ValidationResult
	err    error
}

func (t *testMessageHandler) Published(topic string, raw []byte, msg transport.Message) {
	t.topic = topic
	t.raw = raw
	t.msg = msg
}

func (t *testMessageHandler) Received(topic string, msg *pubsub.Message, result pubsub.ValidationResult) {
	t.topic = topic
	t.raw = msg.Data
	t.psMsg = msg
	t.result = result
}

func (t *testMessageHandler) Broken(topic string, msg *pubsub.Message, err error) {
	t.topic = topic
	t.raw = msg.Data
	t.psMsg = msg
	t.err = err
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
	assert.Equal(t, "foo", mh1.topic)
	assert.Equal(t, "foo", mh2.topic)
	assert.Equal(t, []byte("bar"), mh1.raw)
	assert.Equal(t, []byte("bar"), mh2.raw)
	assert.Equal(t, msg, mh1.msg)
	assert.Equal(t, msg, mh2.msg)
}

func TestMessageHandlerSet_Received(t *testing.T) {
	mhs := NewMessageHandlerSet()

	msg := &pubsub.Message{
		Message: &pb.Message{
			Data: []byte("bar"),
		},
	}
	mh1 := &testMessageHandler{}
	mh2 := &testMessageHandler{}

	mhs.Add(mh1, mh2)
	mhs.Received("foo", msg, pubsub.ValidationAccept)

	// All message handlers should be invoked:
	assert.Equal(t, "foo", mh1.topic)
	assert.Equal(t, "foo", mh2.topic)
	assert.Equal(t, []byte("bar"), mh1.raw)
	assert.Equal(t, []byte("bar"), mh2.raw)
	assert.Equal(t, msg, mh1.psMsg)
	assert.Equal(t, msg, mh2.psMsg)
	assert.Equal(t, pubsub.ValidationAccept, mh1.result)
	assert.Equal(t, pubsub.ValidationAccept, mh2.result)
}
