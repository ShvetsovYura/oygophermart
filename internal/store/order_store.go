package store

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/ShvetsovYura/oygophermart/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderStore struct {
	db *pgxpool.Pool
}

const (
	loyalityOrderTbl = "loyalty_order"
	userTbl          = "user"
)

func NewOrderStore(conn *pgxpool.Pool) (*OrderStore, error) {

	store := &OrderStore{db: conn}
	err := store.Ping(context.TODO())
	if err != nil {
		return nil, err
	}
	return store, nil
}

func (s *OrderStore) Ping(ctx context.Context) error {
	err := s.db.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *OrderStore) GetOrdersById(ctx context.Context, orderId string) ([]models.LoyaltyOrderModel, error) {
	var entities []models.LoyaltyOrderModel
	stmt, args, err := sq.Select("id", "order_id", "type", "status", "value", "user_id", "created_at", "updated_at").
		From(loyalityOrderTbl).
		Where(sq.Eq{"order_id": orderId}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var m models.LoyaltyOrderModel
		rows.Scan(&m.Id, &m.OrderId, &m.Type, &m.Status, &m.Value, &m.UserId, &m.CreatedAt, &m.UpdatedAt)
		entities = append(entities, m)
	}

	return entities, nil
}

func (s *OrderStore) GetUserOrders(ctx context.Context, userLogin string, orderStatus *string, orderType *string) ([]models.LoyaltyOrderModel, error) {
	var entities []models.LoyaltyOrderModel
	query := sq.Select("loyalty_order.id as id", "order_id", "type", "status", "value", "user_id", "created_at", "updated_at").
		From(loyalityOrderTbl).
		Join(`"user" on "user"."id"=loyalty_order.user_id`).
		Where(sq.Eq{"login": userLogin}).
		OrderBy("updated_at desc").
		PlaceholderFormat(sq.Dollar)

	if orderStatus != nil {
		query = query.Where(sq.Eq{"status": *orderStatus})
	}
	if orderType != nil {
		query = query.Where(sq.Eq{"type": *orderType})
	}
	stmt, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var m models.LoyaltyOrderModel
		rows.Scan(&m.Id, &m.OrderId, &m.Type, &m.Status, &m.Value, &m.UserId, &m.CreatedAt, &m.UpdatedAt)
		entities = append(entities, m)
	}

	return entities, nil
}

func (s *OrderStore) AddNewOrder(ctx context.Context, userId int64, orderId string, type_ string, value int64) error {
	stmt, args, _ := sq.Insert("loyalty_order").
		Columns("order_id", "type", "status", "value", "user_id").
		Values(orderId, type_, "NEW", value, userId).
		PlaceholderFormat(sq.Dollar).ToSql()

	_, err := s.db.Exec(ctx, stmt, args...)
	if err != nil {
		return err
	}
	return nil
}

func (s *OrderStore) GetOrdersByIdAndLogin(ctx context.Context, orderId string, login string, type_ string) ([]models.LoyaltyOrderModel, error) {
	var entities []models.LoyaltyOrderModel
	stmt, args, err := sq.Select("id", "order_id", "type", "status", "value", "user_id", "created_at", "updated_at").
		From("loaylty_order lo").
		InnerJoin(`"user" u lo on u."id"=lo.user_id`).
		Where(sq.Eq{
			"lo.order_id": orderId,
			"u.login":     login,
			"lo.type":     type_,
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var m models.LoyaltyOrderModel
		rows.Scan(&m.Id, &m.OrderId, &m.Type, &m.Status, &m.Value, &m.UserId, &m.CreatedAt, &m.UpdatedAt)
		entities = append(entities, m)
	}

	return entities, nil
}

func (s *OrderStore) GetUserBalance(ctx context.Context, login string) models.BalanceModel {
	var m models.BalanceModel
	stmt := `
		with orders as (
			select value, user_id, type
			from loyalty_order lo inner join "user" u on lo.user_id = u."id"
			where u.login = $1 and status = 'PROCESSED'
		)
		select ac.val as accrued, w.val as withdrawal, (ac.val - w.val) as balance
		from (select sum(value) as val  from orders where  type='ACCRUED' group by user_id)  ac,
			 (select sum(value) as val from orders where type='WITHDRAWAL' group by user_id)  w
	`

	s.db.QueryRow(ctx, stmt, login).Scan(&m.Accrued, &m.Withdrawn, &m.Balance)
	return m
}

func (s *OrderStore) Withdraw(ctx context.Context, orderId string, userId int64, value int64) error {
	stmt := `
		insert into loyalty_order(order_id, user_id, "type", status, value)
		values ($1, $2, 'WITHDRAWAL', 'PROCESSED', $3)
	`
	_, err := s.db.Exec(ctx, stmt, orderId, userId, value)
	if err != nil {
		return err
	}
	return nil
}
