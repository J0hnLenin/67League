package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/domain/division"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/infrastructure/persistence/postgres"
)

// Вспомогательная setup-функция: создаёт сезон, дивизион, игрока.
func setupDivisionAndPlayer(t *testing.T) (*division.Division, shared.PlayerID, *postgres.DivisionMemberRepo) {
	t.Helper()
	s := mkSeason(t, 1)
	dr := postgres.NewDivisionRepo(testPool)
	d, err := division.New(s.ID, "A", "A", 0, 2000, 1, time.Now().UTC())
	require.NoError(t, err)
	require.NoError(t, dr.Create(context.Background(), d))

	pr := postgres.NewPlayerRepo(testPool)
	p := mkPlayer(t, "alice", "a@x.com")
	require.NoError(t, pr.Create(context.Background(), p))

	return d, p.ID, postgres.NewDivisionMemberRepo(testPool)
}

func TestDivisionMemberRepo_CRUD(t *testing.T) {
	truncateAll(t)
	d, pid, repo := setupDivisionAndPlayer(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	m := division.NewMember(d.ID, pid, now)
	require.NoError(t, repo.Create(context.Background(), m))
	require.NotZero(t, m.ID)

	got, err := repo.GetByID(context.Background(), m.ID)
	require.NoError(t, err)
	require.Equal(t, pid, got.PlayerID)
	require.Equal(t, division.MemberActive, got.Status)

	// Update: меняем LP и статистику.
	m.LP = 12.5
	m.Wins = 3
	m.MatchesPlayed = 5
	m.StreakMax = 2
	m.UpdatedAt = now.Add(time.Second)
	require.NoError(t, repo.Update(context.Background(), m))

	got, _ = repo.GetByID(context.Background(), m.ID)
	require.Equal(t, shared.LP(12.5), got.LP)
	require.Equal(t, 3, got.Wins)
	require.Equal(t, 5, got.MatchesPlayed)
}

func TestDivisionMemberRepo_GetByPlayerAndSeason(t *testing.T) {
	truncateAll(t)
	d, pid, repo := setupDivisionAndPlayer(t)
	now := time.Now().UTC()

	m := division.NewMember(d.ID, pid, now)
	require.NoError(t, repo.Create(context.Background(), m))

	// Через сезон d.SeasonID — успех.
	got, err := repo.GetByPlayerAndSeason(context.Background(), pid, d.SeasonID)
	require.NoError(t, err)
	require.Equal(t, m.ID, got.ID)

	// Через несуществующий сезон — ErrNotFound.
	_, err = repo.GetByPlayerAndSeason(context.Background(), pid, shared.SeasonID(999))
	require.True(t, errors.Is(err, shared.ErrNotFound))
}

func TestDivisionMemberRepo_ListByDivision(t *testing.T) {
	truncateAll(t)
	d, pid, repo := setupDivisionAndPlayer(t)

	pr := postgres.NewPlayerRepo(testPool)
	p2 := mkPlayer(t, "bob", "b@x.com")
	require.NoError(t, pr.Create(context.Background(), p2))

	now := time.Now().UTC()
	m1 := division.NewMember(d.ID, pid, now)
	m2 := division.NewMember(d.ID, p2.ID, now)
	require.NoError(t, repo.Create(context.Background(), m1))
	require.NoError(t, repo.Create(context.Background(), m2))

	list, err := repo.ListByDivision(context.Background(), d.ID)
	require.NoError(t, err)
	require.Len(t, list, 2)
}

func TestDivisionMemberRepo_DuplicateMembership(t *testing.T) {
	truncateAll(t)
	d, pid, repo := setupDivisionAndPlayer(t)
	now := time.Now().UTC()

	m1 := division.NewMember(d.ID, pid, now)
	m2 := division.NewMember(d.ID, pid, now)
	require.NoError(t, repo.Create(context.Background(), m1))

	err := repo.Create(context.Background(), m2)
	require.True(t, errors.Is(err, shared.ErrAlreadyExists))
}
