package router

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/ShvetsovYura/oygophermart/internal/models"
	"github.com/ShvetsovYura/oygophermart/internal/services"
	"github.com/go-chi/chi/v5"
)

type OrderWorker interface {
	CreateOrder(ctx context.Context, userLogin string, orderId string) error
	GetUserBalance(ctx context.Context, login string) models.BalanceModel
	Withdraw(ctx context.Context, login string, orderId string, value int64) error
	UserWithdrawals(ctx context.Context, login string) ([]models.LoyaltyOrderModel, error)
}

type UserWorker interface {
	CreateUser(ctx context.Context, login string, password string) error
	Login(ctx context.Context, login string, password string) error
}

type Tokener interface {
	GetToken() (string, error)
}

type HTTPRouter struct {
	orderService OrderWorker
	userService  UserWorker
	tokenService Tokener
	rawRouter    *chi.Mux
}

func NewHTTPRouter(orderService OrderWorker, userService UserWorker, tokenService Tokener) *HTTPRouter {
	api := &HTTPRouter{
		orderService: orderService,
		userService:  userService,
		tokenService: tokenService,
	}
	return api
}

func (wa *HTTPRouter) GetRouter() *chi.Mux {
	return wa.rawRouter
}

func (wa *HTTPRouter) InitRouter() {
	r := chi.NewRouter()
	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", wa.userRegister)
			r.Post("/login", wa.userLogin)
			r.Post("/orders", wa.userLoadOrders)
			r.Get("/orders", wa.userListOrders)
			r.Get("/balance", wa.userBalance)
			r.Post("/balance/withdraw", wa.userWithdrawBalance)
			r.Get("/withdrawals", wa.userWithdrawals)
		})
	})
	wa.rawRouter = r
}

func (wa *HTTPRouter) userRegister(w http.ResponseWriter, r *http.Request) {
	var user models.UserReq
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &user)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = wa.userService.CreateUser(r.Context(), user.Login, user.Password)
	if err != nil {
		if errors.Is(err, services.ErrUserAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	token, err := wa.tokenService.GetToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	c := http.Cookie{Name: "token", Value: token, HttpOnly: true, MaxAge: 3600}
	http.SetCookie(w, &c)
	w.WriteHeader(http.StatusOK)
}

func (wa *HTTPRouter) userLogin(w http.ResponseWriter, r *http.Request) {
	var user models.UserReq

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = wa.userService.Login(r.Context(), user.Login, user.Password)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) || errors.Is(err, services.ErrNotValidLoginOrPassword) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	token, err := wa.tokenService.GetToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	c := http.Cookie{
		Name:     "token",
		Value:    token,
		MaxAge:   3600,
		HttpOnly: true,
	}

	http.SetCookie(w, &c)
	w.WriteHeader(http.StatusOK)
}

func (wa *HTTPRouter) userLoadOrders(w http.ResponseWriter, r *http.Request) {
	err := wa.orderService.CreateOrder(r.Context(), "pipa", "q2313")
	if err != nil {
		if errors.Is(err, services.ErrOrderAlreadyAddedByUser) {
			w.WriteHeader(http.StatusOK)
		}
		if errors.Is(err, services.ErrOrderAlreadyAddedByAnotherUser) {
			w.WriteHeader(http.StatusConflict)
		}
		w.WriteHeader(http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusCreated)
}

func (wa *HTTPRouter) userListOrders(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (wa *HTTPRouter) userBalance(w http.ResponseWriter, r *http.Request) {
	userLogin := "pipa"
	balance := wa.orderService.GetUserBalance(r.Context(), userLogin)
	balanceResp := models.BalanceResp{
		Current:   float32(balance.Balance),
		Withdrawn: float32(balance.Withdrawn),
	}
	resp, err := json.Marshal(balanceResp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (wa *HTTPRouter) userWithdrawBalance(w http.ResponseWriter, r *http.Request) {
	var req models.WithdrawReq
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(body, &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Здесь валидация номера заказа
	// проверить номер заказа по алгоритму Луна и вывбростиь 422 в случае ошибки валидации
	err = wa.orderService.Withdraw(r.Context(), "pipa", req.OrderId, req.Sum)
	if err != nil {
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (wa *HTTPRouter) userWithdrawals(w http.ResponseWriter, r *http.Request) {
	orders, err := wa.orderService.UserWithdrawals(r.Context(), "pipa")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var respOrders = make([]models.UserWithdrawalsResp, 0, len(orders))
	for _, o := range orders {
		respOrders = append(respOrders, models.UserWithdrawalsResp{
			OrderId:     o.OrderId,
			Sum:         o.Value,
			ProcessedAt: o.UpdatedAt,
		})
	}
	resp, err := json.Marshal(respOrders)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
