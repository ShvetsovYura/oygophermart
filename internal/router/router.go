package router

import (
	"context"
	"errors"
	"net/http"

	orderservice "github.com/ShvetsovYura/oygophermart/internal/services/order_service"
	"github.com/go-chi/chi/v5"
)

type OrderCreater interface {
	CreateOrder(ctx context.Context, userLogin string, orderId string) error
}

type HTTPRouter struct {
	service   OrderCreater
	rawRouter *chi.Mux
}

func NewHTTPRouter(service OrderCreater) *HTTPRouter {
	api := &HTTPRouter{service: service}
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
	w.WriteHeader(http.StatusOK)
}

func (wa *HTTPRouter) userLogin(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (wa *HTTPRouter) userLoadOrders(w http.ResponseWriter, r *http.Request) {
	err := wa.service.CreateOrder(r.Context(), "pipa", "q2313")
	if err != nil {
		if errors.Is(err, orderservice.ErrOrderAlreadyAddedByUser) {
			w.WriteHeader(http.StatusOK)
		}
		if errors.Is(err, orderservice.ErrOrderAlreadyAddedByAnotherUser) {
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
	w.WriteHeader(http.StatusOK)
}

func (wa *HTTPRouter) userWithdrawBalance(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (wa *HTTPRouter) userWithdrawals(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
