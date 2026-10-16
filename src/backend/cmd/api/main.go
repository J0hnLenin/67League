package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/J0hnLenin/67League/internal/application/usecases/player"
	apphttp "github.com/J0hnLenin/67League/internal/delivery/http"
	"github.com/J0hnLenin/67League/internal/delivery/http/handler"
	"github.com/J0hnLenin/67League/internal/infrastructure/auth"
	"github.com/J0hnLenin/67League/internal/infrastructure/email"
	"github.com/J0hnLenin/67League/internal/infrastructure/persistence/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	dsn := os.Getenv("LIGA_DB_DSN")
	if dsn == "" {
		dsn = "postgres://liga:liga@localhost:5432/liga67?sslmode=disable"
	}
	jwtSecret := os.Getenv("LIGA_JWT_SECRET")
	if jwtSecret == "" {
		logger.Warn("LIGA_JWT_SECRET not set, using insecure default (dev only)")
		jwtSecret = "dev-secret-change-me"
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	db, err := postgres.NewDB(ctx, dsn)
	if err != nil {
		logger.Error("db open", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Репозитории.
	playerRepo := postgres.NewPlayerRepo(db.Pool())
	clubRepo := postgres.NewClubRepo(db.Pool())
	seasonRepo := postgres.NewSeasonRepo(db.Pool())
	divisionRepo := postgres.NewDivisionRepo(db.Pool())
	memberRepo := postgres.NewDivisionMemberRepo(db.Pool())

	// Сервисы.
	hasher := auth.NewPasswordHasher(0)
	tokens := auth.NewJWTIssuer(auth.JWTConfig{Secret: jwtSecret})
	sender := &email.LogSender{Logger: logger}
	txManager := postgres.NewTxManager(db.Pool())

	// Use cases.
	registerUC, err := player.NewRegisterPlayer(player.RegisterPlayerDeps{
		Players:   playerRepo,
		Clubs:     clubRepo,
		Seasons:   seasonRepo,
		Divisions: divisionRepo,
		Members:   memberRepo,
		Hasher:    hasher,
		Clock:     systemClock{},
		Tx:        txManager,
	})
	if err != nil {
		logger.Error("register usecase", "err", err)
		os.Exit(1)
	}
	loginUC, err := player.NewLoginPlayer(player.LoginPlayerDeps{
		Players: playerRepo,
		Hasher:  hasher,
		Tokens:  tokens,
	})
	if err != nil {
		logger.Error("login usecase", "err", err)
		os.Exit(1)
	}
	_ = sender // пока не используется — заглушка

	// Handlers.
	authHandler := handler.NewAuthHandler(registerUC, loginUC, logger)

	// Router.
	router := apphttp.NewRouter(apphttp.RouterConfig{
		Logger: logger,
		Tokens: tokens,
		Auth:   authHandler,
	})

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("api listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown", "err", err)
	}
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }
