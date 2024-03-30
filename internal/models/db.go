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

type OrderGroupedModel struct {
	ID        string    `json:"number"`
	Status    string    `json:"status"`
	Accrual   *float64  `json:"accrual,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
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
