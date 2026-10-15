package season_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/usecases/season"
	"github.com/J0hnLenin/67League/internal/domain/club"
	"github.com/J0hnLenin/67League/internal/domain/division"
	domplayer "github.com/J0hnLenin/67League/internal/domain/player"
	domseason "github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/testsupport/mocks"
)

const (
	csSeasonID    = shared.SeasonID(70)
	csDivisionS   = shared.DivisionID(301)
	csDivisionA   = shared.DivisionID(302)
	csClubShibuya = shared.ClubID(501)
	csClubKyoto   = shared.ClubID(502)
)

type csHarness struct {
	seasons   *mocks.SeasonRepository
	divisions *mocks.DivisionRepository
	members   *mocks.DivisionMemberRepository
	players   *mocks.PlayerRepository
	clubs     *mocks.ClubRepository
	clock     *mocks.Clock
	tx        *mocks.TransactionManager
	now       time.Time
}

func newCSHarness(t *testing.T, now time.Time) *csHarness {
	t.Helper()
	return &csHarness{
		seasons:   mocks.NewSeasonRepository(t),
		divisions: mocks.NewDivisionRepository(t),
		members:   mocks.NewDivisionMemberRepository(t),
		players:   mocks.NewPlayerRepository(t),
		clubs:     mocks.NewClubRepository(t),
		clock:     mocks.NewClock(t),
		tx:        mocks.NewTransactionManager(t),
		now:       now,
	}
}

func (h *csHarness) build(t *testing.T) *season.CalculateSeasonResults {
	t.Helper()
	uc, err := season.NewCalculateSeasonResults(season.CalculateSeasonResultsDeps{
		Seasons:   h.seasons,
		Divisions: h.divisions,
		Members:   h.members,
		Players:   h.players,
		Clubs:     h.clubs,
		Clock:     h.clock,
		Tx:        h.tx,
	})
	require.NoError(t, err)
	return uc
}

func (h *csHarness) expectTx() {
	h.tx.EXPECT().WithinTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
}

// ---- Фабрики ---------------------------------------------------------------

func csMakeSeason(now time.Time, p int) *domseason.Season {
	set := domseason.DefaultSettings()
	set.P = p
	return &domseason.Season{
		ID: csSeasonID, Number: 1, Title: "S1",
		Status: domseason.StatusFinished, Settings: set,
		CreatedAt: now, UpdatedAt: now,
	}
}

func csMakeDivision(now time.Time, id shared.DivisionID, code string, order int) *division.Division {
	return &division.Division{
		ID: id, SeasonID: csSeasonID, Code: code, Order: order,
		Status: division.StatusActive, CreatedAt: now, UpdatedAt: now,
	}
}

func csMakeMember(
	now time.Time,
	id shared.DivisionMemberID,
	pid shared.PlayerID,
	divID shared.DivisionID,
	lp float64,
	matches, wins int,
) *division.DivisionMember {
	m := division.NewMember(divID, pid, now)
	m.ID = id
	m.LP = shared.LP(lp)
	m.MatchesPlayed = matches
	m.Wins = wins
	return m
}

func csMakePlayer(now time.Time, id shared.PlayerID, clubID *shared.ClubID) *domplayer.Player {
	p, _ := domplayer.New(domplayer.NewPlayerParams{
		FullName: "P", Email: shared.Email("p@x.com"),
		PasswordHash: "h", Nickname: shared.Nickname("p"), City: "M",
	}, now)
	p.ID = id
	p.ClubID = clubID
	return p
}

// expectMembersUpdate — общий мок для Update дивизион-мемберов.
func expectMembersUpdate(
	h *csHarness,
	times int,
	assert func(mm *division.DivisionMember),
) {
	h.members.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*division.DivisionMember")).
		Run(func(_ context.Context, mm *division.DivisionMember) {
			if assert != nil {
				assert(mm)
			}
		}).
		Return(nil).Times(times)
}

// ---- Тесты -----------------------------------------------------------------

