package contracts

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type CharityVault struct{}

func DeployCharityVault(
	auth *bind.TransactOpts,
	backend bind.ContractBackend,
	target *big.Int,
) (common.Address, *types.Transaction, *CharityVault, error) {
	return common.Address{}, nil, &CharityVault{}, fmt.Errorf("DeployCharityVault is not yet implemented")
}

