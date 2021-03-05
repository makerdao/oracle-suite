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
	"path/filepath"

	"github.com/sirupsen/logrus"

	ghostConfig "github.com/makerdao/gofer/pkg/ghost/config"
	ghostJSON "github.com/makerdao/gofer/pkg/ghost/config/json"
	"github.com/makerdao/gofer/pkg/gofer"
	goferJSON "github.com/makerdao/gofer/pkg/gofer/config/json"
	"github.com/makerdao/gofer/pkg/log"
	logLogrus "github.com/makerdao/gofer/pkg/log/logrus"
)

func main() {
	var opts options
	rootCmd := NewRootCommand(&opts)

	rootCmd.AddCommand(
		NewRunCmd(&opts),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newLogger(level string) (log.Logger, error) {
	ll, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	lr := logrus.New()
	lr.SetLevel(ll)

	return logLogrus.New(lr), nil
}

func newGofer(opts *options, path string, log log.Logger) (gofer.Gofer, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	err = goferJSON.ParseJSONFile(&opts.GoferConfig, absPath)
	if err != nil {
		return nil, err
	}

	gof, err := opts.GoferConfig.ConfigureGofer(log)
	if err != nil {
		return nil, err
	}

	return gof, nil
}

func newGhost(opts *options, path string, gof gofer.Gofer, log log.Logger) (*ghostConfig.Instances, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	err = ghostJSON.ParseJSONFile(&opts.GhostConfig, absPath)
	if err != nil {
		return nil, err
	}

	i, err := opts.GhostConfig.Configure(ghostConfig.Dependencies{
		Context: context.Background(),
		Gofer:   gof,
		Logger:  log,
	})
	if err != nil {
		return nil, err
	}

	return i, nil
}
