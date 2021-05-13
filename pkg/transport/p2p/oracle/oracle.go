package oracle

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/transport/messages"
	"github.com/makerdao/oracle-suite/pkg/transport/p2p/crypto/ethkey"
	"github.com/makerdao/oracle-suite/pkg/transport/p2p/node"
)

func Oracle(feeders []ethereum.Address, signer ethereum.Signer, logger log.Logger) node.Options {
	return func(n *node.Node) error {
		n.AddValidator(func(ctx context.Context, topic string, id peer.ID, psMsg *pubsub.Message) pubsub.ValidationResult {
			priceMsg, ok := psMsg.ValidatorData.(*messages.Price)
			if !ok {
				return pubsub.ValidationAccept
			}
			// Check is a message signature is valid and extract author's address:
			priceFrom, err := priceMsg.Price.From(signer)
			if err != nil {
				logger.
					WithError(err).
					WithField("peerID", psMsg.String()).
					Info("Rejected price message with invalid signature")
				return pubsub.ValidationReject
			}
			// The libp2p message should be created by the same person who signs the price message:
			if ethkey.AddressToPeerID(*priceFrom) != psMsg.GetFrom() {
				logger.
					WithField("peerID", psMsg.String()).
					Info("Rejected price message, message author and price signature doesn't match")
				return pubsub.ValidationReject
			}
			// Check when message was created, reject if older than 5 min:
			if time.Since(priceMsg.Price.Age) > 5*time.Minute {
				logger.
					WithField("peerID", psMsg.String()).
					Info("Rejected price message, message is older than 5 min")

				if time.Since(priceMsg.Price.Age) > 10*time.Minute {
					return pubsub.ValidationReject
				}
				return pubsub.ValidationIgnore
			}
			// Check is the author is allowed to send price messages:
			for _, addr := range feeders {
				if addr == *priceFrom {
					return pubsub.ValidationAccept
				}
			}
			return pubsub.ValidationReject
		})
		return nil
	}
}
