package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/ShvetsovYura/oygophermart/internal/models"
)

var ErrOrderAlreadyAddedByUser = errors.New("the order has already been added by the user")
var ErrOrderAlreadyAddedByAnotherUser = errors.New("the order has already been added by another user")
var ErrInsufficientFunds = errors.New("insufficient funds")

type OrderStorer interface {
	GetUserOrders(ctx context.Context, login string, orderStatus *string, orderType *string) ([]models.LoyaltyOrderModel, error)
	GetOrdersById(ctx context.Context, orderId string) ([]models.LoyaltyOrderModel, error)
	AddNewOrder(ctx context.Context, userId int64, orderId string, type_ string, value int64) error
	GetOrdersByIdAndLogin(ctx context.Context, orderId string, login string, type_ string) ([]models.LoyaltyOrderModel, error)
	GetUserBalance(ctx context.Context, login string) models.BalanceModel
	Withdraw(ctx context.Context, orderId string, userId int64, value int64) error
}

type stores struct {
	orderStore OrderStorer
	userStore  UserStorer
}

type OrderService struct {
	stores stores
}

func NewOrderService(orderStore OrderStorer, userStore UserStorer) *OrderService {
	s := stores{

		orderStore: orderStore,
		userStore:  userStore,
	}
	service := &OrderService{stores: s}
	return service
}

func (s *OrderService) CreateOrder(ctx context.Context, userLogin string, orderId string) error {
	orders, err := s.stores.orderStore.GetOrdersByIdAndLogin(ctx, orderId, userLogin, "ADD")
	if err != nil {
		return err
	}

	if len(orders) > 0 {
		return fmt.Errorf("%w", ErrOrderAlreadyAddedByUser)
	}

	orders, err = s.stores.orderStore.GetOrdersById(ctx, orderId)
	if err != nil {
		return err
	}
	if len(orders) > 0 {
		return fmt.Errorf("%w", ErrOrderAlreadyAddedByAnotherUser)
	}

	user, err := s.stores.userStore.GetUserByLogin(ctx, userLogin)
	if err != nil {
		return err
	}
	err = s.stores.orderStore.AddNewOrder(ctx, user.Id, orderId, "ADD", 1234)
	if err != nil {
		return err
	}
	return nil
}

func (s *OrderService) GetUserOrders(ctx context.Context, userLogin string) ([]models.LoyaltyOrderModel, error) {
	records, err := s.stores.orderStore.GetUserOrders(ctx, userLogin, nil, nil)
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (s *OrderService) GetUserBalance(ctx context.Context, login string) models.BalanceModel {
	record := s.stores.orderStore.GetUserBalance(ctx, login)
	return record
}

func (s *OrderService) Withdraw(ctx context.Context, login string, orderId string, value int64) error {
	user, err := s.stores.userStore.GetUserByLogin(ctx, login)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrInsufficientFunds
	}

	balance := s.stores.orderStore.GetUserBalance(ctx, login)
	if (balance.Balance - value) < 0 {
		return ErrInsufficientFunds
	}

	err = s.stores.orderStore.Withdraw(ctx, orderId, user.Id, value)
	if err != nil {
		return err
	}

	return nil

}

func (s *OrderService) UserWithdrawals(ctx context.Context, login string) ([]models.LoyaltyOrderModel, error) {
	status := "PROCESSED"
	type_ := "WITHDRAWAL"
	orders, err := s.stores.orderStore.GetUserOrders(ctx, login, &status, &type_)
	if err != nil {
		return nil, err
	}

	return orders, nil
}
