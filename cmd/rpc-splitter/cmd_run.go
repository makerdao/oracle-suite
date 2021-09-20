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
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/makerdao/oracle-suite/internal/httpserver"
	"github.com/makerdao/oracle-suite/internal/rpcsplitter"
	"github.com/spf13/cobra"
)

func NewRunCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:     "run",
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"agent"},
		Short:   "",
		Long:    ``,
		RunE: func(_ *cobra.Command, args []string) error {
			rpc, err := rpcsplitter.NewHandler(args)
			if err != nil {
				return err
			}

			ctx, ctxCancel := context.WithCancel(context.Background())
			defer ctxCancel()

			srv := httpserver.New(ctx, &http.Server{Handler: rpc, Addr: opts.Listen})
			err = srv.ListenAndServe()
			if err != nil {
				return err
			}
			defer srv.Wait()

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c

			return nil
		},
	}
}
