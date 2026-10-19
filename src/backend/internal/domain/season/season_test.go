package season_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

func TestNew(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	start := now.Add(24 * time.Hour)
	end := now.Add(30 * 24 * time.Hour)
	settings := season.Settings{K1: 1, K2: 2, K3: 2, P: 8, N: 4, T: 12}

	t.Run("успех: корректные параметры", func(t *testing.T) {
		t.Parallel()
		s, err := season.New(1, "Сезон 1", start, end, settings, now)
		require.NoError(t, err)
		require.NotNil(t, s)
		require.Equal(t, 1, s.Number)
		require.Equal(t, "Сезон 1", s.Title)
		require.Equal(t, season.StatusAnnounced, s.Status)
		require.Equal(t, settings, s.Settings)
		require.True(t, s.StartDate.Equal(start))
		require.True(t, s.EndDate.Equal(end))
		require.True(t, s.CreatedAt.Equal(now))
		require.True(t, s.UpdatedAt.Equal(now))
		require.Zero(t, s.ID)
	})

	t.Run("ошибка: номер <= 0", func(t *testing.T) {
		t.Parallel()
		_, err := season.New(0, "Сезон 0", start, end, settings, now)
		require.ErrorIs(t, err, shared.ErrValidation)
	})

	t.Run("ошибка: пустое название", func(t *testing.T) {
		t.Parallel()
		_, err := season.New(1, "", start, end, settings, now)
		require.ErrorIs(t, err, shared.ErrValidation)
	})

	t.Run("ошибка: end <= start", func(t *testing.T) {
		t.Parallel()
		_, err := season.New(1, "Сезон 1", end, start, settings, now)
		require.ErrorIs(t, err, shared.ErrValidation)
	})

	t.Run("ошибка: end == start", func(t *testing.T) {
		t.Parallel()
		_, err := season.New(1, "Сезон 1", start, start, settings, now)
		require.ErrorIs(t, err, shared.ErrValidation)
	})
}

func TestSeason_StartRegistration(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("успех: announced → registration", func(t *testing.T) {
		t.Parallel()
		s := mustNewSeason(t, now)
		err := s.StartRegistration(now.Add(time.Hour))
		require.NoError(t, err)
		require.Equal(t, season.StatusRegistration, s.Status)
	})

	t.Run("ошибка: только из announced", func(t *testing.T) {
		t.Parallel()
		s := mustNewSeason(t, now)
		require.NoError(t, s.StartRegistration(now.Add(time.Hour)))
		err := s.StartRegistration(now.Add(2 * time.Hour))
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})
}

func TestSeason_Start(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("успех: registration → active", func(t *testing.T) {
		t.Parallel()
		s := mustNewSeason(t, now)
		require.NoError(t, s.StartRegistration(now.Add(time.Hour)))
		err := s.Start(now.Add(2 * time.Hour))
		require.NoError(t, err)
		require.Equal(t, season.StatusActive, s.Status)
	})

	t.Run("успех: announced → active (skip registration)", func(t *testing.T) {
		t.Parallel()
		s := mustNewSeason(t, now)
		err := s.Start(now.Add(time.Hour))
		require.NoError(t, err)
		require.Equal(t, season.StatusActive, s.Status)
	})

	t.Run("ошибка: из active", func(t *testing.T) {
		t.Parallel()
		s := mustNewSeason(t, now)
		require.NoError(t, s.Start(now.Add(time.Hour)))
		err := s.Start(now.Add(2 * time.Hour))
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})
}

func TestSeason_Close(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("успех: active → closing", func(t *testing.T) {
		t.Parallel()
		s := mustNewSeason(t, now)
		require.NoError(t, s.Start(now.Add(time.Hour)))
		err := s.Close(now.Add(2 * time.Hour))
		require.NoError(t, err)
		require.Equal(t, season.StatusClosing, s.Status)
	})

	t.Run("ошибка: только из active", func(t *testing.T) {
		t.Parallel()
		s := mustNewSeason(t, now)
		err := s.Close(now.Add(time.Hour))
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})
}

