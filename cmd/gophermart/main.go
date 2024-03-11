package main

import (
	"fmt"

	"github.com/ShvetsovYura/oygophermart/internal/webserver"
)

func main() {
	ws, err := webserver.NewWebServer()
	if err != nil {
		fmt.Printf("%e", err)
		return
	}
	ws.Start()
}
