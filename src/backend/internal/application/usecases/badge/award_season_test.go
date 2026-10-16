package badge_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/usecases/badge"
	dombadge "github.com/J0hnLenin/67League/internal/domain/badge"
	"github.com/J0hnLenin/67League/internal/domain/division"
	domplayer "github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/testsupport/mocks"
)

const (
	asSeasonID = shared.SeasonID(70)
	asDivS     = shared.DivisionID(301)
	asDivA     = shared.DivisionID(302)

	asClub1 = shared.ClubID(501)
	asClub2 = shared.ClubID(502)
)

// ---- harness ---------------------------------------------------------------

type asHarness struct {
	badges       *mocks.BadgeRepository
	playerBadges *mocks.PlayerBadgeRepository
	members      *mocks.DivisionMemberRepository
	divisions    *mocks.DivisionRepository
	players      *mocks.PlayerRepository
	clubs        *mocks.ClubRepository
	clock        *mocks.Clock
	tx           *mocks.TransactionManager
	now          time.Time
}

func newASHarness(t *testing.T, now time.Time) *asHarness {
	t.Helper()
	return &asHarness{
		badges:       mocks.NewBadgeRepository(t),
		playerBadges: mocks.NewPlayerBadgeRepository(t),
		members:      mocks.NewDivisionMemberRepository(t),
		divisions:    mocks.NewDivisionRepository(t),
		players:      mocks.NewPlayerRepository(t),
		clubs:        mocks.NewClubRepository(t),
		clock:        mocks.NewClock(t),
		tx:           mocks.NewTransactionManager(t),
		now:          now,
	}
}

