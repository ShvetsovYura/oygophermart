package services

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"github.com/ShvetsovYura/oygophermart/internal/models"
// 	"github.com/ShvetsovYura/oygophermart/mocks"
// 	"github.com/golang/mock/gomock"
// 	"github.com/stretchr/testify/assert"
// )

// func TestCreateOrder(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	m := mocks.NewMockOrderStorer(ctrl)
// 	ctx := context.TODO()
// 	userOrders := []models.LoyaltyOrderModel{
// 		{
// 			Id:        1,
// 			OrderId:   "123456",
// 			Status:    "NEW",
// 			Value:     123,
// 			Type:      "ADD",
// 			UserId:    321,
// 			CreatedAt: time.Now(),
// 			UpdatedAt: time.Now(),
// 		},
// 	}
// 	m.EXPECT().GetUserOrders(ctx, "pipa").Return(userOrders, nil)

// 	m.EXPECT().GetOrdersById(ctx, "123456").Return(userOrders, nil)
// 	s := NewOrderService(m)
// 	err := s.CreateOrder(ctx, "pipa", "123456")

// 	if assert.Error(t, err) {
// 		assert.Equal(t, ErrOrderAlreadyAddedByUser, err)
// 	}
// }
