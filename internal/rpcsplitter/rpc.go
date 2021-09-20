package rpcsplitter

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"reflect"
	"sort"
	"strings"

	gethRPC "github.com/ethereum/go-ethereum/rpc"
)

// maxBlocksBehind is the number of blocks behind the median of the block
// numbers reported by the endpoints that determines the lowest block number
// that can returned by the eth_blockNumber method.
const maxBlocksBehind = 3

type rpcClient interface {
	Call(result interface{}, method string, args ...interface{}) error
}

// rpc is an RPC proxy server. It merges multiple RPC endpoints into one.
type rpc struct {
	rpc *gethRPC.Server // rpc is an RPC server.
	cli []rpcClient     // cli is a list of RPC clients.
	eth *rpcETHAPI      // eth implements procedures with the "eth_" prefix.
	net *rpcNETAPI      // net implements procedures with the "net_" prefix.
}

type rpcETHAPI struct {
	srv *rpc
}

type rpcNETAPI struct {
	srv *rpc
}

func newRPC(endpoints []string) (*rpc, error) {
	srv := &rpc{rpc: gethRPC.NewServer(), cli: make([]rpcClient, len(endpoints))}
	eth := &rpcETHAPI{srv: srv}
	net := &rpcNETAPI{srv: srv}
	srv.eth = eth
	srv.net = net
	for n, e := range endpoints {
		c, err := gethRPC.Dial(e)
		if err != nil {
			return nil, err
		}
		srv.cli[n] = c
	}
	if err := srv.rpc.RegisterName("eth", eth); err != nil {
		return nil, err
	}
	if err := srv.rpc.RegisterName("net", net); err != nil {
		return nil, err
	}
	return srv, nil
}

func (r *rpc) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	r.rpc.ServeHTTP(rw, req)
}

// BlockNumber implements the "eth_blockNumber" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
func (r *rpcETHAPI) BlockNumber() (interface{}, error) {
	return useMedianDist(r.srv.doRPC((*numberType)(nil), "eth_blockNumber"), r.srv.minReq(), -maxBlocksBehind)
}

// GetBlockByHash implements the "eth_getBlockByHash" call.
//
// The number returned by this method is the median of all numbers returned
// by the endpoints.
func (r *rpcETHAPI) GetBlockByHash(blockHash hashType, obj bool) (interface{}, error) {
	return useMostCommon(r.srv.doRPC((*blockType)(nil), "eth_getBlockByHash", blockHash, obj), r.srv.minReq())
}

// GetBlockByNumber implements the "eth_getBlockByNumber" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
func (r *rpcETHAPI) GetBlockByNumber(blockNumber numberType, obj bool) (interface{}, error) {
	return useMostCommon(r.srv.doRPC((*blockType)(nil), "eth_getBlockByNumber", blockNumber, obj), r.srv.minReq())
}

// GetTransactionByHash implements the "eth_getTransactionByHash" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
func (r *rpcETHAPI) GetTransactionByHash(txHash hashType) (interface{}, error) {
	return useMostCommon(r.srv.doRPC((*transactionType)(nil), "eth_getTransactionByHash", txHash), r.srv.minReq())
}

// GetTransactionCount implements the "eth_getTransactionCount" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) GetTransactionCount(addr addressType, blockNumber blockNumberType) (interface{}, error) {
	return useMostCommon(r.srv.doRPC((*numberType)(nil), "eth_getTransactionCount", addr, blockNumber), r.srv.minReq())
}

// GetTransactionReceipt implements the "eth_getTransactionReceipt" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
func (r *rpcETHAPI) GetTransactionReceipt(txHash hashType) (interface{}, error) {
	return useMostCommon(r.srv.doRPC((*transactionReceiptType)(nil), "eth_getTransactionReceipt", txHash), r.srv.minReq())
}

// TODO: eth_getBlockTransactionCountByHash
// TODO: eth_getBlockTransactionCountByNumber
// TODO: eth_getTransactionByBlockHashAndIndex
// TODO: eth_getTransactionByBlockNumberAndIndex

// SendRawTransaction implements the "eth_sendRawTransaction" call.
//
// It returns the most common response.
func (r *rpcETHAPI) SendRawTransaction(data jsonType) (interface{}, error) {
	return useMostCommon(r.srv.doRPC((*hashType)(nil), "eth_sendRawTransaction", data), 1)
}

// GetBalance implements the "eth_getBalance" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) GetBalance(addr addressType, blockNumber blockNumberType) (interface{}, error) {
	return useMostCommon(r.srv.doRPC((*numberType)(nil), "eth_getBalance", addr, blockNumber), r.srv.minReq())
}

// GetCode implements the "eth_getCode" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) GetCode(addr addressType, blockNumber blockNumberType) (interface{}, error) {
	return useMostCommon(r.srv.doRPC((*bytesType)(nil), "eth_getCode", addr, blockNumber), r.srv.minReq())
}

// GetStorageAt implements the "eth_getStorageAt" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) GetStorageAt(data addressType, position numberType, blockNumber blockNumberType) (interface{}, error) {
	return useMostCommon(r.srv.doRPC((*hashType)(nil), "eth_getStorageAt", data, position, blockNumber), r.srv.minReq())
}

