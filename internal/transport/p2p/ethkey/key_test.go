package ethkey

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressToPeerID(t *testing.T) {
	assert.Equal(
		t,
		"1Afqz6rsuyYpr7Dpp12PbftE22nYH3k2Fw5",
		AddressToPeerID("0x69B352cbE6Fc5C130b6F62cc8f30b9d7B0DC27d0").Pretty(),
	)

	assert.Equal(
		t,
		"",
		AddressToPeerID("").Pretty(),
	)
}
