package store

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LoyaltyOrder struct {
	id       int
	order_id string
}

type User struct {
	Id      int64
	Login   string
	PwdHash string
}

type Store struct {
	db *pgxpool.Pool
}

const (
	loyalityOrderTbl = "loyalty_order"
	userTbl          = "user"
)

func NewStore(connString string) (*Store, error) {
	conn, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	store := &Store{
		db: conn,
	}
	err = store.Ping(context.TODO())
	if err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) Ping(ctx context.Context) error {
	err := s.db.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetOrdersById(orderId string) ([]LoyaltyOrder, error) {
	var entities []LoyaltyOrder
	stmt, args, err := sq.Select("*").
		From(loyalityOrderTbl).
		Where(sq.Eq{"order_id": orderId}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(context.TODO(), stmt, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var m LoyaltyOrder
		rows.Scan(&m.id, &m.order_id)
		entities = append(entities, m)
	}

	return entities, nil
}

func (s *Store) GetUserOrders(userLogin string) ([]LoyaltyOrder, error) {
	var entities []LoyaltyOrder
	stmt, args, err := sq.Select("*").
		From(loyalityOrderTbl).
		Join(`"user" on "user"."id"=loyalty_order.user_id`).
		Where(sq.Eq{"login": userLogin}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(context.TODO(), stmt, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var m LoyaltyOrder
		rows.Scan(&m.id, &m.order_id)
		entities = append(entities, m)
	}

	return entities, nil
}

func (s *Store) GetUserByLogin(ctx context.Context, userLogin string) (*User, error) {
	stmt, args, _ := sq.Select(`"id"`, "login", "pwd_hash").
		From(`"user"`).
		Where(sq.Eq{"login": userLogin}).
		PlaceholderFormat(sq.Dollar).ToSql()
	row := s.db.QueryRow(ctx, stmt, args...)
	var u User
	err := row.Scan(&u.Id, &u.Login, &u.PwdHash)
	if err != nil {
		return nil, err
	}
	return &u, nil

}
func (s *Store) AddNewOrder(ctx context.Context, userId int64, orderId string, type_ string, value int64) error {
	stmt, args, _ := sq.Insert("loyalty_order").Columns("order_id", "type", "status", "value", "user_id").
		Values(orderId, type_, "NEW", value, userId).
		PlaceholderFormat(sq.Dollar).ToSql()

	_, err := s.db.Exec(ctx, stmt, args...)
	if err != nil {
		return err
	}
	return nil
}
