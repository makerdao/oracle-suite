package node

import (
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/oracle-suite/pkg/log"
)

type Options func(n *Node) error

func ListenAddrs(addrs []multiaddr.Multiaddr) Options {
	return func(n *Node) error {
		n.listenAddrs = addrs
		return nil
	}
}

func Bootstrap(addrs []multiaddr.Multiaddr) Options {
	return func(n *Node) error {
		for _, maddr := range addrs {
			err := n.Connect(maddr)
			if err != nil {
				n.log.
					WithFields(log.Fields{"addr": maddr.String()}).
					WithError(err).
					Warn("Unable to connect to bootstrap peer")
			}
		}
		return nil
	}
}

func PeerPrivKey(pk crypto.PrivKey) Options {
	return func(n *Node) error {
		n.peerPrivKey = pk
		return nil
	}
}

func MessagePrivKey(pk crypto.PrivKey) Options {
	return func(n *Node) error {
		pid, err := peer.IDFromPublicKey(pk.GetPublic())
		if err != nil {
			return err
		}
		n.messagePrivKey = pk
		n.messageAuthorPID = pid
		return nil
	}
}

func Logger(logger log.Logger) Options {
	return func(n *Node) error {
		n.log = logger
		return nil
	}
}
