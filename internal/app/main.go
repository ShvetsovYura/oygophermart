package app

import (
	"context"
	"fmt"

	accrualagent "github.com/ShvetsovYura/oygophermart/internal/accrual_agent"
	"github.com/ShvetsovYura/oygophermart/internal/options"
	"github.com/ShvetsovYura/oygophermart/internal/store"
	"github.com/ShvetsovYura/oygophermart/internal/webserver"
	"github.com/ShvetsovYura/oygophermart/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Run(ctx context.Context, opts *options.AppOptions) error {
	migrations.RunUpMigration(opts.DatabaseURI)
	conn, err := pgxpool.New(ctx, opts.DatabaseURI)

	if err != nil {
		return err
	}
	ws, err := webserver.NewWebServer(conn, opts)
	if err != nil {
		fmt.Printf("%e", err)
		return err
	}
	context := context.Background()
	orderStore, err := store.NewOrderStore(conn)
	if err != nil {
		return err
	}
	a := accrualagent.NewAccrualAgent(opts.AccrualSystemAddr, orderStore, 1)
	go a.Start(context)
	go ws.Start()
	<-ctx.Done()
	return nil
}
