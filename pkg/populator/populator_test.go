package populator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/pkg/graph"
)

func Test_getMinTTL(t *testing.T) {
	p := graph.Pair{Base: "A", Quote: "B"}
	root := graph.NewMedianAggregatorNode(p, 1)
	ttl := time.Second * time.Duration(time.Now().Unix() + 10)
	on1 := graph.NewOriginNode(graph.OriginPair{Origin: "a", Pair: p}, 12 * time.Second, ttl)
	on2 := graph.NewOriginNode(graph.OriginPair{Origin: "b", Pair: p}, 5 * time.Second, ttl)
	on3 := graph.NewOriginNode(graph.OriginPair{Origin: "b", Pair: p}, 10 * time.Second, ttl)

	root.AddChild(on1)
	root.AddChild(on2)
	root.AddChild(on3)

	assert.Equal(t, 5 * time.Second, getMinTTL([]graph.Node{root}))
}

func Test_getMinTTL_SorterThanOneSecond(t *testing.T) {
	p := graph.Pair{Base: "A", Quote: "B"}
	root := graph.NewMedianAggregatorNode(p, 1)
	ttl := time.Second * time.Duration(time.Now().Unix() + 10)
	on1 := graph.NewOriginNode(graph.OriginPair{Origin: "a", Pair: p}, 12 * time.Second, ttl)
	on2 := graph.NewOriginNode(graph.OriginPair{Origin: "b", Pair: p}, -5 * time.Second, ttl)
	on3 := graph.NewOriginNode(graph.OriginPair{Origin: "b", Pair: p}, 0 * time.Second, ttl)

	root.AddChild(on1)
	root.AddChild(on2)
	root.AddChild(on3)

	assert.Equal(t, 1 * time.Second, getMinTTL([]graph.Node{root}))
}
