package match_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/usecases/match"
	dommatch "github.com/J0hnLenin/67League/internal/domain/match"
	"github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/testsupport/mocks"
)

const (
	trMatchID  = shared.MatchID(4001)
	trBlackID  = shared.PlayerID(1001)
	trWhiteID  = shared.PlayerID(1002)
	trSeasonID = shared.SeasonID(70)
)

// trMakeMatch создаёт партию с заданным статусом и настройками
// «кто вносил результат», «кто победил» и т.д.
func trMakeMatch(now time.Time, status dommatch.Status) *dommatch.Match {
	m, _ := dommatch.New(dommatch.NewMatchParams{
		SeasonID:      trSeasonID,
		DivisionID:    shared.DivisionID(310),
		BlackPlayerID: trBlackID,
		WhitePlayerID: trWhiteID,
		CreatedBy:     trBlackID,
		Server:        dommatch.ServerOGS,
	}, now)
	m.ID = trMatchID
	m.Status = status
	return m
}

// ---- AcceptMatch ------------------------------------------------------------

func TestAcceptMatch(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		initial    dommatch.Status
		actor      shared.PlayerID
		wantErr    error
		wantStatus dommatch.Status
	}{
		{
			name:       "успех: соперник принимает",
			initial:    dommatch.StatusProposed,
			actor:      trWhiteID,
			wantStatus: dommatch.StatusAccepted,
		},
		{
			name:    "создатель не может принять свой вызов",
			initial: dommatch.StatusProposed,
			actor:   trBlackID,
			wantErr: shared.ErrForbidden,
		},
		{
			name:    "не тот статус",
			initial: dommatch.StatusAccepted,
			actor:   trWhiteID,
			wantErr: shared.ErrInvalidState,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			matches := mocks.NewMatchRepository(t)
			events := mocks.NewMatchEventRepository(t)
			notifs := mocks.NewNotificationRepository(t)
			clock := mocks.NewClock(t)
			tx := mocks.NewTransactionManager(t)

			clock.EXPECT().Now().Return(now).Once()
			tx.EXPECT().WithinTransaction(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				}).Once()

			m := trMakeMatch(now, tt.initial)
			matches.EXPECT().GetByID(mock.Anything, trMatchID).Return(m, nil).Once()

			if tt.wantErr == nil {
				matches.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*match.Match")).
					Run(func(_ context.Context, mm *dommatch.Match) {
						require.Equal(t, tt.wantStatus, mm.Status)
						require.NotNil(t, mm.AcceptedAt)
					}).
					Return(nil).Once()
				events.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*match.MatchEvent")).
					Return(nil).Once()
				notifs.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(nil).Once()
			}

			uc, err := match.NewAcceptMatch(match.AcceptMatchDeps{
				Matches: matches, Events: events, Notifs: notifs,
				Clock: clock, Tx: tx,
			})
			require.NoError(t, err)

			got, err := uc.Execute(context.Background(), dto.AcceptMatchInput{
				MatchID: trMatchID, ActorID: tt.actor,
			})

			if tt.wantErr != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.wantErr),
					"expected %v, got %v", tt.wantErr, err)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, tt.wantStatus, got.Match.Status)
		})
	}
}

// ---- StartMatch -------------------------------------------------------------

