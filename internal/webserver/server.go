package webserver

import (
	"context"
	"net/http"

	"github.com/ShvetsovYura/oygophermart/internal/router"
	"github.com/ShvetsovYura/oygophermart/internal/services"
	"github.com/ShvetsovYura/oygophermart/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebServer struct {
	router *router.HTTPRouter
}

const DSN = "postgres://pipa:F,shdfk!@localhost:5432/oy_loyalty?sslmode=disable"

func NewWebServer() (*WebServer, error) {
	conn, err := pgxpool.New(context.Background(), DSN)
	if err != nil {
		return nil, err
	}
	orderStore, err := store.NewOrderStore(conn)
	if err != nil {
		return nil, err
	}
	userStore, err := store.NewUserStore(conn)
	if err != nil {
		return nil, err
	}
	hasher := services.NewHashService()

	router := router.NewHTTPRouter(
		services.NewOrderService(orderStore, userStore),
		services.NewUserService(userStore, hasher),
		hasher,
	)

	return &WebServer{
		router: router,
	}, nil
}

func (ws *WebServer) Start() {
	ws.router.InitRouter()
	http.ListenAndServe(":3001", ws.router.GetRouter())
}
