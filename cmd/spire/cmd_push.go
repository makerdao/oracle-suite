package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/pkg/transport/messages"
)

func NewPushCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
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

	cmd.AddCommand(NewPushPriceCmd())

	return cmd
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

			// Read JSON and parse it:
			input, err := readAll(in)
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
