package store

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/ShvetsovYura/oygophermart/internal/logger"
	"github.com/ShvetsovYura/oygophermart/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrOrdersNotFoundInDB = errors.New("orders not found")
var ErrOrderAlreadyExistsInDB = errors.New("order already exists")

const (
	UniqueViolation = "23505"
)

type OrderStore struct {
	db *pgxpool.Pool
}

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

func (s *OrderStore) GetOrdersByID(ctx context.Context, orderID string) ([]models.OrderModel, error) {
	var entities = make([]models.OrderModel, 0)
	stmt := `
		select id, status, user_id, created_at, updated_at 
		from "order" 
		where id = $1
	`

	rows, err := s.db.Query(ctx, stmt, orderID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var m models.OrderModel
		rows.Scan(&m.Id, &m.Status, &m.UserID, &m.CreateedAt, &m.UpdatedAt)
		entities = append(entities, m)
	}

	return entities, nil
}

func (s *OrderStore) GetUserOrders(ctx context.Context, userID uint64) ([]models.OrderGroupedModel, error) {
	var entities []models.OrderGroupedModel

	stmt := `
	SELECT
		O.ID,
		O.STATUS,
		O.UPDATED_AT,
		SUM("value") as val
	FROM
		"order" O
		LEFT JOIN LOYALTY L ON O.ID = L.ORDER_ID
	WHERE
		O.USER_ID = $1
	GROUP BY
		O.ID,
		O.STATUS,
		O.UPDATED_AT
	ORDER BY
		UPDATED_AT ASC;
	`

	rows, err := s.db.Query(ctx, stmt, userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var m models.OrderGroupedModel
		rows.Scan(&m.ID, &m.Status, &m.UpdatedAt, &m.Accrual)
		entities = append(entities, m)
	}

	return entities, nil
}

func (s *OrderStore) AddNewOrder(ctx context.Context, userID int64, orderID string) error {
	stmt, args, _ := sq.Insert(`"order"`).
		Columns("id", "status", "user_id").
		Values(orderID, "NEW", userID).
		PlaceholderFormat(sq.Dollar).ToSql()

	_, err := s.db.Exec(ctx, stmt, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == UniqueViolation {
				return ErrOrderAlreadyExistsInDB
			}
		}
		return err
	}
	return nil
}

func (s *OrderStore) GetUserOrderByID(ctx context.Context, orderID string, userID int64) (*models.LoyaltyOrderModel, error) {
	var m models.LoyaltyOrderModel
	stmt := `
	SELECT
		ID,
		STATUS,
		CREATED_AT,
		UPDATED_AT
	FROM
		"order" O
	WHERE
		O.USER_ID = $1
		AND O.ID = $2;
	`

	row := s.db.QueryRow(ctx, stmt, userID, orderID)
	err := row.Scan(&m.Id, &m.Status, &m.Value, &m.UserID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrdersNotFoundInDB
		}
		return nil, err
	}

	return &m, nil
}

func (s *OrderStore) GetUserBalance(ctx context.Context, userID uint64) models.BalanceModel {
	var m models.BalanceModel
	stmt := `
	SELECT
		-- U.ID AS UID,
		COALESCE(WD, 0) AS ACCRUED,
		COALESCE(WD1, 0) AS WITHDRAWN,
		CAST(COALESCE(WD, 0.0) + COALESCE(WD1, 0.0) as numeric(10, 4)) AS BALANCE
	FROM
		"user" U
		LEFT JOIN (
			SELECT
				O.USER_ID AS UID,
				SUM(L."value") AS WD
			FROM
				"order" O
				INNER JOIN LOYALTY L ON O."id" = L.ORDER_ID
			WHERE
				O.STATUS = 'PROCESSED'
				AND L."value" > 0
			GROUP BY
				O.USER_ID
		) W ON U."id" = W.UID
		LEFT JOIN (
			SELECT
				O.USER_ID AS UID,
				SUM(L."value") AS WD1
			FROM
				"order" O
				INNER JOIN LOYALTY L ON O."id" = L.ORDER_ID
			WHERE
				O.STATUS = 'PROCESSED'
				AND L."value" < 0
			GROUP BY
				O.USER_ID
		) W1 ON U."id" = W1.UID
	WHERE
		U.ID = $1;
	`
	s.db.QueryRow(ctx, stmt, userID).Scan(&m.Accrued, &m.Withdrawn, &m.Balance)
	logger.Log.Debugf("user %s balance %v", userID, m)
	return m
}

