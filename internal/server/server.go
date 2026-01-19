package server

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/mikeziminio/go-loyalty-system/internal/model"
)

type UserRepository interface {
	Register(ctx context.Context, login string, password string) (*model.User, error)
	AuthByLogin(ctx context.Context, login string, password string) (*model.User, error)
	AuthByToken(ctx context.Context, token string) (*model.User, error)
	AddWithdrawal(ctx context.Context, userID int, orderID string, sum float64) error
	Withdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error)
	AddOrder(ctx context.Context, userID int, orderID string) error
	ProcessOrder(
		ctx context.Context, userID int, orderID string,
		accrual float64, status string,
	) error
	Orders(ctx context.Context, userID int) ([]model.Order, error)
}

type OrderFetcher interface {
	Order(ctx context.Context, orderID string) (*model.Order, error)
}

type API struct {
	logger         *zap.Logger
	httpServer     *http.Server
	router         *chi.Mux
	userRepository UserRepository
	orderFetcher   OrderFetcher
}

func NewAPI(
	address string, userRepository UserRepository,
	orderFetcher OrderFetcher, logger *zap.Logger,
) *API {
	r := chi.NewRouter()

	httpServer := &http.Server{
		Addr:              address,
		Handler:           r,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
	}

	return &API{
		logger:         logger,
		httpServer:     httpServer,
		router:         r,
		userRepository: userRepository,
		orderFetcher:   orderFetcher,
	}
}

func (a *API) RegisterRouters() {
	a.router.Use(middleware.StripSlashes)

	a.router.Post("/api/user/register", a.Register)
	a.router.Post("/api/user/login", a.Login)

	a.router.With(a.authMiddlewareHandler).
		Get("/api/user/balance", a.Balance)
	a.router.With(a.authMiddlewareHandler).
		Get("/api/user/withdrawals", a.Withdrawals)
	a.router.With(a.authMiddlewareHandler).
		Post("/api/user/balance/withdraw", a.AddWithdrawal)
	a.router.With(a.authMiddlewareHandler).
		Get("/api/user/orders", a.Orders)
	a.router.With(a.authMiddlewareHandler).
		Post("/api/user/orders", a.AddOrder)
}

func (a *API) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		a.logger.Info("start listen",
			zap.String("addr", a.httpServer.Addr),
		)
		err := a.httpServer.ListenAndServe()
		if err != nil {
			a.logger.Error("stop listen",
				zap.String("addr", a.httpServer.Addr),
				zap.Error(err),
			)
			cancel()
		}
	}()

	ctx, cancel = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	<-ctx.Done()
}

func (a *API) standardError(w http.ResponseWriter, logText string, err error, code int) {
	a.logger.Error(logText, zap.Error(err))
	http.Error(w, http.StatusText(code), code)
}

func (a *API) customError(w http.ResponseWriter, errText string, err error, code int) {
	a.logger.Error(errText, zap.Error(err))
	http.Error(w, errText, code)
}
