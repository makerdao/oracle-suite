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

package p2p

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/makerdao/oracle-suite/pkg/transport/messages"
)

//nolint:gomnd
// NodeScoring configures node scoring parameters.
func NodeScoring() Options {
	return func(n *Node) error {
		// Create parameters with reasonable default values
		params := &pubsub.PeerScoreParams{
			AppSpecificScore:            func(peer.ID) float64 { return 0 },
			AppSpecificWeight:           1,
			IPColocationFactorWeight:    -1,
			IPColocationFactorThreshold: 4,
			DecayInterval:               1 * time.Minute,
			DecayToZero:                 0.01,
			RetainScore:                 10 * time.Second,
			Topics:                      make(map[string]*pubsub.TopicScoreParams),
		}
		params.Topics[messages.PriceMessageName] = &pubsub.TopicScoreParams{
			TopicWeight:                     1,
			TimeInMeshWeight:                0.0027,
			TimeInMeshQuantum:               time.Second,
			TimeInMeshCap:                   3600,
			MeshMessageDeliveriesWeight:     -0.25,
			MeshMessageDeliveriesDecay:      0.97,
			MeshMessageDeliveriesCap:        400,
			MeshMessageDeliveriesThreshold:  200,
			MeshMessageDeliveriesActivation: 30 * time.Second,
			MeshMessageDeliveriesWindow:     5 * time.Minute,
			MeshFailurePenaltyWeight:        -0.25,
			MeshFailurePenaltyDecay:         0.997,
			InvalidMessageDeliveriesWeight:  -99,
			InvalidMessageDeliveriesDecay:   0.9994,
		}
		thresholds := &pubsub.PeerScoreThresholds{
			GossipThreshold:   -100,
			PublishThreshold:  -200,
			GraylistThreshold: -400,
			AcceptPXThreshold: 0,
		}

		n.pubsubOpts = append(n.pubsubOpts, pubsub.WithPeerScore(params, thresholds))
		return nil
	}
}
