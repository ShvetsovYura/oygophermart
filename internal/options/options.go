package options

import (
	"flag"

	"github.com/caarlos0/env/v10"
)

type AppOptions struct {
	RunAddr           string `env:"RUN_ADDRESS"`
	DatabaseURI       string `env:"DATABASE_URI"`
	AccrualSystemAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func (o *AppOptions) ParseArgs() {
	flag.StringVar(&o.RunAddr, "a", ":3001", "server endpoint address")
	flag.StringVar(&o.DatabaseURI, "d", "", "db connection string")
	flag.StringVar(&o.AccrualSystemAddr, "r", "localhost:8080", "accrual service address")
	flag.Parse()
}

func (o *AppOptions) ParseEnv() error {
	if err := env.Parse(o); err != nil {
		return err
	}
	return nil
}
