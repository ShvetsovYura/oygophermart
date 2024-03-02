package main

import "github.com/ShvetsovYura/oygophermart/internal/webserver"

func main() {
	ws := webserver.NewWebServer()
	ws.Start()
}
