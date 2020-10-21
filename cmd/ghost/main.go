//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/hashicorp/serf/cmd/serf/command"
	"github.com/hashicorp/serf/cmd/serf/command/agent"
	"github.com/mitchellh/cli"
)

func main() {
	wg := &sync.WaitGroup{}

	wg.Add(3)
	go serfAgent(1, wg)
	go serfAgent(2, wg)
	go serfAgent(3, wg)

	for serfInfo(1) > 0 || serfInfo(2) > 0 || serfInfo(3) > 0 {
		log.Println("NO")
		time.Sleep(time.Second)
	}

	serfJoin(1, 2)
	serfJoin(1, 3)

	serfMembers(1)

	go func() {
		for x := 1; x < 10; x++ {
			i := serfEvent(1, "test", "pay1")
			log.Printf("SENT pay%d from %d - %d", x, 1, i)
			time.Sleep(3 * time.Second)
		}
	}()

	log.Println("WAITING for ^C")
	wg.Wait()
	log.Println("DONE")
}

var ui = &cli.ConcurrentUi{Ui: &cli.ColoredUi{
	OutputColor: cli.UiColorNone,
	InfoColor:   cli.UiColorGreen,
	ErrorColor:  cli.UiColorRed,
	WarnColor:   cli.UiColorYellow,
	Ui:          &cli.BasicUi{Writer: os.Stdout},
}}

func serfAgent(i int, wg *sync.WaitGroup) {
	c := &agent.Command{
		Ui: &cli.PrefixedUi{
			AskPrefix:       fmt.Sprintf("%d", i),
			AskSecretPrefix: fmt.Sprintf("%d", i),
			OutputPrefix:    fmt.Sprintf("%d", i),
			InfoPrefix:      fmt.Sprintf("%d", i),
			ErrorPrefix:     fmt.Sprintf("%d", i),
			WarnPrefix:      fmt.Sprintf("%d", i),
			Ui:              ui,
		},
		ShutdownCh: makeShutdownCh(),
	}
	c.Run([]string{
		fmt.Sprintf("-node=agent-%d", i),
		fmt.Sprintf("-bind=127.0.0.%d", i),
		fmt.Sprintf("-rpc-addr=127.0.0.%d:7373", i),
	})
	wg.Done()
}
func serfEvent(i int, name, payload string) int {
	c := &command.EventCommand{
		Ui: ui,
	}
	return c.Run([]string{
		fmt.Sprintf("-rpc-addr=127.0.0.%d:7373", i),
		name,
		payload,
	})
}
func serfInfo(i int) int {
	c := &command.InfoCommand{
		Ui: ui,
	}
	return c.Run([]string{
		fmt.Sprintf("-rpc-addr=127.0.0.%d:7373", i),
	})
}
func serfMembers(i int) int {
	c := &command.MembersCommand{
		Ui: ui,
	}
	return c.Run([]string{
		fmt.Sprintf("-rpc-addr=127.0.0.%d:7373", i),
	})
}
func serfJoin(i, j int) int {
	c := &command.JoinCommand{
		Ui: ui,
	}
	return c.Run([]string{
		fmt.Sprintf("-rpc-addr=127.0.0.%d:7373", i),
		fmt.Sprintf("127.0.0.%d", j),
	})
}
func makeShutdownCh() <-chan struct{} {
	resultCh := make(chan struct{})
	go func() {
		signalCh := make(chan os.Signal, 4)
		signal.Notify(signalCh, os.Interrupt)
		for {
			<-signalCh
			resultCh <- struct{}{}
		}
	}()
	return resultCh
}
