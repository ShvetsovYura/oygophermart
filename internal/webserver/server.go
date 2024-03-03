package webserver

import (
	"net/http"

	"github.com/ShvetsovYura/oygophermart/internal/router"
	orderservice "github.com/ShvetsovYura/oygophermart/internal/services/order_service"
	"github.com/ShvetsovYura/oygophermart/internal/store"
)

type WebServer struct {
	router *router.HTTPRouter
}

const DSN = "postgres://pipa:F,shdfk!@localhost:5432/oy_loyalty?sslmode=disable"

func NewWebServer() *WebServer {

	dbStore, _ := store.NewStore(DSN)

	service := orderservice.NewOrderService(dbStore)
	router := router.NewHTTPRouter(service)
	s := &WebServer{
		router: router,
	}
	return s
}

func (ws *WebServer) Start() {
	ws.router.InitRouter()
	http.ListenAndServe(":3001", ws.router.GetRouter())
}
