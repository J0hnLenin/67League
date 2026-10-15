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
	"github.com/J0hnLenin/67League/internal/domain/division"
	domseason "github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/testsupport/mocks"
)

const (
	arPrevSeasonID = shared.SeasonID(70)
	arNextSeasonID = shared.SeasonID(71)

	arPrevDivS = shared.DivisionID(301)
	arPrevDivA = shared.DivisionID(302)
	arPrevDivB = shared.DivisionID(303)

	arNextDivS = shared.DivisionID(401)
	arNextDivA = shared.DivisionID(402)
	arNextDivB = shared.DivisionID(403)
)

type arHarness struct {
	seasons   *mocks.SeasonRepository
	divisions *mocks.DivisionRepository
	members   *mocks.DivisionMemberRepository
	clock     *mocks.Clock
	tx        *mocks.TransactionManager
	now       time.Time
}

func newARHarness(t *testing.T, now time.Time) *arHarness {
	t.Helper()
	return &arHarness{
		seasons:   mocks.NewSeasonRepository(t),
		divisions: mocks.NewDivisionRepository(t),
		members:   mocks.NewDivisionMemberRepository(t),
		clock:     mocks.NewClock(t),
		tx:        mocks.NewTransactionManager(t),
		now:       now,
	}
}

func (h *arHarness) build(t *testing.T) *season.ApplyRotation {
	t.Helper()
	uc, err := season.NewApplyRotation(season.ApplyRotationDeps{
		Seasons:   h.seasons,
		Divisions: h.divisions,
		Members:   h.members,
		Clock:     h.clock,
		Tx:        h.tx,
	})
	require.NoError(t, err)
	return uc
}

func (h *arHarness) expectTx() {
	h.tx.EXPECT().WithinTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
}

func arMakeSeason(id shared.SeasonID, status domseason.Status, k1, k2 int) *domseason.Season {
	set := domseason.DefaultSettings()
	set.K1 = k1
	set.K2 = k2
	return &domseason.Season{
		ID: id, Number: int(id) - 69, Title: "S",
		Status: status, Settings: set,
	}
}

func arMakeDivision(now time.Time, id shared.DivisionID, seasonID shared.SeasonID, code string, order int) *division.Division {
	return &division.Division{
		ID: id, SeasonID: seasonID, Code: code, Order: order,
		Status: division.StatusActive, CreatedAt: now, UpdatedAt: now,
	}
}

func arMakeMember(now time.Time, id shared.DivisionMemberID, pid shared.PlayerID, divID shared.DivisionID, lp float64) *division.DivisionMember {
	m := division.NewMember(divID, pid, now)
	m.ID = id
	m.LP = shared.LP(lp)
	m.MatchesPlayed = 5
	return m
}

// ---- Тесты -----------------------------------------------------------------

