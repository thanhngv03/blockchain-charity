package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/services"
	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/utils"
)

type DonateRequest struct {
	TxHash string `json:"tx_hash"`
}

func DonationHandler(
	client *ethclient.Client,
	contractAddr common.Address,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var req DonateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", 400)
			return
		}

		donor, amount, err := services.VerifyDonateTx(
			client,
			req.TxHash,
			contractAddr,
		)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		_, err = utils.DB.ExecContext(
			context.Background(),
			`INSERT INTO donations (tx_hash, wallet_address, amount_wei)
			 VALUES ($1,$2,$3)
			 ON CONFLICT (tx_hash) DO NOTHING`,
			req.TxHash,
			donor.Hex(),
			amount.String(),
		)
		if err != nil {
			http.Error(w, "DB error", 500)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"status": "success",
		})
	}
}