func TestStartMatch(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		initial    dommatch.Status
		actor      shared.PlayerID
		wantErr    error
		wantStatus dommatch.Status
	}{
		{
			name:       "успех: чёрный стартует",
			initial:    dommatch.StatusAccepted,
			actor:      trBlackID,
			wantStatus: dommatch.StatusPlaying,
		},
		{
			name:       "успех: белый стартует",
			initial:    dommatch.StatusAccepted,
			actor:      trWhiteID,
			wantStatus: dommatch.StatusPlaying,
		},
		{
			name:    "не тот статус",
			initial: dommatch.StatusProposed,
			actor:   trBlackID,
			wantErr: shared.ErrInvalidState,
		},
		{
			name:    "не участник",
			initial: dommatch.StatusAccepted,
			actor:   shared.PlayerID(9999),
			wantErr: shared.ErrForbidden,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			matches := mocks.NewMatchRepository(t)
			events := mocks.NewMatchEventRepository(t)
			clock := mocks.NewClock(t)
			tx := mocks.NewTransactionManager(t)

			clock.EXPECT().Now().Return(now).Once()
			tx.EXPECT().WithinTransaction(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				}).Once()

			m := trMakeMatch(now, tt.initial)
			matches.EXPECT().GetByID(mock.Anything, trMatchID).Return(m, nil).Once()

			if tt.wantErr == nil {
				matches.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*match.Match")).
					Run(func(_ context.Context, mm *dommatch.Match) {
						require.Equal(t, tt.wantStatus, mm.Status)
						require.NotNil(t, mm.StartedAt)
					}).
					Return(nil).Once()
				events.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*match.MatchEvent")).
					Return(nil).Once()
			}

			uc, err := match.NewStartMatch(match.StartMatchDeps{
				Matches: matches, Events: events, Clock: clock, Tx: tx,
			})
			require.NoError(t, err)

			got, err := uc.Execute(context.Background(), dto.StartMatchInput{
				MatchID: trMatchID, ActorID: tt.actor,
			})

			if tt.wantErr != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.wantErr),
					"expected %v, got %v", tt.wantErr, err)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, tt.wantStatus, got.Match.Status)
		})
	}
}

// ---- SubmitResult -----------------------------------------------------------

func TestSubmitResult(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		initial    dommatch.Status
		actor      shared.PlayerID
		winner     shared.PlayerID
		wantErr    error
		wantStatus dommatch.Status
	}{
		{
			name:       "успех: победитель вносит результат",
			initial:    dommatch.StatusPlaying,
			actor:      trBlackID,
			winner:     trBlackID,
			wantStatus: dommatch.StatusPendingConfirmation,
		},
		{
			name:    "проигравший не может внести",
			initial: dommatch.StatusPlaying,
			actor:   trWhiteID,
			winner:  trBlackID,
			wantErr: shared.ErrForbidden,
		},
		{
			name:    "не тот статус",
			initial: dommatch.StatusProposed,
			actor:   trBlackID,
			winner:  trBlackID,
			wantErr: shared.ErrInvalidState,
		},
		{
			name:    "не участник",
			initial: dommatch.StatusPlaying,
			actor:   shared.PlayerID(9999),
			winner:  trBlackID,
			wantErr: shared.ErrForbidden,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			matches := mocks.NewMatchRepository(t)
			events := mocks.NewMatchEventRepository(t)
			notifs := mocks.NewNotificationRepository(t)
			clock := mocks.NewClock(t)
			tx := mocks.NewTransactionManager(t)

			clock.EXPECT().Now().Return(now).Once()
			tx.EXPECT().WithinTransaction(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				}).Once()

			m := trMakeMatch(now, tt.initial)
			matches.EXPECT().GetByID(mock.Anything, trMatchID).Return(m, nil).Once()

			if tt.wantErr == nil {
				matches.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*match.Match")).
					Run(func(_ context.Context, mm *dommatch.Match) {
						require.Equal(t, tt.wantStatus, mm.Status)
						require.NotNil(t, mm.WinnerID)
						require.Equal(t, tt.winner, *mm.WinnerID)
					}).
					Return(nil).Once()
				events.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*match.MatchEvent")).
					Return(nil).Once()
				notifs.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(nil).Once()
			}

			uc, err := match.NewSubmitResult(match.SubmitResultDeps{
				Matches: matches, Events: events, Notifs: notifs,
				Clock: clock, Tx: tx,
			})
			require.NoError(t, err)

			got, err := uc.Execute(context.Background(), dto.SubmitResultInput{
				MatchID:  trMatchID,
				ActorID:  tt.actor,
				WinnerID: tt.winner,
			})

			if tt.wantErr != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.wantErr),
					"expected %v, got %v", tt.wantErr, err)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, tt.wantStatus, got.Match.Status)
		})
	}
}

// ---- CancelMatch ------------------------------------------------------------

