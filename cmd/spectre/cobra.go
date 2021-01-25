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
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/pkg/log"
	logLogrus "github.com/makerdao/gofer/pkg/log/logrus"
	"github.com/makerdao/gofer/pkg/spectre/config"
	configJSON "github.com/makerdao/gofer/pkg/spectre/config/json"
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

func newSpectre(opts *options, path string, log log.Logger) (*config.Instances, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	err = configJSON.ParseJSONFile(&opts.Config, absPath)
	if err != nil {
		return nil, err
	}

	i, err := opts.Config.Configure(config.Dependencies{
		Context: context.Background(),
		Logger:  log,
	})
	if err != nil {
		return nil, err
	}

	return i, nil
}

func NewRunCmd(o *options) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Args:  cobra.ExactArgs(0),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, _ []string) error {
			absPath, err := filepath.Abs(o.ConfigFilePath)
			if err != nil {
				return err
			}

			l, err := newLogger(o.LogVerbosity)
			if err != nil {
				return err
			}

			ins, err := newSpectre(o, absPath, l)
			if err != nil {
				return err
			}

			err = ins.Spectre.Start()
			if err != nil {
				return err
			}
			defer func() {
				err := ins.Spectre.Stop()
				if err != nil {
					l.Errorf("SPECTRE", "Unable to stop Spectre: %s", err)
				}
			}()

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c

			return nil
		},
	}
}

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "spectre",
		Version:       "DEV",
		Short:         "",
		Long:          ``,
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().StringVarP(
		&opts.LogVerbosity,
		"log.verbosity",
		"v",
		"info",
		"verbosity level",
	)
	rootCmd.PersistentFlags().StringVarP(
		&opts.ConfigFilePath,
		"spectre-config",
		"c",
		"./spectre.json",
		"spectre config file",
	)

	return rootCmd
}
