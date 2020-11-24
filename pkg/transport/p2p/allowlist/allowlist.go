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

package allowlist

import (
	"bytes"
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/gofer/pkg/transport/p2p/sets"
)

type node interface {
	AddValidator(validator sets.Validator)
}

// Register registers p2p.Node extensions required by the Allowlist and returns
// its instance.
func Register(node node) *Allowlist {
	allowlist := &Allowlist{}
	node.AddValidator(allowlist.validator)
	return allowlist
}

// Allowlist allows to define a list of peers allowed to publish messages.
// Until the first peer is added to this list, everyone will be allowed to
// publish messages.
type Allowlist struct {
	peers []peer.ID
}

// Allow adds a peer ID to the list of allowed peers.
func (a *Allowlist) Allow(id peer.ID) {
	a.peers = append(a.peers, id)
}

func (a *Allowlist) validator(topic string, ctx context.Context, id peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
	if len(a.peers) == 0 {
		return pubsub.ValidationAccept
	}
	for _, allowed := range a.peers {
		if bytes.Equal([]byte(allowed), []byte(id)) {
			return pubsub.ValidationAccept
		}
	}
	return pubsub.ValidationIgnore
}
