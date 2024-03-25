package services

import (
	"context"
	"errors"

	"github.com/ShvetsovYura/oygophermart/internal/logger"
	"github.com/ShvetsovYura/oygophermart/internal/models"
)

var ErrOrderAlreadyAddedByUser = errors.New("the order has already been added by the user")
var ErrOrderAlreadyAddedByAnotherUser = errors.New("the order has already been added by another user")
var ErrInsufficientFunds = errors.New("insufficient funds")

type OrderStorer interface {
	GetUserOrders(ctx context.Context, userID uint64) ([]models.OrderGroupedModel, error)
	GetOrdersByID(ctx context.Context, orderID string) ([]models.OrderModel, error)
	AddNewOrder(ctx context.Context, userID int64, orderID string) error
	GetUserOrderByID(ctx context.Context, orderID string, userID int64) (*models.LoyaltyOrderModel, error)
	GetUserBalance(ctx context.Context, userID uint64) models.BalanceModel
	Withdraw(ctx context.Context, orderID string, userID int64, value float64) error
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

func (s *OrderService) CreateOrder(ctx context.Context, userID uint64, orderID string) error {
	records, err := s.stores.orderStore.GetOrdersByID(ctx, orderID)
	if err != nil {
		return err
	}
	if len(records) > 0 {
		for _, r := range records {
			if r.UserId == userID {
				return ErrOrderAlreadyAddedByUser
			}
		}
		return ErrOrderAlreadyAddedByAnotherUser
	}

	err = s.stores.orderStore.AddNewOrder(ctx, int64(userID), orderID)
	if err != nil {
		return err
	}
	return nil

}

func (s *OrderService) GetUserOrders(ctx context.Context, userID uint64) ([]models.OrderGroupedModel, error) {
	records, err := s.stores.orderStore.GetUserOrders(ctx, userID)
	var result = make([]models.OrderGroupedModel, 0, len(records))

	if err != nil {
		return nil, err
	}
	for _, r := range records {
		if r.Accrual == nil {
			result = append(result, r)
		} else {
			if *r.Accrual >= 0.0 {
				result = append(result, r)
			}
		}
	}
	return result, nil
}

func (s *OrderService) GetUserBalance(ctx context.Context, userID uint64) models.BalanceModel {
	record := s.stores.orderStore.GetUserBalance(ctx, userID)
	return record
}

func (s *OrderService) Withdraw(ctx context.Context, userID uint64, orderID string, value float64) error {
	balance := s.stores.orderStore.GetUserBalance(ctx, userID)
	logger.Log.Debugf("withdraw balance: user %s %s %v", userID, orderID, balance)
	if (balance.Balance - value) < 0 {
		logger.Log.Debug("User not funds")
		return ErrInsufficientFunds
	}

	err := s.stores.orderStore.Withdraw(ctx, orderID, int64(userID), value)
	if err != nil {
		logger.Log.Debugf("err on withdrsw %e", err)
		return err
	}

	return nil

}

func (s *OrderService) UserWithdrawals(ctx context.Context, userID uint64) ([]models.OrderGroupedModel, error) {
	status := "PROCESSED"
	var result = make([]models.OrderGroupedModel, 0)
	orders, err := s.stores.orderStore.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, r := range orders {
		if r.Status == status {
			if r.Accrual != nil {
				if *r.Accrual < 0 {
					result = append(result, r)
				}
			}
		}
	}

	return result, nil
}
