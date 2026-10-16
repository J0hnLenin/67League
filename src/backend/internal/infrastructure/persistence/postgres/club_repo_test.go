package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	domclub "github.com/J0hnLenin/67League/internal/domain/club"
	"github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/infrastructure/persistence/postgres"
)

func mkPlayer(t *testing.T, nickname, email string) *player.Player {
	t.Helper()
	p, err := player.New(player.NewPlayerParams{
		FullName:     "Tester",
		Email:        shared.Email(email),
		PasswordHash: "h",
		Nickname:     shared.Nickname(nickname),
		City:         "Moscow",
	}, time.Now().UTC())
	require.NoError(t, err)
	return p
}

func TestClubRepo_CRUD(t *testing.T) {
	truncateAll(t)
	pr := postgres.NewPlayerRepo(testPool)
	cr := postgres.NewClubRepo(testPool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	creator := mkPlayer(t, "alice", "a@x.com")
	require.NoError(t, pr.Create(context.Background(), creator))

	c, err := domclub.New("Shibuya", "Tokyo", "Waseda", creator.ID, now)
	require.NoError(t, err)
	require.NoError(t, cr.Create(context.Background(), c))
	require.NotZero(t, c.ID)

	got, err := cr.GetByID(context.Background(), c.ID)
	require.NoError(t, err)
	require.Equal(t, "Shibuya", got.Name)

	byName, err := cr.GetByName(context.Background(), "Shibuya")
	require.NoError(t, err)
	require.Equal(t, c.ID, byName.ID)

	c.City = "Kyoto"
	c.UpdatedAt = now.Add(time.Second)
	require.NoError(t, cr.Update(context.Background(), c))

	got, _ = cr.GetByID(context.Background(), c.ID)
	require.Equal(t, "Kyoto", got.City)
}

func TestClubRepo_DuplicateName(t *testing.T) {
	truncateAll(t)
	pr := postgres.NewPlayerRepo(testPool)
	cr := postgres.NewClubRepo(testPool)
	now := time.Now().UTC()

	creator := mkPlayer(t, "alice", "a@x.com")
	require.NoError(t, pr.Create(context.Background(), creator))

	c1, _ := domclub.New("Shibuya", "Tokyo", "", creator.ID, now)
	c2, _ := domclub.New("Shibuya", "Kyoto", "", creator.ID, now)
	require.NoError(t, cr.Create(context.Background(), c1))

	err := cr.Create(context.Background(), c2)
	require.True(t, errors.Is(err, shared.ErrAlreadyExists))
}

func TestClubRepo_ExistsByName(t *testing.T) {
	truncateAll(t)
	pr := postgres.NewPlayerRepo(testPool)
	cr := postgres.NewClubRepo(testPool)
	now := time.Now().UTC()

	creator := mkPlayer(t, "alice", "a@x.com")
	require.NoError(t, pr.Create(context.Background(), creator))

	exists, err := cr.ExistsByName(context.Background(), "Nope")
	require.NoError(t, err)
	require.False(t, exists)

	c, _ := domclub.New("Shibuya", "Tokyo", "", creator.ID, now)
	require.NoError(t, cr.Create(context.Background(), c))

	exists, _ = cr.ExistsByName(context.Background(), "Shibuya")
	require.True(t, exists)
}
