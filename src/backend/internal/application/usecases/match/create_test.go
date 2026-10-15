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
	"github.com/J0hnLenin/67League/internal/domain/division"
	dommatch "github.com/J0hnLenin/67League/internal/domain/match"
	"github.com/J0hnLenin/67League/internal/domain/notification"
	"github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/testsupport/mocks"
)

const (
	cmBlackID    = shared.PlayerID(1001)
	cmWhiteID    = shared.PlayerID(1002)
	cmSeasonID   = shared.SeasonID(70)
	cmDivisionID = shared.DivisionID(310)
	cmMatchID    = shared.MatchID(4001)
)

type createMatchHarness struct {
	matches *mocks.MatchRepository
	events  *mocks.MatchEventRepository
	members *mocks.DivisionMemberRepository
	seasons *mocks.SeasonRepository
	notifs  *mocks.NotificationRepository
	clock   *mocks.Clock
	tx      *mocks.TransactionManager
	now     time.Time
}

func newCreateMatchHarness(t *testing.T, now time.Time) *createMatchHarness {
	t.Helper()
	return &createMatchHarness{
		matches: mocks.NewMatchRepository(t),
		events:  mocks.NewMatchEventRepository(t),
		members: mocks.NewDivisionMemberRepository(t),
		seasons: mocks.NewSeasonRepository(t),
		notifs:  mocks.NewNotificationRepository(t),
		clock:   mocks.NewClock(t),
		tx:      mocks.NewTransactionManager(t),
		now:     now,
	}
}

func (h *createMatchHarness) build(t *testing.T) *match.CreateMatch {
	t.Helper()
	uc, err := match.NewCreateMatch(match.CreateMatchDeps{
		Matches: h.matches,
		Events:  h.events,
		Members: h.members,
		Seasons: h.seasons,
		Notifs:  h.notifs,
		Clock:   h.clock,
		Tx:      h.tx,
	})
	require.NoError(t, err)
	return uc
}

func (h *createMatchHarness) expectTx() {
	h.tx.EXPECT().WithinTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
}

