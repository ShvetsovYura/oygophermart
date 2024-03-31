package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ShvetsovYura/oygophermart/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestOrderModel(t *testing.T) {
	m := models.OrderGroupedModel{
		ID:        "123456",
		Accrual:   nil,
		Status:    "new",
		UpdatedAt: time.Date(2024, time.March, 10, 22, 30, 12, 0, time.Now().Location()),
	}

	j, e := json.Marshal(m)
	assert.NoError(t, e)
	assert.Equal(t, `{"number":"123456","status":"new","updated_at":"2024-03-10T22:30:12+03:00"}`, string(j))
}
