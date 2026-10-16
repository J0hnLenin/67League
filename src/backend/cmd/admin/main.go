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

	badgeuc "github.com/J0hnLenin/67League/internal/application/usecases/badge"
	matchuc "github.com/J0hnLenin/67League/internal/application/usecases/match"
	seasonuc "github.com/J0hnLenin/67League/internal/application/usecases/season"
	adminhttp "github.com/J0hnLenin/67League/internal/delivery/adminhttp"
	adminhandler "github.com/J0hnLenin/67League/internal/delivery/adminhttp/handler"
	"github.com/J0hnLenin/67League/internal/infrastructure/auth"
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
	pool := db.Pool()

	// Repositories
	adminRepo := postgres.NewAdminRepo(pool)
	playerRepo := postgres.NewPlayerRepo(pool)
	seasonRepo := postgres.NewSeasonRepo(pool)
	divisionRepo := postgres.NewDivisionRepo(pool)
	memberRepo := postgres.NewDivisionMemberRepo(pool)
	clubRepo := postgres.NewClubRepo(pool)
	matchRepo := postgres.NewMatchRepo(pool)
	matchEventRepo := postgres.NewMatchEventRepo(pool)
	badgeRepo := postgres.NewBadgeRepo(pool)
	playerBadgeRepo := postgres.NewPlayerBadgeRepo(pool)
	clubBadgeRepo := postgres.NewClubBadgeRepo(pool)
	superTitleRepo := postgres.NewSuperTitleRepo(pool)
	notificationRepo := postgres.NewNotificationRepo(pool)
	stoneRepo := postgres.NewStoneTransactionRepo(pool)
	settingsRepo := postgres.NewSettingsRepo(pool)

	hasher := auth.NewPasswordHasher(0)
	tokens := auth.NewJWTIssuer(auth.JWTConfig{Secret: jwtSecret})
	txManager := postgres.NewTxManager(pool)
	clock := systemClock{}

	// Use cases: season
	calculateUC, err := seasonuc.NewCalculateSeasonResults(seasonuc.CalculateSeasonResultsDeps{
		Seasons:   seasonRepo,
		Divisions: divisionRepo,
		Members:   memberRepo,
		Players:   playerRepo,
		Clubs:     clubRepo,
		Clock:     clock,
		Tx:        txManager,
	})
	if err != nil {
		os.Exit(1)
	}

	rotateUC, err := seasonuc.NewApplyRotation(seasonuc.ApplyRotationDeps{
		Seasons:   seasonRepo,
		Divisions: divisionRepo,
		Members:   memberRepo,
		Clock:     clock,
		Tx:        txManager,
	})
	if err != nil {
		os.Exit(1)
	}

	// Use cases: badge
	awardPersonalUC, err := badgeuc.NewAwardPersonalBadges(badgeuc.AwardPersonalBadgesDeps{
		Badges:       badgeRepo,
		PlayerBadges: playerBadgeRepo,
		Members:      memberRepo,
		Matches:      matchRepo,
		Players:      playerRepo,
		Clubs:        clubRepo,
		Clock:        clock,
		Tx:           txManager,
	})
	if err != nil {
		os.Exit(1)
	}

	awardSeasonUC, err := badgeuc.NewAwardSeasonBadges(badgeuc.AwardSeasonBadgesDeps{
		Badges:       badgeRepo,
		PlayerBadges: playerBadgeRepo,
		Members:      memberRepo,
		Divisions:    divisionRepo,
		Players:      playerRepo,
		Clubs:        clubRepo,
		Clock:        clock,
		Tx:           txManager,
	})
	if err != nil {
		os.Exit(1)
	}

	awardSuperUC, err := badgeuc.NewAwardSuperTitle(badgeuc.AwardSuperTitleDeps{
		Badges:       badgeRepo,
		PlayerBadges: playerBadgeRepo,
		SuperTitles:  superTitleRepo,
		Members:      memberRepo,
		Divisions:    divisionRepo,
		Players:      playerRepo,
		Clubs:        clubRepo,
		Clock:        clock,
		Tx:           txManager,
	})
	if err != nil {
		os.Exit(1)
	}

	// Use case: cancel (admin)
	cancelUC, err := matchuc.NewCancelMatch(matchuc.CancelMatchDeps{
		Matches: matchRepo,
		Events:  matchEventRepo,
		Seasons: seasonRepo,
		Notifs:  notificationRepo,
		Clock:   clock,
		Tx:      txManager,
	})
	if err != nil {
		os.Exit(1)
	}

	// Handlers
	authH := adminhandler.NewAdminAuthHandler(adminRepo, hasher, tokens)
	seasonH := adminhandler.NewAdminSeasonHandler(
		seasonRepo, divisionRepo,
		calculateUC, rotateUC,
		awardPersonalUC, awardSeasonUC, awardSuperUC,
	)
	playerH := adminhandler.NewAdminPlayerHandler(playerRepo)
	matchH := adminhandler.NewAdminMatchHandler(cancelUC)
	settingsH := adminhandler.NewAdminSettingsHandler(settingsRepo)

	_ = stoneRepo // понадобится для будущих админских операций
	_ = clubBadgeRepo

	// Router
	router := adminhttp.NewRouter(adminhttp.RouterConfig{
		Logger:   logger,
		Tokens:   tokens,
		Auth:     authH,
		Season:   seasonH,
		Player:   playerH,
		Match:    matchH,
		Settings: settingsH,
	})

	srv := &http.Server{
		Addr:              ":8081",
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("admin listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdown()
	_ = srv.Shutdown(shutdownCtx)
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }
