package badge_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/usecases/badge"
	dombadge "github.com/J0hnLenin/67League/internal/domain/badge"
	"github.com/J0hnLenin/67League/internal/domain/division"
	"github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/testsupport/mocks"
)

const (
	stSeasonID = shared.SeasonID(70)
	stDivID    = shared.DivisionID(301)
	stClubX    = shared.ClubID(501)
	stClubY    = shared.ClubID(502)
)

type stHarness struct {
	badges       *mocks.BadgeRepository
	playerBadges *mocks.PlayerBadgeRepository
	superTitles  *mocks.SuperTitleRepository
	members      *mocks.DivisionMemberRepository
	divisions    *mocks.DivisionRepository
	players      *mocks.PlayerRepository
	clubs        *mocks.ClubRepository
	clock        *mocks.Clock
	tx           *mocks.TransactionManager
	now          time.Time
}

func newSTHarness(t *testing.T, now time.Time) *stHarness {
	t.Helper()
	return &stHarness{
		badges:       mocks.NewBadgeRepository(t),
		playerBadges: mocks.NewPlayerBadgeRepository(t),
		superTitles:  mocks.NewSuperTitleRepository(t),
		members:      mocks.NewDivisionMemberRepository(t),
		divisions:    mocks.NewDivisionRepository(t),
		players:      mocks.NewPlayerRepository(t),
		clubs:        mocks.NewClubRepository(t),
		clock:        mocks.NewClock(t),
		tx:           mocks.NewTransactionManager(t),
		now:          now,
	}
}

func (h *stHarness) build(t *testing.T) *badge.AwardSuperTitle {
	t.Helper()
	uc, err := badge.NewAwardSuperTitle(badge.AwardSuperTitleDeps{
		Badges:       h.badges,
		PlayerBadges: h.playerBadges,
		SuperTitles:  h.superTitles,
		Members:      h.members,
		Divisions:    h.divisions,
		Players:      h.players,
		Clubs:        h.clubs,
		Clock:        h.clock,
		Tx:           h.tx,
	})
	require.NoError(t, err)
	return uc
}

func (h *stHarness) expectTx() {
	h.tx.EXPECT().WithinTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
}

func stMakeMember(now time.Time, id shared.DivisionMemberID, pid shared.PlayerID, lp float64, matches int) *division.DivisionMember {
	m := division.NewMember(stDivID, pid, now)
	m.ID = id
	m.LP = shared.LP(lp)
	m.MatchesPlayed = matches
	return m
}

func TestAwardSuperTitle_ChampionSelected(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	h := newSTHarness(t, now)

	h.superTitles.EXPECT().GetBySeason(mock.Anything, stSeasonID).
		Return(nil, shared.ErrNotFound).Once()

	h.divisions.EXPECT().ListBySeason(mock.Anything, stSeasonID).
		Return([]*division.Division{
			{ID: stDivID, SeasonID: stSeasonID, Code: "S", Order: 0},
		}, nil).Once()

	// Клуб X: 101 (20 LP) + 102 (10 LP) = 30
	// Клуб Y: 103 (15 LP) + 104 (20 LP) = 35 ← лучший
	// Внутри Y лучший — 104 (20 LP).
	m101 := stMakeMember(now, 1, 101, 20, 5)
	m102 := stMakeMember(now, 2, 102, 10, 5)
	m103 := stMakeMember(now, 3, 103, 15, 5)
	m104 := stMakeMember(now, 4, 104, 20, 5)

	h.members.EXPECT().ListByDivision(mock.Anything, stDivID).
		Return([]*division.DivisionMember{m101, m102, m103, m104}, nil).Once()

	clubX := stClubX
	clubY := stClubY
	for pid, c := range map[shared.PlayerID]*shared.ClubID{
		101: &clubX, 102: &clubX, 103: &clubY, 104: &clubY,
	} {
		h.players.EXPECT().GetByID(mock.Anything, pid).
			Return(&player.Player{ID: pid, ClubID: c}, nil).Once()
	}

	h.superTitles.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*badge.SuperTitle")).
		Run(func(_ context.Context, st *dombadge.SuperTitle) {
			require.Equal(t, stSeasonID, st.SeasonID)
			require.Equal(t, shared.PlayerID(104), st.PlayerID)
			require.Equal(t, stClubY, st.ClubID)
		}).
		Return(nil).Once()

	h.badges.EXPECT().GetByCode(mock.Anything, badge.CodeSuperChampion67).
		Return(&dombadge.Badge{ID: 900, Code: badge.CodeSuperChampion67}, nil).Once()

	h.playerBadges.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*badge.PlayerBadge")).
		Return(nil).Once()

	h.clock.EXPECT().Now().Return(now).Once()
	h.expectTx()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.AwardSuperTitleInput{SeasonID: stSeasonID})

	require.NoError(t, err)
	require.True(t, got.Awarded)
	require.NotNil(t, got.Player)
	require.Equal(t, shared.PlayerID(104), *got.Player)
	require.NotNil(t, got.Club)
	require.Equal(t, stClubY, *got.Club)
}

func TestAwardSuperTitle_Idempotent(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	h := newSTHarness(t, now)

	existing := &dombadge.SuperTitle{
		ID: 1, SeasonID: stSeasonID,
		PlayerID: 101, ClubID: stClubX, AwardedAt: now,
	}
	h.superTitles.EXPECT().GetBySeason(mock.Anything, stSeasonID).
		Return(existing, nil).Once()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.AwardSuperTitleInput{SeasonID: stSeasonID})

	require.NoError(t, err)
	require.False(t, got.Awarded)
	require.Same(t, existing, got.SuperTitle)
}

func TestAwardSuperTitle_Errors(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		setup   func(t *testing.T, h *stHarness)
		wantErr error
	}{
		{
			name: "нет дивизионов",
			setup: func(t *testing.T, h *stHarness) {
				h.superTitles.EXPECT().GetBySeason(mock.Anything, stSeasonID).
					Return(nil, shared.ErrNotFound).Once()
				h.divisions.EXPECT().ListBySeason(mock.Anything, stSeasonID).
					Return(nil, nil).Once()
			},
			wantErr: shared.ErrInvalidState,
		},
		{
			name: "нет игроков в клубах",
			setup: func(t *testing.T, h *stHarness) {
				h.superTitles.EXPECT().GetBySeason(mock.Anything, stSeasonID).
					Return(nil, shared.ErrNotFound).Once()
				h.divisions.EXPECT().ListBySeason(mock.Anything, stSeasonID).
					Return([]*division.Division{{ID: stDivID, SeasonID: stSeasonID, Code: "S"}}, nil).Once()
				m := stMakeMember(now, 1, 101, 20, 5)
				h.members.EXPECT().ListByDivision(mock.Anything, stDivID).
					Return([]*division.DivisionMember{m}, nil).Once()
				h.players.EXPECT().GetByID(mock.Anything, shared.PlayerID(101)).
					Return(&player.Player{ID: 101, ClubID: nil}, nil).Once()
			},
			wantErr: shared.ErrInvalidState,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newSTHarness(t, now)
			tt.setup(t, h)

			uc := h.build(t)
			got, err := uc.Execute(context.Background(), dto.AwardSuperTitleInput{SeasonID: stSeasonID})

			require.Error(t, err)
			require.True(t, errors.Is(err, tt.wantErr),
				"expected %v, got %v", tt.wantErr, err)
			require.Nil(t, got)
		})
	}
}
