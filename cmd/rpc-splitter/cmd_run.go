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
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/makerdao/oracle-suite/internal/httpserver"
	"github.com/makerdao/oracle-suite/internal/httpserver/middleware"
	"github.com/makerdao/oracle-suite/internal/rpcsplitter"
	"github.com/spf13/cobra"
)

const httpTimeout = 10 * time.Second

func NewRunCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:     "run",
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"agent"},
		Short:   "",
		Long:    ``,
		RunE: func(_ *cobra.Command, args []string) error {
			ctx, ctxCancel := context.WithCancel(context.Background())
			defer ctxCancel()

			log, err := logger(opts)
			if err != nil {
				return err
			}

			handler, err := rpcsplitter.NewHandler(args)
			if err != nil {
				return err
			}

			srv := httpserver.New(ctx, &http.Server{
				Addr:    opts.Listen,
				Handler: handler,
			})

			srv.Use(&middleware.Recover{
				Recover: func(err interface{}) {
					log.WithField("panic", fmt.Sprintf("%s", err)).Error("Server handler crashed")
				},
			})

			if opts.EnableCORS {
				srv.Use(&middleware.CORS{
					Origin:  func(r *http.Request) string { return r.Header.Get("Origin") },
					Headers: func(*http.Request) string { return "Content-Type" },
					Methods: func(*http.Request) string { return "POST" },
				})
			}

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
