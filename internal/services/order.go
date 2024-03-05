package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/ShvetsovYura/oygophermart/internal/models"
)

var ErrOrderAlreadyAddedByUser = errors.New("the order has already been added by the user")
var ErrOrderAlreadyAddedByAnotherUser = errors.New("the order has already been added by another user")

type OrderStorer interface {
	GetUserByLogin(ctx context.Context, userLogin string) (*models.UserModel, error)
	GetUserOrders(ctx context.Context, login string) ([]models.LoyaltyOrderModel, error)
	GetOrdersById(ctx context.Context, orderId string) ([]models.LoyaltyOrderModel, error)
	AddNewOrder(ctx context.Context, userId int64, orderId string, type_ string, value int64) error
	GetOrdersByIdAndLogin(ctx context.Context, orderId string, login string, type_ string) ([]models.LoyaltyOrderModel, error)
	GetUserBalance(ctx context.Context, login string) models.BalanceModel
	Withdraw(ctx context.Context, orderId string, userId int64, value int64) error
}

type OrderService struct {
	store OrderStorer
}

func NewOrderService(store OrderStorer) *OrderService {
	service := &OrderService{store: store}
	return service
}

func (s *OrderService) CreateOrder(ctx context.Context, userLogin string, orderId string) error {
	orders, err := s.store.GetOrdersByIdAndLogin(ctx, orderId, userLogin, "ADD")
	if err != nil {
		return err
	}

	if len(orders) > 0 {
		return fmt.Errorf("%w", ErrOrderAlreadyAddedByUser)
	}

	orders, err = s.store.GetOrdersById(ctx, orderId)
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
	if err != nil {
		return err
	}
	return nil
}

func (s *OrderService) GetUserOrders(ctx context.Context, userLogin string) ([]models.LoyaltyOrderModel, error) {
	records, err := s.store.GetUserOrders(ctx, userLogin)
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (s *OrderService) GetUserBalance(ctx context.Context, login string) models.BalanceModel {
	record := s.store.GetUserBalance(ctx, login)
	return record
}

func (s *OrderService) Withdraw(ctx context.Context, login string, orderId string, value int64) error {
	user, err := s.store.GetUserByLogin(ctx, login)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("User not found")
	}

	balance := s.store.GetUserBalance(ctx, login)
	if (balance.Balance - value) < 0 {
		return fmt.Errorf("insufficient funds")
	}

	err = s.store.Withdraw(ctx, orderId, user.Id, value)
	if err != nil {
		return err
	}

	return nil

}