func TestSeason_Finish(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("успех: closing → finished", func(t *testing.T) {
		t.Parallel()
		s := mustNewSeason(t, now)
		require.NoError(t, s.Start(now.Add(time.Hour)))
		require.NoError(t, s.Close(now.Add(2*time.Hour)))
		err := s.Finish(now.Add(3 * time.Hour))
		require.NoError(t, err)
		require.Equal(t, season.StatusFinished, s.Status)
	})

	t.Run("успех: active → finished (skip closing)", func(t *testing.T) {
		t.Parallel()
		s := mustNewSeason(t, now)
		require.NoError(t, s.Start(now.Add(time.Hour)))
		err := s.Finish(now.Add(2 * time.Hour))
		require.NoError(t, err)
		require.Equal(t, season.StatusFinished, s.Status)
	})

	t.Run("ошибка: только из active/closing", func(t *testing.T) {
		t.Parallel()
		s := mustNewSeason(t, now)
		err := s.Finish(now.Add(time.Hour))
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})
}

func TestSeason_Archive(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("успех: finished → archived", func(t *testing.T) {
		t.Parallel()
		s := mustNewSeason(t, now)
		require.NoError(t, s.Start(now.Add(time.Hour)))
		require.NoError(t, s.Finish(now.Add(2*time.Hour)))
		err := s.Archive(now.Add(3 * time.Hour))
		require.NoError(t, err)
		require.Equal(t, season.StatusArchived, s.Status)
	})

	t.Run("ошибка: только из finished", func(t *testing.T) {
		t.Parallel()
		s := mustNewSeason(t, now)
		require.NoError(t, s.Start(now.Add(time.Hour)))
		err := s.Archive(now.Add(2 * time.Hour))
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})
}

func TestStatus_CanJoin(t *testing.T) {
	t.Parallel()
	require.True(t, season.StatusAnnounced.CanJoin())
	require.True(t, season.StatusRegistration.CanJoin())
	require.True(t, season.StatusActive.CanJoin())
	require.False(t, season.StatusClosing.CanJoin())
	require.False(t, season.StatusFinished.CanJoin())
	require.False(t, season.StatusArchived.CanJoin())
}

func TestStatus_CanCreateMatch(t *testing.T) {
	t.Parallel()
	require.False(t, season.StatusAnnounced.CanCreateMatch())
	require.False(t, season.StatusRegistration.CanCreateMatch())
	require.True(t, season.StatusActive.CanCreateMatch())
	require.False(t, season.StatusClosing.CanCreateMatch())
	require.False(t, season.StatusFinished.CanCreateMatch())
	require.False(t, season.StatusArchived.CanCreateMatch())
}

func TestStatus_CanConfirmMatch(t *testing.T) {
	t.Parallel()
	require.False(t, season.StatusAnnounced.CanConfirmMatch())
	require.False(t, season.StatusRegistration.CanConfirmMatch())
	require.True(t, season.StatusActive.CanConfirmMatch())
	require.True(t, season.StatusClosing.CanConfirmMatch())
	require.False(t, season.StatusFinished.CanConfirmMatch())
	require.False(t, season.StatusArchived.CanConfirmMatch())
}

func TestStatus_Valid(t *testing.T) {
	t.Parallel()
	require.True(t, season.StatusAnnounced.Valid())
	require.True(t, season.StatusRegistration.Valid())
	require.True(t, season.StatusActive.Valid())
	require.True(t, season.StatusClosing.Valid())
	require.True(t, season.StatusFinished.Valid())
	require.True(t, season.StatusArchived.Valid())
	require.False(t, season.Status("unknown").Valid())
}

func TestSeason_IsDateWithinSeason(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	start := now.Add(24 * time.Hour)
	end := now.Add(30 * 24 * time.Hour)
	s := mustNewSeasonWithDates(t, now, start, end)

	require.True(t, s.IsDateWithinSeason(start))
	require.True(t, s.IsDateWithinSeason(end))
	require.True(t, s.IsDateWithinSeason(start.Add(15*24*time.Hour)))
	require.False(t, s.IsDateWithinSeason(start.Add(-time.Hour)))
	require.False(t, s.IsDateWithinSeason(end.Add(time.Hour)))
}

// ---- helpers ----------------------------------------------------------------

func mustNewSeason(t *testing.T, now time.Time) *season.Season {
	t.Helper()
	start := now.Add(24 * time.Hour)
	end := now.Add(30 * 24 * time.Hour)
	return mustNewSeasonWithDates(t, now, start, end)
}

func mustNewSeasonWithDates(t *testing.T, now, start, end time.Time) *season.Season {
	t.Helper()
	settings := season.Settings{K1: 1, K2: 2, K3: 2, P: 8, N: 4, T: 12}
	s, err := season.New(1, "Сезон 1", start, end, settings, now)
	require.NoError(t, err)
	return s
}
