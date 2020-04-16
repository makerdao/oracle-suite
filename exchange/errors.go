package exchange

import "fmt"

var errNoPotentialPricePoint = fmt.Errorf("failed to make request to nil PricePoint")

var errNoExchangeInPotentialPricePoint = fmt.Errorf("failed to make request for nil exchange in given PotentialPricePoint")

var errUnknownExchange = fmt.Errorf("unknown exchange name given")
