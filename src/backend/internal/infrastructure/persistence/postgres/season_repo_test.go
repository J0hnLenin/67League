package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	domseason "github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/infrastructure/persistence/postgres"
)

func TestSeasonRepo_CRUDAndSettings(t *testing.T) {
	truncateAll(t)
	repo := postgres.NewSeasonRepo(testPool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	settings := domseason.DefaultSettings()
	settings.K1 = 2
	settings.N = 5

	s, err := domseason.New(1, "S1", now, now.AddDate(0, 1, 0), settings, now)
	require.NoError(t, err)
	require.NoError(t, repo.Create(context.Background(), s))
	require.NotZero(t, s.ID)

	got, err := repo.GetByID(context.Background(), s.ID)
	require.NoError(t, err)
	require.Equal(t, "S1", got.Title)
	require.Equal(t, 2, got.Settings.K1)
	require.Equal(t, 5, got.Settings.N)
	require.Equal(t, domseason.StatusAnnounced, got.Status)
}

func TestSeasonRepo_GetCurrentAndLatest(t *testing.T) {
	truncateAll(t)
	repo := postgres.NewSeasonRepo(testPool)
	now := time.Now().UTC()

	mk := func(n int, status domseason.Status) *domseason.Season {
		s, _ := domseason.New(n, "S", now, now.AddDate(0, 1, 0), domseason.DefaultSettings(), now)
		s.Status = status
		require.NoError(t, repo.Create(context.Background(), s))
		return s
	}

	s1 := mk(1, domseason.StatusArchived)
	s2 := mk(2, domseason.StatusFinished)
	s3 := mk(3, domseason.StatusActive)

	cur, err := repo.GetCurrent(context.Background())
	require.NoError(t, err)
	require.Equal(t, s3.ID, cur.ID)

	latest, err := repo.GetLatest(context.Background())
	require.NoError(t, err)
	require.Equal(t, s3.ID, latest.ID)

	// Отмечаем 3 как архив — текущим должен стать 2.
	s3.Status = domseason.StatusArchived
	s3.UpdatedAt = now.Add(time.Second)
	require.NoError(t, repo.Update(context.Background(), s3))

	cur, _ = repo.GetCurrent(context.Background())
	require.Equal(t, s2.ID, cur.ID)

	// Latest всё равно 3.
	latest, _ = repo.GetLatest(context.Background())
	require.Equal(t, s3.ID, latest.ID)

	_ = s1
}

func TestSeasonRepo_NotFound(t *testing.T) {
	truncateAll(t)
	repo := postgres.NewSeasonRepo(testPool)

	_, err := repo.GetByID(context.Background(), shared.SeasonID(999))
	require.True(t, errors.Is(err, shared.ErrNotFound))

	_, err = repo.GetCurrent(context.Background())
	require.True(t, errors.Is(err, shared.ErrNotFound))

	_, err = repo.GetLatest(context.Background())
	require.True(t, errors.Is(err, shared.ErrNotFound))
}
