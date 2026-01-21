package models

import "time"

type Donation struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	WalletAddress string    `json:"wallet_address"`
	TxHash        string    `json:"tx_hash"`
	AmountWei     string    `json:"amount_wei"`
	BlockNumber   uint64    `json:"block_number"`
	DonatedAt     time.Time `json:"donated_at"`
	CreatedAt     time.Time `json:"created_at"`
}
