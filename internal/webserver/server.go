package webserver

import (
	"net/http"

	"github.com/ShvetsovYura/oygophermart/internal/router"
)

type WebServer struct {
	router *router.HTTPRouter
}

func NewWebServer() *WebServer {
	router := router.NewHTTPRouter()
	s := &WebServer{
		router: router,
	}
	return s
}

func (ws *WebServer) Start() {
	ws.router.InitRouter()
	http.ListenAndServe(":3001", ws.router.GetRouter())
}
