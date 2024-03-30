package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/ShvetsovYura/oygophermart/internal/app"
	"github.com/ShvetsovYura/oygophermart/internal/logger"
	"github.com/ShvetsovYura/oygophermart/internal/options"
)

func main() {
	logger.InitLogger("info")

	opt := options.AppOptions{}
	opt.ParseArgs()
	err := opt.ParseEnv()
	if err != nil {
		fmt.Println(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	err = app.Run(ctx, &opt)
	if err != nil {
		logger.Log.Fatalf("Error on run: %e", err)
	}
	<-ctx.Done()
}
