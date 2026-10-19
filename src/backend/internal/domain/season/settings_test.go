package season_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/domain/season"
)

func TestDefaultSettings(t *testing.T) {
	t.Parallel()
	s := season.DefaultSettings()

	require.Equal(t, 1, s.K1)
	require.Equal(t, 2, s.K2)
	require.Equal(t, 2, s.K3)
	require.Equal(t, 8, s.P)
	require.Equal(t, 4, s.N)
	require.Equal(t, 12, s.T)
	require.Equal(t, 3, s.ChallengeAcceptDays)
	require.Equal(t, 3, s.MatchCancelDays)
	require.Equal(t, 30, s.SeasonDurationDays)
	require.Equal(t, 3, s.ClosingDays)
	require.Equal(t, 3, s.LPWin)
	require.Equal(t, 0, s.LPLoss)
	require.Equal(t, 3, s.StreakBonus3)
	require.Equal(t, 5, s.StreakBonus5)
	require.Equal(t, 6, s.StreakBonus67)
	require.Equal(t, 3.0, s.StonesWin)
	require.Equal(t, 0.0, s.StonesLoss)
	require.Equal(t, 5.0, s.StonesClubTop1)
	require.Equal(t, 67.0, s.StonesSuperTitle)
	require.Equal(t, 0.5, s.StonesPerExtraMatch)
}

func TestSettings_Validate(t *testing.T) {
	t.Parallel()
	valid := season.DefaultSettings()

	t.Run("успех: дефолтные настройки", func(t *testing.T) {
		t.Parallel()
		err := valid.Validate()
		require.NoError(t, err)
	})

	t.Run("ошибка: отрицательный P", func(t *testing.T) {
		t.Parallel()
		s := valid
		s.P = -1
		require.ErrorIs(t, s.Validate(), season.ErrInvalidSettings)
	})

	t.Run("ошибка: N = 0", func(t *testing.T) {
		t.Parallel()
		s := valid
		s.N = 0
		require.ErrorIs(t, s.Validate(), season.ErrInvalidSettings)
	})

	t.Run("ошибка: отрицательный T", func(t *testing.T) {
		t.Parallel()
		s := valid
		s.T = -1
		require.ErrorIs(t, s.Validate(), season.ErrInvalidSettings)
	})

	t.Run("ошибка: отрицательный K1", func(t *testing.T) {
		t.Parallel()
		s := valid
		s.K1 = -1
		require.ErrorIs(t, s.Validate(), season.ErrInvalidSettings)
	})

	t.Run("ошибка: streak_bonus_67 = 5", func(t *testing.T) {
		t.Parallel()
		s := valid
		s.StreakBonus67 = 5
		require.ErrorIs(t, s.Validate(), season.ErrInvalidSettings)
	})

	t.Run("успех: streak_bonus_67 = 7", func(t *testing.T) {
		t.Parallel()
		s := valid
		s.StreakBonus67 = 7
		require.NoError(t, s.Validate())
	})
}
