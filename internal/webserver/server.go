package webserver

import (
	"context"
	"net/http"

	"github.com/ShvetsovYura/oygophermart/internal/logger"
	"github.com/ShvetsovYura/oygophermart/internal/options"
	"github.com/ShvetsovYura/oygophermart/internal/router"
	"github.com/ShvetsovYura/oygophermart/internal/services"
	"github.com/ShvetsovYura/oygophermart/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebServer struct {
	router  *router.HTTPRouter
	options *options.AppOptions
}

func NewWebServer(ctx context.Context, dbConn *pgxpool.Pool, opt *options.AppOptions) (*WebServer, error) {
	orderStore, err := store.NewOrderStore(dbConn)
	if err != nil {
		return nil, err
	}
	userStore, err := store.NewUserStore(dbConn)
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
		router:  router,
		options: opt,
	}, nil
}

func (ws *WebServer) Start() {
	ws.router.InitRouter()
	logger.Log.Debugf("start on: %s", ws.options.RunAddr)
	http.ListenAndServe(ws.options.RunAddr, ws.router.GetRouter())
}
