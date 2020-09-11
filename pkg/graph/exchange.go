package graph

// ExchangeNode contains a Tick fetched directly from an exchange.
type ExchangeNode struct {
	exchangePair ExchangePair
	tick         ExchangeTick
}

func NewExchangeNode(exchangePair ExchangePair) *ExchangeNode {
	return &ExchangeNode{
		exchangePair: exchangePair,
	}
}

func (n *ExchangeNode) ExchangePair() ExchangePair {
	return n.exchangePair
}

func (n *ExchangeNode) Feed(tick ExchangeTick) {
	n.tick = tick
}

func (n *ExchangeNode) Tick() ExchangeTick {
	return n.tick
}

func (n ExchangeNode) Children() []Node {
	return []Node{}
}
