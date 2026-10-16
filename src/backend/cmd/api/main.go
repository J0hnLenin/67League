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

	"github.com/J0hnLenin/67League/internal/application/usecases/club"
	"github.com/J0hnLenin/67League/internal/application/usecases/player"
	"github.com/J0hnLenin/67League/internal/application/usecases/season"
	apphttp "github.com/J0hnLenin/67League/internal/delivery/http"
	"github.com/J0hnLenin/67League/internal/delivery/http/handler"
	"github.com/J0hnLenin/67League/internal/infrastructure/auth"
	"github.com/J0hnLenin/67League/internal/infrastructure/email"
	"github.com/J0hnLenin/67League/internal/infrastructure/persistence/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

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

	// ---- База данных --------------------------------------------------
	db, err := postgres.NewDB(ctx, dsn)
	if err != nil {
		logger.Error("db open", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	pool := db.Pool()

	// ---- Репозитории --------------------------------------------------
	playerRepo := postgres.NewPlayerRepo(pool)
	adminRepo := postgres.NewAdminRepo(pool)
	clubRepo := postgres.NewClubRepo(pool)
	seasonRepo := postgres.NewSeasonRepo(pool)
	divisionRepo := postgres.NewDivisionRepo(pool)
	memberRepo := postgres.NewDivisionMemberRepo(pool)
	matchRepo := postgres.NewMatchRepo(pool)
	matchEventRepo := postgres.NewMatchEventRepo(pool)
	badgeRepo := postgres.NewBadgeRepo(pool)
	titleRepo := postgres.NewTitleRepo(pool)
	playerBadgeRepo := postgres.NewPlayerBadgeRepo(pool)
	clubBadgeRepo := postgres.NewClubBadgeRepo(pool)
	superTitleRepo := postgres.NewSuperTitleRepo(pool)
	stoneRepo := postgres.NewStoneTransactionRepo(pool)
	purchaseRepo := postgres.NewPurchaseRepo(pool)
	notificationRepo := postgres.NewNotificationRepo(pool)
	settingsRepo := postgres.NewSettingsRepo(pool)

	// _ = adminRepo и т.д. — здесь просто чтобы показать, что они созданы.
	// Часть пока не используется в use cases, но понадобится в админ-слое.
	_ = adminRepo
	_ = matchRepo
	_ = matchEventRepo
	_ = superTitleRepo
	_ = purchaseRepo
	_ = notificationRepo
	_ = settingsRepo

	// ---- Инфраструктурные сервисы -------------------------------------
	hasher := auth.NewPasswordHasher(0)
	tokens := auth.NewJWTIssuer(auth.JWTConfig{Secret: jwtSecret})
	sender := &email.LogSender{Logger: logger}
	txManager := postgres.NewTxManager(pool)

	// _ = sender — заглушка, чтобы показать: в use case'ах он не нужен.
	_ = sender

	clock := systemClock{}

	// ---- Use cases ----------------------------------------------------
	registerUC, err := player.NewRegisterPlayer(player.RegisterPlayerDeps{
		Players:   playerRepo,
		Clubs:     clubRepo,
		Seasons:   seasonRepo,
		Divisions: divisionRepo,
		Members:   memberRepo,
		Hasher:    hasher,
		Clock:     clock,
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

	createClubUC, err := club.NewCreateClub(club.CreateClubDeps{
		Clubs:     clubRepo,
		Players:   playerRepo,
		Badges:    badgeRepo,
		ClubBadge: clubBadgeRepo,
		Clock:     clock,
		Tx:        txManager,
	})
	if err != nil {
		logger.Error("create club usecase", "err", err)
		os.Exit(1)
	}

	joinClubUC, err := club.NewJoinClub(club.JoinClubDeps{
		Clubs:      clubRepo,
		Players:    playerRepo,
		Badges:     badgeRepo,
		ClubBadges: clubBadgeRepo,
		Clock:      clock,
		Tx:         txManager,
	})
	if err != nil {
		logger.Error("join club usecase", "err", err)
		os.Exit(1)
	}

	joinSeasonUC, err := season.NewJoinSeason(season.JoinSeasonDeps{
		Players:   playerRepo,
		Seasons:   seasonRepo,
		Divisions: divisionRepo,
		Members:   memberRepo,
		Clock:     clock,
		Tx:        txManager,
	})
	if err != nil {
		logger.Error("join season usecase", "err", err)
		os.Exit(1)
	}

	// ---- HTTP handlers ------------------------------------------------
	authHandler := handler.NewAuthHandler(registerUC, loginUC, logger)
	playerHandler := handler.NewPlayerHandler(
		playerRepo,
		badgeRepo,
		playerBadgeRepo,
		titleRepo,
		stoneRepo,
	)
	clubHandler := handler.NewClubHandler(
		clubRepo,
		playerRepo,
		clubBadgeRepo,
		badgeRepo,
		createClubUC,
		joinClubUC,
	)
	seasonHandler := handler.NewSeasonHandler(seasonRepo, joinSeasonUC)
	divisionHandler := handler.NewDivisionHandler(divisionRepo, memberRepo)

	// ---- Router -------------------------------------------------------
	router := apphttp.NewRouter(apphttp.RouterConfig{
		Logger:   logger,
		Tokens:   tokens,
		Auth:     authHandler,
		Player:   playerHandler,
		Club:     clubHandler,
		Season:   seasonHandler,
		Division: divisionHandler,
	})

	// ---- HTTP server --------------------------------------------------
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

// systemClock реализует ports.Clock через time.Now().
type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }
