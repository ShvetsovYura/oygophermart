package models

import (
	"encoding/json"
	"time"
)

type OrderGroupedModel struct {
	ID        string    `json:"number"`
	Status    string    `json:"status"`
	Accrual   *float64  `json:"accrual,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (m OrderModel) MarshalJSON() ([]byte, error) {
	type OrderModelAlias OrderModel
	value := struct {
		OrderModelAlias
		UpdatedAt string `json:"updated_at"`
	}{
		OrderModelAlias: OrderModelAlias(m),
		UpdatedAt:       m.UpdatedAt.Format(time.RFC3339),
	}
	return json.Marshal(value)
}
