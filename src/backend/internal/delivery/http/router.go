package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/delivery/http/handler"
	"github.com/J0hnLenin/67League/internal/delivery/http/middleware"
)

type RouterConfig struct {
	Logger *slog.Logger
	Tokens ports.TokenIssuer

	Auth     *handler.AuthHandler
	Player   *handler.PlayerHandler
	Club     *handler.ClubHandler
	Season   *handler.SeasonHandler
	Division *handler.DivisionHandler
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
		// Публичные.
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", cfg.Auth.Register)
			r.Post("/login", cfg.Auth.Login)
		})
		r.Get("/players/{id}", cfg.Player.GetByID)
		r.Get("/clubs", cfg.Club.List)
		r.Get("/clubs/{id}", cfg.Club.GetByID)
		r.Get("/clubs/{id}/badges", cfg.Club.Badges)
		r.Get("/seasons", cfg.Season.List)
		r.Get("/seasons/current", cfg.Season.Current)
		r.Get("/seasons/{id}", cfg.Season.GetByID)
		r.Get("/divisions/{id}", cfg.Division.GetByID)
		r.Get("/divisions/{id}/members", cfg.Division.Members)

		// Приватные — игрок.
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.Tokens))
			r.Use(middleware.RequirePlayer())

			r.Get("/me", cfg.Player.Me)
			r.Get("/me/badges", cfg.Player.Badges)
			r.Get("/me/titles", cfg.Player.Titles)
			r.Get("/me/stones/history", cfg.Player.StoneHistory)

			r.Post("/clubs", cfg.Club.Create)
			r.Post("/clubs/{id}/join", cfg.Club.Join)
			r.Post("/clubs/{id}/leave", cfg.Club.Leave)

			r.Post("/seasons/{id}/join", cfg.Season.Join)
		})
	})

	return r
}
