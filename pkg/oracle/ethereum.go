package oracle

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Ethereum struct {
	ethClient *ethclient.Client
	wallet    *Wallet
}

func NewEthereum(ethClient *ethclient.Client, wallet *Wallet) *Ethereum {
	return &Ethereum{
		ethClient: ethClient,
		wallet: wallet,
	}
}

func (e *Ethereum) Call(ctx context.Context, address common.Address, data []byte) ([]byte, error) {
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

	return e.ethClient.CallContract(ctx, cm, new(big.Int).SetUint64(bn))
}

func (e *Ethereum) Storage(ctx context.Context, address common.Address, key common.Hash) ([]byte, error) {
	bn, err := e.ethClient.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}

	return e.ethClient.StorageAt(ctx, address, key, new(big.Int).SetUint64(bn))
}

func (e *Ethereum) GetTransaction(ctx context.Context, address common.Address, gasLimit uint64, data []byte) (*types.Transaction, error) {
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

	return signedTx, nil
}

func (e *Ethereum) SendTransaction(ctx context.Context, address common.Address, gasLimit uint64, data []byte) (*common.Hash, error) {
	tx, err := e.GetTransaction(ctx, address, gasLimit, data)
	if err != nil {
		return nil, err
	}

	hash := tx.Hash()
	return &hash, e.ethClient.SendTransaction(ctx, tx)
}
