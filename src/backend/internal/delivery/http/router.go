package http

import (
	"log/slog"
	"net/http"

	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/delivery/http/handler"
	"github.com/J0hnLenin/67League/internal/delivery/http/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type RouterConfig struct {
	Logger *slog.Logger
	Tokens ports.TokenIssuer
	Auth   *handler.AuthHandler
	// добавьте сюда остальные handlers по мере реализации
}

func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RealIP)
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(cfg.Logger))
	r.Use(middleware.Recover(cfg.Logger))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", cfg.Auth.Register)
			r.Post("/login", cfg.Auth.Login)
		})

		// Приватные роуты.
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.Tokens))
			r.Use(middleware.RequirePlayer())

			// TODO: /me, /me/badges, /me/title, /matches, /clubs, ...
		})
	})

	return r
}
