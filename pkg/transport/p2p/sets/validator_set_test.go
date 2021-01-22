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
	"context"
	"testing"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
)

func TestValidatorSet_Validator_EmptySet(t *testing.T) {
	vs := NewValidatorSet()

	// If there is no validators, message should be accepted:
	assert.Equal(t, pubsub.ValidationAccept, vs.Validator("foo")(context.Background(), "", &pubsub.Message{}))
}

func TestValidatorSet_Validator_Accept(t *testing.T) {
	vs := NewValidatorSet()

	vs.Add(func(ctx context.Context, topic string, id peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
		assert.Equal(t, "foo", topic)
		return pubsub.ValidationAccept
	})

	// Message should be accepted:
	assert.Equal(t, pubsub.ValidationAccept, vs.Validator("foo")(context.Background(), "", &pubsub.Message{}))
}

func TestValidatorSet_Validator_RejectFirst(t *testing.T) {
	vs := NewValidatorSet()

	vs.Add(func(ctx context.Context, topic string, id peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
		assert.Equal(t, "foo", topic)
		return pubsub.ValidationReject
	})

	vs.Add(func(ctx context.Context, topic string, id peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
		assert.Equal(t, "foo", topic)
		return pubsub.ValidationAccept
	})

	// Message should be rejected if at least one validator rejects it:
	assert.Equal(t, pubsub.ValidationReject, vs.Validator("foo")(context.Background(), "", &pubsub.Message{}))
}

func TestValidatorSet_Validator_RejectLast(t *testing.T) {
	vs := NewValidatorSet()

	vs.Add(func(ctx context.Context, topic string, id peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
		assert.Equal(t, "foo", topic)
		return pubsub.ValidationAccept
	})

	vs.Add(func(ctx context.Context, topic string, id peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
		assert.Equal(t, "foo", topic)
		return pubsub.ValidationReject
	})

	// Message should be rejected if at least one validator rejects it:
	assert.Equal(t, pubsub.ValidationReject, vs.Validator("foo")(context.Background(), "", &pubsub.Message{}))
}
