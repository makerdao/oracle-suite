package main

import (
	"fmt"
	"time"

	"github.com/kRoqmoq/oracle-suite/internal/query"

	"github.com/kRoqmoq/oracle-suite/pkg/gofer"
	"github.com/kRoqmoq/oracle-suite/pkg/gofer/graph"
	"github.com/kRoqmoq/oracle-suite/pkg/gofer/graph/feeder"
	"github.com/kRoqmoq/oracle-suite/pkg/gofer/graph/nodes"
	"github.com/kRoqmoq/oracle-suite/pkg/gofer/origins"
	"github.com/kRoqmoq/oracle-suite/pkg/log/null"
)

func main() {
	// Price model model for the BTC/USD pair:
	//btcusd := gofer.Pair{Base: "BTC", Quote: "USD"}
	//bitfinexBTCUSD := nodes.OriginPair{Origin: "bitfinex", Pair: btcusd}
	//binanceBTCUSD := nodes.OriginPair{Origin: "binance", Pair: btcusd}
	//medianNode := nodes.NewMedianAggregatorNode(btcusd, 2)
	//medianNode.AddChild(nodes.NewOriginNode(bitfinexBTCUSD, time.Minute, 2*time.Minute))
	//medianNode.AddChild(nodes.NewOriginNode(binanceBTCUSD, time.Minute, 2*time.Minute))
	//m := map[gofer.Pair]nodes.Aggregator{btcusd: medianNode}

	ethusd := gofer.Pair{Base: "ETH", Quote: "USD"}
	usdjpy := gofer.Pair{Base: "USD", Quote: "JPY"}

	bitstampETHUSD := nodes.OriginPair{Origin: "bitstamp", Pair: ethusd}
	binanceETHUSD := nodes.OriginPair{Origin: "binance", Pair: ethusd}

	fxUSDJPY := nodes.OriginPair{Origin: "fx", Pair: usdjpy}

	medianNodeETHUSD := nodes.NewMedianAggregatorNode(ethusd, 2)
	medianNodeETHUSD.AddChild(nodes.NewOriginNode(bitstampETHUSD, time.Minute, 2*time.Minute))
	medianNodeETHUSD.AddChild(nodes.NewOriginNode(binanceETHUSD, time.Minute, 2*time.Minute))
	mETHUSD := map[gofer.Pair]nodes.Aggregator{ethusd: medianNodeETHUSD}

	medianNodeUSDJPY := nodes.NewMedianAggregatorNode(usdjpy, 1)
	medianNodeUSDJPY.AddChild(nodes.NewOriginNode(fxUSDJPY, time.Minute, 2*time.Minute))
	mUSDJPY := map[gofer.Pair]nodes.Aggregator{ethusd: medianNodeETHUSD}

	const defaultWorkerCount = 5
	httpWorkerPool := query.NewHTTPWorkerPool(defaultWorkerCount)

	// Feeder is used to fetch prices:
	f := feeder.NewFeeder(origins.DefaultOriginSet(httpWorkerPool, 1), null.New())

	// Initialize gofer and ask for BTC/USD price:
	//g := graph.NewGofer(m, f)
	//p, err := g.Price(btcusd)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("%s: %f", btcusd, p.Price)

	//test, _ := gofer.Gofer.Pairs
	gETHUSD := graph.NewGofer(mETHUSD, f)
	pETHUSD, err := gETHUSD.Price(ethusd)
	if err != nil {
		panic(err)
	}
	gUSDJPY := graph.NewGofer(mUSDJPY, f)
	pUSDJPY, err := gUSDJPY.Price(usdjpy)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s: %f", ethusd, pETHUSD.Price*pUSDJPY.Price)
}
