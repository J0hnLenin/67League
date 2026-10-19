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

	clubuc "github.com/J0hnLenin/67League/internal/application/usecases/club"
	matchuc "github.com/J0hnLenin/67League/internal/application/usecases/match"
	playeruc "github.com/J0hnLenin/67League/internal/application/usecases/player"
	seasonuc "github.com/J0hnLenin/67League/internal/application/usecases/season"
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

	// TODO: подключить adminRepo к admin HTTP handlers (cmd/admin/main.go уже использует).
	// TODO: подключить superTitleRepo — выдача супертитула через admin API.
	// TODO: подключить purchaseRepo — магазин и покупки (фаза 2).
	// TODO: подключить settingsRepo — глобальные настройки через API.
	_ = adminRepo
	_ = superTitleRepo
	_ = purchaseRepo
	_ = settingsRepo

	// ---- Инфраструктурные сервисы -------------------------------------
	hasher := auth.NewPasswordHasher(0)
	tokens := auth.NewJWTIssuer(auth.JWTConfig{Secret: jwtSecret})
	sender := &email.LogSender{Logger: logger}
	txManager := postgres.NewTxManager(pool)
	clock := systemClock{}

	_ = sender // понадобится в воркере рассылки

	// ---- Use cases: player --------------------------------------------
	registerUC, err := playeruc.NewRegisterPlayer(playeruc.RegisterPlayerDeps{
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

	loginUC, err := playeruc.NewLoginPlayer(playeruc.LoginPlayerDeps{
		Players: playerRepo,
		Hasher:  hasher,
		Tokens:  tokens,
	})
	if err != nil {
		logger.Error("login usecase", "err", err)
		os.Exit(1)
	}

	// ---- Use cases: club ----------------------------------------------
	createClubUC, err := clubuc.NewCreateClub(clubuc.CreateClubDeps{
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

	joinClubUC, err := clubuc.NewJoinClub(clubuc.JoinClubDeps{
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

	// ---- Use cases: season --------------------------------------------
	joinSeasonUC, err := seasonuc.NewJoinSeason(seasonuc.JoinSeasonDeps{
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

	// ---- Use cases: match ---------------------------------------------
	createMatchUC, err := matchuc.NewCreateMatch(matchuc.CreateMatchDeps{
		Matches: matchRepo,
		Events:  matchEventRepo,
		Members: memberRepo,
		Seasons: seasonRepo,
		Notifs:  notificationRepo,
		Clock:   clock,
		Tx:      txManager,
	})
	if err != nil {
		logger.Error("create match usecase", "err", err)
		os.Exit(1)
	}

	acceptMatchUC, err := matchuc.NewAcceptMatch(matchuc.AcceptMatchDeps{
		Matches: matchRepo,
		Events:  matchEventRepo,
		Notifs:  notificationRepo,
		Clock:   clock,
		Tx:      txManager,
	})
	if err != nil {
		logger.Error("accept match usecase", "err", err)
		os.Exit(1)
	}

	startMatchUC, err := matchuc.NewStartMatch(matchuc.StartMatchDeps{
		Matches: matchRepo,
		Events:  matchEventRepo,
		Clock:   clock,
		Tx:      txManager,
	})
	if err != nil {
		logger.Error("start match usecase", "err", err)
		os.Exit(1)
	}

	submitMatchUC, err := matchuc.NewSubmitResult(matchuc.SubmitResultDeps{
		Matches: matchRepo,
		Events:  matchEventRepo,
		Notifs:  notificationRepo,
		Clock:   clock,
		Tx:      txManager,
	})
	if err != nil {
		logger.Error("submit result usecase", "err", err)
		os.Exit(1)
	}

	confirmMatchUC, err := matchuc.NewConfirmResult(matchuc.ConfirmResultDeps{
		Matches: matchRepo,
		Events:  matchEventRepo,
		Members: memberRepo,
		Players: playerRepo,
		Seasons: seasonRepo,
		Stones:  stoneRepo,
		Notifs:  notificationRepo,
		Clock:   clock,
		Tx:      txManager,
	})
	if err != nil {
		logger.Error("confirm result usecase", "err", err)
		os.Exit(1)
	}

	cancelMatchUC, err := matchuc.NewCancelMatch(matchuc.CancelMatchDeps{
		Matches: matchRepo,
		Events:  matchEventRepo,
		Seasons: seasonRepo,
		Notifs:  notificationRepo,
		Clock:   clock,
		Tx:      txManager,
	})
	if err != nil {
		logger.Error("cancel match usecase", "err", err)
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
	matchHandler := handler.NewMatchHandler(
		matchRepo,
		matchEventRepo,
		createMatchUC,
		acceptMatchUC,
		startMatchUC,
		submitMatchUC,
		confirmMatchUC,
		cancelMatchUC,
	)

	// ---- Router -------------------------------------------------------
	router := apphttp.NewRouter(apphttp.RouterConfig{
		Logger:   logger,
		Tokens:   tokens,
		Auth:     authHandler,
		Player:   playerHandler,
		Club:     clubHandler,
		Season:   seasonHandler,
		Division: divisionHandler,
		Match:    matchHandler,
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
