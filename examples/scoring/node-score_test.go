package scoring

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"

	logging "github.com/ipfs/go-log"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-msgio/protoio"
)

const baseHostAddr = "/ip4/0.0.0.0/tcp/0"

func startNode(t *testing.T, ctx context.Context) host.Host {
	//p := tnet.RandPeerNetParamsOrFatal(t)
	// create a new libp2p Host that listens on a random TCP port
	h, err := libp2p.New(ctx, libp2p.ListenAddrStrings(baseHostAddr))
	if err != nil {
		panic(err)
	}
	return h
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
	legit := startNode(t, ctx)
	attacker := startNode(t, ctx)

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

	connect(t, legit, attacker)

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
