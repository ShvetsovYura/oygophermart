package models

import "time"

type BalanceResp struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type UserWithdrawalsResp struct {
	OrderId     string    `json:"order"`
	Sum         int64     `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