func (s *OrderStore) Withdraw(ctx context.Context, orderID string, userID int64, value float64) error {
	insertOrderStmt := `
		insert into "order"(id, user_id, status)
		values ($1, $2, $3);
	`

	insertLoyaltyStmt := `
		insert into loyalty(order_id, value)
		values ($1, $2)
	`
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Log.Debugf("error begintx withdraw: %e", err)

		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, insertOrderStmt, orderID, userID, "PROCESSED")
	if err != nil {
		// tx.Rollback(ctx)
		logger.Log.Debugf("error on exec insert order withdraw: %e", err)
		return ErrOrderAlreadyExistsInDB
	}
	_, err = tx.Exec(ctx, insertLoyaltyStmt, orderID, -1*value)
	if err != nil {
		logger.Log.Debugf("error on exec insert loyalty withdraw: %e", err)
		tx.Rollback(ctx)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		logger.Log.Debugf("eror on commit withdraw: %e", err)
		// tx.Rollback(ctx)
		return err
	}
	return nil
}

func (s *OrderStore) UpdateOrdersStatus(ctx context.Context, processRecords ...models.AccrualResult) error {
	statusMap := map[string]string{
		"REGISTERED": "PROCESSING",
		"PROCESSING": "PROCESSING",
		"INVALID":    "INVALID",
		"PROCESSED":  "PROCESSED",
	}
	stmtUpdOrder := `update "order" set status = $1 where id = $2`
	stmtInsLyalty := `insert into loyalty (order_id, value) values($1, $2) `
	tx, err := s.db.Begin(ctx)
	if err != nil {
		logger.Log.Debug("erro on tx")
		return err
	}
	defer tx.Rollback(ctx)
	logger.Log.Debugf("Orderw to write in DB %v", len(processRecords))
	// b := pgx.Batch{}
	for _, inRec := range processRecords {
		if status, ok := statusMap[inRec.Status]; ok {
			logger.Log.Debugf("New status: %s", status)
			// b.Queue(stmtUpdOrder, status, inRecorder_id.OrderId)
			// for _, o := range orders {
			// if o.Status != inRec.Status {
			tx.Exec(ctx, stmtUpdOrder, status, inRec.OrderId)
			if inRec.Accrual != nil {
				// b.Queue(stmtInsLyalty, inRec.OrderId, inRec.Accrual)
				tx.Exec(ctx, stmtInsLyalty, inRec.OrderId, *inRec.Accrual)
			}
			// }

			// }
		}
	}
	// tx.SendBatch(ctx, &b)
	err = tx.Commit(ctx)
	if err != nil {
		logger.Log.Debugf("err commint: %e", err)
		tx.Rollback(ctx)
		return err
	}
	return nil
}

func (s *OrderStore) GetOrdersToAccrualProcess(ctx context.Context) ([]models.OrderModel, error) {
	var records = make([]models.OrderModel, 0, 10)
	stmt := `
		select id, user_id, status, created_at, updated_at 
		from "order" o 
		where o.status not in ('INVALID', 'PROCESSED')
	`

	rows, err := s.db.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var m models.OrderModel
		rows.Scan(&m.Id, &m.UserID, &m.Status, &m.CreateedAt, &m.UpdatedAt)
		records = append(records, m)
	}
	return records, nil
}