func TestCalculateSeasonResults_SingleDivision(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC)
	h := newCSHarness(t, now)
	h.clock.EXPECT().Now().Return(now).Once()

	seasonObj := csMakeSeason(now, 1)
	h.seasons.EXPECT().GetByID(mock.Anything, csSeasonID).Return(seasonObj, nil).Once()

	div := csMakeDivision(now, csDivisionS, "S", 0)
	h.divisions.EXPECT().ListBySeason(mock.Anything, csSeasonID).Return([]*division.Division{div}, nil).Once()

	aliceClub := csClubShibuya
	bobClub := csClubShibuya

	alice := csMakeMember(now, 1, 101, csDivisionS, 12, 5, 4)
	bob := csMakeMember(now, 2, 102, csDivisionS, 9, 5, 3)
	carol := csMakeMember(now, 3, 103, csDivisionS, 9, 3, 2) // тот же LP, меньше партий

	h.members.EXPECT().ListByDivision(mock.Anything, csDivisionS).
		Return([]*division.DivisionMember{alice, bob, carol}, nil).Once()

	// 3 Update для 3 участников (все прошли порог p=1).
	expectMembersUpdate(h, 3, func(mm *division.DivisionMember) {
		require.NotNil(t, mm.FinalPlace)
	})

	// Каждый — клубный игрок.
	h.players.EXPECT().GetByID(mock.Anything, shared.PlayerID(101)).
		Return(csMakePlayer(now, 101, &aliceClub), nil).Once()
	h.players.EXPECT().GetByID(mock.Anything, shared.PlayerID(102)).
		Return(csMakePlayer(now, 102, &bobClub), nil).Once()
	h.players.EXPECT().GetByID(mock.Anything, shared.PlayerID(103)).
		Return(csMakePlayer(now, 103, nil), nil).Once() // без клуба

	h.clubs.EXPECT().GetByID(mock.Anything, csClubShibuya).
		Return(&club.Club{ID: csClubShibuya, Name: "Shibuya"}, nil).Once()

	h.expectTx()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.CalculateSeasonResultsInput{SeasonID: csSeasonID})

	require.NoError(t, err)
	require.NotNil(t, got)
	require.Len(t, got.Standings, 1)
	require.Len(t, got.Standings[0].Rows, 3)

	// Порядок: Alice (12) > Bob (9, 5 партий) > Carol (9, 3 партии).
	require.Equal(t, 1, got.Standings[0].Rows[0].Place)
	require.Equal(t, shared.PlayerID(101), got.Standings[0].Rows[0].PlayerID)
	require.Equal(t, 2, got.Standings[0].Rows[1].Place)
	require.Equal(t, shared.PlayerID(102), got.Standings[0].Rows[1].PlayerID)
	require.Equal(t, 3, got.Standings[0].Rows[2].Place)
	require.Equal(t, shared.PlayerID(103), got.Standings[0].Rows[2].PlayerID)

	require.Len(t, got.ClubTable, 1)
	require.Equal(t, csClubShibuya, got.ClubTable[0].ClubID)
	require.Equal(t, shared.LP(21), got.ClubTable[0].TotalLP) // 12+9
	require.Equal(t, 2, got.ClubTable[0].Members)
	require.NotNil(t, got.BestClub)
	require.Equal(t, csClubShibuya, got.BestClub.ClubID)
}

func TestCalculateSeasonResults_ExcludedBelowThreshold(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC)
	h := newCSHarness(t, now)
	h.clock.EXPECT().Now().Return(now).Once()

	seasonObj := csMakeSeason(now, 3) // порог 3 партии
	h.seasons.EXPECT().GetByID(mock.Anything, csSeasonID).Return(seasonObj, nil).Once()

	div := csMakeDivision(now, csDivisionS, "S", 0)
	h.divisions.EXPECT().ListBySeason(mock.Anything, csSeasonID).
		Return([]*division.Division{div}, nil).Once()

	active := csMakeMember(now, 1, 101, csDivisionS, 5, 5, 4)
	excluded := csMakeMember(now, 2, 102, csDivisionS, 3, 1, 1) // 1 партия < 3

	h.members.EXPECT().ListByDivision(mock.Anything, csDivisionS).
		Return([]*division.DivisionMember{active, excluded}, nil).Once()

	// Update идёт по каждому: active — FinalPlace, excluded — Exclude.
	h.members.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*division.DivisionMember")).
		Run(func(_ context.Context, mm *division.DivisionMember) {
			switch mm.ID {
			case 1:
				require.NotNil(t, mm.FinalPlace)
				require.Equal(t, 1, *mm.FinalPlace)
			case 2:
				require.Nil(t, mm.FinalPlace)
				require.Equal(t, division.MemberExcluded, mm.Status)
			}
		}).
		Return(nil).Twice()

	// Только active идёт в клубный зачёт.
	shibuya := csClubShibuya
	h.players.EXPECT().GetByID(mock.Anything, shared.PlayerID(101)).
		Return(csMakePlayer(now, 101, &shibuya), nil).Once()
	h.clubs.EXPECT().GetByID(mock.Anything, csClubShibuya).
		Return(&club.Club{ID: csClubShibuya, Name: "Shibuya"}, nil).Once()

	h.expectTx()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.CalculateSeasonResultsInput{SeasonID: csSeasonID})

	require.NoError(t, err)
	require.Len(t, got.Standings[0].Rows, 2)
	require.True(t, got.Standings[0].Rows[1].Excluded)
	require.Equal(t, shared.PlayerID(102), got.Standings[0].Rows[1].PlayerID)
	require.Len(t, got.ClubTable, 1)
	require.Equal(t, 1, got.ClubTable[0].Members)
}

