package reducer

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"

	"makerdao/gofer/store/leveldb"
)

func BenchmarkListRW(b *testing.B) {
	// setup
	path := "./benchmark-list.leveldb"
	db := leveldb.NewLevelDbStore(3 << 30)
	_ = db.Open(path)

	// benchmarks
	for i := 0; i < b.N; i++ {
		prices := make([]*Price, 5)
		for j := 0; j < len(prices); j++ {
			prices[j] = &Price{
				Timestamp: time.Now().Unix(),
				Exchange:  "mjaucoinmarket",
				Base:      "bench",
				Quote:     "mark",
				Price:     uint64(i + j + 1337),
				Volume:    uint64(i + j + 1),
			}
		}
		priceAverage := &PriceAverage{
			Base:   "bench",
			Quote:  "mark",
			Prices: prices,
			Mean:   prices[0].Price,
			Median: prices[1].Price,
			Cccagg: prices[2].Price,
		}
		buf, _ := proto.Marshal(priceAverage)
		_ = db.Put(fmt.Sprintf("list-%d", i), buf)
	}

	for i := 0; i < b.N; i++ {
		priceAverage := &PriceAverage{}
		buf, _ := db.Get(fmt.Sprintf("list-%f", math.Floor(rand.Float64()*float64(b.N))))
		_ = proto.Unmarshal(buf, priceAverage)
	}

	// tear-down
	db.Close()
	os.RemoveAll(path)
}
