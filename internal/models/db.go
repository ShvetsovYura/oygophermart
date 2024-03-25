package models

import "time"

type OrderModel struct {
	Id         string
	UserId     uint64
	Status     string
	CreateedAt time.Time
	UpdatedAt  time.Time
}

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

type OrderGroupedModel struct {
	Id        string    `json:"number"`
	Status    string    `json:"status"`
	Accrual   *float64  `json:"accrual,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserModel struct {
	Id      int64
	Login   string
	PwdHash string
}

type BalanceModel struct {
	Accrued   float64
	Withdrawn float64
	Balance   float64
}
