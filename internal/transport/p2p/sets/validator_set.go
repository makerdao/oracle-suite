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