func TestApplyRotation_ThreeDivisions(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	h := newARHarness(t, now)
	h.clock.EXPECT().Now().Return(now).Once()

	// k1=1 (S↔A), k2=2 (A↔B)
	prev := arMakeSeason(arPrevSeasonID, domseason.StatusFinished, 1, 2)
	next := arMakeSeason(arNextSeasonID, domseason.StatusAnnounced, 1, 2)

	h.seasons.EXPECT().GetByID(mock.Anything, arPrevSeasonID).Return(prev, nil).Once()
	h.seasons.EXPECT().GetByID(mock.Anything, arNextSeasonID).Return(next, nil).Once()

	prevDivs := []*division.Division{
		arMakeDivision(now, arPrevDivS, arPrevSeasonID, "S", 0),
		arMakeDivision(now, arPrevDivA, arPrevSeasonID, "A", 1),
		arMakeDivision(now, arPrevDivB, arPrevSeasonID, "B", 2),
	}
	nextDivs := []*division.Division{
		arMakeDivision(now, arNextDivS, arNextSeasonID, "S", 0),
		arMakeDivision(now, arNextDivA, arNextSeasonID, "A", 1),
		arMakeDivision(now, arNextDivB, arNextSeasonID, "B", 2),
	}
	h.divisions.EXPECT().ListBySeason(mock.Anything, arPrevSeasonID).Return(prevDivs, nil).Once()
	h.divisions.EXPECT().ListBySeason(mock.Anything, arNextSeasonID).Return(nextDivs, nil).Once()

	// S: 2 игрока, k1=1 → нет повышения, нижний 1 вниз в A.
	// A: 3 игрока, k2=2 → топ-2 вверх в S, нижний 1 вниз в B.
	// B: 2 игрока, k2=2 → топ-2 вверх в A, нет понижения.
	sPlayers := []*division.DivisionMember{
		arMakeMember(now, 1, 101, arPrevDivS, 20),
		arMakeMember(now, 2, 102, arPrevDivS, 15),
	}
	aPlayers := []*division.DivisionMember{
		arMakeMember(now, 3, 103, arPrevDivA, 18),
		arMakeMember(now, 4, 104, arPrevDivA, 12),
		arMakeMember(now, 5, 105, arPrevDivA, 6),
	}
	bPlayers := []*division.DivisionMember{
		arMakeMember(now, 6, 106, arPrevDivB, 10),
		arMakeMember(now, 7, 107, arPrevDivB, 8),
	}
	h.members.EXPECT().ListByDivision(mock.Anything, arPrevDivS).Return(sPlayers, nil).Once()
	h.members.EXPECT().ListByDivision(mock.Anything, arPrevDivA).Return(aPlayers, nil).Once()
	h.members.EXPECT().ListByDivision(mock.Anything, arPrevDivB).Return(bPlayers, nil).Once()

	// Для каждого игрока — проверка, что ещё не в новом сезоне, и создание.
	for _, pid := range []shared.PlayerID{101, 102, 103, 104, 105, 106, 107} {
		h.members.EXPECT().
			GetByPlayerAndSeason(mock.Anything, pid, arNextSeasonID).
			Return(nil, shared.ErrNotFound).Once()
	}
	h.members.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*division.DivisionMember")).
		Return(nil).Times(7)

	h.expectTx()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.ApplyRotationInput{
		PreviousSeasonID: arPrevSeasonID,
		NextSeasonID:     arNextSeasonID,
	})

	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, 7, got.Created)

	dirs := map[shared.PlayerID]dto.RotationDirection{}
	toDiv := map[shared.PlayerID]shared.DivisionID{}
	for _, tr := range got.Transfers {
		dirs[tr.PlayerID] = tr.Direction
		toDiv[tr.PlayerID] = tr.ToDivID
	}

	// S: 101 остаётся в S, 102 вниз в A.
	require.Equal(t, dto.RotationStay, dirs[101])
	require.Equal(t, arNextDivS, toDiv[101])
	require.Equal(t, dto.RotationDown, dirs[102])
	require.Equal(t, arNextDivA, toDiv[102])

	// A: 103 и 104 вверх в S, 105 вниз в B.
	require.Equal(t, dto.RotationUp, dirs[103])
	require.Equal(t, arNextDivS, toDiv[103])
	require.Equal(t, dto.RotationUp, dirs[104])
	require.Equal(t, arNextDivS, toDiv[104])
	require.Equal(t, dto.RotationDown, dirs[105])
	require.Equal(t, arNextDivB, toDiv[105])

	// B: 106 и 107 вверх в A.
	require.Equal(t, dto.RotationUp, dirs[106])
	require.Equal(t, arNextDivA, toDiv[106])
	require.Equal(t, dto.RotationUp, dirs[107])
	require.Equal(t, arNextDivA, toDiv[107])
}

// TestApplyRotation_SkipsExcluded проверяет, что исключённые игроки
// не переносятся в новый сезон — им нужна отдельная регистрация.
func TestApplyRotation_SkipsExcluded(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	h := newARHarness(t, now)
	h.clock.EXPECT().Now().Return(now).Once()

	prev := arMakeSeason(arPrevSeasonID, domseason.StatusFinished, 1, 1)
	next := arMakeSeason(arNextSeasonID, domseason.StatusRegistration, 1, 1)

	h.seasons.EXPECT().GetByID(mock.Anything, arPrevSeasonID).Return(prev, nil).Once()
	h.seasons.EXPECT().GetByID(mock.Anything, arNextSeasonID).Return(next, nil).Once()

	prevDivs := []*division.Division{
		arMakeDivision(now, arPrevDivS, arPrevSeasonID, "S", 0),
		arMakeDivision(now, arPrevDivA, arPrevSeasonID, "A", 1),
	}
	nextDivs := []*division.Division{
		arMakeDivision(now, arNextDivS, arNextSeasonID, "S", 0),
		arMakeDivision(now, arNextDivA, arNextSeasonID, "A", 1),
	}
	h.divisions.EXPECT().ListBySeason(mock.Anything, arPrevSeasonID).Return(prevDivs, nil).Once()
	h.divisions.EXPECT().ListBySeason(mock.Anything, arNextSeasonID).Return(nextDivs, nil).Once()

	excluded := arMakeMember(now, 1, 101, arPrevDivS, 0)
	excluded.Exclude(now)

	h.members.EXPECT().ListByDivision(mock.Anything, arPrevDivS).
		Return([]*division.DivisionMember{excluded}, nil).Once()
	h.members.EXPECT().ListByDivision(mock.Anything, arPrevDivA).
		Return([]*division.DivisionMember{}, nil).Once()

	// Ни GetByPlayerAndSeason, ни Create для исключённого не вызываются.
	// Это проверяет сама mockery: любые лишние вызовы дадут ошибку.

	h.expectTx()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.ApplyRotationInput{
		PreviousSeasonID: arPrevSeasonID,
		NextSeasonID:     arNextSeasonID,
	})

	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, 0, got.Created)
	require.Empty(t, got.Transfers)
}