func TestCalculateSeasonResults_MultipleDivisionsAndBestClubTiebreak(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC)
	h := newCSHarness(t, now)
	h.clock.EXPECT().Now().Return(now).Once()

	seasonObj := csMakeSeason(now, 1)
	h.seasons.EXPECT().GetByID(mock.Anything, csSeasonID).Return(seasonObj, nil).Once()

	divS := csMakeDivision(now, csDivisionS, "S", 0)
	divA := csMakeDivision(now, csDivisionA, "A", 1)
	h.divisions.EXPECT().ListBySeason(mock.Anything, csSeasonID).
		Return([]*division.Division{divS, divA}, nil).Once()

	// Дивизион S
	a1 := csMakeMember(now, 1, 101, csDivisionS, 10, 5, 4) // Shibuya
	a2 := csMakeMember(now, 2, 102, csDivisionS, 8, 5, 3)  // Kyoto

	// Дивизион A
	b1 := csMakeMember(now, 3, 103, csDivisionA, 9, 5, 4) // Shibuya
	b2 := csMakeMember(now, 4, 104, csDivisionA, 7, 5, 2) // Kyoto

	h.members.EXPECT().ListByDivision(mock.Anything, csDivisionS).
		Return([]*division.DivisionMember{a1, a2}, nil).Once()
	h.members.EXPECT().ListByDivision(mock.Anything, csDivisionA).
		Return([]*division.DivisionMember{b1, b2}, nil).Once()

	// Update: 4 мембера.
	expectMembersUpdate(h, 4, nil)

	// Игроки.
	shibuya := csClubShibuya
	kyoto := csClubKyoto
	h.players.EXPECT().GetByID(mock.Anything, shared.PlayerID(101)).
		Return(csMakePlayer(now, 101, &shibuya), nil).Once()
	h.players.EXPECT().GetByID(mock.Anything, shared.PlayerID(102)).
		Return(csMakePlayer(now, 102, &kyoto), nil).Once()
	h.players.EXPECT().GetByID(mock.Anything, shared.PlayerID(103)).
		Return(csMakePlayer(now, 103, &shibuya), nil).Once()
	h.players.EXPECT().GetByID(mock.Anything, shared.PlayerID(104)).
		Return(csMakePlayer(now, 104, &kyoto), nil).Once()

	// Клубы: по одному разу.
	h.clubs.EXPECT().GetByID(mock.Anything, csClubShibuya).
		Return(&club.Club{ID: csClubShibuya, Name: "Shibuya"}, nil).Once()
	h.clubs.EXPECT().GetByID(mock.Anything, csClubKyoto).
		Return(&club.Club{ID: csClubKyoto, Name: "Kyoto"}, nil).Once()

	h.expectTx()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.CalculateSeasonResultsInput{SeasonID: csSeasonID})

	require.NoError(t, err)
	require.Len(t, got.ClubTable, 2)

	// Shibuya: 10+9 = 19. Kyoto: 8+7 = 15. Shibuya — лучший.
	require.Equal(t, csClubShibuya, got.ClubTable[0].ClubID)
	require.Equal(t, shared.LP(19), got.ClubTable[0].TotalLP)
	require.Equal(t, csClubKyoto, got.ClubTable[1].ClubID)
	require.Equal(t, shared.LP(15), got.ClubTable[1].TotalLP)
	require.NotNil(t, got.BestClub)
	require.Equal(t, csClubShibuya, got.BestClub.ClubID)
}

func TestCalculateSeasonResults_Errors(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		setup   func(t *testing.T, h *csHarness)
		wantErr error
	}{
		{
			name: "сезон не в статусе finished",
			setup: func(t *testing.T, h *csHarness) {
				s := csMakeSeason(h.now, 1)
				s.Status = domseason.StatusActive
				h.seasons.EXPECT().GetByID(mock.Anything, csSeasonID).Return(s, nil).Once()
			},
			wantErr: shared.ErrInvalidState,
		},
		{
			name: "нет дивизионов",
			setup: func(t *testing.T, h *csHarness) {
				h.seasons.EXPECT().GetByID(mock.Anything, csSeasonID).
					Return(csMakeSeason(h.now, 1), nil).Once()
				h.divisions.EXPECT().ListBySeason(mock.Anything, csSeasonID).
					Return(nil, nil).Once()
			},
			wantErr: shared.ErrInvalidState,
		},
		{
			name: "сезон не найден",
			setup: func(t *testing.T, h *csHarness) {
				h.seasons.EXPECT().GetByID(mock.Anything, csSeasonID).
					Return(nil, shared.ErrNotFound).Once()
			},
			wantErr: shared.ErrNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newCSHarness(t, now)
			h.clock.EXPECT().Now().Return(now).Maybe()
			tt.setup(t, h)

			uc := h.build(t)
			got, err := uc.Execute(context.Background(),
				dto.CalculateSeasonResultsInput{SeasonID: csSeasonID})

			require.Error(t, err)
			require.True(t, errors.Is(err, tt.wantErr),
				"expected %v, got %v", tt.wantErr, err)
			require.Nil(t, got)
		})
	}
}
