package allowlist

import (
	"bytes"
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/gofer/internal/transport/p2p/sets"
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
