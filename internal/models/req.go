package models

type WithdrawReq struct {
	OrderId string  `json:"order"`
	Sum     float32 `json:"sum"`
}

type UserReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
