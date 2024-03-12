package models

type WithdrawReq struct {
	OrderId string `json:"order"`
	Sum     int64  `json:"sum"`
}

type UserReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
