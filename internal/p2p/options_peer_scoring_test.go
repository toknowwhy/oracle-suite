package p2p

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type scoringEventTracer struct {
	prunes int
	mu     sync.Mutex
}

func (s *scoringEventTracer) Trace(evt *pubsub_pb.TraceEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if evt.GetType() == pubsub_pb.TraceEvent_PRUNE {
		s.prunes++
	}
}

func (s *scoringEventTracer) Prunes() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.prunes
}

func TestNode_PeerScoring(t *testing.T) {
	// This test verifies that the integration with peer scoring is working
	// correctly.

	var peerScoreParams = &pubsub.PeerScoreParams{
		AppSpecificScore:            func(id peer.ID) float64 { return 0 },
		AppSpecificWeight:           1,
		IPColocationFactorWeight:    -1,
		IPColocationFactorThreshold: 1,
		DecayInterval:               1 * time.Minute,
		DecayToZero:                 0.01,
		RetainScore:                 10 * time.Second,
		Topics:                      make(map[string]*pubsub.TopicScoreParams),
	}

	var topicScoreParams = &pubsub.TopicScoreParams{
		TimeInMeshCap:                  0,
		TimeInMeshQuantum:              1 * time.Second,
		TopicWeight:                    1,
		FirstMessageDeliveriesWeight:   100,
		FirstMessageDeliveriesDecay:    0.999,
		FirstMessageDeliveriesCap:      300,
		InvalidMessageDeliveriesWeight: -100,
		InvalidMessageDeliveriesDecay:  0.999,
	}

	var thresholds = &pubsub.PeerScoreThresholds{
		GossipThreshold:   -100,
		PublishThreshold:  -200,
		GraylistThreshold: -300,
		AcceptPXThreshold: 0,
	}

	peers, err := getNodeInfo(2)
	require.NoError(t, err)

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	et := &scoringEventTracer{}
	n0, err := NewNode(
		ctx,
		PeerPrivKey(peers[0].PrivKey),
		ListenAddrs(peers[0].ListenAddrs),
		PubsubEventTracer(et),
		PeerScoring(peerScoreParams, thresholds, func(topic string) *pubsub.TopicScoreParams {
			return topicScoreParams
		}),
	)
	require.NoError(t, err)
	require.NoError(t, n0.Start())

	n1, err := NewNode(
		ctx,
		PeerPrivKey(peers[1].PrivKey),
		ListenAddrs(peers[1].ListenAddrs),
	)
	require.NoError(t, err)
	require.NoError(t, n1.Start())

	// Add validator to the n0 node which will reject all received messages
	// from the second node:
	n0.AddValidator(func(ctx context.Context, topic string, id peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
		if n0.Host().ID() == id {
			return pubsub.ValidationAccept
		}
		if bytes.Equal(msg.Data, []byte("valid")) {
			return pubsub.ValidationAccept
		}
		return pubsub.ValidationReject
	})

	require.NoError(t, n1.Connect(peers[0].PeerAddrs[0]))
	require.NoError(t, n0.Subscribe("test", (*message)(nil)))
	require.NoError(t, n1.Subscribe("test", (*message)(nil)))

	s0, err := n0.Subscription("test")
	require.NoError(t, err)
	s1, err := n1.Subscription("test")
	require.NoError(t, err)

	// Wait for the peers to connect to each other:
	waitFor(t, func() bool {
		return len(n0.PubSub().ListPeers("test")) > 0 && len(n1.PubSub().ListPeers("test")) > 0
	})

	s0wait := countMessages(s0, 2*time.Second)

	// Send a few valid messages to boost a peer score:
	for i := 0; i < 5; i++ {
		err := s1.Publish(newMessage("valid"))
		if err != nil {
			panic(err)
		}
	}

	// Now send a 4 invalid messages, that should be enough to lower the score below 0:
	for i := 0; i < 4; i++ {
		err := s1.Publish(newMessage("invalid"))
		if err != nil {
			panic(err)
		}
	}
	<-s0wait

	// Wait 1 second for the "heartbeat".
	time.Sleep(1 * time.Second)

	// The n1 node should be pruned:
	assert.Equal(t, 1, et.Prunes())
}
