package ethereum

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"regexp"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type Client struct {
	ethClient *ethclient.Client
	wallet    *Wallet
}

func NewClient(ethClient *ethclient.Client, wallet *Wallet) *Client {
	return &Client{
		ethClient: ethClient,
		wallet:    wallet,
	}
}

func (e *Client) Call(ctx context.Context, address common.Address, data []byte) ([]byte, error) {
	bn, err := e.ethClient.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}

	cm := ethereum.CallMsg{
		From:     e.wallet.Address(),
		To:       &address,
		Gas:      0,
		GasPrice: nil,
		Value:    nil,
		Data:     data,
	}

	resp, err := e.ethClient.CallContract(ctx, cm, new(big.Int).SetUint64(bn))
	if err != nil {
		return nil, parseError(err)
	}

	return resp, err
}

func (e *Client) Storage(ctx context.Context, address common.Address, key common.Hash) ([]byte, error) {
	bn, err := e.ethClient.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}

	return e.ethClient.StorageAt(ctx, address, key, new(big.Int).SetUint64(bn))
}

func (e *Client) SendTransaction(ctx context.Context, address common.Address, gasLimit uint64, data []byte) (*common.Hash, error) {
	nonce, err := e.ethClient.PendingNonceAt(ctx, e.wallet.Address())
	if err != nil {
		return nil, err
	}

	gas, err := e.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	tx := types.NewTransaction(
		nonce,
		address,
		nil,
		gasLimit,
		gas,
		data,
	)

	chainID, err := e.ethClient.NetworkID(ctx)
	if err != nil {
		return nil, err
	}

	signedTx, err := e.wallet.EthWallet().SignTx(*e.wallet.EthAccount(), tx, chainID)
	if err != nil {
		return nil, err
	}

	hash := signedTx.Hash()

	return &hash, e.ethClient.SendTransaction(ctx, signedTx)
}

func parseError(vmErr error) error {
	switch terr := vmErr.(type) {
	// rpc.jsonError:
	case rpc.DataError:
		// Some RPC servers returns "revert" data as a hex encoded string, here
		// we're trying to parse it. If any error occurs during it, then original
		// error will be returned.
		if str, ok := terr.ErrorData().(string); ok {
			re := regexp.MustCompile("(0x[a-zA-Z0-9]+)")
			match := re.FindStringSubmatch(str)

			if len(match) == 2 && len(match[1]) > 2 {
				bytes, err := hex.DecodeString(match[1][2:])
				if err != nil {
					return vmErr
				}

				revert, err := abi.UnpackRevert(bytes)
				if err != nil {
					return vmErr
				}

				return errors.New(revert)
			}
		}
	}

	return vmErr
}
