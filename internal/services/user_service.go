package services

import (
	"context"
	"errors"

	"github.com/ShvetsovYura/oygophermart/internal/models"
)

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrNotValidLoginOrPassword = errors.New("not valid login/password")
var ErrUserNotFound = errors.New("user not found")

type UserStorer interface {
	AddUser(ctx context.Context, login string, pwdHash string) error
	GetUserByLogin(ctx context.Context, userLogin string) (*models.UserModel, error)
}

type Hasher interface {
	Hash(val string) string
}

type UserServcie struct {
	store   UserStorer
	hashSvc Hasher
}

func NewUserService(store UserStorer, hash Hasher) *UserServcie {
	return &UserServcie{store: store, hashSvc: hash}
}

func (u *UserServcie) CreateUser(ctx context.Context, login string, password string) (int64, error) {
	user, err := u.store.GetUserByLogin(ctx, login)
	if err != nil {
		return 0, err
	}
	if user != nil {
		return 0, ErrUserAlreadyExists
	}

	hashPwd := u.hashSvc.Hash(password)
	err = u.store.AddUser(ctx, login, hashPwd)
	if err != nil {
		return 0, err
	}
	m, err := u.store.GetUserByLogin(ctx, login)
	if err != nil {
		return 0, err
	}
	return m.Id, nil
}

func (u *UserServcie) Login(ctx context.Context, login string, password string) (int64, error) {
	user, err := u.store.GetUserByLogin(ctx, login)
	if err != nil {
		return 0, err
	}
	if user == nil {
		return 0, ErrUserNotFound
	}

	pwdHash := u.hashSvc.Hash(password)
	if pwdHash != user.PwdHash {
		return 0, ErrNotValidLoginOrPassword
	}
	return user.Id, nil
}
