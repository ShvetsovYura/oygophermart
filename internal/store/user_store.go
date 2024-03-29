package store

import (
	"context"
	"errors"

	"github.com/ShvetsovYura/oygophermart/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStore struct {
	db *pgxpool.Pool
}

func NewUserStore(db *pgxpool.Pool) (*UserStore, error) {
	return &UserStore{db: db}, nil
}

func (s *UserStore) AddUser(ctx context.Context, login string, pwdHash string) error {
	stmt := `insert into "user"(login, pwd_hash) values($1, $2);`
	_, err := s.db.Exec(ctx, stmt, login, pwdHash)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) GetUserByLogin(ctx context.Context, userLogin string) (*models.UserModel, error) {
	stmt := `
		SELECT
			"id",
			login,
			pwd_hash
		FROM
			"user"
		WHERE
			login = $1
	`

	row := s.db.QueryRow(ctx, stmt, userLogin)

	var u models.UserModel
	err := row.Scan(&u.ID, &u.Login, &u.PwdHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil

}
