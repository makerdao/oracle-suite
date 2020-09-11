package graph

// OriginNode contains a Tick fetched directly from an origin.
type OriginNode struct {
	originPair OriginPair
	tick       OriginTick
}

func NewOriginNode(originPair OriginPair) *OriginNode {
	return &OriginNode{
		originPair: originPair,
	}
}

func (n *OriginNode) OriginPair() OriginPair {
	return n.originPair
}

func (n *OriginNode) Feed(tick OriginTick) {
	n.tick = tick
}

func (n *OriginNode) Tick() OriginTick {
	return n.tick
}

func (n OriginNode) Children() []Node {
	return []Node{}
}
