package store

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
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

func (s *UserStore) GetUserByLogin(ctx context.Context, serLogin string) (*models.UserModel, error) {
	stmt, args, _ := sq.Select(`"id"`, "login", "pwd_hash").
		From(`"user"`).
		Where(sq.Eq{"login": serLogin}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	row := s.db.QueryRow(ctx, stmt, args...)
	var u models.UserModel
	err := row.Scan(&u.ID, &u.Login, &u.PwdHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &u, nil

}