// TestApplyRotation_SkipsExcludedAndRotatesValid проверяет, что при
// смешанном составе дивизиона исключённые игнорируются, а валидные
// ротируются в обычном порядке.
func TestApplyRotation_SkipsExcludedAndRotatesValid(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	h := newARHarness(t, now)
	h.clock.EXPECT().Now().Return(now).Once()

	prev := arMakeSeason(arPrevSeasonID, domseason.StatusFinished, 1, 1)
	next := arMakeSeason(arNextSeasonID, domseason.StatusRegistration, 1, 1)

	h.seasons.EXPECT().GetByID(mock.Anything, arPrevSeasonID).Return(prev, nil).Once()
	h.seasons.EXPECT().GetByID(mock.Anything, arNextSeasonID).Return(next, nil).Once()

	prevDivs := []*division.Division{
		arMakeDivision(now, arPrevDivS, arPrevSeasonID, "S", 0),
		arMakeDivision(now, arPrevDivA, arPrevSeasonID, "A", 1),
	}
	nextDivs := []*division.Division{
		arMakeDivision(now, arNextDivS, arNextSeasonID, "S", 0),
		arMakeDivision(now, arNextDivA, arNextSeasonID, "A", 1),
	}
	h.divisions.EXPECT().ListBySeason(mock.Anything, arPrevSeasonID).Return(prevDivs, nil).Once()
	h.divisions.EXPECT().ListBySeason(mock.Anything, arNextSeasonID).Return(nextDivs, nil).Once()

	excluded := arMakeMember(now, 1, 101, arPrevDivS, 0)
	excluded.Exclude(now)

	top := arMakeMember(now, 2, 102, arPrevDivS, 20)
	bottom := arMakeMember(now, 3, 103, arPrevDivS, 10)

	// Возвращаем в неотсортированном виде, use case сам сортирует.
	h.members.EXPECT().ListByDivision(mock.Anything, arPrevDivS).
		Return([]*division.DivisionMember{bottom, excluded, top}, nil).Once()
	h.members.EXPECT().ListByDivision(mock.Anything, arPrevDivA).
		Return([]*division.DivisionMember{}, nil).Once()

	// k1=1 в S: повышения нет (idx=0). Понижается только bottom.
	// Исключённый не переносится.
	h.members.EXPECT().
		GetByPlayerAndSeason(mock.Anything, shared.PlayerID(102), arNextSeasonID).
		Return(nil, shared.ErrNotFound).Once()
	h.members.EXPECT().
		GetByPlayerAndSeason(mock.Anything, shared.PlayerID(103), arNextSeasonID).
		Return(nil, shared.ErrNotFound).Once()

	var transfers []shared.PlayerID
	h.members.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*division.DivisionMember")).
		Run(func(_ context.Context, m *division.DivisionMember) {
			transfers = append(transfers, m.PlayerID)
		}).
		Return(nil).Twice()

	h.expectTx()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.ApplyRotationInput{
		PreviousSeasonID: arPrevSeasonID,
		NextSeasonID:     arNextSeasonID,
	})

	require.NoError(t, err)
	require.Equal(t, 2, got.Created)
	require.Len(t, got.Transfers, 2)

	// Исключённого нет среди переносов.
	for _, tr := range got.Transfers {
		require.NotEqual(t, shared.PlayerID(101), tr.PlayerID,
			"excluded player must not be transferred")
	}

	// 102 (верхний по LP) остаётся в S, 103 понижается в A.
	byPlayer := map[shared.PlayerID]dto.PlayerTransfer{}
	for _, tr := range got.Transfers {
		byPlayer[tr.PlayerID] = tr
	}
	require.Equal(t, dto.RotationStay, byPlayer[102].Direction)
	require.Equal(t, arNextDivS, byPlayer[102].ToDivID)
	require.Equal(t, dto.RotationDown, byPlayer[103].Direction)
	require.Equal(t, arNextDivA, byPlayer[103].ToDivID)
}

