package main

import (
	"context"
	stdlog "log"

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

	l, err := log.New(conf.LogLevel)
	if err != nil {
		stdlog.Fatalf("failed to init config: %v", err)
	}

	connURL := "postgres://postgres:postgres@localhost:5432/main?sslmode=disable"
	minConns := 2
	maxConns := 10

	d, err := db.NewDB(ctx, connURL, minConns, maxConns, l)
	if err != nil {
		l.Fatal("failed to init db", zap.Error(err))
	}
	defer d.Close()

	// accrualClient := accrual.NewClient(fmt.Sprintf("http://%s/", conf.AccrualAddress))

	// ids := []string{
	// 	"9278923470",
	// 	"12345678903",
	// 	"346436439",
	// 	"12345678900",
	// 	"346436488",
	// }
	// for i := range 5 {
	// 	oi, err := accrualClient.OrderInfo(ctx, ids[i])
	// 	log.Info("order info", zap.String("order", fmt.Sprintf("%+v", oi)), zap.Error(err))
	// }

	a := server.NewAPI(conf.Address, d, l)

	a.RegisterRouters()
	a.Run(ctx)
}
