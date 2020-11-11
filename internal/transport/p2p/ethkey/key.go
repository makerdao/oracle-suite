package ethkey

import (
	"github.com/libp2p/go-libp2p-core/crypto"
	crypto_pb "github.com/libp2p/go-libp2p-core/crypto/pb"
)

// Eth key type uses the ethereum wallet to sign and verify messages.
const KeyType_Eth crypto_pb.KeyType = 10

func init() {
	crypto.PubKeyUnmarshallers[KeyType_Eth] = UnmarshalEthPublicKey
	crypto.PrivKeyUnmarshallers[KeyType_Eth] = UnmarshalEthPrivateKey
}
