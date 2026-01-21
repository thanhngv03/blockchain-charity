// Sau này chạy để đọc contract: go run scripts/read_contract.go

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/contract"
)

// Cấu trúc đọc JSON
type DeploymentInfo struct {
	CharityVault string `json:"CharityVault"`
	Network      string `json:"network"`
	Deployer     string `json:"deployer"`
}

func loadDeploymentInfo() DeploymentInfo {
	data, err := os.ReadFile("../deployments/deployed-address.json")
	if err != nil {
		log.Fatalf("Không đọc được file deployed-address.json: %v", err)
	}

	var info DeploymentInfo
	if err := json.Unmarshal(data, &info); err != nil {
		log.Fatalf("Không parse được JSON: %v", err)
	}
	return info
}

func main() {

	// Đọc file JSON
	data, err := os.ReadFile("../deployments/deployed-address.json")
	if err != nil {
		log.Fatalf("Không đọc được file JSON: %v", err)
	}

	var deployment DeploymentInfo
	if err := json.Unmarshal(data, &deployment); err != nil {
		log.Fatalf("Không parse được JSON: %v", err)
	}

	fmt.Println("Địa chỉ contract:", deployment.CharityVault)

	// Kết nối đến Hardhat node
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatal("Lỗi kết nối Ethereum client:", err)
	}
	defer client.Close()

	// Tạo instance contract
	contractAddr := common.HexToAddress(deployment.CharityVault)
	vault, err := contract.NewCharityVault(contractAddr, client)
	if err != nil {
		log.Fatal("Không load contract:", err)
	}

	// Gọi hàm view trong contract
	total, err := vault.TotalDonated(&bind.CallOpts{Context: context.Background()})
	if err != nil {
		log.Fatal("Lỗi khi gọi totalDonated:", err)
	}

	fmt.Println("Tổng tiền donate hiện tại:", total.String(), "wei")

	ethValue := new(big.Float).Quo(new(big.Float).SetInt(total), big.NewFloat(1e18))
	fmt.Println("Tổng tiền donate (ETH):", ethValue)
}
