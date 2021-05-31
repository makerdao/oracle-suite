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

package datastore

import (
	"github.com/makerdao/oracle-suite/pkg/log"
)

type LoggingNullDatastore struct {
	priceStore *PriceStore
	log        log.Logger
}

func NewLoggingNullDatastore(logger log.Logger) *LoggingNullDatastore {
	return &LoggingNullDatastore{
		priceStore: NewPriceStore(),
		log:        logger.WithField("tag", "NULL_"+LoggerTag),
	}
}

func (n *LoggingNullDatastore) Prices() *PriceStore {
	return n.priceStore
}

func (n *LoggingNullDatastore) Start() error {
	n.log.Info("Starting")
	return nil
}

func (n *LoggingNullDatastore) Stop() error {
	n.log.Info("Stopping")
	return nil
}
