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
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func NewAgentCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "agent",
		Args:  cobra.NoArgs,
		Short: "Start an RPC server",
		Long:  `Start an RPC server.`,
		RunE: func(_ *cobra.Command, args []string) (err error) {
			log, err := newLogger(opts.LogVerbosity)
			if err != nil {
				return
			}

			srv, err := newServer(opts, opts.ConfigFilePath, log)
			if err != nil {
				return
			}

			// Start RPC server:
			err = srv.Start()
			if err != nil {
				return
			}
			defer func() {
				serr := srv.Stop()
				if err == nil {
					err = serr
				}
			}()

			// Wait for interrupt signal
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c

			return
		},
	}
}
