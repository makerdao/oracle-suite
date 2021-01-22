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
