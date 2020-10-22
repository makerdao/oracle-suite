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

package agent

import (
	"fmt"
	"time"

	"github.com/armon/go-metrics"
	"github.com/mitchellh/cli"
)

func (c *Command) RunWithHandlers(args []string, handlers ...EventHandler) int {
	c.Ui = &cli.PrefixedUi{
		OutputPrefix: "==> ",
		InfoPrefix:   "    ",
		ErrorPrefix:  "==> ",
		Ui:           c.Ui,
	}

	// Parse our configs
	c.args = args
	config := c.readConfig()
	if config == nil {
		return 1
	}

	// Setup the log outputs
	logGate, logWriter, logOutput := c.setupLoggers(config)
	if logWriter == nil {
		return 1
	}

	/*
		Setup telemetry
		Aggregate on 10 second intervals for 1 minute. Expose the
		metrics over stderr when there is a SIGUSR1 received.
	*/
	inm := metrics.NewInmemSink(10*time.Second, time.Minute)
	metrics.DefaultInmemSignal(inm)
	metricsConf := metrics.DefaultConfig("serf-agent")

	// Configure the statsite sink
	var fanout metrics.FanoutSink
	if config.StatsiteAddr != "" {
		sink, err := metrics.NewStatsiteSink(config.StatsiteAddr)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to start statsite sink. Got: %s", err))
			return 1
		}
		fanout = append(fanout, sink)
	}

	// Configure the statsd sink
	if config.StatsdAddr != "" {
		sink, err := metrics.NewStatsdSink(config.StatsdAddr)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to start statsd sink. Got: %s", err))
			return 1
		}
		fanout = append(fanout, sink)
	}

	// Initialize the global sink
	if len(fanout) > 0 {
		fanout = append(fanout, inm)
		metrics.NewGlobal(metricsConf, fanout)
	} else {
		metricsConf.EnableHostname = false
		metrics.NewGlobal(metricsConf, inm)
	}

	// Setup serf
	agent := c.setupAgent(config, logOutput)
	if agent == nil {
		return 1
	}
	defer agent.Shutdown()

	for _, handler := range handlers {
		agent.RegisterEventHandler(handler)
	}
	// Start the agent
	ipc := c.startAgent(config, agent, logWriter, logOutput)
	if ipc == nil {
		return 1
	}
	defer ipc.Shutdown()

	// Join startup nodes if specified
	if err := c.startupJoin(config, agent); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	// Enable log streaming
	c.Ui.Info("")
	c.Ui.Output("Log data will now stream in as it occurs:\n")
	logGate.Flush()

	// Start the retry joins
	retryJoinCh := make(chan struct{})
	go c.retryJoin(config, agent, retryJoinCh)

	// Wait for exit
	return c.handleSignals(config, agent, retryJoinCh)
}
