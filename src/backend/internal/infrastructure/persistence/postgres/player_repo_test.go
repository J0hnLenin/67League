package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/infrastructure/persistence/postgres"
)

func TestPlayerRepo_CreateAndGet(t *testing.T) {
	truncateAll(t)

	repo := postgres.NewPlayerRepo(testPool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	p, err := player.New(player.NewPlayerParams{
		FullName:     "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed:secret",
		Nickname:     "alice",
		City:         "Moscow",
	}, now)
	require.NoError(t, err)

	require.NoError(t, repo.Create(context.Background(), p))
	require.NotZero(t, p.ID)

	got, err := repo.GetByID(context.Background(), p.ID)
	require.NoError(t, err)
	require.Equal(t, p.Email, got.Email)
	require.Equal(t, p.Nickname, got.Nickname)
	require.Equal(t, player.StatusActive, got.Status)
	require.Equal(t, shared.Rating(0), got.Rating)
}

func TestPlayerRepo_DuplicateEmail(t *testing.T) {
	truncateAll(t)

	repo := postgres.NewPlayerRepo(testPool)
	now := time.Now().UTC()

	mk := func(nick string) *player.Player {
		p, _ := player.New(player.NewPlayerParams{
			FullName: "A", Email: "a@x.com", PasswordHash: "h",
			Nickname: shared.Nickname(nick), City: "M",
		}, now)
		return p
	}

	require.NoError(t, repo.Create(context.Background(), mk("alice")))

	err := repo.Create(context.Background(), mk("another"))
	require.Error(t, err)
	require.True(t, errors.Is(err, shared.ErrAlreadyExists))
}

func TestPlayerRepo_ExistsByEmail(t *testing.T) {
	truncateAll(t)

	repo := postgres.NewPlayerRepo(testPool)
	now := time.Now().UTC()

	exists, err := repo.ExistsByEmail(context.Background(), "missing@x.com")
	require.NoError(t, err)
	require.False(t, exists)

	p, _ := player.New(player.NewPlayerParams{
		FullName: "A", Email: "a@x.com", PasswordHash: "h",
		Nickname: "a", City: "M",
	}, now)
	require.NoError(t, repo.Create(context.Background(), p))

	exists, err = repo.ExistsByEmail(context.Background(), "a@x.com")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestPlayerRepo_Update(t *testing.T) {
	truncateAll(t)

	repo := postgres.NewPlayerRepo(testPool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	p, _ := player.New(player.NewPlayerParams{
		FullName: "A", Email: "a@x.com", PasswordHash: "h",
		Nickname: "a", City: "Moscow",
	}, now)
	require.NoError(t, repo.Create(context.Background(), p))

	p.City = "Kyoto"
	p.Stones = 42.5
	p.UpdatedAt = now.Add(time.Second)
	require.NoError(t, repo.Update(context.Background(), p))

	got, err := repo.GetByID(context.Background(), p.ID)
	require.NoError(t, err)
	require.Equal(t, "Kyoto", got.City)
	require.Equal(t, shared.Stones(42.5), got.Stones)
}
