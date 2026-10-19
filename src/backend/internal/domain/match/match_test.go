package match_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/domain/match"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

func TestNewMatch(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	validParams := match.NewMatchParams{
		SeasonID:      1,
		DivisionID:    2,
		BlackPlayerID: 10,
		WhitePlayerID: 20,
		CreatedBy:     10,
		Server:        match.ServerOGS,
		TimeControl:   "byoyomi:30:3x30",
		Handicap:      0,
		Komi:          6.5,
	}

	t.Run("успех: корректные параметры", func(t *testing.T) {
		t.Parallel()
		m, err := match.New(validParams, now)
		require.NoError(t, err)
		require.NotNil(t, m)
		require.Equal(t, match.StatusProposed, m.Status)
		require.Equal(t, match.ServerOGS, m.Server)
		require.Equal(t, shared.PlayerID(10), m.BlackPlayerID)
		require.Equal(t, shared.PlayerID(20), m.WhitePlayerID)
		require.Equal(t, shared.PlayerID(10), m.CreatedBy)
		require.Equal(t, 0, m.Handicap)
		require.Equal(t, 6.5, m.Komi)
		require.True(t, m.CreatedAt.Equal(now))
		require.Nil(t, m.WinnerID)
		require.Nil(t, m.ScoreDiff)
		require.Nil(t, m.AcceptedAt)
		require.Nil(t, m.StartedAt)
		require.Nil(t, m.FinishedAt)
		require.Nil(t, m.ConfirmedAt)
		require.Nil(t, m.CancelledAt)
	})

	t.Run("ошибка: одинаковые игроки", func(t *testing.T) {
		t.Parallel()
		p := validParams
		p.WhitePlayerID = p.BlackPlayerID
		_, err := match.New(p, now)
		require.ErrorIs(t, err, shared.ErrValidation)
	})

	t.Run("ошибка: невалидный сервер", func(t *testing.T) {
		t.Parallel()
		p := validParams
		p.Server = "unknown"
		_, err := match.New(p, now)
		require.ErrorIs(t, err, shared.ErrValidation)
	})

	t.Run("ошибка: создатель не участник", func(t *testing.T) {
		t.Parallel()
		p := validParams
		p.CreatedBy = 99
		_, err := match.New(p, now)
		require.ErrorIs(t, err, shared.ErrForbidden)
	})

	t.Run("успех: создатель = белые", func(t *testing.T) {
		t.Parallel()
		p := validParams
		p.CreatedBy = 20
		m, err := match.New(p, now)
		require.NoError(t, err)
		require.Equal(t, shared.PlayerID(20), m.CreatedBy)
	})

	t.Run("успех: offline сервер", func(t *testing.T) {
		t.Parallel()
		p := validParams
		p.Server = match.ServerOffline
		m, err := match.New(p, now)
		require.NoError(t, err)
		require.Equal(t, match.ServerOffline, m.Server)
	})

	t.Run("успех: KGS сервер", func(t *testing.T) {
		t.Parallel()
		p := validParams
		p.Server = match.ServerKGS
		m, err := match.New(p, now)
		require.NoError(t, err)
		require.Equal(t, match.ServerKGS, m.Server)
	})

	t.Run("успех: Fox сервер", func(t *testing.T) {
		t.Parallel()
		p := validParams
		p.Server = match.ServerFox
		m, err := match.New(p, now)
		require.NoError(t, err)
		require.Equal(t, match.ServerFox, m.Server)
	})
}

