package models

type WithdrawReq struct {
	OrderID string  `json:"order"`
	Sum     float32 `json:"sum"`
}

type UserReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
