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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/pkg/log"
	"github.com/makerdao/gofer/pkg/node"
	"github.com/makerdao/gofer/pkg/node/config"
	"github.com/makerdao/gofer/pkg/transport/messages"
)

func newLogger(level string, tags []string) (logrus.FieldLogger, error) {
	ll, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	lr := logrus.New()
	lr.SetLevel(ll)

	return log.WrapLogger(lr, nil), nil
}

func newServer(path string, log log.Logger) (*node.Server, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	j, err := config.ParseJSONFile(absPath)
	if err != nil {
		return nil, err
	}

	s, err := j.ConfigureServer(config.Dependencies{
		Context: context.Background(),
		Logger:  log,
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}

func newClient(path string, log log.Logger) (*node.Client, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	j, err := config.ParseJSONFile(absPath)
	if err != nil {
		return nil, err
	}

	c, err := j.ConfigureClient()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func NewAgentCmd(o *options) *cobra.Command {
	return &cobra.Command{
		Use:   "agent",
		Args:  cobra.ExactArgs(0),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, _ []string) error {
			absPath, err := filepath.Abs(o.ConfigFilePath)
			if err != nil {
				return err
			}

			l, err := newLogger(o.LogVerbosity, o.LogTags)
			if err != nil {
				return err
			}

			srv, err := newServer(absPath, l)
			if err != nil {
				return err
			}
			defer func() {
				err := srv.Stop()
				if err != nil {
					l.Errorf("RPC", "Unable to stop RPC Server: %s", err)
				}
			}()

			err = srv.Start()
			if err != nil {
				return err
			}

			c := make(chan os.Signal)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c

			return nil
		},
	}
}

func NewGetPricesCmd(o *options) *cobra.Command {
	return &cobra.Command{
		Use:   "get-prices",
		Args:  cobra.ExactArgs(1),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			absPath, err := filepath.Abs(o.ConfigFilePath)
			if err != nil {
				return err
			}

			l, err := newLogger(o.LogVerbosity, o.LogTags)
			if err != nil {
				return err
			}

			cli, err := newClient(absPath, l)
			if err != nil {
				return err
			}

			err = cli.Start()
			if err != nil {
				return err
			}
			defer func() {
				err := cli.Stop()
				if err != nil {
					l.Errorf("RPC", "Unable to stop RPC Client: %s", err)
				}
			}()

			p, err := cli.GetPrices(args[0])
			if err != nil {
				return err
			}

			bts, err := json.Marshal(p)
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", string(bts))

			return nil
		},
	}
}

func NewBroadcastPriceCmd(o *options) *cobra.Command {
	return &cobra.Command{
		Use:   "broadcast-price",
		Args:  cobra.ExactArgs(0),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, _ []string) error {
			absPath, err := filepath.Abs(o.ConfigFilePath)
			if err != nil {
				return err
			}

			l, err := newLogger(o.LogVerbosity, o.LogTags)
			if err != nil {
				return err
			}

			cli, err := newClient(absPath, l)
			if err != nil {
				return err
			}

			err = cli.Start()
			if err != nil {
				return err
			}
			defer func() {
				err := cli.Stop()
				if err != nil {
					l.Errorf("RPC", "Unable to stop RPC Client: %s", err)
				}
			}()

			// Fetch json from stdin for parse it:
			input, err := ReadAll(os.Stdin)
			if err != nil {
				return err
			}

			msg := &messages.Price{}
			err = msg.Unmarshall(input)
			if err != nil {
				return err
			}

			// Send price message to RPC client:
			err = cli.BroadcastPrice(msg)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "node",
		Version:       "DEV",
		Short:         "",
		Long:          ``,
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().StringVarP(&opts.LogVerbosity, "log.verbosity", "v", "info", "verbosity level")
	rootCmd.PersistentFlags().StringSliceVar(&opts.LogTags, "log.tags", nil, "list of log tags to be printed")
	rootCmd.PersistentFlags().StringVarP(&opts.ConfigFilePath, "config", "c", "./node.json", "node config file")

	return rootCmd
}

func ReadAll(r io.Reader) ([]byte, error) {
	b := make([]byte, 0, 512)
	for {
		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return b, err
		}
	}
}
