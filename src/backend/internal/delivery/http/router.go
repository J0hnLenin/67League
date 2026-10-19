package http

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

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
	Match    *handler.MatchHandler
}

func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RealIP)
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(cfg.Logger))
	r.Use(middleware.Recover(cfg.Logger))
	r.Use(corsMiddleware())

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		// ---- Публичные роуты -------------------------------------------
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

		r.Get("/matches/{id}", cfg.Match.GetByID)

		// ---- Приватные роуты (игрок) -----------------------------------
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.Tokens))
			r.Use(middleware.RequirePlayer())

			// Профиль.
			r.Get("/me", cfg.Player.Me)
			r.Get("/me/badges", cfg.Player.Badges)
			r.Get("/me/titles", cfg.Player.Titles)
			r.Get("/me/stones/history", cfg.Player.StoneHistory)

			// Клубы.
			r.Post("/clubs", cfg.Club.Create)
			r.Post("/clubs/{id}/join", cfg.Club.Join)
			r.Post("/clubs/{id}/leave", cfg.Club.Leave)

			// Сезоны.
			r.Post("/seasons/{id}/join", cfg.Season.Join)

			// Партии.
			r.Post("/matches", cfg.Match.Create)
			r.Post("/matches/{id}/accept", cfg.Match.Accept)
			r.Post("/matches/{id}/start", cfg.Match.Start)
			r.Post("/matches/{id}/result", cfg.Match.SubmitResult)
			r.Post("/matches/{id}/confirm", cfg.Match.Confirm)
			r.Post("/matches/{id}/cancel", cfg.Match.Cancel)

			r.Get("/me/matches", cfg.Match.ListMine)
		})
	})

	return r
}

// corsMiddleware возвращает CORS middleware.
// В dev-режиме разрешает запросы с localhost (фронтенд).
// В production — только с указанного ORIGIN.
func corsMiddleware() func(http.Handler) http.Handler {
	origins := []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	if env := os.Getenv("CORS_ALLOWED_ORIGINS"); env != "" {
		origins = strings.Split(env, ",")
	}

	return cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
