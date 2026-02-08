package models

import "time"

type Project struct {
	ID                 int64     `json:"id"`
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	TargetAmountWei    string    `json:"target_amount_wei"`
	CollectedAmountWei string    `json:"collected_amount_wei"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
