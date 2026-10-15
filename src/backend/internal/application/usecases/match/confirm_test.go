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
	domplayer "github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/testsupport/mocks"
)

// ---- Детерминированные ID --------------------------------------------------

const (
	cfWinnerID   = shared.PlayerID(2001)
	cfLoserID    = shared.PlayerID(2002)
	cfMatchID    = shared.MatchID(3001)
	cfSeasonID   = shared.SeasonID(70)
	cfDivisionID = shared.DivisionID(310)

	cfWinnerMemberID = shared.DivisionMemberID(4001)
	cfLoserMemberID  = shared.DivisionMemberID(4002)
)

// ---- harness ---------------------------------------------------------------

type confirmHarness struct {
	matches *mocks.MatchRepository
	events  *mocks.MatchEventRepository
	members *mocks.DivisionMemberRepository
	players *mocks.PlayerRepository
	seasons *mocks.SeasonRepository
	stones  *mocks.StoneTransactionRepository
	notifs  *mocks.NotificationRepository
	clock   *mocks.Clock
	tx      *mocks.TransactionManager
	now     time.Time
}

func newConfirmHarness(t *testing.T, now time.Time) *confirmHarness {
	t.Helper()
	return &confirmHarness{
		matches: mocks.NewMatchRepository(t),
		events:  mocks.NewMatchEventRepository(t),
		members: mocks.NewDivisionMemberRepository(t),
		players: mocks.NewPlayerRepository(t),
		seasons: mocks.NewSeasonRepository(t),
		stones:  mocks.NewStoneTransactionRepository(t),
		notifs:  mocks.NewNotificationRepository(t),
		clock:   mocks.NewClock(t),
		tx:      mocks.NewTransactionManager(t),
		now:     now,
	}
}

func (h *confirmHarness) build(t *testing.T) *match.ConfirmResult {
	t.Helper()
	uc, err := match.NewConfirmResult(match.ConfirmResultDeps{
		Matches: h.matches,
		Events:  h.events,
		Members: h.members,
		Players: h.players,
		Seasons: h.seasons,
		Stones:  h.stones,
		Notifs:  h.notifs,
		Clock:   h.clock,
		Tx:      h.tx,
	})
	require.NoError(t, err)
	return uc
}

// ---- fixture ---------------------------------------------------------------

// confirmFixture — заранее подготовленная партия в pending_confirmation,
// где победитель уже внёс результат, а соперник ещё не подтвердил.
type confirmFixture struct {
	winner, loser *domplayer.Player
	winnerMember  *division.DivisionMember
	loserMember   *division.DivisionMember
	m             *dommatch.Match
	season        *season.Season
}

func buildConfirmFixture(t *testing.T, now time.Time, loserStreak int, settings season.Settings) *confirmFixture {
	t.Helper()

	winner, err := domplayer.New(domplayer.NewPlayerParams{
		FullName: "Winner", Email: "w@x.com", PasswordHash: "h",
		Nickname: "w", City: "M",
	}, now)
	require.NoError(t, err)
	winner.ID = cfWinnerID

	loser, err := domplayer.New(domplayer.NewPlayerParams{
		FullName: "Loser", Email: "l@x.com", PasswordHash: "h",
		Nickname: "l", City: "M",
	}, now)
	require.NoError(t, err)
	loser.ID = cfLoserID

	s := &season.Season{
		ID:       cfSeasonID,
		Number:   1,
		Title:    "S1",
		Status:   season.StatusActive,
		Settings: settings,
	}

	m, err := dommatch.New(dommatch.NewMatchParams{
		SeasonID:      cfSeasonID,
		DivisionID:    cfDivisionID,
		BlackPlayerID: winner.ID,
		WhitePlayerID: loser.ID,
		CreatedBy:     winner.ID,
		Server:        dommatch.ServerOGS,
	}, now)
	require.NoError(t, err)
	m.ID = cfMatchID
	m.Status = dommatch.StatusPendingConfirmation
	m.WinnerID = &winner.ID
	m.ResultSubmittedBy = &winner.ID

	wm := division.NewMember(cfDivisionID, winner.ID, now)
	wm.ID = cfWinnerMemberID

	lm := division.NewMember(cfDivisionID, loser.ID, now)
	lm.ID = cfLoserMemberID
	lm.StreakCurrent = loserStreak
	lm.StreakMax = loserStreak

	return &confirmFixture{
		winner: winner, loser: loser,
		winnerMember: wm, loserMember: lm,
		m: m, season: s,
	}
}

// ---- тест ------------------------------------------------------------------

