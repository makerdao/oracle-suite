package main

import (
	"context"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"

	bhost "github.com/libp2p/go-libp2p-blankhost"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	swarmt "github.com/libp2p/go-libp2p-swarm/testing"

	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-msgio/protoio"
)

func getNetHosts(t *testing.T, ctx context.Context, n int) []host.Host {
	var out []host.Host

	for i := 0; i < n; i++ {
		netw := swarmt.GenSwarm(t, ctx)
		h := bhost.NewBlankHost(netw)
		out = append(out, h)
	}

	return out
}

func connect(t *testing.T, a, b host.Host) {
	pinfo := a.Peerstore().PeerInfo(a.ID())
	err := b.Connect(context.Background(), pinfo)
	if err != nil {
		t.Fatal(err)
	}
}

func rpcWithControl(msgs []*pb.Message,
	ihave []*pb.ControlIHave,
	iwant []*pb.ControlIWant,
	graft []*pb.ControlGraft,
	prune []*pb.ControlPrune) *pubsub.RPC {
	return &pubsub.RPC{
		RPC: pb.RPC{
			Publish: msgs,
			Control: &pb.ControlMessage{
				Ihave: ihave,
				Iwant: iwant,
				Graft: graft,
				Prune: prune,
			},
		},
	}
}

// Test that when Gossipsub receives too many IWANT messages from a peer
// for the same message ID, it cuts off the peer
func TestGossipsubAttackSpamIWANT(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create legitimate and attacker hosts
	hosts := getNetHosts(t, ctx, 2)
	legit := hosts[0]
	attacker := hosts[1]

	// Set up gossipsub on the legit host
	ps, err := pubsub.NewGossipSub(ctx, legit)
	if err != nil {
		t.Fatal(err)
	}

	// Subscribe to mytopic on the legit host
	mytopic := "mytopic"
	_, err = ps.Subscribe(mytopic)
	if err != nil {
		t.Fatal(err)
	}

	// Used to publish a message with random data
	publishMsg := func() {
		data := make([]byte, 16)
		rand.Read(data)

		if err = ps.Publish(mytopic, data); err != nil {
			t.Fatal(err)
		}
	}

	// Wait a bit after the last message before checking we got the
	// right number of messages
	msgWaitMax := time.Second
	msgCount := 0
	msgTimer := time.NewTimer(msgWaitMax)

	// Checks we received the right number of messages
	checkMsgCount := func() {
		// After the original message from the legit host, we keep sending
		// IWANT until it stops replying. So the number of messages is
		// <original message> + GossipSubGossipRetransmission
		exp := 1 + pubsub.GossipSubGossipRetransmission
		if msgCount != exp {
			t.Fatalf("Expected %d messages, got %d", exp, msgCount)
		}
	}

	// Wait for the timer to expire
	go func() {
		select {
		case <-msgTimer.C:
			checkMsgCount()
			cancel()
			return
		case <-ctx.Done():
			checkMsgCount()
		}
	}()

	newMockGS(ctx, t, attacker, func(writeMsg func(*pb.RPC), irpc *pb.RPC) {
		// When the legit host connects it will send us its subscriptions
		for _, sub := range irpc.GetSubscriptions() {
			if sub.GetSubscribe() {
				// Reply by subcribing to the topic and grafting to the peer
				writeMsg(&pb.RPC{
					Subscriptions: []*pb.RPC_SubOpts{&pb.RPC_SubOpts{Subscribe: sub.Subscribe, Topicid: sub.Topicid}},
					Control:       &pb.ControlMessage{Graft: []*pb.ControlGraft{&pb.ControlGraft{TopicID: sub.Topicid}}},
				})

				go func() {
					// Wait for a short interval to make sure the legit host
					// received and processed the subscribe + graft
					time.Sleep(100 * time.Millisecond)

					// Publish a message from the legit host
					publishMsg()
				}()
			}
		}

		// Each time the legit host sends a message
		for _, msg := range irpc.GetPublish() {
			// Increment the number of messages and reset the timer
			msgCount++
			msgTimer.Reset(msgWaitMax)

			// Shouldn't get more than the expected number of messages
			exp := 1 + pubsub.GossipSubGossipRetransmission
			if msgCount > exp {
				cancel()
				t.Fatal("Received too many responses")
			}

			// Send an IWANT with the message ID, causing the legit host
			// to send another message (until it cuts off the attacker for
			// being spammy)
			iwantlst := []string{pubsub.DefaultMsgIdFn(msg)}
			iwant := []*pb.ControlIWant{&pb.ControlIWant{MessageIDs: iwantlst}}
			orpc := rpcWithControl(nil, nil, iwant, nil, nil)
			writeMsg(&orpc.RPC)
		}
	})

	connect(t, hosts[0], hosts[1])

	<-ctx.Done()
}