func TestMatch_Accept(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("успех: соперник принимает вызов", func(t *testing.T) {
		t.Parallel()
		m := mustNewMatch(t, now)
		err := m.Accept(20, now.Add(time.Hour))
		require.NoError(t, err)
		require.Equal(t, match.StatusAccepted, m.Status)
		require.NotNil(t, m.AcceptedAt)
		require.True(t, m.AcceptedAt.Equal(now.Add(time.Hour)))
	})

	t.Run("ошибка: создатель не может принять свой вызов", func(t *testing.T) {
		t.Parallel()
		m := mustNewMatch(t, now)
		err := m.Accept(10, now.Add(time.Hour))
		require.ErrorIs(t, err, shared.ErrForbidden)
	})

	t.Run("ошибка: посторонний не может принять", func(t *testing.T) {
		t.Parallel()
		m := mustNewMatch(t, now)
		err := m.Accept(99, now.Add(time.Hour))
		require.ErrorIs(t, err, shared.ErrForbidden)
	})

	t.Run("ошибка: повторный accept", func(t *testing.T) {
		t.Parallel()
		m := mustNewMatch(t, now)
		require.NoError(t, m.Accept(20, now.Add(time.Hour)))
		err := m.Accept(20, now.Add(2*time.Hour))
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})

	t.Run("ошибка: accept после start", func(t *testing.T) {
		t.Parallel()
		m := mustNewMatch(t, now)
		require.NoError(t, m.Accept(20, now.Add(time.Hour)))
		require.NoError(t, m.Start(10, now.Add(2*time.Hour)))
		err := m.Accept(20, now.Add(3*time.Hour))
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})
}

func TestMatch_Start(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("успех: чёрные отмечают начало", func(t *testing.T) {
		t.Parallel()
		m := mustNewMatch(t, now)
		require.NoError(t, m.Accept(20, now.Add(time.Hour)))
		err := m.Start(10, now.Add(2*time.Hour))
		require.NoError(t, err)
		require.Equal(t, match.StatusPlaying, m.Status)
		require.NotNil(t, m.StartedAt)
	})

	t.Run("успех: белые отмечают начало", func(t *testing.T) {
		t.Parallel()
		m := mustNewMatch(t, now)
		require.NoError(t, m.Accept(20, now.Add(time.Hour)))
		err := m.Start(20, now.Add(2*time.Hour))
		require.NoError(t, err)
		require.Equal(t, match.StatusPlaying, m.Status)
	})

	t.Run("ошибка: start без accept", func(t *testing.T) {
		t.Parallel()
		m := mustNewMatch(t, now)
		err := m.Start(10, now.Add(time.Hour))
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})

	t.Run("ошибка: посторонний отмечает начало", func(t *testing.T) {
		t.Parallel()
		m := mustNewMatch(t, now)
		require.NoError(t, m.Accept(20, now.Add(time.Hour)))
		err := m.Start(99, now.Add(2*time.Hour))
		require.ErrorIs(t, err, shared.ErrForbidden)
	})
}

func TestMatch_SubmitResult(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	scoreDiff := 15.5

	t.Run("успех: чёрные победили, результат вносят чёрные", func(t *testing.T) {
		t.Parallel()
		m := mustNewAcceptedAndStarted(t, now)
		err := m.SubmitResult(10, 10, &scoreDiff, "sgf-url", "link", now.Add(3*time.Hour))
		require.NoError(t, err)
		require.Equal(t, match.StatusPendingConfirmation, m.Status)
		require.NotNil(t, m.WinnerID)
		require.Equal(t, shared.PlayerID(10), *m.WinnerID)
		require.NotNil(t, m.ScoreDiff)
		require.Equal(t, 15.5, *m.ScoreDiff)
		require.Equal(t, "sgf-url", m.SGFURL)
		require.Equal(t, "link", m.ExternalLink)
		require.NotNil(t, m.ResultSubmittedBy)
		require.Equal(t, shared.PlayerID(10), *m.ResultSubmittedBy)
		require.NotNil(t, m.FinishedAt)
	})

	t.Run("успех: белые победили, результат вносят белые", func(t *testing.T) {
		t.Parallel()
		m := mustNewAcceptedAndStarted(t, now)
		err := m.SubmitResult(20, 20, &scoreDiff, "", "", now.Add(3*time.Hour))
		require.NoError(t, err)
		require.NotNil(t, m.WinnerID)
		require.Equal(t, shared.PlayerID(20), *m.WinnerID)
	})

	t.Run("ошибка: результат вносит проигравший", func(t *testing.T) {
		t.Parallel()
		m := mustNewAcceptedAndStarted(t, now)
		err := m.SubmitResult(20, 10, &scoreDiff, "", "", now.Add(3*time.Hour))
		require.ErrorIs(t, err, shared.ErrForbidden)
	})

	t.Run("ошибка: победитель не участник", func(t *testing.T) {
		t.Parallel()
		m := mustNewAcceptedAndStarted(t, now)
		err := m.SubmitResult(10, 99, &scoreDiff, "", "", now.Add(3*time.Hour))
		require.ErrorIs(t, err, shared.ErrValidation)
	})

	t.Run("ошибка: посторонний вносит результат", func(t *testing.T) {
		t.Parallel()
		m := mustNewAcceptedAndStarted(t, now)
		err := m.SubmitResult(99, 10, &scoreDiff, "", "", now.Add(3*time.Hour))
		require.ErrorIs(t, err, shared.ErrForbidden)
	})

	t.Run("успех: submit сразу после accept (без start)", func(t *testing.T) {
		t.Parallel()
		// SubmitResult разрешает StatusAccepted И StatusPlaying
		m := mustNewMatch(t, now)
		require.NoError(t, m.Accept(20, now.Add(time.Hour)))
		err := m.SubmitResult(10, 10, &scoreDiff, "", "", now.Add(2*time.Hour))
		require.NoError(t, err)
	})
}

