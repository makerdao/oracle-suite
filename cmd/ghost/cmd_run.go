package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
)

func NewRunCmd(o *options) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Args:  cobra.ExactArgs(0),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, _ []string) error {
			ghostAbsPath, err := filepath.Abs(o.GhostConfigFilePath)
			if err != nil {
				return err
			}

			goferAbsPath, err := filepath.Abs(o.GoferConfigFilePath)
			if err != nil {
				return err
			}

			l, err := newLogger(o.LogVerbosity)
			if err != nil {
				return err
			}

			gof, err := newGofer(o, goferAbsPath, l)
			if err != nil {
				return err
			}

			ins, err := newGhost(o, ghostAbsPath, gof, l)
			if err != nil {
				return err
			}

			err = ins.Ghost.Start()
			if err != nil {
				return err
			}
			defer func() {
				err := ins.Ghost.Stop()
				if err != nil {
					l.Errorf("Unable to stop Ghost: %s", err)
				}
			}()

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c

			return nil
		},
	}
}
