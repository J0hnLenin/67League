package adminhttp

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/J0hnLenin/67League/internal/application/ports"
	adminhandler "github.com/J0hnLenin/67League/internal/delivery/adminhttp/handler"
	"github.com/J0hnLenin/67League/internal/delivery/http/middleware"
)

type RouterConfig struct {
	Logger *slog.Logger
	Tokens ports.TokenIssuer

	Auth     *adminhandler.AdminAuthHandler
	Season   *adminhandler.AdminSeasonHandler
	Player   *adminhandler.AdminPlayerHandler
	Match    *adminhandler.AdminMatchHandler
	Settings *adminhandler.AdminSettingsHandler
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

	r.Route("/admin", func(r chi.Router) {
		// Публичный логин.
		r.Post("/login", cfg.Auth.Login)

		// Приватная группа.
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.Tokens))
			r.Use(middleware.RequireAdmin())

			r.Route("/seasons", func(r chi.Router) {
				r.Post("/", cfg.Season.Create)

				r.Post("/{id}/start-registration", cfg.Season.StartRegistration)
				r.Post("/{id}/start", cfg.Season.Start)
				r.Post("/{id}/close", cfg.Season.Close)
				r.Post("/{id}/finish", cfg.Season.Finish)
				r.Post("/{id}/archive", cfg.Season.Archive)

				r.Post("/{id}/calculate-results", cfg.Season.CalculateResults)
				r.Post("/{id}/apply-rotation", cfg.Season.ApplyRotation)
				r.Post("/{id}/award-season-badges", cfg.Season.AwardSeasonBadges)
				r.Post("/{id}/award-super-title", cfg.Season.AwardSuperTitle)
			})

			r.Route("/players", func(r chi.Router) {
				r.Get("/", cfg.Player.List)
				r.Post("/{id}/ban", cfg.Player.Ban)
				r.Post("/{id}/unban", cfg.Player.Unban)
			})

			r.Route("/matches", func(r chi.Router) {
				r.Post("/{id}/cancel", cfg.Match.Cancel)
			})

			r.Route("/settings", func(r chi.Router) {
				r.Get("/", cfg.Settings.Get)
				r.Patch("/", cfg.Settings.Patch)
			})
		})
	})

	return r
}