func TestMatch_Confirm(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	scoreDiff := 15.5

	t.Run("успех: соперник подтверждает", func(t *testing.T) {
		t.Parallel()
		m := mustNewAcceptedAndStarted(t, now)
		require.NoError(t, m.SubmitResult(10, 10, &scoreDiff, "", "", now.Add(3*time.Hour)))
		err := m.Confirm(20, now.Add(4*time.Hour))
		require.NoError(t, err)
		require.Equal(t, match.StatusConfirmed, m.Status)
		require.NotNil(t, m.ConfirmedBy)
		require.Equal(t, shared.PlayerID(20), *m.ConfirmedBy)
		require.NotNil(t, m.ConfirmedAt)
	})

	t.Run("ошибка: тот же кто внёс не может подтвердить", func(t *testing.T) {
		t.Parallel()
		m := mustNewAcceptedAndStarted(t, now)
		require.NoError(t, m.SubmitResult(10, 10, &scoreDiff, "", "", now.Add(3*time.Hour)))
		err := m.Confirm(10, now.Add(4*time.Hour))
		require.ErrorIs(t, err, shared.ErrForbidden)
	})

	t.Run("ошибка: подтверждение без submit", func(t *testing.T) {
		t.Parallel()
		m := mustNewAcceptedAndStarted(t, now)
		err := m.Confirm(20, now.Add(3*time.Hour))
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})

	t.Run("ошибка: посторонний подтверждает", func(t *testing.T) {
		t.Parallel()
		m := mustNewAcceptedAndStarted(t, now)
		require.NoError(t, m.SubmitResult(10, 10, &scoreDiff, "", "", now.Add(3*time.Hour)))
		err := m.Confirm(99, now.Add(4*time.Hour))
		require.ErrorIs(t, err, shared.ErrForbidden)
	})
}

func TestMatch_Cancel(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	cancelAfterDays := 3

	t.Run("успех: игрок отменяет proposed", func(t *testing.T) {
		t.Parallel()
		m := mustNewMatch(t, now)
		err := m.Cancel(10, false, cancelAfterDays, now.Add(time.Hour))
		require.NoError(t, err)
		require.Equal(t, match.StatusCancelled, m.Status)
		require.NotNil(t, m.CancelledAt)
	})

	t.Run("успех: админ отменяет confirmed", func(t *testing.T) {
		t.Parallel()
		m := mustNewConfirmedMatch(t, now)
		err := m.Cancel(1, true, cancelAfterDays, now.Add(10*time.Hour))
		require.NoError(t, err)
		require.Equal(t, match.StatusCancelled, m.Status)
	})

	t.Run("ошибка: игрок не может отменить confirmed", func(t *testing.T) {
		t.Parallel()
		m := mustNewConfirmedMatch(t, now)
		err := m.Cancel(10, false, cancelAfterDays, now.Add(10*time.Hour))
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})

	t.Run("ошибка: игрок не может отменить после срока", func(t *testing.T) {
		t.Parallel()
		m := mustNewMatch(t, now)
		// Симулируем finished_at в прошлом
		m.Status = match.StatusPendingConfirmation
		finishedAt := now.Add(-7 * 24 * time.Hour) // 7 дней назад
		m.FinishedAt = &finishedAt
		err := m.Cancel(10, false, cancelAfterDays, now)
		require.ErrorIs(t, err, shared.ErrForbidden)
	})

	t.Run("ошибка: повторная отмена", func(t *testing.T) {
		t.Parallel()
		m := mustNewMatch(t, now)
		require.NoError(t, m.Cancel(10, false, cancelAfterDays, now.Add(time.Hour)))
		err := m.Cancel(10, true, cancelAfterDays, now.Add(2*time.Hour))
		require.ErrorIs(t, err, shared.ErrInvalidState)
	})

	t.Run("ошибка: посторонний отменяет", func(t *testing.T) {
		t.Parallel()
		m := mustNewMatch(t, now)
		err := m.Cancel(99, false, cancelAfterDays, now.Add(time.Hour))
		require.ErrorIs(t, err, shared.ErrForbidden)
	})
}

func TestMatch_OpponentOf(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	m := mustNewMatch(t, now)

	opp, ok := m.OpponentOf(10)
	require.True(t, ok)
	require.Equal(t, shared.PlayerID(20), opp)

	opp, ok = m.OpponentOf(20)
	require.True(t, ok)
	require.Equal(t, shared.PlayerID(10), opp)

	_, ok = m.OpponentOf(99)
	require.False(t, ok)
}

func TestMatch_IsParticipant(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	m := mustNewMatch(t, now)

	require.True(t, m.IsParticipant(10))
	require.True(t, m.IsParticipant(20))
	require.False(t, m.IsParticipant(99))
}

func TestStatus_IsFinal(t *testing.T) {
	t.Parallel()
	require.True(t, match.StatusConfirmed.IsFinal())
	require.True(t, match.StatusCancelled.IsFinal())
	require.False(t, match.StatusProposed.IsFinal())
	require.False(t, match.StatusAccepted.IsFinal())
	require.False(t, match.StatusPlaying.IsFinal())
	require.False(t, match.StatusPendingConfirmation.IsFinal())
}

func TestStatus_Valid(t *testing.T) {
	t.Parallel()
	require.True(t, match.StatusProposed.Valid())
	require.True(t, match.StatusAccepted.Valid())
	require.True(t, match.StatusPlaying.Valid())
	require.True(t, match.StatusPendingConfirmation.Valid())
	require.True(t, match.StatusConfirmed.Valid())
	require.True(t, match.StatusCancelled.Valid())
	require.False(t, match.Status("unknown").Valid())
}

func TestServer_Valid(t *testing.T) {
	t.Parallel()
	require.True(t, match.ServerOGS.Valid())
	require.True(t, match.ServerKGS.Valid())
	require.True(t, match.ServerFox.Valid())
	require.True(t, match.ServerOffline.Valid())
	require.False(t, match.Server("unknown").Valid())
}

// ---- helpers ----------------------------------------------------------------

func mustNewMatch(t *testing.T, now time.Time) *match.Match {
	t.Helper()
	m, err := match.New(match.NewMatchParams{
		SeasonID:      1,
		DivisionID:    2,
		BlackPlayerID: 10,
		WhitePlayerID: 20,
		CreatedBy:     10,
		Server:        match.ServerOGS,
	}, now)
	require.NoError(t, err)
	return m
}

func mustNewAcceptedAndStarted(t *testing.T, now time.Time) *match.Match {
	t.Helper()
	m := mustNewMatch(t, now)
	require.NoError(t, m.Accept(20, now.Add(time.Hour)))
	require.NoError(t, m.Start(10, now.Add(2*time.Hour)))
	return m
}

func mustNewConfirmedMatch(t *testing.T, now time.Time) *match.Match {
	t.Helper()
	m := mustNewAcceptedAndStarted(t, now)
	scoreDiff := 15.5
	require.NoError(t, m.SubmitResult(10, 10, &scoreDiff, "", "", now.Add(3*time.Hour)))
	require.NoError(t, m.Confirm(20, now.Add(4*time.Hour)))
	return m
}
