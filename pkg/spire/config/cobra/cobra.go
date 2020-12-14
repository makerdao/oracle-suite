package cobra

import (
	"github.com/spf13/pflag"

	"github.com/makerdao/gofer/pkg/spire/config"
)

func RegisterFlags(cfg *config.Config, flags *pflag.FlagSet) {
	flags.StringVar(&cfg.Ethereum.From, "config.ethereum.from", "", "")
	flags.StringVar(&cfg.Ethereum.Keystore, "config.ethereum.keystore", "", "")
	flags.StringVar(&cfg.Ethereum.Password, "config.ethereum.password", "", "")

	flags.StringSliceVar(&cfg.P2P.Listen, "config.p2p.listen", nil, "")
	flags.StringSliceVar(&cfg.P2P.BootstrapAddrs, "config.p2p.boostrap", nil, "")
	flags.StringSliceVar(&cfg.P2P.BlockedAddrs, "config.p2p.blocked", nil, "")

	flags.StringSliceVar(&cfg.Feeds, "config.feeds", nil, "")
	flags.StringSliceVar(&cfg.Feeds, "config.pairs", nil, "")
}
