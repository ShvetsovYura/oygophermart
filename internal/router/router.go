package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type HTTPRouter struct {
	data      any
	rawRouter *chi.Mux
}

func NewHTTPRouter() *HTTPRouter {
	api := &HTTPRouter{data: "data"}
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
	w.WriteHeader(http.StatusOK)
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
