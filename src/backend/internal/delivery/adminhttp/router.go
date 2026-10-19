package adminhttp

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

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
	r.Use(corsMiddleware())

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

func corsMiddleware() func(http.Handler) http.Handler {
	origins := []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	allowAll := false
	if env := os.Getenv("CORS_ALLOWED_ORIGINS"); env != "" {
		origins = strings.Split(env, ",")
		for _, o := range origins {
			if strings.TrimSpace(o) == "*" {
				allowAll = true
				break
			}
		}
	}

	opts := cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}
	if allowAll {
		opts.AllowedOrigins = nil
		opts.AllowOriginFunc = func(r *http.Request, origin string) bool { return true }
	}
	return cors.Handler(opts)
}
