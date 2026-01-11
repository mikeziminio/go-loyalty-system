package server

import (
	"context"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

var userContextKey = "user"

func (a *API) authMiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(auth, "Bearer ")
		if !ok {
			a.logger.Error("failed to fetch bearer token from header", zap.String("Authorization", auth))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		user, err := a.userRepository.AuthByToken(token)
		if err != nil {
			a.logger.Error("failed to auth by token", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
