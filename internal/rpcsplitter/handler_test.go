package rpcsplitter

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newRPCErrorType(msg string) *errorsType {
	return &errorsType{err: errors.New(msg)}
}

func newRPCNumberType(v uint64) *numberType {
	return (*numberType)(new(big.Int).SetUint64(v))
}


func Test_useMostCommon(t *testing.T) {
	tests := []struct {
		res        []interface{}
		minReq     int
		want       interface{}
		wantErrMsg string
	}{
		{
			res: []interface{}{
				newRPCNumberType(10),
			},
			minReq:     1,
			want:       newRPCNumberType(10),
			wantErrMsg: "",
		},
		{
			res: []interface{}{
				newRPCNumberType(10),
				newRPCNumberType(10),
				newRPCNumberType(30),
			},
			minReq:     1,
			want:       newRPCNumberType(10),
			wantErrMsg: "",
		},
		{
			res: []interface{}{
				newRPCNumberType(10),
				newRPCNumberType(10),
			},
			minReq:     2,
			want:       newRPCNumberType(10),
			wantErrMsg: "",
		},
		{
			res: []interface{}{
				newRPCNumberType(10),
				newRPCNumberType(10),
				newRPCNumberType(30),
			},
			minReq:     2,
			want:       newRPCNumberType(10),
			wantErrMsg: "",
		},
		{
			res: []interface{}{
				newRPCNumberType(10),
				newRPCNumberType(10),
				newRPCNumberType(10),
			},
			minReq:     3,
			want:       newRPCNumberType(10),
			wantErrMsg: "",
		},
		// Fails because there is not enough same responses:
		{
			res: []interface{}{
				newRPCNumberType(10),
				newRPCNumberType(10),
				newRPCNumberType(15),
			},
			minReq:     3,
			want:       nil,
			wantErrMsg: NoResponsesErr.Error(),
		},
		// Fails because there are multiple responses and any of them occurs only once:
		{
			res: []interface{}{
				newRPCNumberType(10),
				newRPCNumberType(15),
				newRPCNumberType(30),
			},
			minReq:     1,
			want:       nil,
			wantErrMsg: NoResponsesErr.Error(),
		},
		// Fails because there are multiple responses and any of them occurs only once:
		{
			res: []interface{}{
				newRPCNumberType(10),
				newRPCErrorType("a"),
				newRPCErrorType("b"),
			},
			minReq:     1,
			want:       newRPCNumberType(10),
			wantErrMsg: NoResponsesErr.Error(),
		},
		// Fails because error is the most common answer:
		{
			res: []interface{}{
				newRPCNumberType(10),
				newRPCErrorType("a"),
				newRPCErrorType("a"),
			},
			minReq:     1,
			want:       nil,
			wantErrMsg: "a",
		},
		// Fails because we got only errors:
		{
			res: []interface{}{
				newRPCErrorType("a"),
				newRPCErrorType("a"),
				newRPCErrorType("a"),
			},
			minReq:     3,
			want:       nil,
			wantErrMsg: "a",
		},
		// Fails because error is the most common answer:
		{
			res: []interface{}{
				newRPCErrorType("a"),
			},
			minReq:     3,
			want:       nil,
			wantErrMsg: "a",
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := useMostCommon(tt.res, tt.minReq)
			if tt.wantErrMsg != "" {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else {
				assert.NoError(t, err)
				assert.True(t, compare(got, tt.want))
			}
		})
	}
}

