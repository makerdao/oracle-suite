package p2p

import (
	"math"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// Peer scoring:
// https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.1.md#peer-scoring

const decayInterval = time.Minute
const decayToZero = 0.01

// maxScore is a maximum positive score that peers can gain.
const maxScore = priceMaxP1Score + priceMaxP2Score

// silentPeerScore score is the lowest score silent peer can get without
// receiving any penalties other than P₃ and P₃b. Because only very few
// peers can produce messages, most peers will have a score equal to this
// number.
const silentPeerScore = priceMaxP3Score + priceMaxP3BScore

// In our network, most peers will have a score equal to the silentPeerScore.
// A lower score indicates that the peer is violating some rules. Since there
// will not be many peers with a score between 0 and silentPeerScore, we can
// set all threshold values to silentPeerScore.
var thresholds = &pubsub.PeerScoreThresholds{
	GossipThreshold:             silentPeerScore,
	PublishThreshold:            silentPeerScore,
	GraylistThreshold:           silentPeerScore,
	AcceptPXThreshold:           0,
	OpportunisticGraftThreshold: 0,
}

var peerScoreParams = &pubsub.PeerScoreParams{
	AppSpecificScore:            func(id peer.ID) float64 { return 0 },
	AppSpecificWeight:           1,
	IPColocationFactorWeight:    -10,
	IPColocationFactorThreshold: 2,
	BehaviourPenaltyWeight:      -1,
	BehaviourPenaltyThreshold:   1,
	BehaviourPenaltyDecay:       decay(1, time.Hour),
	DecayInterval:               decayInterval,
	DecayToZero:                 decayToZero,
	RetainScore:                 15 * time.Minute,
	Topics:                      make(map[string]*pubsub.TopicScoreParams),
}

// decay calculates a decay parameter for a peer scoring. It finds a number X
// that satisfies the equation: from*X^(duration/decayInterval)=decayToZero.
// In other words, it finds a decay value for which a scoring will drop to
// the decayToZero value after a specified time.
func decay(from float64, duration time.Duration) float64 {
	return math.Pow(decayToZero/from, 1/(float64(duration)/float64(decayInterval)))
}

// Values for the "price" topic:

// Maximum scores for score parameters:
const priceMaxP1Score float64 = 500
const priceMaxP2Score float64 = 500
const priceMaxP3Score float64 = -1000
const priceMaxP3BScore float64 = -1000

// Length values used to calculate cap and decay values. They shouldn't
// be longer than the RetainScore parameter.
const priceP1Length = 15 * time.Minute
const priceP2Length = 15 * time.Minute
const priceP3Length = 15 * time.Minute
const priceP3BLength = 15 * time.Minute

func priceTopicScoreParams(cfg Config) *pubsub.TopicScoreParams {
	var maxPeers = float64(pubsub.GossipSubDhi)
	// Minimum and maximum expected number of feeders connected to the network:
	var minFeederCount = float64(len(cfg.FeedersAddrs)) / 2 // assume that 50% of feeders are offline
	var maxFeederCount = float64(len(cfg.FeedersAddrs))
	// Minimum and maximum expected number of asset pairs to be broadcast by each feeder:
	var minAssetPairCount = float64(1)
	var maxAssetPairCount = maxBytesPerSecond / maxMsgSizeInBytes * priceUpdateInterval.Seconds() / maxFeederCount
	// Minimum and maximum expected number of messages to be received from a single peer in a mesh:
	var minMsgsPerSecond = (minFeederCount * minAssetPairCount) / maxPeers / priceUpdateInterval.Seconds()
	var maxMsgsPerSecond = (maxFeederCount * maxAssetPairCount) / priceUpdateInterval.Seconds()

	return &pubsub.TopicScoreParams{
		TopicWeight: 1,

		// P₁
		TimeInMeshWeight:  priceMaxP1Score / priceP1Length.Seconds(),
		TimeInMeshQuantum: time.Second,
		TimeInMeshCap:     priceP1Length.Seconds(),

		// P₂
		FirstMessageDeliveriesWeight: priceMaxP2Score / (minMsgsPerSecond * priceP2Length.Seconds()),
		FirstMessageDeliveriesDecay:  decay(minMsgsPerSecond*priceP2Length.Seconds(), priceP2Length),
		FirstMessageDeliveriesCap:    minMsgsPerSecond * priceP2Length.Seconds(),

		// P₃
		// Note that because of the difference between the cap value and the threshold value, the decay time may be
		// longer than priceP3Length. The priceP3Length parameter describes the decay time for the minimum
		// number of messages.
		MeshMessageDeliveriesWeight:     priceMaxP3Score / math.Pow(minMsgsPerSecond*priceP3Length.Seconds(), 2),
		MeshMessageDeliveriesDecay:      decay(minMsgsPerSecond*priceP3Length.Seconds(), priceP3Length),
		MeshMessageDeliveriesCap:        maxMsgsPerSecond * priceP3Length.Seconds(),
		MeshMessageDeliveriesThreshold:  minMsgsPerSecond * priceP3Length.Seconds(),
		MeshMessageDeliveriesActivation: priceP3Length,
		MeshMessageDeliveriesWindow:     10 * time.Millisecond,

		// P₃b
		MeshFailurePenaltyWeight: priceMaxP3BScore / math.Pow(minMsgsPerSecond*priceP3BLength.Seconds(), 2),
		MeshFailurePenaltyDecay:  decay(minMsgsPerSecond*priceP3BLength.Seconds(), priceP3BLength),

		// P₄
		InvalidMessageDeliveriesWeight: maxScore / math.Pow(maxInvalidMsgsPerHour, 2) * -1,
		InvalidMessageDeliveriesDecay:  decay(math.Pow(maxInvalidMsgsPerHour, 2), time.Hour),
	}
}