func TestApplyRotation_SkipsAlreadyMembers(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	h := newARHarness(t, now)
	h.clock.EXPECT().Now().Return(now).Once()

	prev := arMakeSeason(arPrevSeasonID, domseason.StatusFinished, 1, 1)
	next := arMakeSeason(arNextSeasonID, domseason.StatusAnnounced, 1, 1)

	h.seasons.EXPECT().GetByID(mock.Anything, arPrevSeasonID).Return(prev, nil).Once()
	h.seasons.EXPECT().GetByID(mock.Anything, arNextSeasonID).Return(next, nil).Once()

	prevDivs := []*division.Division{
		arMakeDivision(now, arPrevDivS, arPrevSeasonID, "S", 0),
	}
	nextDivs := []*division.Division{
		arMakeDivision(now, arNextDivS, arNextSeasonID, "S", 0),
	}
	h.divisions.EXPECT().ListBySeason(mock.Anything, arPrevSeasonID).Return(prevDivs, nil).Once()
	h.divisions.EXPECT().ListBySeason(mock.Anything, arNextSeasonID).Return(nextDivs, nil).Once()

	m1 := arMakeMember(now, 1, 101, arPrevDivS, 20)
	h.members.EXPECT().ListByDivision(mock.Anything, arPrevDivS).
		Return([]*division.DivisionMember{m1}, nil).Once()

	// Игрок уже есть в новом сезоне — пропускаем без Create.
	existing := division.NewMember(arNextDivS, 101, now)
	h.members.EXPECT().
		GetByPlayerAndSeason(mock.Anything, shared.PlayerID(101), arNextSeasonID).
		Return(existing, nil).Once()

	h.expectTx()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.ApplyRotationInput{
		PreviousSeasonID: arPrevSeasonID,
		NextSeasonID:     arNextSeasonID,
	})

	require.NoError(t, err)
	require.Equal(t, 0, got.Created)
	require.Empty(t, got.Transfers)
}

func TestApplyRotation_Errors(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		setup   func(t *testing.T, h *arHarness, in dto.ApplyRotationInput)
		input   dto.ApplyRotationInput
		wantErr error
	}{
		{
			name: "одинаковые сезоны",
			input: dto.ApplyRotationInput{
				PreviousSeasonID: arPrevSeasonID,
				NextSeasonID:     arPrevSeasonID,
			},
			wantErr: shared.ErrValidation,
		},
		{
			name: "предыдущий сезон не finished",
			setup: func(t *testing.T, h *arHarness, in dto.ApplyRotationInput) {
				s := arMakeSeason(arPrevSeasonID, domseason.StatusActive, 1, 1)
				h.seasons.EXPECT().GetByID(mock.Anything, arPrevSeasonID).Return(s, nil).Once()
			},
			input: dto.ApplyRotationInput{
				PreviousSeasonID: arPrevSeasonID,
				NextSeasonID:     arNextSeasonID,
			},
			wantErr: shared.ErrInvalidState,
		},
		{
			name: "новый сезон уже активен",
			setup: func(t *testing.T, h *arHarness, in dto.ApplyRotationInput) {
				h.seasons.EXPECT().GetByID(mock.Anything, arPrevSeasonID).
					Return(arMakeSeason(arPrevSeasonID, domseason.StatusFinished, 1, 1), nil).Once()
				h.seasons.EXPECT().GetByID(mock.Anything, arNextSeasonID).
					Return(arMakeSeason(arNextSeasonID, domseason.StatusActive, 1, 1), nil).Once()
			},
			input: dto.ApplyRotationInput{
				PreviousSeasonID: arPrevSeasonID,
				NextSeasonID:     arNextSeasonID,
			},
			wantErr: shared.ErrInvalidState,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newARHarness(t, now)
			if tt.setup != nil {
				tt.setup(t, h, tt.input)
			}

			uc := h.build(t)
			got, err := uc.Execute(context.Background(), tt.input)

			require.Error(t, err)
			require.True(t, errors.Is(err, tt.wantErr),
				"expected %v, got %v", tt.wantErr, err)
			require.Nil(t, got)
		})
	}
}
