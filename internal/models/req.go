package models

type WithdrawReq struct {
	OrderId string `json:"order"`
	Sum     int64  `json:"sum"`
}