// TODO: eth_accounts
// TODO: eth_getProof

// Call implements the "eth_call" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) Call(args jsonType, blockNumber blockNumberType, overrides *jsonType) (interface{}, error) {
	return useMostCommon(r.srv.doRPC((*bytesType)(nil), "eth_call", args, blockNumber, overrides), r.srv.minReq())
}

// TODO: eth_getLogs
// TODO: eth_protocolVersion

// GasPrice implements the "eth_gasPrice" call.
//
// The number returned by this method is the median of all numbers returned
// by the endpoints.
func (r *rpcETHAPI) GasPrice() (interface{}, error) {
	return useMedian(r.srv.doRPC((*numberType)(nil), "eth_gasPrice"), r.srv.minReq())
}

// EstimateGas implements the "eth_estimateGas" call.
//
// The number returned by this method is the median of all numbers returned
// by the endpoints.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) EstimateGas(args jsonType, blockNumber blockNumberType) (interface{}, error) {
	return useMedian(r.srv.doRPC((*numberType)(nil), "eth_estimateGas", args, blockNumber), r.srv.minReq())
}

// TODO: eth_feeHistory

// MaxPriorityFeePerGas implements the "eth_maxPriorityFeePerGas" call.
//
// The number returned by this method is the median of all numbers returned
// by the endpoints.
func (r *rpcETHAPI) MaxPriorityFeePerGas() (interface{}, error) {
	return useMedian(r.srv.doRPC((*numberType)(nil), "eth_maxPriorityFeePerGas"), r.srv.minReq())
}

// ChainId implements the "eth_chainId" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
func (r *rpcETHAPI) ChainId() (interface{}, error) {
	return useMostCommon(r.srv.doRPC((*numberType)(nil), "eth_chainId"), r.srv.minReq())
}

// TODO: eth_getUncleByBlockNumberAndIndex
// TODO: eth_getUncleByBlockHashAndIndex
// TODO: eth_getUncleCountByBlockHash
// TODO: eth_getUncleCountByBlockNumber
// TODO: eth_getFilterChanges
// TODO: eth_getFilterLogs
// TODO: eth_newBlockFilter
// TODO: eth_newFilter
// TODO: eth_newPendingTransactionFilter
// TODO: eth_uninstallFilter

// Version implements the "net_version" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
func (r *rpcNETAPI) Version() (interface{}, error) {
	return useMostCommon(r.srv.doRPC((*jsonType)(nil), "net_version"), r.srv.minReq())
}

// doRPC executes RPC on all endpoints and returns a slice with all results.
// The typ argument must be an empty pointer with a type to which the results
// will be converted.
func (r *rpc) doRPC(typ interface{}, method string, args ...interface{}) (res []interface{}) {
	err := r.processRPCArgs(&args)
	if err != nil {
		return []interface{}{err}
	}
	ch := make(chan interface{})
	rt := reflect.TypeOf(typ).Elem()
	for _, cli := range r.cli {
		cli := cli
		go func() {
			var val interface{}
			var err error
			defer func() {
				if e := recover(); e != nil {
					ch <- fmt.Errorf("panic: %s", e)
				} else if err != nil {
					ch <- err
				} else {
					ch <- val.(interface{})
				}
			}()
			val = reflect.New(rt).Interface()
			err = cli.Call(val, method, args...)
		}()
	}
	for {
		res = append(res, <-ch)
		if len(res) == len(r.cli) {
			break
		}
	}
	return
}

// processRPCArgs removes trailing nil arguments from the args slice and
// replaces tagged blocks to block numbers.
func (r *rpc) processRPCArgs(args *[]interface{}) error {
	for i := len(*args) - 1; i >= 0; i-- {
		// Remove null arguments from the end of the args list. Some RPC
		// servers do not like null parameters and will return a "bad request"
		// error if they occur.
		if len(*args)-1 == i && isNil((*args)[i]) {
			*args = (*args)[0:i]
			continue
		}
		// Replace tagged blocks with block numbers. This is necessary because
		// different RPC endpoints may convert these tags to different block
		// numbers.
		if arg, ok := (*args)[i].(blockNumberType); ok && arg.IsTag() {
			if arg.IsEarliest() {
				// The earliest block will be completely different on different
				// endpoints. It is impossible to reliably support it.
				return errors.New("earliest tag is not supported")
			}
			// The latest and pending blocks are handled in the same way.
			bn, err := r.eth.BlockNumber()
			if err != nil {
				return err
			}
			(*args)[i] = bn
		}
	}
	return nil
}

// minReq returns a number indicating how many times the same response
// must be returned by different endpoints to be considered valid.
func (r *rpc) minReq() int {
	l := len(r.cli)
	if l <= 2 {
		return l
	}
	return l - 1
}

type rpcErrors []error

func (e rpcErrors) Error() string {
	switch len(e) {
	case 0:
		return "unknown error"
	case 1:
		return e[0].Error()
	default:
		s := strings.Builder{}
		s.WriteString("the following errors occurred: ")
		s.WriteString("[")
		for n, err := range e {
			s.WriteString(err.Error())
			if n < len(e)-1 {
				s.WriteString(", ")
			}
		}
		s.WriteString("]")
		return s.String()
	}
}

