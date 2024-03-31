package models

import "time"

type OrderModel struct {
	ID         string
	UserID     uint64
	Status     string
	CreateedAt time.Time
	UpdatedAt  time.Time
}

type LoyaltyOrderModel struct {
	ID        int
	OrderID   string
	Status    string
	Value     int64
	Type      string
	UserID    int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserModel struct {
	ID      int64
	Login   string
	PwdHash string
}

type BalanceModel struct {
	Accrued   float64
	Withdrawn float64
	Balance   float64
}
