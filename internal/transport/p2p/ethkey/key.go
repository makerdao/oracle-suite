package ethkey

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p-core/crypto"
	crypto_pb "github.com/libp2p/go-libp2p-core/crypto/pb"
	"github.com/libp2p/go-libp2p-core/peer"
)

// Eth key type uses the ethereum wallet to sign and verify messages.
const KeyType_Eth crypto_pb.KeyType = 10

func init() {
	crypto.PubKeyUnmarshallers[KeyType_Eth] = UnmarshalEthPublicKey
}

// AddressToPeerID converts an Ethereum address to a peer ID. If address is
// invalid then empty ID will be returned.
func AddressToPeerID(a string) peer.ID {
	null := common.Address{}
	addr := common.HexToAddress(a)
	if addr == null {
		return ""
	}
	id, err := peer.IDFromPublicKey(NewPubKey(addr))
	if err != nil {
		return ""
	}
	return id
}
