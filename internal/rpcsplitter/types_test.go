package splitter

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)


func Test_uint64ToHex(t *testing.T) {
	tests := []struct {
		arg  uint64
		want []byte
	}{
		{arg: 0, want: []byte("0x0")},
		{arg: 1, want: []byte("0x1")},
		{arg: math.MaxUint64, want: []byte("0xffffffffffffffff")},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			assert.Equal(t, tt.want, bigIntToHex(new(big.Int).SetUint64(tt.arg)))
		})
	}
}

func Test_hexToUint64(t *testing.T) {
	tests := []struct {
		arg     []byte
		want    uint64
		wantErr bool
	}{
		{arg: []byte("0x0"), want: 0, wantErr: false},
		{arg: []byte("0x1"), want: 1, wantErr: false},
		{arg: []byte("0xffffffffffffffff"), want: math.MaxUint64, wantErr: false},
		{arg: []byte("0"), want: 0, wantErr: false},
		{arg: []byte("1"), want: 1, wantErr: false},
		{arg: []byte("ffffffffffffffff"), want: math.MaxUint64, wantErr: false},
		{arg: []byte("foo"), want: 0, wantErr: true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := hexToBigInt(tt.arg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got.Uint64())
			}
		})
	}
}
