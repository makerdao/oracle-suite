package feeds

import "github.com/makerdao/oracle-suite/pkg/ethereum"

type Feeds []string

func (f *Feeds) Addresses() []ethereum.Address {
	var addrs []ethereum.Address
	for _, addr := range *f {
		addrs = append(addrs, ethereum.HexToAddress(addr))
	}
	return addrs
}
