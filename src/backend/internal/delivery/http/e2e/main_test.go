// Package e2e содержит интеграционные тесты HTTP API 6-7 Лиги.
//
// Один общий Postgres-контейнер на весь пакет. Перед каждым тестом
// таблицы обрезаются через truncateAll. Тесты не параллельны.
package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/J0hnLenin/67League/internal/application/usecases/club"
	matchuc "github.com/J0hnLenin/67League/internal/application/usecases/match"
	"github.com/J0hnLenin/67League/internal/application/usecases/player"
	usecaseseason "github.com/J0hnLenin/67League/internal/application/usecases/season"
	apphttp "github.com/J0hnLenin/67League/internal/delivery/http"
	"github.com/J0hnLenin/67League/internal/delivery/http/handler"
	"github.com/J0hnLenin/67League/internal/domain/division"
	domseason "github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/infrastructure/auth"
	"github.com/J0hnLenin/67League/internal/infrastructure/persistence/postgres"
	"github.com/J0hnLenin/67League/migrations"
)

var (
	testApp  *app
	testSrv  *httptest.Server
	testPool *pgxpool.Pool
)

// app — собранный стек, доступный тестам.
type app struct {
	handler   http.Handler
	pool      *pgxpool.Pool
	players   *postgres.PlayerRepo
	clubs     *postgres.ClubRepo
	seasons   *postgres.SeasonRepo
	divisions *postgres.DivisionRepo
	members   *postgres.DivisionMemberRepo
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("liga67_e2e"),
		tcpostgres.WithUsername("liga"),
		tcpostgres.WithPassword("liga"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start postgres: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = pgContainer.Terminate(ctx) }()

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "get dsn: %v\n", err)
		os.Exit(1)
	}

	if err := runMigrations(dsn); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}

	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pool: %v\n", err)
		os.Exit(1)
	}
	defer testPool.Close()

	built, err := buildApp(testPool)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build app: %v\n", err)
		os.Exit(1)
	}
	testApp = built

	testSrv = httptest.NewServer(testApp.handler)
	defer testSrv.Close()

	os.Exit(m.Run())
}

// buildApp собирает роутер как в cmd/api/main.go, только без слушателя порта.
func buildApp(pool *pgxpool.Pool) (*app, error) {
	playerRepo := postgres.NewPlayerRepo(pool)
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
	stoneRepo := postgres.NewStoneTransactionRepo(pool)
	notificationRepo := postgres.NewNotificationRepo(pool)

	hasher := auth.NewPasswordHasher(4)
	tokens := auth.NewJWTIssuer(auth.JWTConfig{
		Secret:    "test-secret",
		PlayerTTL: time.Hour,
	})
	txManager := postgres.NewTxManager(pool)
	clock := systemClock{}
	logger := realLogger()

	// ---- Use cases: player ----
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
		return nil, err
	}
	loginUC, err := player.NewLoginPlayer(player.LoginPlayerDeps{
		Players: playerRepo,
		Hasher:  hasher,
		Tokens:  tokens,
	})
	if err != nil {
		return nil, err
	}

	// ---- Use cases: club ----
	createClubUC, err := club.NewCreateClub(club.CreateClubDeps{
		Clubs:     clubRepo,
		Players:   playerRepo,
		Badges:    badgeRepo,
		ClubBadge: clubBadgeRepo,
		Clock:     clock,
		Tx:        txManager,
	})
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// ---- Use cases: season ----
	joinSeasonUC, err := usecaseseason.NewJoinSeason(usecaseseason.JoinSeasonDeps{
		Players:   playerRepo,
		Seasons:   seasonRepo,
		Divisions: divisionRepo,
		Members:   memberRepo,
		Clock:     clock,
		Tx:        txManager,
	})
	if err != nil {
		return nil, err
	}

	// ---- Use cases: match ----
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
		return nil, err
	}
	acceptMatchUC, err := matchuc.NewAcceptMatch(matchuc.AcceptMatchDeps{
		Matches: matchRepo,
		Events:  matchEventRepo,
		Notifs:  notificationRepo,
		Clock:   clock,
		Tx:      txManager,
	})
	if err != nil {
		return nil, err
	}
	startMatchUC, err := matchuc.NewStartMatch(matchuc.StartMatchDeps{
		Matches: matchRepo,
		Events:  matchEventRepo,
		Clock:   clock,
		Tx:      txManager,
	})
	if err != nil {
		return nil, err
	}
	submitMatchUC, err := matchuc.NewSubmitResult(matchuc.SubmitResultDeps{
		Matches: matchRepo,
		Events:  matchEventRepo,
		Notifs:  notificationRepo,
		Clock:   clock,
		Tx:      txManager,
	})
	if err != nil {
		return nil, err
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
		return nil, err
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
		return nil, err
	}

	// ---- Handlers ----
	authH := handler.NewAuthHandler(registerUC, loginUC, logger)
	playerH := handler.NewPlayerHandler(
		playerRepo, badgeRepo, playerBadgeRepo, titleRepo, stoneRepo,
	)
	clubH := handler.NewClubHandler(
		clubRepo, playerRepo, clubBadgeRepo, badgeRepo,
		createClubUC, joinClubUC,
	)
	seasonH := handler.NewSeasonHandler(seasonRepo, joinSeasonUC)
	divisionH := handler.NewDivisionHandler(divisionRepo, memberRepo)
	matchH := handler.NewMatchHandler(
		matchRepo, matchEventRepo,
		createMatchUC, acceptMatchUC, startMatchUC,
		submitMatchUC, confirmMatchUC, cancelMatchUC,
	)

	router := apphttp.NewRouter(apphttp.RouterConfig{
		Logger:   logger,
		Tokens:   tokens,
		Auth:     authH,
		Player:   playerH,
		Club:     clubH,
		Season:   seasonH,
		Division: divisionH,
		Match:    matchH,
	})

	return &app{
		handler:   router,
		pool:      pool,
		players:   playerRepo,
		clubs:     clubRepo,
		seasons:   seasonRepo,
		divisions: divisionRepo,
		members:   memberRepo,
	}, nil
}

