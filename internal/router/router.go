package router

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/ShvetsovYura/oygophermart/internal/logger"
	"github.com/ShvetsovYura/oygophermart/internal/middlewares"
	"github.com/ShvetsovYura/oygophermart/internal/models"
	"github.com/ShvetsovYura/oygophermart/internal/services"
	"github.com/ShvetsovYura/oygophermart/internal/store"
	"github.com/ShvetsovYura/oygophermart/internal/utils"
	"github.com/go-chi/chi/v5"
)

type OrderWorker interface {
	CreateOrder(ctx context.Context, userId uint64, orderId string) error
	GetUserBalance(ctx context.Context, userId uint64) models.BalanceModel
	Withdraw(ctx context.Context, userId uint64, orderId string, value float64) error
	UserWithdrawals(ctx context.Context, userId uint64) ([]models.OrderGroupedModel, error)
	GetUserOrders(ctx context.Context, userId uint64) ([]models.OrderGroupedModel, error)
}

type UserWorker interface {
	CreateUser(ctx context.Context, login string, password string) (int64, error)
	Login(ctx context.Context, login string, password string) (int64, error)
}

type Tokener interface {
	GenerateToken(id uint64) (string, error)
	ValidateSign(token string) (bool, error)
	ExtractUserId(token string) (uint64, error)
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
	ms := []func(http.Handler) http.Handler{
		middlewares.CheckAuthCookie(wa.tokenService),
		middlewares.ExtractUserId(wa.tokenService),
	}

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", wa.userRegister)
			r.Post("/login", wa.userLogin)
			r.With(ms...).Post("/orders", wa.userLoadOrders)
			r.With(ms...).Get("/orders", wa.userListOrders)
			r.With(ms...).Get("/balance", wa.userBalance)
			r.With(ms...).Post("/balance/withdraw", wa.userWithdraw)
			r.With(ms...).Get("/withdrawals", wa.userWithdrawals)
		})
	})
	wa.rawRouter = r
}

func (wa *HTTPRouter) userRegister(w http.ResponseWriter, r *http.Request) {
	// TODO:  Добавить проверку content-type
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

	id, err := wa.userService.CreateUser(r.Context(), user.Login, user.Password)
	if err != nil {
		if errors.Is(err, services.ErrUserAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	token, err := wa.tokenService.GenerateToken(uint64(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	c := http.Cookie{Name: "token", Value: token, HttpOnly: true, MaxAge: 3600}
	http.SetCookie(w, &c)
	w.WriteHeader(http.StatusOK)
}

func (wa *HTTPRouter) userLogin(w http.ResponseWriter, r *http.Request) {
	// TODO:  Добавить проверку content-type

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

	uid, err := wa.userService.Login(r.Context(), user.Login, user.Password)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) || errors.Is(err, services.ErrNotValidLoginOrPassword) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	token, err := wa.tokenService.GenerateToken(uint64(uid))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	c := http.Cookie{
		Name:     "token",
		Value:    token,
		MaxAge:   0,
		Expires:  time.Now().Add(time.Minute * 30),
		HttpOnly: true,
	}

	http.SetCookie(w, &c)
	w.WriteHeader(http.StatusOK)
}

func (wa *HTTPRouter) userLoadOrders(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("uid").(uint64)

	contentType := r.Header.Get("Content-Type")
	if contentType != "text/plain" {
		w.WriteHeader(http.StatusTeapot)
		logger.Log.Debugf("bad content-type: %s", contentType)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer r.Body.Close()
	orderId := string(body)
	isValid, err := utils.CheckLuhnFromStr(orderId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !isValid {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = wa.orderService.CreateOrder(r.Context(), userId, orderId)
	if err != nil {
		logger.Log.Debugf("Error on create order, %v", err)
		if errors.Is(err, services.ErrOrderAlreadyAddedByUser) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, services.ErrOrderAlreadyAddedByAnotherUser) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		logger.Log.Debugf("error on create order: %v", err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (wa *HTTPRouter) userListOrders(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("uid").(uint64)
	w.Header().Add("Content-Type", "application/json")
	orders, err := wa.orderService.GetUserOrders(r.Context(), userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(orders) < 1 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	response, err := json.Marshal(orders)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(response)
	// w.WriteHeader(http.StatusOK)
}

func (wa *HTTPRouter) userBalance(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("uid")
	w.Header().Add("Content-Type", "application/json")

	balance := wa.orderService.GetUserBalance(r.Context(), userId.(uint64))
	balanceResp := models.BalanceResp{
		Current:   balance.Balance,
		Withdrawn: math.Abs(balance.Withdrawn),
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

func (wa *HTTPRouter) userWithdraw(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("uid").(uint64)
	var req models.WithdrawReq
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Debugf("err: %e", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	err = json.Unmarshal(body, &req)
	if err != nil {
		logger.Log.Debugf("err on withdraw: %e", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Log.Debugf("withdraw req: %v", req)

	isValid, err := utils.CheckLuhnFromStr(req.OrderId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !isValid {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	err = wa.orderService.Withdraw(r.Context(), userId, req.OrderId, float64(req.Sum))
	if err != nil {
		logger.Log.Debugf("err ubu : %e", err)
		if errors.Is(err, store.ErrOrderAlreadyExistsInDb) {
			http.Error(w, "Order already exists", http.StatusUnprocessableEntity)
		}
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (wa *HTTPRouter) userWithdrawals(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("uid").(uint64)
	w.Header().Add("Content-Type", "application/json")
	orders, err := wa.orderService.UserWithdrawals(r.Context(), userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(orders) < 1 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var respOrders = make([]models.UserWithdrawalsResp, 0, len(orders))
	for _, o := range orders {
		respOrders = append(respOrders, models.UserWithdrawalsResp{
			OrderId:     o.Id,
			Sum:         math.Abs(*o.Accrual),
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