// Test that Gossipsub only responds to IHAVE with IWANT once per heartbeat
func TestGossipsubAttackSpamIHAVE(t *testing.T) {
	originalGossipSubIWantFollowupTime := pubsub.GossipSubIWantFollowupTime
	pubsub.GossipSubIWantFollowupTime = 10 * time.Second
	defer func() {
		pubsub.GossipSubIWantFollowupTime = originalGossipSubIWantFollowupTime
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create legitimate and attacker hosts
	hosts := getNetHosts(t, ctx, 2)
	legit := hosts[0]
	attacker := hosts[1]

	// Set up gossipsub on the legit host
	ps, err := pubsub.NewGossipSub(ctx, legit,
		pubsub.WithPeerScore(
			&pubsub.PeerScoreParams{
				AppSpecificScore:       func(peer.ID) float64 { return 0 },
				BehaviourPenaltyWeight: -1,
				BehaviourPenaltyDecay:  pubsub.ScoreParameterDecay(time.Minute),
				DecayInterval:          pubsub.DefaultDecayInterval,
				DecayToZero:            pubsub.DefaultDecayToZero,
			},
			&pubsub.PeerScoreThresholds{
				GossipThreshold:   -100,
				PublishThreshold:  -500,
				GraylistThreshold: -1000,
			}))
	if err != nil {
		t.Fatal(err)
	}

	// Subscribe to mytopic on the legit host
	mytopic := "mytopic"
	_, err = ps.Subscribe(mytopic)
	if err != nil {
		t.Fatal(err)
	}

	iWantCount := 0
	iWantCountMx := sync.Mutex{}
	getIWantCount := func() int {
		iWantCountMx.Lock()
		defer iWantCountMx.Unlock()
		return iWantCount
	}
	addIWantCount := func(i int) {
		iWantCountMx.Lock()
		defer iWantCountMx.Unlock()
		iWantCount += i
	}

	newMockGS(ctx, t, attacker, func(writeMsg func(*pb.RPC), irpc *pb.RPC) {
		// When the legit host connects it will send us its subscriptions
		for _, sub := range irpc.GetSubscriptions() {
			if sub.GetSubscribe() {
				// Reply by subcribing to the topic and grafting to the peer
				writeMsg(&pb.RPC{
					Subscriptions: []*pb.RPC_SubOpts{&pb.RPC_SubOpts{Subscribe: sub.Subscribe, Topicid: sub.Topicid}},
					Control:       &pb.ControlMessage{Graft: []*pb.ControlGraft{&pb.ControlGraft{TopicID: sub.Topicid}}},
				})

				go func() {
					defer cancel()

					// Wait for a short interval to make sure the legit host
					// received and processed the subscribe + graft
					time.Sleep(20 * time.Millisecond)

					// Send a bunch of IHAVEs
					for i := 0; i < 3*pubsub.GossipSubMaxIHaveLength; i++ {
						ihavelst := []string{"someid" + strconv.Itoa(i)}
						ihave := []*pb.ControlIHave{&pb.ControlIHave{TopicID: sub.Topicid, MessageIDs: ihavelst}}
						orpc := rpcWithControl(nil, ihave, nil, nil, nil)
						writeMsg(&orpc.RPC)
					}

					time.Sleep(pubsub.GossipSubHeartbeatInterval)

					// Should have hit the maximum number of IWANTs per peer
					// per heartbeat
					iwc := getIWantCount()
					if iwc > pubsub.GossipSubMaxIHaveLength {
						t.Fatalf("Expecting max %d IWANTs per heartbeat but received %d", pubsub.GossipSubMaxIHaveLength, iwc)
					}
					firstBatchCount := iwc

					// the score should still be 0 because we haven't broken any promises yet
					score := ps.rt.(*pubsub.GossipSubRouter).score.Score(attacker.ID())
					if score != 0 {
						t.Fatalf("Expected 0 score, but got %f", score)
					}

					// Send a bunch of IHAVEs
					for i := 0; i < 3*pubsub.GossipSubMaxIHaveLength; i++ {
						ihavelst := []string{"someid" + strconv.Itoa(i+100)}
						ihave := []*pb.ControlIHave{&pb.ControlIHave{TopicID: sub.Topicid, MessageIDs: ihavelst}}
						orpc := rpcWithControl(nil, ihave, nil, nil, nil)
						writeMsg(&orpc.RPC)
					}

					time.Sleep(pubsub.GossipSubHeartbeatInterval)

					// Should have sent more IWANTs after the heartbeat
					iwc = getIWantCount()
					if iwc == firstBatchCount {
						t.Fatal("Expecting to receive more IWANTs after heartbeat but did not")
					}
					// Should not be more than the maximum per heartbeat
					if iwc-firstBatchCount > pubsub.GossipSubMaxIHaveLength {
						t.Fatalf("Expecting max %d IWANTs per heartbeat but received %d", pubsub.GossipSubMaxIHaveLength, iwc-firstBatchCount)
					}

					time.Sleep(pubsub.GossipSubIWantFollowupTime)

					// The score should now be negative because of broken promises
					score = ps.rt.(*pubsub.GossipSubRouter).score.Score(attacker.ID())
					if score >= 0 {
						t.Fatalf("Expected negative score, but got %f", score)
					}
				}()
			}
		}

		// Record the count of received IWANT messages
		if ctl := irpc.GetControl(); ctl != nil {
			addIWantCount(len(ctl.GetIwant()))
		}
	})

	connect(t, hosts[0], hosts[1])

	<-ctx.Done()
}

// Test that when Gossipsub receives GRAFT for an unknown topic, it ignores
// the request
func TestGossipsubAttackGRAFTNonExistentTopic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create legitimate and attacker hosts
	hosts := getNetHosts(t, ctx, 2)
	legit := hosts[0]
	attacker := hosts[1]

	// Set up gossipsub on the legit host
	ps, err := pubsub.NewGossipSub(ctx, legit)
	if err != nil {
		t.Fatal(err)
	}

	// Subscribe to mytopic on the legit host
	mytopic := "mytopic"
	_, err = ps.Subscribe(mytopic)
	if err != nil {
		t.Fatal(err)
	}

	// Checks that we haven't received any PRUNE message
	pruneCount := 0
	checkForPrune := func() {
		// We send a GRAFT for a non-existent topic so we shouldn't
		// receive a PRUNE in response
		if pruneCount != 0 {
			t.Fatalf("Got %d unexpected PRUNE messages", pruneCount)
		}
	}

	newMockGS(ctx, t, attacker, func(writeMsg func(*pb.RPC), irpc *pb.RPC) {
		// When the legit host connects it will send us its subscriptions
		for _, sub := range irpc.GetSubscriptions() {
			if sub.GetSubscribe() {
				// Reply by subcribing to the topic and grafting to the peer
				writeMsg(&pb.RPC{
					Subscriptions: []*pb.RPC_SubOpts{&pb.RPC_SubOpts{Subscribe: sub.Subscribe, Topicid: sub.Topicid}},
					Control:       &pb.ControlMessage{Graft: []*pb.ControlGraft{&pb.ControlGraft{TopicID: sub.Topicid}}},
				})

				// Graft to the peer on a non-existent topic
				nonExistentTopic := "non-existent"
				writeMsg(&pb.RPC{
					Control: &pb.ControlMessage{Graft: []*pb.ControlGraft{&pb.ControlGraft{TopicID: &nonExistentTopic}}},
				})

				go func() {
					// Wait for a short interval to make sure the legit host
					// received and processed the subscribe + graft
					time.Sleep(100 * time.Millisecond)

					// We shouldn't get any prune messages becaue the topic
					// doesn't exist
					checkForPrune()
					cancel()
				}()
			}
		}

		// Record the count of received PRUNE messages
		if ctl := irpc.GetControl(); ctl != nil {
			pruneCount += len(ctl.GetPrune())
		}
	})

	connect(t, hosts[0], hosts[1])

	<-ctx.Done()
}

