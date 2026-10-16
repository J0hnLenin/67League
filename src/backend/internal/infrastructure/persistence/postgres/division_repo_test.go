package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/domain/division"
	domseason "github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/infrastructure/persistence/postgres"
)

func mkSeason(t *testing.T, number int) *domseason.Season {
	t.Helper()
	sr := postgres.NewSeasonRepo(testPool)
	now := time.Now().UTC()
	s, err := domseason.New(number, "S", now, now.AddDate(0, 1, 0), domseason.DefaultSettings(), now)
	require.NoError(t, err)
	require.NoError(t, sr.Create(context.Background(), s))
	return s
}

func TestDivisionRepo_CRUDAndQueries(t *testing.T) {
	truncateAll(t)
	repo := postgres.NewDivisionRepo(testPool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	s := mkSeason(t, 1)

	mk := func(order int, code string, min, max int) *division.Division {
		d, err := division.New(s.ID, code, code, shared.Rating(min), shared.Rating(max), order, now)
		require.NoError(t, err)
		require.NoError(t, repo.Create(context.Background(), d))
		require.NotZero(t, d.ID)
		return d
	}

	sDiv := mk(0, "S", 1000, 2000)
	aDiv := mk(1, "A", 500, 999)
	bDiv := mk(2, "B", 0, 499)

	list, err := repo.ListBySeason(context.Background(), s.ID)
	require.NoError(t, err)
	require.Len(t, list, 3)
	require.Equal(t, "S", list[0].Code)
	require.Equal(t, "B", list[2].Code)

	// FindForRating
	got, err := repo.FindForRating(context.Background(), s.ID, shared.Rating(1500))
	require.NoError(t, err)
	require.Equal(t, sDiv.ID, got.ID)

	got, _ = repo.FindForRating(context.Background(), s.ID, shared.Rating(700))
	require.Equal(t, aDiv.ID, got.ID)

	got, _ = repo.FindForRating(context.Background(), s.ID, shared.Rating(100))
	require.Equal(t, bDiv.ID, got.ID)

	// Не найдено
	_, err = repo.FindForRating(context.Background(), s.ID, shared.Rating(5000))
	require.True(t, errors.Is(err, shared.ErrNotFound))

	// GetLowest / GetHighest
	low, err := repo.GetLowest(context.Background(), s.ID)
	require.NoError(t, err)
	require.Equal(t, bDiv.ID, low.ID)

	high, err := repo.GetHighest(context.Background(), s.ID)
	require.NoError(t, err)
	require.Equal(t, sDiv.ID, high.ID)

	// Update
	aDiv.RatingMax = 1000
	aDiv.UpdatedAt = now.Add(time.Second)
	require.NoError(t, repo.Update(context.Background(), aDiv))

	got, _ = repo.GetByID(context.Background(), aDiv.ID)
	require.Equal(t, shared.Rating(1000), got.RatingMax)
}

func TestDivisionRepo_DuplicateCodeInSeason(t *testing.T) {
	truncateAll(t)
	repo := postgres.NewDivisionRepo(testPool)
	now := time.Now().UTC()

	s := mkSeason(t, 1)

	d1, _ := division.New(s.ID, "S", "S", 0, 0, 0, now)
	require.NoError(t, repo.Create(context.Background(), d1))

	d2, _ := division.New(s.ID, "S", "S again", 0, 0, 1, now)
	err := repo.Create(context.Background(), d2)
	require.True(t, errors.Is(err, shared.ErrAlreadyExists))
}
