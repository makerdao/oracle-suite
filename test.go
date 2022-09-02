package main

import (
	"fmt"
	"github.com/kRoqmoq/oracle-suite/internal/query"
	"github.com/kRoqmoq/oracle-suite/pkg/gofer"
	"github.com/kRoqmoq/oracle-suite/pkg/gofer/graph"
	"github.com/kRoqmoq/oracle-suite/pkg/gofer/graph/feeder"
	"github.com/kRoqmoq/oracle-suite/pkg/gofer/graph/nodes"
	"github.com/kRoqmoq/oracle-suite/pkg/gofer/origins"
	"github.com/kRoqmoq/oracle-suite/pkg/log/null"
	"time"
)

func main() {

	ethusd := gofer.Pair{Base: "ETH", Quote: "USD"}

	bitstampETHUSD := nodes.OriginPair{Origin: "bitstamp", Pair: ethusd}
	binanceETHUSD := nodes.OriginPair{Origin: "binance", Pair: ethusd}

	medianNodeETHUSD := nodes.NewMedianAggregatorNode(ethusd, 2)
	medianNodeETHUSD.AddChild(nodes.NewOriginNode(bitstampETHUSD, time.Minute, 2*time.Minute))
	medianNodeETHUSD.AddChild(nodes.NewOriginNode(binanceETHUSD, time.Minute, 2*time.Minute))
	mETHUSD := map[gofer.Pair]nodes.Aggregator{ethusd: medianNodeETHUSD}

	const defaultWorkerCount = 5
	httpWorkerPool := query.NewHTTPWorkerPool(defaultWorkerCount)

	// Feeder is used to fetch prices:
	f := feeder.NewFeeder(origins.DefaultOriginSet(httpWorkerPool, 2), null.New())

	gETHUSD := graph.NewGofer(mETHUSD, f)
	pETHUSD, err := gETHUSD.Price(ethusd)
	if err != nil {
		panic(err)
	}

	//url := "https://api.apilayer.com/fixer/latest?base=JPY&symbols=USD"
	//req, _ := http.NewRequest("GET", url, nil)
	//req.Header.Set("apikey", "u1Za62g8gi1TILNAsIQ1H4AKR5Ah7iya")
	//client := new(http.Client)
	//resp, err := client.Do(req)
	//defer func(Body io.ReadCloser) {
	//	err := Body.Close()
	//	if err != nil {
	//
	//	}
	//}(resp.Body)
	//
	//body, _ := io.ReadAll(resp.Body)
	//
	////test := map[string]float64{}
	//
	////rData := map[string]interface{}{}
	//
	//rData := make(map[string]interface{})
	////rData := map[string]test
	//err = json.Unmarshal(body, &rData)
	//if err != nil {
	//	return
	//}
	////fmt.Printf("%v", rData)
	//r_d := rData["rates"]
	//
	//fmt.Printf("%v", r_d)
	//if err != nil {
	//	panic(err)
	//}
	fmt.Printf("%s: %f", ethusd, pETHUSD.Price/0.007135)
}
