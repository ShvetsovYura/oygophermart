package orderservice

import (
	"context"
	"errors"
	"fmt"

	"github.com/ShvetsovYura/oygophermart/internal/store"
)

type OrderService struct {
	store *store.Store
}

func NewOrderService(store *store.Store) *OrderService {
	service := &OrderService{store: store}
	return service
}

var ErrOrderAlreadyAddedByUser = errors.New("the order has already been added by the user")
var ErrOrderAlreadyAddedByAnotherUser = errors.New("the order has already been added by another user")

func (s *OrderService) CreateOrder(ctx context.Context, userLogin string, orderId string) error {
	orders, err := s.store.GetUserOrders(userLogin)
	if err != nil {
		return err
	}
	if len(orders) > 0 {
		return fmt.Errorf("%w", ErrOrderAlreadyAddedByUser)
	}
	orders, err = s.store.GetOrdersById(orderId)
	if err != nil {
		return err
	}
	if len(orders) > 0 {
		return fmt.Errorf("%w", ErrOrderAlreadyAddedByAnotherUser)
	}

	user, err := s.store.GetUserByLogin(ctx, userLogin)
	if err != nil {
		return err
	}

	err = s.store.AddNewOrder(ctx, user.Id, orderId, "ADD", 1234)
	return nil
}
