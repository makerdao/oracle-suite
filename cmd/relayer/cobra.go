package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

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

			rel, err := newRelayer(absPath)
			if err != nil {
				return err
			}

			err = rel.Start(nil, nil)
			if err != nil {
				return err
			}
			defer rel.Stop()

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

	rootCmd.PersistentFlags().StringVarP(&opts.ConfigFilePath, "relayer-config", "c", "./relayer.json", "config file")

	return rootCmd
}
