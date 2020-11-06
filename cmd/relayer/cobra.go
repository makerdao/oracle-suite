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

	"github.com/makerdao/gofer/internal/log"
	"github.com/makerdao/gofer/pkg/relayer/config"
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

func newRelayer(path string, log logrus.FieldLogger) (*config.Instances, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	j, err := config.ParseJSONFile(absPath)
	if err != nil {
		return nil, err
	}

	i, err := j.Configure(config.Dependencies{
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

			log, err := newLogger(o.LogVerbosity, o.LogTags)
			if err != nil {
				return err
			}

			ins, err := newRelayer(absPath, log)
			if err != nil {
				return err
			}

			err = ins.Relayer.Start()
			if err != nil {
				return err
			}
			defer func() {
				err := ins.Relayer.Stop()
				if err != nil {
					log.Error("RELAYER", "Unable to stop relayer: %s", err)
				}
			}()

			c := make(chan os.Signal)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c

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

	rootCmd.PersistentFlags().StringVarP(&opts.LogVerbosity, "log.verbosity", "v", "info", "verbosity level")
	rootCmd.PersistentFlags().StringSliceVar(&opts.LogTags, "log.tags", nil, "list of log tags to be printed")
	rootCmd.PersistentFlags().StringVarP(&opts.ConfigFilePath, "relayer-config", "c", "./relayer.json", "relayer config file")

	return rootCmd
}
