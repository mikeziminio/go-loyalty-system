package main

import (
	"context"
	stdlog "log"
	"time"

	"github.com/mikeziminio/go-loyalty-system/internal/clients/accrual"
	"github.com/mikeziminio/go-loyalty-system/internal/config"
	"github.com/mikeziminio/go-loyalty-system/internal/db"
	"github.com/mikeziminio/go-loyalty-system/internal/log"
	"github.com/mikeziminio/go-loyalty-system/internal/server"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := config.NewFromEnvsAndFlags()
	if err != nil {
		stdlog.Fatalf("failed to load config: %v", err)
	}

	logger, err := log.New(conf.LogLevel)
	if err != nil {
		stdlog.Fatalf("failed to init config: %v", err)
	}

	var minConns int32 = 2
	var maxConns int32 = 10

	d, err := db.NewDB(ctx, conf.DatabaseURI, minConns, maxConns, logger)
	if err != nil {
		logger.Fatal("failed to init db", zap.Error(err))
	}
	defer d.Close()

	var accrualTimeout = 10 * time.Second
	accrualClient := accrual.NewClient(
		conf.AccrualAddress,
		accrualTimeout,
	)

	a := server.NewAPI(conf.Address, d, accrualClient, logger)

	a.RegisterRouters()
	a.Run(ctx)
}
