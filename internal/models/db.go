package models

import "time"

type LoyaltyOrderModel struct {
	Id        int
	OrderId   string
	Status    string
	Value     int64
	Type      string
	UserId    int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserModel struct {
	Id      int64
	Login   string
	PwdHash string
}

type BalanceModel struct {
	Accrued   int64
	Withdrawn int64
	Balance   int64
}
