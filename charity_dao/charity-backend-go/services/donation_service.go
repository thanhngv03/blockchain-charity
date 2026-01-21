package services

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type DonationStat struct {
	TotalEth       string `json:"total_eth"`
	TotalDonations int    `json:"total_donations"`
}

func GetStats() DonationStat {
	// TẠM THỜI MOCK – sau sẽ lấy DB
	return DonationStat{
		TotalEth:       "12.45",
		TotalDonations: 324,
	}
}

func WeiToEth(wei *big.Int) string {
	f := new(big.Float).SetInt(wei)
	eth := new(big.Float).Quo(f, big.NewFloat(1e18))
	return eth.Text('f', 6)
}

func IsZeroAddress(addr common.Address) bool {
	return addr == common.HexToAddress("0x0000000000000000000000000000000000000000")
}
