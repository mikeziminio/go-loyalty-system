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
	Register(login string, password string) (*model.User, error)
	AuthByLogin(login string, password string) (*model.User, error)
	AuthByToken(token string) (*model.User, error)
}

type API struct {
	logger         *zap.Logger
	userRepository UserRepository
	httpServer     *http.Server
	router         *chi.Mux
}

func NewAPI(address string, userRepository UserRepository, logger *zap.Logger) *API {
	r := chi.NewRouter()

	httpServer := &http.Server{
		Addr:              address,
		Handler:           r,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
	}

	return &API{
		logger:         logger,
		userRepository: userRepository,
		router:         r,
		httpServer:     httpServer,
	}
}

func (a *API) RegisterRouters() {
	a.router.Use(middleware.StripSlashes)

	a.router.Post("/api/user/register", a.Register)
	a.router.Post("/api/user/login", a.Login)
}

func (a *API) Run(ctx context.Context) {
	go func() {
		_ = a.httpServer.ListenAndServe()
	}()

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
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
