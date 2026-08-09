package http

import (
	"net/http"

	"github.com/opsflow/common/middleware"
)

func RegisterRoutes(mux *http.ServeMux, handler *AuthHandler, jwtSecret string) {
	mux.HandleFunc("/register", handler.Register)
	mux.HandleFunc("/login", handler.Login)
	mux.HandleFunc("/refresh", handler.Refresh)

	// Protected routes
	authMw := middleware.AuthMiddleware(jwtSecret)
	mux.Handle("/me", authMw(http.HandlerFunc(handler.Me)))
}