func Test_useMedian(t *testing.T) {
	tests := []struct {
		res        []interface{}
		minReq     int
		want       *numberType
		wantErrMsg string
	}{
		{
			res: []interface{}{
				newRPCNumberType(10),
			},
			minReq:     1,
			want:       newRPCNumberType(10),
			wantErrMsg: "",
		},

		{
			res: []interface{}{
				newRPCNumberType(10),
				newRPCNumberType(30),
			},
			minReq:     1,
			want:       newRPCNumberType(20),
			wantErrMsg: "",
		},
		{
			res: []interface{}{
				newRPCNumberType(10),
				newRPCNumberType(30),
			},
			minReq:     2,
			want:       newRPCNumberType(20),
			wantErrMsg: "",
		},
		{
			res: []interface{}{
				newRPCNumberType(1),
				newRPCNumberType(10),
				newRPCNumberType(100),
			},
			minReq:     3,
			want:       newRPCNumberType(10),
			wantErrMsg: "",
		},
		{
			res: []interface{}{
				newRPCNumberType(10),
				newRPCErrorType("a"),
				newRPCErrorType("a"),
			},
			minReq:     3,
			want:       nil,
			wantErrMsg: "a",
		},
		{
			res: []interface{}{
				newRPCNumberType(10),
				newRPCErrorType("a"),
				newRPCErrorType("b"),
			},
			minReq:     3,
			want:       nil,
			wantErrMsg: NoResponsesErr.Error(),
		},
		{
			res: []interface{}{
				newRPCErrorType("a"),
			},
			minReq:     3,
			want:       nil,
			wantErrMsg: "a",
		},
		{
			res: []interface{}{
				newRPCNumberType(math.MaxUint64),
				newRPCNumberType(math.MaxUint64),
			},
			minReq:     2,
			want:       newRPCNumberType(math.MaxUint64),
			wantErrMsg: "",
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := useMedian(tt.res, tt.minReq)
			if tt.wantErrMsg != "" {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.want.Compare(got))
			}
		})
	}
}

func Test_useMedianDist(t *testing.T) {
	tests := []struct {
		res        []interface{}
		dist       int64
		minReq     int
		want       *numberType
		wantErrMsg string
	}{
		{
			res: []interface{}{
				newRPCNumberType(10),
			},
			dist:       1,
			minReq:     1,
			want:       newRPCNumberType(10),
			wantErrMsg: "",
		},
		{
			res: []interface{}{
				newRPCNumberType(4),
				newRPCNumberType(5),
				newRPCNumberType(6),
				newRPCNumberType(7),
				newRPCNumberType(8),
				newRPCNumberType(9),
				newRPCNumberType(10),
			},
			dist:       -2,
			minReq:     7,
			want:       newRPCNumberType(5),
			wantErrMsg: "",
		},
		{
			res: []interface{}{
				newRPCNumberType(4),
				newRPCNumberType(5),
				newRPCNumberType(6),
				newRPCNumberType(7),
				newRPCNumberType(8),
				newRPCNumberType(9),
				newRPCNumberType(10),
			},
			dist:       2,
			minReq:     7,
			want:       newRPCNumberType(9),
			wantErrMsg: "",
		},
		{
			res: []interface{}{
				newRPCNumberType(10),
				newRPCNumberType(9),
				newRPCNumberType(8),
				newRPCNumberType(7),
				newRPCNumberType(6),
				newRPCNumberType(5),
				newRPCNumberType(4),
			},
			dist:       -2,
			minReq:     7,
			want:       newRPCNumberType(5),
			wantErrMsg: "",
		},
		{
			res: []interface{}{
				newRPCNumberType(10),
				newRPCNumberType(9),
				newRPCNumberType(8),
				newRPCNumberType(7),
				newRPCNumberType(6),
				newRPCNumberType(5),
				newRPCNumberType(4),
			},
			dist:       2,
			minReq:     7,
			want:       newRPCNumberType(9),
			wantErrMsg: "",
		},
		{
			res: []interface{}{
				newRPCNumberType(10),
				newRPCErrorType("a"),
				newRPCErrorType("a"),
			},
			dist:       1,
			minReq:     3,
			want:       nil,
			wantErrMsg: "a",
		},
		{
			res: []interface{}{
				newRPCNumberType(10),
				newRPCErrorType("a"),
				newRPCErrorType("b"),
			},
			dist:       1,
			minReq:     3,
			want:       nil,
			wantErrMsg: NoResponsesErr.Error(),
		},
		{
			res: []interface{}{
				newRPCErrorType("a"),
			},
			dist:       1,
			minReq:     3,
			want:       nil,
			wantErrMsg: "a",
		},
		{
			res: []interface{}{
				newRPCNumberType(math.MaxUint64),
				newRPCNumberType(math.MaxUint64),
			},
			dist:       1,
			minReq:     2,
			want:       newRPCNumberType(math.MaxUint64),
			wantErrMsg: "",
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := useMedianDist(tt.res, tt.minReq, tt.dist)
			if tt.wantErrMsg != "" {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.want.Compare(got))
			}
		})
	}
}
