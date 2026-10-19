package division_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/domain/division"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

func TestNew(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("успех: минимальные параметры", func(t *testing.T) {
		t.Parallel()
		d, err := division.New(1, "A", "Дивизион A", 1000, 1500, 1, now)
		require.NoError(t, err)
		require.NotNil(t, d)
		require.Equal(t, shared.SeasonID(1), d.SeasonID)
		require.Equal(t, "A", d.Code)
		require.Equal(t, "Дивизион A", d.Alias)
		require.Equal(t, shared.Rating(1000), d.RatingMin)
		require.Equal(t, shared.Rating(1500), d.RatingMax)
		require.Equal(t, 1, d.Order)
		require.Equal(t, division.StatusActive, d.Status)
	})

	t.Run("успех: без максимального рейтинга", func(t *testing.T) {
		t.Parallel()
		d, err := division.New(1, "B", "Дивизион B", 0, 0, 2, now)
		require.NoError(t, err)
		require.Equal(t, shared.Rating(0), d.RatingMin)
		require.Equal(t, shared.Rating(0), d.RatingMax)
	})

	t.Run("ошибка: пустой код", func(t *testing.T) {
		t.Parallel()
		_, err := division.New(1, "", "Empty", 0, 0, 0, now)
		require.ErrorIs(t, err, shared.ErrValidation)
	})

	t.Run("ошибка: отрицательный order", func(t *testing.T) {
		t.Parallel()
		_, err := division.New(1, "A", "A", 0, 0, -1, now)
		require.ErrorIs(t, err, shared.ErrValidation)
	})

	t.Run("ошибка: max < min", func(t *testing.T) {
		t.Parallel()
		_, err := division.New(1, "A", "A", 1500, 1000, 0, now)
		require.ErrorIs(t, err, shared.ErrValidation)
	})
}

func TestDivision_MatchesRating(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("в диапазоне", func(t *testing.T) {
		t.Parallel()
		d, err := division.New(1, "A", "A", 1000, 1500, 0, now)
		require.NoError(t, err)
		require.True(t, d.MatchesRating(1000))
		require.True(t, d.MatchesRating(1250))
		require.True(t, d.MatchesRating(1500))
	})

	t.Run("ниже минимума", func(t *testing.T) {
		t.Parallel()
		d, err := division.New(1, "A", "A", 1000, 1500, 0, now)
		require.NoError(t, err)
		require.False(t, d.MatchesRating(999))
		require.False(t, d.MatchesRating(0))
	})

	t.Run("выше максимума", func(t *testing.T) {
		t.Parallel()
		d, err := division.New(1, "A", "A", 1000, 1500, 0, now)
		require.NoError(t, err)
		require.False(t, d.MatchesRating(1501))
		require.False(t, d.MatchesRating(2000))
	})

	t.Run("без максимума — любой >= min", func(t *testing.T) {
		t.Parallel()
		d, err := division.New(1, "S", "S", 2000, 0, 0, now)
		require.NoError(t, err)
		require.False(t, d.MatchesRating(1999))
		require.True(t, d.MatchesRating(2000))
		require.True(t, d.MatchesRating(3000))
	})

	t.Run("min=0, max=0 — любой рейтинг", func(t *testing.T) {
		t.Parallel()
		d, err := division.New(1, "B", "B", 0, 0, 0, now)
		require.NoError(t, err)
		require.True(t, d.MatchesRating(0))
		require.True(t, d.MatchesRating(500))
		require.True(t, d.MatchesRating(9999))
	})
}

func TestStatus_Valid(t *testing.T) {
	t.Parallel()
	require.True(t, division.StatusActive.Valid())
	require.True(t, division.StatusFinished.Valid())
	require.False(t, division.Status("unknown").Valid())
}