func TestConfirmResult(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	settings := season.DefaultSettings()

	tests := []struct {
		name             string
		loserStreak      int
		actorIsWinner    bool
		wantErr          error
		wantWinnerLP     shared.LP
		wantLoserLP      shared.LP
		wantStreakBroken bool
	}{
		{
			name:         "обычная победа",
			loserStreak:  0,
			wantWinnerLP: shared.LP(settings.LPWin),
			wantLoserLP:  shared.LP(settings.LPLoss),
		},
		{
			name:             "снятие стрика 3",
			loserStreak:      3,
			wantWinnerLP:     shared.LP(settings.LPWin + settings.StreakBonus3),
			wantLoserLP:      shared.LP(settings.LPLoss),
			wantStreakBroken: true,
		},
		{
			name:             "снятие стрика 5",
			loserStreak:      5,
			wantWinnerLP:     shared.LP(settings.LPWin + settings.StreakBonus5),
			wantLoserLP:      shared.LP(settings.LPLoss),
			wantStreakBroken: true,
		},
		{
			name:             "снятие стрика 7",
			loserStreak:      7,
			wantWinnerLP:     shared.LP(settings.LPWin + settings.StreakBonus67),
			wantLoserLP:      shared.LP(settings.LPLoss),
			wantStreakBroken: true,
		},
		{
			name:          "нельзя подтвердить свой результат",
			actorIsWinner: true,
			wantErr:       shared.ErrForbidden,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := newConfirmHarness(t, now)
			fx := buildConfirmFixture(t, now, tt.loserStreak, settings)

			// Clock.Now() вызывается до входа в транзакцию — фиксируем один раз.
			h.clock.EXPECT().Now().Return(now).Maybe()

			// WithinTransaction — прозрачный passthrough.
			h.tx.EXPECT().WithinTransaction(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				}).Once()

			// Партия всегда читается первой.
			h.matches.EXPECT().GetByID(mock.Anything, cfMatchID).
				Return(fx.m, nil).Once()

			var capturedWinner, capturedLoser *division.DivisionMember

			if tt.wantErr == nil {
				h.seasons.EXPECT().GetByID(mock.Anything, cfSeasonID).
					Return(fx.season, nil).Once()

				h.members.EXPECT().
					GetByPlayerAndSeason(mock.Anything, cfWinnerID, cfSeasonID).
					Return(fx.winnerMember, nil).Once()
				h.members.EXPECT().
					GetByPlayerAndSeason(mock.Anything, cfLoserID, cfSeasonID).
					Return(fx.loserMember, nil).Once()

				// Update по одному разу на каждого мембера.
				h.members.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*division.DivisionMember")).
					Run(func(_ context.Context, m *division.DivisionMember) {
						switch m.ID {
						case cfWinnerMemberID:
							capturedWinner = m
						case cfLoserMemberID:
							capturedLoser = m
						}
					}).
					Return(nil).Twice()

				// t=12, у обоих 0 подтверждённых → бонуса «сверх t» нет.
				h.matches.EXPECT().
					CountConfirmedByPlayerInSeason(mock.Anything, cfSeasonID, cfWinnerID).
					Return(0, nil).Once()
				h.matches.EXPECT().
					CountConfirmedByPlayerInSeason(mock.Anything, cfSeasonID, cfLoserID).
					Return(0, nil).Once()

				h.players.EXPECT().GetByID(mock.Anything, cfWinnerID).
					Return(fx.winner, nil).Once()
				h.players.EXPECT().GetByID(mock.Anything, cfLoserID).
					Return(fx.loser, nil).Once()

				// Update игроков: один раз на победителя, один раз на проигравшего.
				h.players.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*player.Player")).
					Return(nil).Twice()

				// Save обновлённой партии.
				h.matches.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*match.Match")).
					Run(func(_ context.Context, m *dommatch.Match) {
						require.Equal(t, dommatch.StatusConfirmed, m.Status)
					}).
					Return(nil).Once()

				// Проигравший при stones_loss=0 транзакции не создаёт.
				h.stones.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*economy.StoneTransaction")).
					Return(nil).Once()

				// Журнал события.
				h.events.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*match.MatchEvent")).
					Return(nil).Once()

				// Уведомление тому, кто вносил результат.
				h.notifs.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*notification.Notification")).
					Return(nil).Once()
			}

			uc := h.build(t)
			actor := cfLoserID
			if tt.actorIsWinner {
				actor = cfWinnerID
			}

			got, err := uc.Execute(context.Background(), dto.ConfirmResultInput{
				MatchID: cfMatchID,
				ActorID: actor,
			})

			if tt.wantErr != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.wantErr),
					"expected %v, got %v", tt.wantErr, err)
				require.Nil(t, got)
				// Партия не должна быть сохранена.
				require.Equal(t, dommatch.StatusPendingConfirmation, fx.m.Status)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)

			require.Equal(t, tt.wantWinnerLP, got.WinnerLP)
			require.Equal(t, tt.wantLoserLP, got.LoserLP)
			require.Equal(t, tt.wantStreakBroken, got.StreakBroken)
			require.Equal(t, dommatch.StatusConfirmed, got.Match.Status)
			require.Equal(t, cfMatchID, got.Match.ID)

			require.NotNil(t, capturedWinner, "winner member was not updated")
			require.NotNil(t, capturedLoser, "loser member was not updated")
			require.Equal(t, tt.wantWinnerLP, capturedWinner.LP)
			require.Equal(t, tt.wantLoserLP, capturedLoser.LP)
			require.Equal(t, 1, capturedWinner.StreakCurrent)
			require.Equal(t, 0, capturedLoser.StreakCurrent)
		})
	}
}