func (h *asHarness) build(t *testing.T) *badge.AwardSeasonBadges {
	t.Helper()
	uc, err := badge.NewAwardSeasonBadges(badge.AwardSeasonBadgesDeps{
		Badges:       h.badges,
		PlayerBadges: h.playerBadges,
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

func (h *asHarness) expectTx() {
	h.tx.EXPECT().WithinTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
}

// ---- Фабрики ---------------------------------------------------------------

func asMakeMember(
	now time.Time,
	id shared.DivisionMemberID,
	pid shared.PlayerID,
	divID shared.DivisionID,
	lp float64,
	place int,
	status division.MemberStatus,
) *division.DivisionMember {
	m := division.NewMember(divID, pid, now)
	m.ID = id
	m.LP = shared.LP(lp)
	m.Status = status
	if status == division.MemberActive {
		pl := place
		m.FinalPlace = &pl
	}
	return m
}

func asMakePlayer(now time.Time, id shared.PlayerID, clubID *shared.ClubID) *domplayer.Player {
	p, _ := domplayer.New(domplayer.NewPlayerParams{
		FullName:     "P",
		Email:        shared.Email("p@x.com"),
		PasswordHash: "h",
		Nickname:     shared.Nickname("p"),
		City:         "M",
	}, now)
	p.ID = id
	p.ClubID = clubID
	return p
}

// ---- Хелперы ожиданий ------------------------------------------------------

// expectSeasonalAward — badge существует, у игрока за сезон его нет → Create.
//
// Внутри Run ничего не проверяем: порядок обхода клубов недетерминирован
// (map), а assertion внутри Run абортирует весь тест. Состав проверяется
// после Execute через collectAwardedByPlayer.
func (h *asHarness) expectSeasonalAward(_ shared.PlayerID, code string, id shared.BadgeID) {
	// pid оставлен в сигнатуре для явности: тест всё равно ждёт,
	// что именно этот игрок получит бейдж, но проверка идёт после Execute.
	h.badges.EXPECT().
		GetByCode(mock.Anything, code).
		Return(&dombadge.Badge{ID: id, Code: code}, nil).Once()

	h.playerBadges.EXPECT().
		ListByPlayerAndSeason(mock.Anything, mock.Anything, asSeasonID).
		Return(nil, nil).Once()

	h.playerBadges.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*badge.PlayerBadge")).
		Return(nil).Once()
}

// expectSeasonalSkip — badge существует, у игрока за сезон он уже есть.
func (h *asHarness) expectSeasonalSkip(pid shared.PlayerID, code string, id shared.BadgeID) {
	h.badges.EXPECT().
		GetByCode(mock.Anything, code).
		Return(&dombadge.Badge{ID: id, Code: code}, nil).Once()

	h.playerBadges.EXPECT().
		ListByPlayerAndSeason(mock.Anything, pid, asSeasonID).
		Return([]*dombadge.PlayerBadge{{BadgeID: id}}, nil).Once()
}

// expectSeasonalMissing — badge отсутствует в справочнике.
func (h *asHarness) expectSeasonalMissing(code string) {
	h.badges.EXPECT().
		GetByCode(mock.Anything, code).
		Return(nil, shared.ErrNotFound).Once()
}

// collectAwardedByPlayer — превращает плоский список бейджей в множество
// (playerID → set of badgeIDs). Удобно для проверки состава независимо
// от порядка выдачи.
func collectAwardedByPlayer(badges []*dombadge.PlayerBadge) map[shared.PlayerID]map[shared.BadgeID]bool {
	out := map[shared.PlayerID]map[shared.BadgeID]bool{}
	for _, pb := range badges {
		if out[pb.PlayerID] == nil {
			out[pb.PlayerID] = map[shared.BadgeID]bool{}
		}
		out[pb.PlayerID][pb.BadgeID] = true
	}
	return out
}

// ---- Тесты -----------------------------------------------------------------

func TestAwardSeasonBadges_DivisionPlaces(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	h := newASHarness(t, now)

	h.divisions.EXPECT().ListBySeason(mock.Anything, asSeasonID).
		Return([]*division.Division{
			{ID: asDivS, SeasonID: asSeasonID, Code: "S", Order: 0},
		}, nil).Once()

	// 3 активных + 1 исключённый.
	top1 := asMakeMember(now, 1, 101, asDivS, 30, 1, division.MemberActive)
	top2 := asMakeMember(now, 2, 102, asDivS, 20, 2, division.MemberActive)
	top3 := asMakeMember(now, 3, 103, asDivS, 10, 3, division.MemberActive)
	excl := asMakeMember(now, 4, 104, asDivS, 5, 0, division.MemberExcluded)

	h.members.EXPECT().ListByDivision(mock.Anything, asDivS).
		Return([]*division.DivisionMember{top1, top2, top3, excl}, nil).Once()

	// GetByID для всех участников (в awardClubPlaces). У всех ClubID=nil,
	// клубные бейджи не выдаются.
	for _, pid := range []shared.PlayerID{101, 102, 103, 104} {
		h.players.EXPECT().GetByID(mock.Anything, pid).
			Return(asMakePlayer(now, pid, nil), nil).Once()
	}

	// Дивизионные бейджи для топ-3.
	h.expectSeasonalAward(101, badge.CodeSeasonDivGold, 201)
	h.expectSeasonalAward(102, badge.CodeSeasonDivSilver, 202)
	h.expectSeasonalAward(103, badge.CodeSeasonDivBronze, 203)

	h.clock.EXPECT().Now().Return(now).Once()
	h.expectTx()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.AwardSeasonBadgesInput{SeasonID: asSeasonID})

	require.NoError(t, err)
	require.Len(t, got.Awarded, 3)

	byPlayer := collectAwardedByPlayer(got.Awarded)
	require.True(t, byPlayer[101][201], "101 must get division gold")
	require.True(t, byPlayer[102][202], "102 must get division silver")
	require.True(t, byPlayer[103][203], "103 must get division bronze")
	require.False(t, byPlayer[104][203], "excluded player must not get a badge")
}

func TestAwardSeasonBadges_ClubPlaces(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	h := newASHarness(t, now)

	h.divisions.EXPECT().ListBySeason(mock.Anything, asSeasonID).
		Return([]*division.Division{
			{ID: asDivS, SeasonID: asSeasonID, Code: "S", Order: 0},
		}, nil).Once()

	// 4 игрока, все активны.
	// 101 (30) — Club1
	// 102 (25) — Club2
	// 103 (20) — Club1
	// 104 (15) — Club2
	m1 := asMakeMember(now, 1, 101, asDivS, 30, 1, division.MemberActive)
	m2 := asMakeMember(now, 2, 102, asDivS, 25, 2, division.MemberActive)
	m3 := asMakeMember(now, 3, 103, asDivS, 20, 3, division.MemberActive)
	m4 := asMakeMember(now, 4, 104, asDivS, 15, 4, division.MemberActive)

	h.members.EXPECT().ListByDivision(mock.Anything, asDivS).
		Return([]*division.DivisionMember{m1, m2, m3, m4}, nil).Once()

	club1 := asClub1
	club2 := asClub2
	h.players.EXPECT().GetByID(mock.Anything, shared.PlayerID(101)).
		Return(asMakePlayer(now, 101, &club1), nil).Once()
	h.players.EXPECT().GetByID(mock.Anything, shared.PlayerID(102)).
		Return(asMakePlayer(now, 102, &club2), nil).Once()
	h.players.EXPECT().GetByID(mock.Anything, shared.PlayerID(103)).
		Return(asMakePlayer(now, 103, &club1), nil).Once()
	h.players.EXPECT().GetByID(mock.Anything, shared.PlayerID(104)).
		Return(asMakePlayer(now, 104, &club2), nil).Once()

	// Дивизионные бейджи: топ-3.
	h.expectSeasonalAward(101, badge.CodeSeasonDivGold, 201)
	h.expectSeasonalAward(102, badge.CodeSeasonDivSilver, 202)
	h.expectSeasonalAward(103, badge.CodeSeasonDivBronze, 203)

	// Клубные бейджи:
	//   Club1: 101 (30, gold), 103 (20, silver)
	//   Club2: 102 (25, gold), 104 (15, silver)
	h.expectSeasonalAward(101, badge.CodeSeasonClubGold, 211)
	h.expectSeasonalAward(103, badge.CodeSeasonClubSilver, 212)
	h.expectSeasonalAward(102, badge.CodeSeasonClubGold, 211)
	h.expectSeasonalAward(104, badge.CodeSeasonClubSilver, 212)

	h.clock.EXPECT().Now().Return(now).Once()
	h.expectTx()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.AwardSeasonBadgesInput{SeasonID: asSeasonID})

	require.NoError(t, err)
	require.Len(t, got.Awarded, 7)

	byPlayer := collectAwardedByPlayer(got.Awarded)

	// Дивизионные
	require.True(t, byPlayer[101][201], "101: division gold")
	require.True(t, byPlayer[102][202], "102: division silver")
	require.True(t, byPlayer[103][203], "103: division bronze")

	// Клубные
	require.True(t, byPlayer[101][211], "101: club1 gold")
	require.True(t, byPlayer[103][212], "103: club1 silver")
	require.True(t, byPlayer[102][211], "102: club2 gold")
	require.True(t, byPlayer[104][212], "104: club2 silver")
}

func TestAwardSeasonBadges_Idempotent(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	h := newASHarness(t, now)

	h.divisions.EXPECT().ListBySeason(mock.Anything, asSeasonID).
		Return([]*division.Division{
			{ID: asDivS, SeasonID: asSeasonID, Code: "S", Order: 0},
		}, nil).Once()

	top1 := asMakeMember(now, 1, 101, asDivS, 30, 1, division.MemberActive)
	h.members.EXPECT().ListByDivision(mock.Anything, asDivS).
		Return([]*division.DivisionMember{top1}, nil).Once()

	// Без клуба — только дивизионный пайплайн.
	h.players.EXPECT().GetByID(mock.Anything, shared.PlayerID(101)).
		Return(asMakePlayer(now, 101, nil), nil).Once()

	// Бейдж уже выдан → Create не вызывается.
	h.expectSeasonalSkip(101, badge.CodeSeasonDivGold, 201)

	h.clock.EXPECT().Now().Return(now).Once()
	h.expectTx()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.AwardSeasonBadgesInput{SeasonID: asSeasonID})

	require.NoError(t, err)
	require.Empty(t, got.Awarded)
}

func TestAwardSeasonBadges_MissingBadgeInCatalog(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	h := newASHarness(t, now)

	h.divisions.EXPECT().ListBySeason(mock.Anything, asSeasonID).
		Return([]*division.Division{
			{ID: asDivS, SeasonID: asSeasonID, Code: "S", Order: 0},
		}, nil).Once()

	top1 := asMakeMember(now, 1, 101, asDivS, 30, 1, division.MemberActive)
	h.members.EXPECT().ListByDivision(mock.Anything, asDivS).
		Return([]*division.DivisionMember{top1}, nil).Once()

	h.players.EXPECT().GetByID(mock.Anything, shared.PlayerID(101)).
		Return(asMakePlayer(now, 101, nil), nil).Once()

	// Дивизионный бейдж отсутствует в справочнике — не ошибка.
	h.expectSeasonalMissing(badge.CodeSeasonDivGold)

	h.clock.EXPECT().Now().Return(now).Once()
	h.expectTx()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.AwardSeasonBadgesInput{SeasonID: asSeasonID})

	require.NoError(t, err)
	require.Empty(t, got.Awarded)
}
