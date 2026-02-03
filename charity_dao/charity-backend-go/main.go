package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/services"
	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/utils"
)

// ===== Struct đọc JSON deploy =====
type DeploymentInfo struct {
	CharityVault string `json:"CharityVault"`
}

func main() {

	// 1. Connect PostgreSQL
	utils.ConnectDB()

	// 2. Load contract address
	data, err := os.ReadFile("../deployments/deployed-address.json")
	if err != nil {
		log.Fatal("Cannot read deploy info:", err)
	}

	var deployment DeploymentInfo
	if err := json.Unmarshal(data, &deployment); err != nil {
		log.Fatal("Invalid deploy JSON:", err)
	}

	expectedContract := common.HexToAddress(deployment.CharityVault)

	// 3. Connect Ethereum client
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatal("Cannot connect eth node:", err)
	}

	// 4. Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 5. Donate verify endpoint
	http.HandleFunc("/api/donations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			Wallet string `json:"wallet_address"`
			TxHash string `json:"tx_hash"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		donor, amount, err := services.VerifyDonateTx(
			client,
			body.TxHash,
			expectedContract,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Save DB
		_, err = utils.DB.ExecContext(
			context.Background(),
			`INSERT INTO donations (tx_hash, wallet_address, amount_wei)
			VALUES ($1, $2, $3)
			ON CONFLICT (tx_hash) DO NOTHING`,
			body.TxHash,
			donor.Hex(),
			amount.String(),
		)

		if err != nil {
			http.Error(w, "Database error", 500)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"status": "success",
		})
	})

	// 6. Start server
	fmt.Println("Backend running at http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
