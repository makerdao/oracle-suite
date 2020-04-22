package exchange

import "fmt"

var errNoPotentialPricePoint = fmt.Errorf("failed to make request to nil PotentialPricePoint")

var errWrongPotentialPricePoint = fmt.Errorf("failed to make request using wrong PotentialPricePoint")

var errNoExchangeInPotentialPricePoint = fmt.Errorf("failed to make request for nil exchange in given PotentialPricePoint")

var errUnknownExchange = fmt.Errorf("unknown exchange name given")

var errNoPoolPassed = fmt.Errorf("no query worker pool passed")

var errEmptyExchangeResponse = fmt.Errorf("empty exchange response received")
