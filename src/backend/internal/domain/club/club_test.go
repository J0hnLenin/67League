package club_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/domain/club"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

func TestNew(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("успех: минимальные параметры", func(t *testing.T) {
		t.Parallel()
		c, err := club.New("Shibuya", "Moscow", "", 10, now)
		require.NoError(t, err)
		require.NotNil(t, c)
		require.Equal(t, "Shibuya", c.Name)
		require.Equal(t, "Moscow", c.City)
		require.Equal(t, "", c.University)
		require.Equal(t, shared.PlayerID(10), c.CreatedBy)
		require.Equal(t, club.StatusActive, c.Status)
		require.True(t, c.CreatedAt.Equal(now))
		require.True(t, c.UpdatedAt.Equal(now))
		require.Zero(t, c.ID)
	})

	t.Run("успех: со всеми полями", func(t *testing.T) {
		t.Parallel()
		c, err := club.New("Go Club", "SPb", "SPbSU", 20, now)
		require.NoError(t, err)
		require.Equal(t, "Go Club", c.Name)
		require.Equal(t, "SPb", c.City)
		require.Equal(t, "SPbSU", c.University)
	})

	t.Run("ошибка: пустое название", func(t *testing.T) {
		t.Parallel()
		_, err := club.New("", "Moscow", "", 10, now)
		require.ErrorIs(t, err, shared.ErrValidation)
	})
}

func TestClub_Archive(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("успех: архивация", func(t *testing.T) {
		t.Parallel()
		c, err := club.New("Shibuya", "Moscow", "", 10, now)
		require.NoError(t, err)
		err = c.Archive(now.Add(time.Hour))
		require.NoError(t, err)
		require.Equal(t, club.StatusArchived, c.Status)
	})

	t.Run("ошибка: повторная архивация", func(t *testing.T) {
		t.Parallel()
		c, err := club.New("Shibuya", "Moscow", "", 10, now)
		require.NoError(t, err)
		require.NoError(t, c.Archive(now.Add(time.Hour)))
		err = c.Archive(now.Add(2 * time.Hour))
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})
}

func TestStatus_Valid(t *testing.T) {
	t.Parallel()
	require.True(t, club.StatusActive.Valid())
	require.True(t, club.StatusArchived.Valid())
	require.False(t, club.Status("unknown").Valid())
}
