package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/application/ports"
	dombadge "github.com/J0hnLenin/67League/internal/domain/badge"
	domclub "github.com/J0hnLenin/67League/internal/domain/club"
	"github.com/J0hnLenin/67League/internal/domain/economy"
	"github.com/J0hnLenin/67League/internal/domain/notification"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/infrastructure/persistence/postgres"
)

// ---- Badge / Title --------------------------------------------------------

func TestBadgeRepo_CRUD(t *testing.T) {
	truncateAll(t)
	repo := postgres.NewBadgeRepo(testPool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	b := &dombadge.Badge{
		Code: "first_stone", Icon: "⚫", Name: "Первый камень",
		Description: "Первая партия", Type: dombadge.TypePersonal,
		Rarity: dombadge.RarityCommon, Condition: map[string]any{"matches": 1},
		CreatedAt: now,
	}
	require.NoError(t, repo.Create(context.Background(), b))
	require.NotZero(t, b.ID)

	got, err := repo.GetByID(context.Background(), b.ID)
	require.NoError(t, err)
	require.Equal(t, "first_stone", got.Code)

	byCode, err := repo.GetByCode(context.Background(), "first_stone")
	require.NoError(t, err)
	require.Equal(t, b.ID, byCode.ID)

	_, err = repo.GetByCode(context.Background(), "nope")
	require.True(t, errors.Is(err, shared.ErrNotFound))

	// Duplicate
	dup := *b
	dup.ID = 0
	err = repo.Create(context.Background(), &dup)
	require.True(t, errors.Is(err, shared.ErrAlreadyExists))
}

func TestTitleRepo_CRUD(t *testing.T) {
	truncateAll(t)
	repo := postgres.NewTitleRepo(testPool)

	t1 := &dombadge.Title{Code: "stone_mason", Name: "Каменотёс"}
	require.NoError(t, repo.Create(context.Background(), t1))
	require.NotZero(t, t1.ID)

	got, err := repo.GetByCode(context.Background(), "stone_mason")
	require.NoError(t, err)
	require.Equal(t, t1.ID, got.ID)

	_, err = repo.GetByID(context.Background(), shared.TitleID(999))
	require.True(t, errors.Is(err, shared.ErrNotFound))
}

// ---- PlayerBadge / ClubBadge / SuperTitle ---------------------------------

func TestPlayerBadgeRepo(t *testing.T) {
	truncateAll(t)
	pr := postgres.NewPlayerRepo(testPool)
	br := postgres.NewBadgeRepo(testPool)
	repo := postgres.NewPlayerBadgeRepo(testPool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	p := mkPlayer(t, "alice", "a@x.com")
	require.NoError(t, pr.Create(context.Background(), p))

	b := &dombadge.Badge{
		Code: "x", Name: "X", Type: dombadge.TypePersonal,
		Rarity: dombadge.RarityCommon, CreatedAt: now,
	}
	require.NoError(t, br.Create(context.Background(), b))

	pb := dombadge.NewPlayerBadge(p.ID, b.ID, nil, now)
	require.NoError(t, repo.Create(context.Background(), pb))
	require.NotZero(t, pb.ID)

	has, err := repo.Has(context.Background(), p.ID, b.ID)
	require.NoError(t, err)
	require.True(t, has)

	list, err := repo.ListByPlayer(context.Background(), p.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
}

func TestClubBadgeRepo(t *testing.T) {
	truncateAll(t)
	pr := postgres.NewPlayerRepo(testPool)
	cr := postgres.NewClubRepo(testPool)
	br := postgres.NewBadgeRepo(testPool)
	repo := postgres.NewClubBadgeRepo(testPool)
	now := time.Now().UTC()

	p := mkPlayer(t, "alice", "a@x.com")
	require.NoError(t, pr.Create(context.Background(), p))

	c, _ := domclubNew("Shibuya", "Tokyo", p.ID, now)
	require.NoError(t, cr.Create(context.Background(), c))

	b := &dombadge.Badge{
		Code: "founding", Name: "Основание",
		Type: dombadge.TypeClub, Rarity: dombadge.RarityCommon, CreatedAt: now,
	}
	require.NoError(t, br.Create(context.Background(), b))

	cb := dombadge.NewClubBadge(c.ID, b.ID, nil, now)
	require.NoError(t, repo.Create(context.Background(), cb))

	has, _ := repo.Has(context.Background(), c.ID, b.ID)
	require.True(t, has)
}

func TestSuperTitleRepo(t *testing.T) {
	truncateAll(t)
	pr := postgres.NewPlayerRepo(testPool)
	cr := postgres.NewClubRepo(testPool)
	sr := postgres.NewSeasonRepo(testPool)
	repo := postgres.NewSuperTitleRepo(testPool)
	now := time.Now().UTC()

	p := mkPlayer(t, "alice", "a@x.com")
	require.NoError(t, pr.Create(context.Background(), p))

	c, _ := domclubNew("Shibuya", "Tokyo", p.ID, now)
	require.NoError(t, cr.Create(context.Background(), c))

	s := mkSeason(t, 1)
	_ = sr

	st := dombadge.NewSuperTitle(s.ID, p.ID, c.ID, now)
	require.NoError(t, repo.Create(context.Background(), st))
	require.NotZero(t, st.ID)

	got, err := repo.GetBySeason(context.Background(), s.ID)
	require.NoError(t, err)
	require.Equal(t, st.ID, got.ID)

	// Duplicate сезон
	dup := dombadge.NewSuperTitle(s.ID, p.ID, c.ID, now)
	err = repo.Create(context.Background(), dup)
	require.True(t, errors.Is(err, shared.ErrAlreadyExists))
}

// ---- Economy --------------------------------------------------------------

func TestStoneTransactionRepo(t *testing.T) {
	truncateAll(t)
	pr := postgres.NewPlayerRepo(testPool)
	repo := postgres.NewStoneTransactionRepo(testPool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	p := mkPlayer(t, "alice", "a@x.com")
	require.NoError(t, pr.Create(context.Background(), p))

	tx := economy.NewTransaction(p.ID, 5.5, "match:1", nil, now)
	require.NoError(t, repo.Create(context.Background(), tx))
	require.NotZero(t, tx.ID)

	list, err := repo.ListByPlayer(context.Background(), p.ID, portsDefaultPagination())
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, shared.Stones(5.5), list[0].Amount)
}

func TestPurchaseRepo(t *testing.T) {
	truncateAll(t)
	pr := postgres.NewPlayerRepo(testPool)
	repo := postgres.NewPurchaseRepo(testPool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	p := mkPlayer(t, "alice", "a@x.com")
	require.NoError(t, pr.Create(context.Background(), p))

	pur, err := economy.NewPurchase(p.ID, "avatar_pack_1", 100, now)
	require.NoError(t, err)
	require.NoError(t, repo.Create(context.Background(), pur))

	list, _ := repo.ListByPlayer(context.Background(), p.ID, portsDefaultPagination())
	require.Len(t, list, 1)
	require.Equal(t, "avatar_pack_1", list[0].ItemCode)
}

// ---- Notification ---------------------------------------------------------

func TestNotificationRepo(t *testing.T) {
	truncateAll(t)
	pr := postgres.NewPlayerRepo(testPool)
	repo := postgres.NewNotificationRepo(testPool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	p := mkPlayer(t, "alice", "a@x.com")
	require.NoError(t, pr.Create(context.Background(), p))

	n := notification.New(p.ID, notification.TypeNewChallenge,
		map[string]any{"match_id": 42}, notification.ChannelEmail, now)
	require.NoError(t, repo.Create(context.Background(), n))
	require.NotZero(t, n.ID)

	unsent, err := repo.ListUnsent(context.Background(), 10)
	require.NoError(t, err)
	require.Len(t, unsent, 1)
	require.Equal(t, shared.PlayerID(p.ID), unsent[0].PlayerID)

	// Отмечаем отправленным.
	sent := now.Add(time.Second)
	n.SentAt = &sent
	require.NoError(t, repo.Update(context.Background(), n))

	unsent, _ = repo.ListUnsent(context.Background(), 10)
	require.Empty(t, unsent)

	list, _ := repo.ListByPlayer(context.Background(), p.ID, portsDefaultPagination())
	require.Len(t, list, 1)
	require.NotNil(t, list[0].SentAt)
}

// ---- Settings -------------------------------------------------------------

func TestSettingsRepo(t *testing.T) {
	truncateAll(t)
	repo := postgres.NewSettingsRepo(testPool)
	ctx := context.Background()

	// Отсутствует
	_, err := repo.Get(ctx, "site_name")
	require.True(t, errors.Is(err, shared.ErrNotFound))

	// Set + Get
	require.NoError(t, repo.Set(ctx, "site_name", "6-7 Лига"))
	v, err := repo.Get(ctx, "site_name")
	require.NoError(t, err)
	require.Equal(t, "6-7 Лига", v)

	// Update
	require.NoError(t, repo.Set(ctx, "site_name", "6-7"))
	v, _ = repo.Get(ctx, "site_name")
	require.Equal(t, "6-7", v)

	// SetMany + GetAll
	require.NoError(t, repo.SetMany(ctx, map[string]any{
		"a": 1, "b": true,
	}))
	all, err := repo.GetAll(ctx)
	require.NoError(t, err)
	require.Equal(t, "6-7", all["site_name"])
	require.EqualValues(t, 1, all["a"])
	require.Equal(t, true, all["b"])
}

func portsDefaultPagination() ports.Pagination {
	return ports.DefaultPagination()
}

func domclubNew(name, city string, creator shared.PlayerID, now time.Time) (*domclub.Club, error) {
	return domclub.New(name, city, "", creator, now)
}
