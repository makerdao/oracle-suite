package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

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
