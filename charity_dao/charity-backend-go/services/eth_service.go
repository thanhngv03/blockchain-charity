package services

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func VerifyDonateTx(
	client *ethclient.Client,
	txHash string,
	expectedContract common.Address,
) (common.Address, *big.Int, error) {

	ctx := context.Background()
	hash := common.HexToHash(txHash)

	// 1. Lấy transaction
	tx, ísPending, err := client.TransactionByHash(ctx, hash)
	if err != nil {
		return common.Address{}, nil, err
	}
	if ísPending {
		return common.Address{}, nil, errors.New("transaction is still pending")
	}

	// 2. Lấy receipt
	receipt, err := client.TransactionReceipt(ctx, hash)
	if err != nil {
		return common.Address{}, nil, err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return common.Address{}, nil, errors.New("receipt status is not successful")
	}

	// Verify gui toi CharityVault
	if tx.To() == nil || *tx.To() != expectedContract {
		return common.Address{}, nil, errors.New("tx not sent to CharityVault")
	}

	// Lay amount
	amount := tx.Value()
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return common.Address{}, nil, errors.New("donation amount must be greater than zero")
	}

	// 5. Lấy donor (from)
	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return common.Address{}, nil, err
	}

	signer := types.LatestSignerForChainID(chainID)
	from, err := types.Sender(signer, tx)
	if err != nil {
		return common.Address{}, nil, err
	}

	return from, amount, nil
}
