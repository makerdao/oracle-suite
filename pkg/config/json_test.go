package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/makerdao/gofer/pkg/exchange"
	"github.com/makerdao/gofer/pkg/graph"
)

func TestAll(t *testing.T) {
	// Parse file from JSON
	f, err := ParseJSONFile("../../gofer.json")
	if err != nil {
		panic(err)
	}

	// Build Graph from JSON
	graphs, err := f.BuildGraphs()
	if err != nil {
		panic(err)
	}

	// Print graphs
	for p, g := range graphs {
		fmt.Printf("Pair %s:\n", p)
		ingestor := graph.NewIngestor(exchange.DefaultSet(), 10)
		ingestor.Ingest(g)

		// Get tick
		tick := g.Tick()

		// Print result:
		tickRecurPrinter(tick)

		fmt.Printf(strings.Repeat("-", 80) + "\n\n")
	}
}

func tickRecurPrinter(tick graph.IndirectTick) {
	var recur func(tick interface{}, lvl int)
	recur = func(tick interface{}, lvl int) {
		lvlstr := strings.Repeat("  ", lvl)
		switch typedTick := tick.(type) {
		case graph.IndirectTick:
			if len(typedTick.ExchangeTicks) == 1 && len(typedTick.IndirectTick) == 0 {
				recur(typedTick.ExchangeTicks[0], lvl)
			} else {
				fmt.Printf("%s%s (aggreagate): %f\n", lvlstr, typedTick.Pair, typedTick.Price)
				if typedTick.Error != nil {
					fmt.Printf("%s⌙ errors: %s", lvlstr, formatError(typedTick.Error))
				}

				for _, t := range typedTick.ExchangeTicks {
					recur(t, lvl+1)
				}
				for _, t := range typedTick.IndirectTick {
					recur(t, lvl+1)
				}
			}
		case graph.ExchangeTick:
			fmt.Printf("%s%s (%s): %f\n", lvlstr, typedTick.Pair, typedTick.Exchange, typedTick.Price)
			if typedTick.Error != nil {
				fmt.Printf("%s⌙ errors: %s", lvlstr, formatError(typedTick.Error))
			}
		}
	}
	recur(tick, 0)
}

func formatError(err error) string {
	if err == nil {
		return ""
	}

	return strings.TrimSpace(err.Error()) + "\n"
}