// runMigrations применяет .up.sql из встроенной папки migrations.
func runMigrations(dsn string) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("iofs source: %w", err)
	}
	sqlDB, err := openSQLDB(dsn)
	if err != nil {
		return fmt.Errorf("open sql: %w", err)
	}
	defer sqlDB.Close()

	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func openSQLDB(dsn string) (*sql.DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	return stdlib.OpenDB(*cfg.ConnConfig), nil
}

// truncateAll очищает все таблицы. Вызывается в начале каждого теста.
func truncateAll(t *testing.T) {
	t.Helper()
	const q = `
		TRUNCATE TABLE
			site_settings,
			notifications,
			purchases,
			stone_transactions,
			super_titles,
			club_badges,
			player_badges,
			titles,
			badges,
			match_events,
			matches,
			division_members,
			divisions,
			seasons,
			clubs,
			admins,
			players
		RESTART IDENTITY CASCADE
	`
	if _, err := testPool.Exec(context.Background(), q); err != nil {
		t.Fatalf("truncate all: %v", err)
	}
}

// ---- HTTP helpers ---------------------------------------------------------

type httpResp struct {
	Status int
	Body   []byte
}

func (r *httpResp) JSON(t *testing.T, v any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(r.Body, v), "body: %s", string(r.Body))
}

func (r *httpResp) RequireStatus(t *testing.T, want int) *httpResp {
	t.Helper()
	require.Equal(t, want, r.Status, "body: %s", string(r.Body))
	return r
}

// do выполняет HTTP-запрос к тестовому серверу.
func do(t *testing.T, method, path, token string, body any) *httpResp {
	t.Helper()

	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, testSrv.URL+path, reader)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return &httpResp{Status: resp.StatusCode, Body: data}
}

// ---- Доменные helpers -----------------------------------------------------

// registerAndLogin — самый частый сценарий: регистрация + вход.
// Возвращает playerID и токен.
func registerAndLogin(t *testing.T, email, nickname string) (int64, string) {
	t.Helper()

	reg := do(t, "POST", "/api/v1/auth/register", "", map[string]any{
		"full_name": "Tester",
		"email":     email,
		"password":  "secret123",
		"nickname":  nickname,
		"city":      "Moscow",
	}).RequireStatus(t, 201)

	var regResp struct {
		PlayerID int64 `json:"player_id"`
	}
	reg.JSON(t, &regResp)
	require.NotZero(t, regResp.PlayerID)

	login := do(t, "POST", "/api/v1/auth/login", "", map[string]any{
		"email":    email,
		"password": "secret123",
	}).RequireStatus(t, 200)

	var loginResp struct {
		Token string `json:"token"`
	}
	login.JSON(t, &loginResp)
	require.NotEmpty(t, loginResp.Token)

	return regResp.PlayerID, loginResp.Token
}

// seedSeasonWithDivisions создаёт сезон в статусе active и три дивизиона.
func seedSeasonWithDivisions(t *testing.T) (seasonID, divS, divA, divB int64) {
	t.Helper()
	ctx := context.Background()
	now := time.Now().UTC()

	s := newSeason(t, 1, now)
	require.NoError(t, testApp.seasons.Create(ctx, s))

	divs := []struct {
		code     string
		order    int
		min, max int
	}{
		{"S", 0, 1000, 2000},
		{"A", 1, 500, 999},
		{"B", 2, 0, 499},
	}
	ids := []int64{}
	for _, d := range divs {
		dd := newDivision(t, s.ID, d.code, d.order, d.min, d.max, now)
		require.NoError(t, testApp.divisions.Create(ctx, dd))
		ids = append(ids, int64(dd.ID))
	}
	return int64(s.ID), ids[0], ids[1], ids[2]
}

func newSeason(t *testing.T, number int, now time.Time) *domseason.Season {
	t.Helper()
	s, err := domseason.New(
		number, fmt.Sprintf("Season %d", number),
		now, now.AddDate(0, 1, 0),
		domseason.DefaultSettings(), now,
	)
	require.NoError(t, err)
	s.Status = domseason.StatusActive
	return s
}

func newDivision(
	t *testing.T,
	seasonID shared.SeasonID,
	code string,
	order, min, max int,
	now time.Time,
) *division.Division {
	t.Helper()
	d, err := division.New(
		seasonID, code, code,
		shared.Rating(min), shared.Rating(max),
		order, now,
	)
	require.NoError(t, err)
	return d
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func realLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }
