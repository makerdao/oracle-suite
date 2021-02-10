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
	"net/http"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/internal/gofer/web"
)

func NewServerCmd(o *options) *cobra.Command {
	return &cobra.Command{
		Use:   "server",
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

			g, err := newGofer(o, absPath, l)
			if err != nil {
				return err
			}

			webLogger := l.WithField("TAG", "WEB")
			http.HandleFunc("/v1/pairs/", web.PairsHandler(g, webLogger))
			http.HandleFunc("/v1/origins/", web.OriginsHandler(g, webLogger))
			http.HandleFunc("/v1/prices/", web.PricesHandler(g, webLogger))

			err = g.StartFeeder(g.Pairs()...)
			if err != nil {
				return err
			}
			defer g.StopFeeder()

			return web.StartServer(":8080", webLogger)
		},
	}
}