func cmMakeSeason(now time.Time, status season.Status, n int) *season.Season {
	settings := season.DefaultSettings()
	settings.N = n
	return &season.Season{
		ID:        cmSeasonID,
		Number:    1,
		Title:     "S1",
		Status:    status,
		Settings:  settings,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func cmMakeMember(pid shared.PlayerID, divID shared.DivisionID, now time.Time) *division.DivisionMember {
	m := division.NewMember(divID, pid, now)
	m.ID = shared.DivisionMemberID(int64(pid) + 5000)
	return m
}

func TestCreateMatch(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	validInput := func() dto.CreateMatchInput {
		return dto.CreateMatchInput{
			SeasonID:      cmSeasonID,
			DivisionID:    cmDivisionID,
			BlackPlayerID: cmBlackID,
			WhitePlayerID: cmWhiteID,
			CreatedBy:     cmBlackID,
			Server:        dommatch.ServerOGS,
		}
	}

	tests := []struct {
		name           string
		seasonStatus   season.Status
		settingsN      int
		existingCount  int
		blackDiv       shared.DivisionID
		whiteDiv       shared.DivisionID
		input          func() dto.CreateMatchInput
		wantErr        error
		wantNotifiedTo shared.PlayerID
	}{
		{
			name:           "успех: чёрный создаёт вызов",
			seasonStatus:   season.StatusActive,
			settingsN:      4,
			existingCount:  0,
			blackDiv:       cmDivisionID,
			whiteDiv:       cmDivisionID,
			input:          validInput,
			wantNotifiedTo: cmWhiteID,
		},
		{
			name:          "успех: белый создаёт вызов",
			seasonStatus:  season.StatusActive,
			settingsN:     4,
			existingCount: 1,
			blackDiv:      cmDivisionID,
			whiteDiv:      cmDivisionID,
			input: func() dto.CreateMatchInput {
				in := validInput()
				in.CreatedBy = cmWhiteID
				return in
			},
			wantNotifiedTo: cmBlackID,
		},
		{
			name:         "один и тот же игрок чёрный и белый",
			seasonStatus: season.StatusActive,
			settingsN:    4,
			blackDiv:     cmDivisionID,
			whiteDiv:     cmDivisionID,
			input: func() dto.CreateMatchInput {
				in := validInput()
				in.WhitePlayerID = cmBlackID
				return in
			},
			wantErr: shared.ErrValidation,
		},
		{
			name:         "создатель не участник",
			seasonStatus: season.StatusActive,
			settingsN:    4,
			blackDiv:     cmDivisionID,
			whiteDiv:     cmDivisionID,
			input: func() dto.CreateMatchInput {
				in := validInput()
				in.CreatedBy = shared.PlayerID(9999)
				return in
			},
			wantErr: shared.ErrForbidden,
		},
		{
			name:         "сезон не active — invalid state",
			seasonStatus: season.StatusClosing,
			settingsN:    4,
			blackDiv:     cmDivisionID,
			whiteDiv:     cmDivisionID,
			input:        validInput,
			wantErr:      shared.ErrInvalidState,
		},
		{
			name:         "игроки в разных дивизионах",
			seasonStatus: season.StatusActive,
			settingsN:    4,
			blackDiv:     cmDivisionID,
			whiteDiv:     shared.DivisionID(9999),
			input:        validInput,
			wantErr:      shared.ErrValidation,
		},
		{
			name:          "лимит партий с соперником исчерпан",
			seasonStatus:  season.StatusActive,
			settingsN:     4,
			existingCount: 4,
			blackDiv:      cmDivisionID,
			whiteDiv:      cmDivisionID,
			input:         validInput,
			wantErr:       shared.ErrLimitExceeded,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := newCreateMatchHarness(t, now)
			h.clock.EXPECT().Now().Return(now).Maybe()

			in := tt.input()

			// Ранние валидации до обращения к репозиториям.
			if in.BlackPlayerID == in.WhitePlayerID {
				h.buildAndRun(t, in, tt.wantErr)
				return
			}
			if in.CreatedBy != in.BlackPlayerID && in.CreatedBy != in.WhitePlayerID {
				h.buildAndRun(t, in, tt.wantErr)
				return
			}

			h.seasons.EXPECT().GetByID(mock.Anything, cmSeasonID).
				Return(cmMakeSeason(now, tt.seasonStatus, tt.settingsN), nil).Once()

			// Если сезон не active — выходим раньше.
			if !tt.seasonStatus.CanCreateMatch() {
				h.buildAndRun(t, in, tt.wantErr)
				return
			}

			h.members.EXPECT().
				GetByPlayerAndSeason(mock.Anything, cmBlackID, cmSeasonID).
				Return(cmMakeMember(cmBlackID, tt.blackDiv, now), nil).Once()
			h.members.EXPECT().
				GetByPlayerAndSeason(mock.Anything, cmWhiteID, cmSeasonID).
				Return(cmMakeMember(cmWhiteID, tt.whiteDiv, now), nil).Once()

			if tt.blackDiv != tt.whiteDiv {
				h.buildAndRun(t, in, tt.wantErr)
				return
			}

			h.matches.EXPECT().
				CountBetweenPlayersInSeason(mock.Anything, cmSeasonID, cmBlackID, cmWhiteID).
				Return(tt.existingCount, nil).Once()

			if tt.existingCount >= tt.settingsN {
				h.buildAndRun(t, in, tt.wantErr)
				return
			}

			var capturedMatch *dommatch.Match
			var notifiedTo shared.PlayerID
			h.matches.EXPECT().
				Create(mock.Anything, mock.AnythingOfType("*match.Match")).
				Run(func(_ context.Context, m *dommatch.Match) {
					m.ID = cmMatchID
					capturedMatch = m
				}).
				Return(nil).Once()
			h.events.EXPECT().
				Create(mock.Anything, mock.AnythingOfType("*match.MatchEvent")).
				Return(nil).Once()
			h.notifs.EXPECT().
				Create(mock.Anything, mock.AnythingOfType("*notification.Notification")).
				Run(func(_ context.Context, n *notification.Notification) {
					notifiedTo = n.PlayerID
				}).
				Return(nil).Once()
			h.expectTx()

			uc := h.build(t)
			got, err := uc.Execute(context.Background(), in)

			require.NoError(t, err)
			require.NotNil(t, got)
			require.NotNil(t, got.Match)
			require.Equal(t, cmMatchID, got.Match.ID)
			require.Equal(t, dommatch.StatusProposed, got.Match.Status)
			require.Equal(t, tt.wantNotifiedTo, notifiedTo)
			require.NotNil(t, capturedMatch)
		})
	}
}

// buildAndRun — для коротких веток: собрать и проверить ошибку.
func (h *createMatchHarness) buildAndRun(t *testing.T, in dto.CreateMatchInput, wantErr error) {
	t.Helper()
	uc, err := match.NewCreateMatch(match.CreateMatchDeps{
		Matches: h.matches,
		Events:  h.events,
		Members: h.members,
		Seasons: h.seasons,
		Notifs:  h.notifs,
		Clock:   h.clock,
		Tx:      h.tx,
	})
	require.NoError(t, err)
	got, err := uc.Execute(context.Background(), in)
	require.Error(t, err)
	require.True(t, errors.Is(err, wantErr),
		"expected %v, got %v", wantErr, err)
	require.Nil(t, got)
}
