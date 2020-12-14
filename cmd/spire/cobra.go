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
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/pkg/log"
	logLogrus "github.com/makerdao/gofer/pkg/log/logrus"
	"github.com/makerdao/gofer/pkg/spire"
	"github.com/makerdao/gofer/pkg/spire/config"
	configCobra "github.com/makerdao/gofer/pkg/spire/config/cobra"
	configJSON "github.com/makerdao/gofer/pkg/spire/config/json"
	"github.com/makerdao/gofer/pkg/transport/messages"
)

var (
	logger log.Logger
	client *spire.Client
)

func newLogger(level string) (log.Logger, error) {
	ll, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	lr := logrus.New()
	lr.SetLevel(ll)

	return logLogrus.New(lr), nil
}

func newServer(opts *options, log log.Logger) (*spire.Server, error) {
	if opts.ConfigPath != "" {
		absPath, err := filepath.Abs(opts.ConfigPath)
		if err != nil {
			return nil, err
		}

		err = configJSON.ParseJSONFile(&opts.Config, absPath)
		if err != nil {
			return nil, err
		}
	}

	s, err := opts.Config.ConfigureServer(config.Dependencies{
		Context: context.Background(),
		Logger:  log,
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}

func newClient(opts *options, log log.Logger) (*spire.Client, error) {
	if opts.ConfigPath != "" {
		absPath, err := filepath.Abs(opts.ConfigPath)
		if err != nil {
			return nil, err
		}

		err = configJSON.ParseJSONFile(&opts.Config, absPath)
		if err != nil {
			return nil, err
		}
	}

	c, err := opts.Config.ConfigureClient(config.Dependencies{
		Context: context.Background(),
		Logger:  log,
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "spire",
		Version:       "DEV",
		Short:         "",
		Long:          ``,
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().StringVarP(
		&opts.Verbosity,
		"verbosity",
		"v",
		"error",
		"log verbosity level",
	)

	rootCmd.PersistentFlags().StringVarP(
		&opts.ConfigPath,
		"config",
		"c",
		"./spire.json",
		"spire config file",
	)

	configCobra.RegisterFlags(&opts.Config, rootCmd.PersistentFlags())

	return rootCmd
}

func NewAgentCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "agent",
		Args:  cobra.ExactArgs(0),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, _ []string) error {
			var err error

			logger, err = newLogger(opts.Verbosity)
			if err != nil {
				return err
			}
			srv, err := newServer(opts, logger)
			if err != nil {
				return err
			}

			err = srv.Start()
			if err != nil {
				return err
			}
			defer func() {
				err := srv.Stop()
				if err != nil {
					logger.WithError(err).Error("Unable to stop RPC Server")
				}
			}()

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c

			return nil
		},
	}
}

func NewPullCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "pull",
		Args:  cobra.ExactArgs(1),
		Short: "",
		Long:  ``,
		PersistentPreRunE: func(_ *cobra.Command, args []string) error {
			var err error

			logger, err = newLogger(opts.Verbosity)
			if err != nil {
				return err
			}
			client, err = newClient(opts, logger)
			if err != nil {
				return err
			}

			return client.Start()
		},
		PersistentPostRunE: func(_ *cobra.Command, args []string) error {
			err := client.Stop()
			if err != nil {
				logger.WithError(err).Error("Unable to stop RPC Client")
			}
			return nil
		},
	}
}

func NewPushCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Args:  cobra.ExactArgs(1),
		Short: "",
		Long:  ``,
		PersistentPreRunE: func(_ *cobra.Command, args []string) error {
			var err error

			logger, err = newLogger(opts.Verbosity)
			if err != nil {
				return err
			}
			client, err = newClient(opts, logger)
			if err != nil {
				return err
			}

			return client.Start()
		},
		PersistentPostRunE: func(_ *cobra.Command, args []string) error {
			err := client.Stop()
			if err != nil {
				logger.WithError(err).Error("Unable to stop RPC Client")
			}
			return nil
		},
	}
}

func NewPullPricesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prices",
		Args:  cobra.ExactArgs(1),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			p, err := client.PullPrices(args[0])
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

func NewPullPriceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "price",
		Args:  cobra.ExactArgs(2),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			p, err := client.PullPrice(args[0], args[1])
			if err != nil {
				return err
			}
			if p == nil {
				return errors.New("there is no price in the datastore for a given feeder and asset pair")
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

func NewPushPriceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "price",
		Args:  cobra.MaximumNArgs(1),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			var err error

			in := os.Stdin
			if len(args) == 1 {
				in, err = os.Open(args[0])
				if err != nil {
					return err
				}
			}

			// Fetch json from stdin for parse it:
			input, err := ReadAll(in)
			if err != nil {
				return err
			}

			msg := &messages.Price{}
			err = msg.Unmarshall(input)
			if err != nil {
				return err
			}

			// Send price message to RPC client:
			err = client.PublishPrice(msg)
			if err != nil {
				return err
			}

			return nil
		},
	}
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
