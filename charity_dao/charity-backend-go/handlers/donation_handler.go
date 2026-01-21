package handlers

import (
	"encoding/json"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/db"
	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/services"
)

type DonateRequest struct {
	TxHash string `json:"txHash"`
}

func CreateDonation(w http.ResponseWriter, r *http.Request) {

	var req DonateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", 400)
		return
	}

	contractAddr := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")

	tx, err := services.VerifyDonationTx(req.TxHash, contractAddr)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	ethValue := new(big.Float).Quo(
		new(big.Float).SetInt(tx.AmountWei),
		big.NewFloat(1e18),
	)

	_, err = db.DB.Exec(`
		INSERT INTO donations (tx_hash, donor_address, amount_wei, amount_eth, block_number)
		VALUES ($1,$2,$3,$4,$5)
	`,
		tx.TxHash,
		tx.From,
		tx.AmountWei.String(),
		ethValue.String(),
		tx.BlockNumber,
	)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"ok"}`))
}
