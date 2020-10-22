package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/serf/cmd/serf/command"
	"github.com/hashicorp/serf/serf"
	"github.com/mitchellh/cli"
	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/cmd/ghost/agent"
	"github.com/makerdao/gofer/pkg/relayer"
	"github.com/makerdao/gofer/pkg/relayer/config"
)

func newRelayer(path string) (*relayer.Relayer, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	j, err := config.ParseJSONFile(absPath)
	if err != nil {
		return nil, err
	}

	r, err := j.MakeRelayer()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func NewRelayerCmd(o *options) *cobra.Command {
	return &cobra.Command{
		Use:   "relayer",
		Args:  cobra.ExactArgs(0),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, _ []string) error {
			// absPath, err := filepath.Abs(o.ConfigFilePath)
			// if err != nil {
			// 	return err
			// }
			//
			// r, err := newRelayer(absPath)
			// if err != nil {
			// 	return err
			// }
			//
			// r.Start(nil, nil)
			// defer r.Stop()

			if err := runSerf(9); err != nil {
				return err
			}

			return nil
		},
	}
}

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "relayer",
		Version:       "DEV",
		Short:         "",
		Long:          ``,
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().StringVarP(&opts.ConfigFilePath, "config", "c", "./relayer.json", "config file")

	return rootCmd
}

type handler string

func (x handler) HandleEvent(event serf.Event) {
	if event.EventType() != serf.EventUser {
		return
	}

	u, ok := event.(serf.UserEvent)
	if !ok {
		return
	}

	if u.Name != string(x) {
		log.Printf("ignoring %s", u.Name)
		return
	}
	log.Println(u.String(), string(u.Payload))
}

func runSerf(id int) error {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	var s int
	go func() {
		defer wg.Done()
		c := &agent.Command{
			Ui: &cli.PrefixedUi{
				AskPrefix:       fmt.Sprintf("Agent %d Ask", id),
				AskSecretPrefix: fmt.Sprintf("Agent %d AskSecret", id),
				OutputPrefix:    fmt.Sprintf("Agent %d Output", id),
				InfoPrefix:      fmt.Sprintf("Agent %d Info", id),
				ErrorPrefix:     fmt.Sprintf("Agent %d Error", id),
				WarnPrefix:      fmt.Sprintf("Agent %d Warn", id),
				Ui:              ui,
			},
			ShutdownCh: makeShutdownCh(),
		}
		s = c.RunWithHandlers([]string{
			fmt.Sprintf("-node=agent-%d", id),
			fmt.Sprintf("-bind=127.0.0.%d", id),
			fmt.Sprintf("-rpc-addr=127.0.0.%d:7373", id),
		}, handler("xxx"))
	}()

	for serfInfo(id) > 0 {
		time.Sleep(time.Second)
	}

	if status := serfJoin(id, 1); status != 0 {
		return fmt.Errorf("cannot join node #%d", 1)
	}

	wg.Wait()
	if s != 0 {
		return fmt.Errorf("node #%d - failed", id)
	}
	return nil
}

func serfInfo(i int) int {
	c := &command.InfoCommand{
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

var ui = &cli.ConcurrentUi{Ui: &cli.ColoredUi{
	OutputColor: cli.UiColorNone,
	InfoColor:   cli.UiColorGreen,
	ErrorColor:  cli.UiColorRed,
	WarnColor:   cli.UiColorYellow,
	Ui:          &cli.BasicUi{Writer: os.Stdout},
}}
