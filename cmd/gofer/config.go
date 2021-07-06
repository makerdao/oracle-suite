package main

import (
	goferConfig "github.com/makerdao/oracle-suite/internal/config/gofer"
	pkgGofer "github.com/makerdao/oracle-suite/pkg/gofer"
	"github.com/makerdao/oracle-suite/pkg/gofer/rpc"
	"github.com/makerdao/oracle-suite/pkg/log"
)

type Config struct {
	Gofer goferConfig.Gofer `json:"gofer"`
}

func (c *Config) Configure(logger log.Logger, noRPC bool) (pkgGofer.Gofer, error) {
	return c.Gofer.ConfigureGofer(logger, noRPC)
}

func (c *Config) ConfigureRPCAgent(logger log.Logger) (*rpc.Agent, error) {
	return c.Gofer.ConfigureRPCAgent(logger)
}
