package services

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type DonationTx struct {
	TxHash      string
	From        string
	AmountWei   *big.Int
	BlockNumber uint64
}

func VerifyDonationTx(txHash string, contractAddr common.Address) (*DonationTx, error) {

	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	tx, isPending, err := client.TransactionByHash(ctx, common.HexToHash(txHash))
	if err != nil {
		return nil, err
	}
	if isPending {
		return nil, errors.New("transaction pending")
	}

	receipt, err := client.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return nil, err
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, errors.New("transaction failed")
	}

	// 1️⃣ Check gửi đến CharityVault
	if tx.To() == nil || *tx.To() != contractAddr {
		return nil, errors.New("transaction not sent to CharityVault")
	}

	// 2️⃣ Lấy chainID để recover sender
	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}

	signer := types.LatestSignerForChainID(chainID)

	from, err := types.Sender(signer, tx)
	if err != nil {
		return nil, err
	}

	return &DonationTx{
		TxHash:      txHash,
		From:        from.Hex(),
		AmountWei:   tx.Value(),
		BlockNumber: receipt.BlockNumber.Uint64(),
	}, nil
}