func TestCancelMatch(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	cancelAfterDays := 3

	makeSeason := func(now time.Time) *season.Season {
		s := &season.Season{
			ID:       trSeasonID,
			Number:   1,
			Title:    "S1",
			Status:   season.StatusActive,
			Settings: season.DefaultSettings(),
		}
		s.Settings.MatchCancelDays = cancelAfterDays
		return s
	}

	tests := []struct {
		name        string
		initial     dommatch.Status
		actor       shared.PlayerID
		isAdmin     bool
		finishedAgo time.Duration // 0 = не завершена
		wantErr     error
		wantStatus  dommatch.Status
	}{
		{
			name:       "успех: игрок отменяет proposed",
			initial:    dommatch.StatusProposed,
			actor:      trBlackID,
			wantStatus: dommatch.StatusCancelled,
		},
		{
			name:       "успех: игрок отменяет playing",
			initial:    dommatch.StatusPlaying,
			actor:      trWhiteID,
			wantStatus: dommatch.StatusCancelled,
		},
		{
			name:       "успех: админ отменяет confirmed",
			initial:    dommatch.StatusConfirmed,
			actor:      trBlackID,
			isAdmin:    true,
			wantStatus: dommatch.StatusCancelled,
		},
		{
			name:        "успех: игрок отменяет в пределах match_cancel_days",
			initial:     dommatch.StatusPendingConfirmation,
			actor:       trBlackID,
			finishedAgo: 2 * 24 * time.Hour,
			wantStatus:  dommatch.StatusCancelled,
		},
		{
			name:        "игрок не может отменить старую партию",
			initial:     dommatch.StatusPendingConfirmation,
			actor:       trBlackID,
			finishedAgo: 10 * 24 * time.Hour,
			wantErr:     shared.ErrForbidden,
		},
		{
			name:    "не участник",
			initial: dommatch.StatusProposed,
			actor:   shared.PlayerID(9999),
			wantErr: shared.ErrForbidden,
		},
		{
			name:    "уже отменена",
			initial: dommatch.StatusCancelled,
			actor:   trBlackID,
			wantErr: shared.ErrInvalidState,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			matches := mocks.NewMatchRepository(t)
			events := mocks.NewMatchEventRepository(t)
			seasons := mocks.NewSeasonRepository(t)
			notifs := mocks.NewNotificationRepository(t)
			clock := mocks.NewClock(t)
			tx := mocks.NewTransactionManager(t)

			// now — «текущее время». FinishedAt задаём в прошлом,
			// чтобы deadline = FinishedAt + match_cancel_days уже прошёл.
			clock.EXPECT().Now().Return(now).Once()

			tx.EXPECT().WithinTransaction(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				}).Once()

			m := trMakeMatch(now, tt.initial)
			if tt.finishedAgo > 0 {
				fin := now.Add(-tt.finishedAgo) // в прошлом!
				m.FinishedAt = &fin
			}
			matches.EXPECT().GetByID(mock.Anything, trMatchID).Return(m, nil).Once()

			// Сезон читается всегда — до Cancel.
			seasons.EXPECT().GetByID(mock.Anything, trSeasonID).
				Return(makeSeason(now), nil).Once()

			if tt.wantErr == nil {
				matches.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*match.Match")).
					Run(func(_ context.Context, mm *dommatch.Match) {
						require.Equal(t, tt.wantStatus, mm.Status)
						require.NotNil(t, mm.CancelledAt)
					}).
					Return(nil).Once()
				events.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*match.MatchEvent")).
					Return(nil).Once()
				if !tt.isAdmin {
					notifs.EXPECT().
						Create(mock.Anything, mock.Anything).
						Return(nil).Once()
				}
			}

			uc, err := match.NewCancelMatch(match.CancelMatchDeps{
				Matches: matches, Events: events, Seasons: seasons,
				Notifs: notifs, Clock: clock, Tx: tx,
			})
			require.NoError(t, err)

			got, err := uc.Execute(context.Background(), dto.CancelMatchInput{
				MatchID: trMatchID,
				ActorID: tt.actor,
				IsAdmin: tt.isAdmin,
				Reason:  "test",
			})

			if tt.wantErr != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.wantErr),
					"expected %v, got %v", tt.wantErr, err)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, tt.wantStatus, got.Match.Status)
			require.Equal(t, "test", got.Match.CancelReason)
		})
	}
}
