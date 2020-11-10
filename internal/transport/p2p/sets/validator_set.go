package sets

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// ValidatorSet stores multiple instances of validators that implements
// the pubsub.ValidatorEx functions. Validators are groped by topic.
type ValidatorSet struct {
	validators map[string][]pubsub.ValidatorEx
}

// NewValidatorSet creates new instance of the ValidatorSet.
func NewValidatorSet() *ValidatorSet {
	return &ValidatorSet{
		validators: make(map[string][]pubsub.ValidatorEx, 0),
	}
}

// Add adds new pubsub.ValidatorEx to the set.
func (n *ValidatorSet) Add(topic string, validator ...pubsub.ValidatorEx) {
	if _, ok := n.validators[topic]; !ok {
		n.validators[topic] = []pubsub.ValidatorEx{}
	}
	n.validators[topic] = append(n.validators[topic], validator...)
}

// Validator returns function that implements pubsub.ValidatorEx. That function
// will invoke all registered validators for given topic.
func (n *ValidatorSet) Validator(topic string) pubsub.ValidatorEx {
	if _, ok := n.validators[topic]; !ok {
		n.validators[topic] = []pubsub.ValidatorEx{}
	}
	return func(ctx context.Context, id peer.ID, message *pubsub.Message) pubsub.ValidationResult {
		for _, validator := range n.validators[topic] {
			if result := validator(ctx, id, message); result != pubsub.ValidationAccept {
				return result
			}
		}
		return pubsub.ValidationAccept
	}
}