// Test that when Gossipsub receives GRAFT for a peer that has been PRUNED,
// it penalizes through P7 and eventually graylists and ignores the requests if the
// GRAFTs are coming too fast
func TestGossipsubAttackGRAFTDuringBackoff(t *testing.T) {
	originalGossipSubPruneBackoff := pubsub.GossipSubPruneBackoff
	pubsub.GossipSubPruneBackoff = 200 * time.Millisecond
	originalGossipSubGraftFloodThreshold := pubsub.GossipSubGraftFloodThreshold
	pubsub.GossipSubGraftFloodThreshold = 100 * time.Millisecond
	defer func() {
		pubsub.GossipSubPruneBackoff = originalGossipSubPruneBackoff
		pubsub.GossipSubGraftFloodThreshold = originalGossipSubGraftFloodThreshold
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create legitimate and attacker hosts
	hosts := getNetHosts(t, ctx, 2)
	legit := hosts[0]
	attacker := hosts[1]

	// Set up gossipsub on the legit host
	ps, err := pubsub.NewGossipSub(ctx, legit,
		pubsub.WithPeerScore(
			&pubsub.PeerScoreParams{
				AppSpecificScore:       func(peer.ID) float64 { return 0 },
				BehaviourPenaltyWeight: -100,
				BehaviourPenaltyDecay:  pubsub.ScoreParameterDecay(time.Minute),
				DecayInterval:          pubsub.DefaultDecayInterval,
				DecayToZero:            pubsub.DefaultDecayToZero,
			},
			&pubsub.PeerScoreThresholds{
				GossipThreshold:   -100,
				PublishThreshold:  -500,
				GraylistThreshold: -1000,
			}))
	if err != nil {
		t.Fatal(err)
	}

	// Subscribe to mytopic on the legit host
	mytopic := "mytopic"
	_, err = ps.Subscribe(mytopic)
	if err != nil {
		t.Fatal(err)
	}

	pruneCount := 0
	pruneCountMx := sync.Mutex{}
	getPruneCount := func() int {
		pruneCountMx.Lock()
		defer pruneCountMx.Unlock()
		return pruneCount
	}
	addPruneCount := func(i int) {
		pruneCountMx.Lock()
		defer pruneCountMx.Unlock()
		pruneCount += i
	}

	newMockGS(ctx, t, attacker, func(writeMsg func(*pb.RPC), irpc *pb.RPC) {
		// When the legit host connects it will send us its subscriptions
		for _, sub := range irpc.GetSubscriptions() {
			if sub.GetSubscribe() {
				// Reply by subcribing to the topic and grafting to the peer
				graft := []*pb.ControlGraft{&pb.ControlGraft{TopicID: sub.Topicid}}
				writeMsg(&pb.RPC{
					Subscriptions: []*pb.RPC_SubOpts{&pb.RPC_SubOpts{Subscribe: sub.Subscribe, Topicid: sub.Topicid}},
					Control:       &pb.ControlMessage{Graft: graft},
				})

				go func() {
					defer cancel()

					// Wait for a short interval to make sure the legit host
					// received and processed the subscribe + graft
					time.Sleep(20 * time.Millisecond)

					// No PRUNE should have been sent at this stage
					pc := getPruneCount()
					if pc != 0 {
						t.Fatalf("Expected %d PRUNE messages but got %d", 0, pc)
					}

					// Send a PRUNE to remove the attacker node from the legit
					// host's mesh
					var prune []*pb.ControlPrune
					prune = append(prune, &pb.ControlPrune{TopicID: sub.Topicid})
					writeMsg(&pb.RPC{
						Control: &pb.ControlMessage{Prune: prune},
					})

					time.Sleep(20 * time.Millisecond)

					// No PRUNE should have been sent at this stage
					pc = getPruneCount()
					if pc != 0 {
						t.Fatalf("Expected %d PRUNE messages but got %d", 0, pc)
					}

					// wait for the GossipSubGraftFloodThreshold to pass before attempting another graft
					time.Sleep(pubsub.GossipSubGraftFloodThreshold + time.Millisecond)

					// Send a GRAFT to attempt to rejoin the mesh
					writeMsg(&pb.RPC{
						Control: &pb.ControlMessage{Graft: graft},
					})

					time.Sleep(20 * time.Millisecond)

					// We should have been peanalized by the peer for sending before the backoff has expired
					// but should still receive a PRUNE because we haven't dropped below GraylistThreshold
					// yet.
					pc = getPruneCount()
					if pc != 1 {
						t.Fatalf("Expected %d PRUNE messages but got %d", 1, pc)
					}

					score1 := ps.rt.(*pubsub.GossipSubRouter).score.Score(attacker.ID())
					if score1 >= 0 {
						t.Fatalf("Expected negative score, but got %f", score1)
					}

					// Send a GRAFT again to attempt to rejoin the mesh
					writeMsg(&pb.RPC{
						Control: &pb.ControlMessage{Graft: graft},
					})

					time.Sleep(20 * time.Millisecond)

					// we are before the flood threshold so we should be penalized twice, but still get
					// a PRUNE because we are before the flood threshold
					pc = getPruneCount()
					if pc != 2 {
						t.Fatalf("Expected %d PRUNE messages but got %d", 2, pc)
					}

					score2 := ps.rt.(*pubsub.GossipSubRouter).score.Score(attacker.ID())
					if score2 >= score1 {
						t.Fatalf("Expected score below %f, but got %f", score1, score2)
					}

					// Send another GRAFT; this should get us a PRUNE, but penalize us below the graylist threshold
					writeMsg(&pb.RPC{
						Control: &pb.ControlMessage{Graft: graft},
					})

					time.Sleep(20 * time.Millisecond)

					pc = getPruneCount()
					if pc != 3 {
						t.Fatalf("Expected %d PRUNE messages but got %d", 3, pc)
					}

					score3 := ps.rt.(*pubsub.GossipSubRouter).score.Score(attacker.ID())
					if score3 >= score2 {
						t.Fatalf("Expected score below %f, but got %f", score2, score3)
					}
					if score3 >= -1000 {
						t.Fatalf("Expected score below %f, but got %f", -1000.0, score3)
					}

					// Wait for the PRUNE backoff to expire and try again; this time we should fail
					// because we are below the graylist threshold, so our RPC should be ignored and
					// we should get no PRUNE back
					time.Sleep(pubsub.GossipSubPruneBackoff + time.Millisecond)

					writeMsg(&pb.RPC{
						Control: &pb.ControlMessage{Graft: graft},
					})

					time.Sleep(20 * time.Millisecond)

					pc = getPruneCount()
					if pc != 3 {
						t.Fatalf("Expected %d PRUNE messages but got %d", 3, pc)
					}

					// make sure we are _not_ in the mesh
					res := make(chan bool)
					ps.eval <- func() {
						mesh := ps.rt.(*pubsub.GossipSubRouter).mesh[mytopic]
						_, inMesh := mesh[attacker.ID()]
						res <- inMesh
					}

					inMesh := <-res
					if inMesh {
						t.Fatal("Expected to not be in the mesh of the legitimate host")
					}
				}()
			}
		}

		if ctl := irpc.GetControl(); ctl != nil {
			addPruneCount(len(ctl.GetPrune()))
		}
	})

	connect(t, hosts[0], hosts[1])

	<-ctx.Done()
}

type gsAttackInvalidMsgTracer struct {
	rejectCount int
}

func (t *gsAttackInvalidMsgTracer) Trace(evt *pb.TraceEvent) {
	// fmt.Printf("    %s %s\n", evt.Type, evt)
	if evt.GetType() == pb.TraceEvent_REJECT_MESSAGE {
		t.rejectCount++
	}
}

// Test that when Gossipsub receives a lot of invalid messages from
// a peer it should graylist the peer
func TestGossipsubAttackInvalidMessageSpam(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create legitimate and attacker hosts
	hosts := getNetHosts(t, ctx, 2)
	legit := hosts[0]
	attacker := hosts[1]

	mytopic := "mytopic"

	// Create parameters with reasonable default values
	params := &pubsub.PeerScoreParams{
		AppSpecificScore:            func(peer.ID) float64 { return 0 },
		IPColocationFactorWeight:    0,
		IPColocationFactorThreshold: 1,
		DecayInterval:               5 * time.Second,
		DecayToZero:                 0.01,
		RetainScore:                 10 * time.Second,
		Topics:                      make(map[string]*pubsub.TopicScoreParams),
	}
	params.Topics[mytopic] = &pubsub.TopicScoreParams{
		TopicWeight:                     0.25,
		TimeInMeshWeight:                0.0027,
		TimeInMeshQuantum:               time.Second,
		TimeInMeshCap:                   3600,
		FirstMessageDeliveriesWeight:    0.664,
		FirstMessageDeliveriesDecay:     0.9916,
		FirstMessageDeliveriesCap:       1500,
		MeshMessageDeliveriesWeight:     -0.25,
		MeshMessageDeliveriesDecay:      0.97,
		MeshMessageDeliveriesCap:        400,
		MeshMessageDeliveriesThreshold:  100,
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
		GraylistThreshold: -300,
		AcceptPXThreshold: 0,
	}

	// Set up gossipsub on the legit host
	tracer := &gsAttackInvalidMsgTracer{}
	ps, err := pubsub.NewGossipSub(ctx, legit,
		pubsub.WithEventTracer(tracer),
		pubsub.WithPeerScore(params, thresholds),
	)
	if err != nil {
		t.Fatal(err)
	}

	attackerScore := func() float64 {
		return ps.rt.(*pubsub.GossipSubRouter).score.Score(attacker.ID())
	}

	// Subscribe to mytopic on the legit host
	_, err = ps.Subscribe(mytopic)
	if err != nil {
		t.Fatal(err)
	}

	pruneCount := 0
	pruneCountMx := sync.Mutex{}
	getPruneCount := func() int {
		pruneCountMx.Lock()
		defer pruneCountMx.Unlock()
		return pruneCount
	}
	addPruneCount := func(i int) {
		pruneCountMx.Lock()
		defer pruneCountMx.Unlock()
		pruneCount += i
	}

	newMockGS(ctx, t, attacker, func(writeMsg func(*pb.RPC), irpc *pb.RPC) {
		// When the legit host connects it will send us its subscriptions
		for _, sub := range irpc.GetSubscriptions() {
			if sub.GetSubscribe() {
				// Reply by subcribing to the topic and grafting to the peer
				writeMsg(&pb.RPC{
					Subscriptions: []*pb.RPC_SubOpts{&pb.RPC_SubOpts{Subscribe: sub.Subscribe, Topicid: sub.Topicid}},
					Control:       &pb.ControlMessage{Graft: []*pb.ControlGraft{&pb.ControlGraft{TopicID: sub.Topicid}}},
				})

				go func() {
					defer cancel()

					// Attacker score should start at zero
					if attackerScore() != 0 {
						t.Fatalf("Expected attacker score to be zero but it's %f", attackerScore())
					}

					// Send a bunch of messages with no signature (these will
					// fail validation and reduce the attacker's score)
					for i := 0; i < 100; i++ {
						msg := &pb.Message{
							Data:  []byte("some data" + strconv.Itoa(i)),
							Topic: &mytopic,
							From:  []byte(attacker.ID()),
							Seqno: []byte{byte(i + 1)},
						}
						writeMsg(&pb.RPC{
							Publish: []*pb.Message{msg},
						})
					}

					// Wait for the initial heartbeat, plus a bit of padding
					time.Sleep(100*time.Millisecond + pubsub.GossipSubHeartbeatInitialDelay)

					// The attackers score should now have fallen below zero
					if attackerScore() >= 0 {
						t.Fatalf("Expected attacker score to be less than zero but it's %f", attackerScore())
					}
					// There should be several rejected messages (because the signature was invalid)
					if tracer.rejectCount == 0 {
						t.Fatal("Expected message rejection but got none")
					}
					// The legit node should have sent a PRUNE message
					pc := getPruneCount()
					if pc == 0 {
						t.Fatal("Expected attacker node to be PRUNED when score drops low enough")
					}
				}()
			}
		}

		if ctl := irpc.GetControl(); ctl != nil {
			addPruneCount(len(ctl.GetPrune()))
		}
	})

	connect(t, hosts[0], hosts[1])

	<-ctx.Done()
}

func turnOnPubsubDebug() {
	logging.SetLogLevel("pubsub", "debug")
}

type mockGSOnRead func(writeMsg func(*pb.RPC), irpc *pb.RPC)

func newMockGS(ctx context.Context, t *testing.T, attacker host.Host, onReadMsg mockGSOnRead) {
	// Listen on the gossipsub protocol
	const gossipSubID = protocol.ID("/meshsub/1.0.0")
	const maxMessageSize = 1024 * 1024
	attacker.SetStreamHandler(gossipSubID, func(stream network.Stream) {
		// When an incoming stream is opened, set up an outgoing stream
		p := stream.Conn().RemotePeer()
		ostream, err := attacker.NewStream(ctx, p, gossipSubID)
		if err != nil {
			t.Fatal(err)
		}

		r := protoio.NewDelimitedReader(stream, maxMessageSize)
		w := protoio.NewDelimitedWriter(ostream)

		var irpc pb.RPC

		writeMsg := func(rpc *pb.RPC) {
			if err = w.WriteMsg(rpc); err != nil {
				t.Fatalf("error writing RPC: %s", err)
			}
		}

		// Keep reading messages and responding
		for {
			// Bail out when the test finishes
			if ctx.Err() != nil {
				return
			}

			irpc.Reset()

			err := r.ReadMsg(&irpc)

			// Bail out when the test finishes
			if ctx.Err() != nil {
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			onReadMsg(writeMsg, &irpc)
		}
	})
}
