package services

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/contracts"
)

func DeployProjectContract(targetWei string) (common.Address, error) {

	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		return common.Address{}, err
	}

	privateKey, err := crypto.HexToECDSA(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		return common.Address{}, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, _ := client.PendingNonceAt(context.Background(), fromAddress)

	gasPrice, _ := client.SuggestGasPrice(context.Background())

	auth, _ := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(31337))

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	target := new(big.Int)
	target.SetString(targetWei, 10)

	address, _, _, err := contracts.DeployCharityVault(
		auth,
		client,
		target,
	)

	return address, err
}