// addErr adds an error to an error slice. If errs is not an error slice it will
// be converted into one. If there is already an error with the same message in
// the slice, it will not be added.
func addErr(errs error, err error, prepend bool) error {
	if errs, ok := errs.(rpcErrors); ok {
		msg := err.Error()
		for _, e := range errs {
			if e.Error() == msg {
				return errs
			}
		}
		if prepend {
			return append(rpcErrors{err}, errs...)
		}
		return append(errs, err)
	}
	if errs == nil {
		return rpcErrors{err}
	}
	return addErr(rpcErrors{errs}, err, prepend)
}

// useMostCommon compares all responses returned from RPC endpoints and chooses
// the one that was repeated at least as many times as indicated by the minReq
// arg. Errors in the slice are not counted as responses and will be returned
// as one error if no valid response can be found.
func useMostCommon(s []interface{}, minReq int) (interface{}, error) {
	var err error
	// Count the number of occurrences of each item by comparing each item
	// in the slice with every other item. The result is stored in a map,
	// where the key is the item itself and the value is the number of
	// occurrences.
	maxCount := 0
	counters := map[interface{}]int{}
	for _, a := range s {
		// Errors are handled separately.
		if e, ok := a.(error); ok {
			err = addErr(err, e, false)
			continue
		}
		// Check if there is an item same as the a var already added to
		// the counters map. If so, skip it.
		f := false
		for b, _ := range counters {
			if compare(a, b) {
				f = true
				break
			}
		}
		if f {
			continue
		}
		// Count occurrences of the item.
		for _, b := range s {
			if compare(a, b) {
				counters[a]++
				if counters[a] > maxCount {
					maxCount = counters[a]
				}
			}
		}
	}
	// Check if there are enough occurrences of the most common item.
	if maxCount < minReq {
		err = addErr(err, errors.New("not enough occurrences of the same response from RPC servers"), true)
		return nil, err
	}
	// Find the item with the maximum number of occurrences.
	var res interface{}
	for v, c := range counters {
		if c == maxCount {
			if res != nil {
				// If res is not nil it means, that there are multiple items
				// that occurred maxCount times. In this case, we cannot
				// determine which one should be chosen.
				err = addErr(err, errors.New("RPC servers returned different responses"), true)
				return nil, err
			}
			res = v
			// We do not want to "break" here because we still have to check
			// it there are no more items with the same number of occurrences.
		}
	}
	return res, nil
}

// useMedian calculates the median value of all numberType items in the given
// slice. There must be at least minReq items of type numberType in the slice,
// otherwise an error is returned.
func useMedian(s []interface{}, minReq int) (*numberType, error) {
	// Collect errors from responses.
	var err error
	for _, v := range s {
		if e, ok := v.(error); ok {
			err = addErr(err, e, false)
		}
	}
	// Filter out anything that is not a number.
	s = filter(s, (*numberType)(nil))
	if len(s) < minReq {
		err = addErr(err, errors.New("not enough responses from RPC servers"), true)
		return nil, err
	}
	// Calculate the median.
	sort.Slice(s, func(i, j int) bool {
		return s[i].(*numberType).Big().Cmp(s[j].(*numberType).Big()) < 0
	})
	if len(s)%2 == 0 {
		m := len(s) / 2
		bx := s[m-1].(*numberType).Big()
		by := s[m].(*numberType).Big()
		bm := new(big.Int).Div(new(big.Int).Add(bx, by), big.NewInt(2))
		return (*numberType)(bm), nil
	}
	return s[len(s)/2].(*numberType), nil
}

// useMedianDist works similarly to the useMedian function, but instead of
// median, it will return the lowest value that is greater than or equal to
// median+distance (when distance is negative) and the highest value that is
// less than or equal to median+distance (when distance is positive).
func useMedianDist(s []interface{}, minReq int, distance int64) (*numberType, error) {
	m, err := useMedian(s, minReq)
	if err != nil {
		return nil, err
	}
	s = filter(s, (*numberType)(nil))
	bd := big.NewInt(distance)
	bm := m.Big()
	bx := m.Big()
	for _, n := range s {
		bn := n.(*numberType).Big()
		bs := new(big.Int).Sub(bn, bm)
		if distance < 0 && new(big.Int).Sub(bs, bd).Sign() >= 0 && bn.Cmp(bx) < 0 {
			bx = bn
		} else if distance > 0 && new(big.Int).Sub(bs, bd).Sign() <= 0 && bn.Cmp(bx) > 0 {
			bx = bn
		}
	}
	return (*numberType)(bx), nil
}

// filter returns values from a slice that have the same type as a typ arg.
func filter(s []interface{}, typ interface{}) []interface{} {
	var r []interface{}
	var t = reflect.TypeOf(typ)
	for _, v := range s {
		if reflect.TypeOf(v) == t {
			r = append(r, v)
		}
	}
	return r
}

func isNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}
