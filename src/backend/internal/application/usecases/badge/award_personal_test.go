package badge_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/application/usecases/badge"
	dombadge "github.com/J0hnLenin/67League/internal/domain/badge"
	"github.com/J0hnLenin/67League/internal/domain/club"
	"github.com/J0hnLenin/67League/internal/domain/division"
	dommatch "github.com/J0hnLenin/67League/internal/domain/match"
	domplayer "github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/testsupport/mocks"
)

const (
	apPlayerID = shared.PlayerID(1001)
	apSeasonID = shared.SeasonID(70)
	apClubID   = shared.ClubID(501)
)

type apHarness struct {
	badges       *mocks.BadgeRepository
	playerBadges *mocks.PlayerBadgeRepository
	members      *mocks.DivisionMemberRepository
	matches      *mocks.MatchRepository
	players      *mocks.PlayerRepository
	clubs        *mocks.ClubRepository
	clock        *mocks.Clock
	tx           *mocks.TransactionManager
	now          time.Time
}

func newAPHarness(t *testing.T, now time.Time) *apHarness {
	t.Helper()
	return &apHarness{
		badges:       mocks.NewBadgeRepository(t),
		playerBadges: mocks.NewPlayerBadgeRepository(t),
		members:      mocks.NewDivisionMemberRepository(t),
		matches:      mocks.NewMatchRepository(t),
		players:      mocks.NewPlayerRepository(t),
		clubs:        mocks.NewClubRepository(t),
		clock:        mocks.NewClock(t),
		tx:           mocks.NewTransactionManager(t),
		now:          now,
	}
}

func (h *apHarness) build(t *testing.T) *badge.AwardPersonalBadges {
	t.Helper()
	uc, err := badge.NewAwardPersonalBadges(badge.AwardPersonalBadgesDeps{
		Badges:       h.badges,
		PlayerBadges: h.playerBadges,
		Members:      h.members,
		Matches:      h.matches,
		Players:      h.players,
		Clubs:        h.clubs,
		Clock:        h.clock,
		Tx:           h.tx,
	})
	require.NoError(t, err)
	return uc
}

func (h *apHarness) expectTx() {
	h.tx.EXPECT().WithinTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
}

// apMakeMember собирает мембера с заданной статистикой.
func apMakeMember(now time.Time, matches, wins, streakMax, streaksBroken int) *division.DivisionMember {
	m := division.NewMember(shared.DivisionID(310), apPlayerID, now)
	m.ID = 1
	m.MatchesPlayed = matches
	m.Wins = wins
	m.StreakMax = streakMax
	m.StreaksBroken = streaksBroken
	return m
}

func apMakePlayer(now time.Time, clubID *shared.ClubID) *domplayer.Player {
	p, _ := domplayer.New(domplayer.NewPlayerParams{
		FullName: "P", Email: shared.Email("p@x.com"), PasswordHash: "h",
		Nickname: "p", City: "Moscow",
	}, now)
	p.ID = apPlayerID
	p.ClubID = clubID
	return p
}

// expectBadgeAward — навешивает ожидания GetByCode + Has + Create
// (Has=false, значит Create вызывается).
func (h *apHarness) expectBadgeAward(code string, id shared.BadgeID) {
	h.badges.EXPECT().GetByCode(mock.Anything, code).
		Return(&dombadge.Badge{ID: id, Code: code}, nil).Once()
	h.playerBadges.EXPECT().Has(mock.Anything, apPlayerID, id).
		Return(false, nil).Once()
	h.playerBadges.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*badge.PlayerBadge")).
		Return(nil).Once()
}

// expectBadgeSkip — бейджа нет в справочнике.
func (h *apHarness) expectBadgeSkip(code string) {
	h.badges.EXPECT().GetByCode(mock.Anything, code).
		Return(nil, shared.ErrNotFound).Once()
}

// expectBadgeAlready — бейдж уже есть у игрока.
func (h *apHarness) expectBadgeAlready(code string, id shared.BadgeID) {
	h.badges.EXPECT().GetByCode(mock.Anything, code).
		Return(&dombadge.Badge{ID: id, Code: code}, nil).Once()
	h.playerBadges.EXPECT().Has(mock.Anything, apPlayerID, id).
		Return(true, nil).Once()
}

func TestAwardPersonalBadges_FirstStoneOnly(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	h := newAPHarness(t, now)

	// Игрок: 1 партия, 0 побед, без клуба, без партий.
	member := apMakeMember(now, 1, 0, 0, 0)
	player := apMakePlayer(now, nil)

	h.members.EXPECT().GetByPlayerAndSeason(mock.Anything, apPlayerID, apSeasonID).
		Return(member, nil).Once()
	h.players.EXPECT().GetByID(mock.Anything, apPlayerID).Return(player, nil).Once()
	h.matches.EXPECT().ListByPlayer(mock.Anything, apPlayerID, ports.MatchFilter{}).
		Return([]*dommatch.Match{}, nil).Once()

	// Квалифицирован только «Первый камень» — остальные бейджи
	// use case даже не запрашивает.
	h.expectBadgeAward(badge.CodeFirstStone, 101)

	h.clock.EXPECT().Now().Return(now).Once()
	h.expectTx()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.AwardPersonalBadgesInput{
		PlayerID: apPlayerID, SeasonID: apSeasonID,
	})

	require.NoError(t, err)
	require.Len(t, got.Awarded, 1)
	require.Equal(t, shared.BadgeID(101), got.Awarded[0].BadgeID)
}

func TestAwardPersonalBadges_MultiStack(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	h := newAPHarness(t, now)

	// Игрок: 10 партий, 10 побед, стрик 7, снял 2 стрика.
	member := apMakeMember(now, 10, 10, 7, 2)
	player := apMakePlayer(now, nil)

	h.members.EXPECT().GetByPlayerAndSeason(mock.Anything, apPlayerID, apSeasonID).
		Return(member, nil).Once()
	h.players.EXPECT().GetByID(mock.Anything, apPlayerID).Return(player, nil).Once()
	h.matches.EXPECT().ListByPlayer(mock.Anything, apPlayerID, ports.MatchFilter{}).
		Return([]*dommatch.Match{}, nil).Once()

	// 9 бейджей квалифицированы.
	h.expectBadgeAward(badge.CodeFirstStone, 101)
	h.expectBadgeAward(badge.CodeFirstBlood, 102)
	h.expectBadgeAward(badge.CodeLizard, 103)
	h.expectBadgeAward(badge.CodeMarathoner, 104)
	h.expectBadgeAward(badge.CodeStoneEater, 105)
	h.expectBadgeAward(badge.CodeFire, 106)
	h.expectBadgeAward(badge.CodeUnstoppable, 107)
	h.expectBadgeAward(badge.CodeSixSeven, 108)
	h.expectBadgeAward(badge.CodeHeadhunter, 109)

	h.clock.EXPECT().Now().Return(now).Once()
	h.expectTx()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.AwardPersonalBadgesInput{
		PlayerID: apPlayerID, SeasonID: apSeasonID,
	})

	require.NoError(t, err)
	require.Len(t, got.Awarded, 9)
}

func TestAwardPersonalBadges_LuckyAndFriendlyAndGlobalist(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	h := newAPHarness(t, now)

	member := apMakeMember(now, 3, 1, 1, 0)
	ownClub := apClubID
	oppClubID := shared.ClubID(502)
	player := apMakePlayer(now, &ownClub)

	h.members.EXPECT().GetByPlayerAndSeason(mock.Anything, apPlayerID, apSeasonID).
		Return(member, nil).Once()
	h.players.EXPECT().GetByID(mock.Anything, apPlayerID).Return(player, nil).Once()

	// 1 партия с разницей 0.5 и с соперником из другого клуба/города.
	oppID := shared.PlayerID(2002)
	score := 0.5
	winnerID := apPlayerID // локальная переменная, чтобы взять адрес
	m := &dommatch.Match{
		ID: shared.MatchID(4001), SeasonID: apSeasonID,
		BlackPlayerID: apPlayerID, WhitePlayerID: oppID,
		WinnerID: &winnerID, ScoreDiff: &score,
		Status: dommatch.StatusConfirmed,
	}

	h.matches.EXPECT().ListByPlayer(mock.Anything, apPlayerID, ports.MatchFilter{}).
		Return([]*dommatch.Match{m}, nil).Once()

	h.players.EXPECT().GetByID(mock.Anything, oppID).
		Return(&domplayer.Player{ID: oppID, ClubID: &oppClubID}, nil).Once()

	h.clubs.EXPECT().GetByID(mock.Anything, ownClub).
		Return(&club.Club{ID: ownClub, City: "Moscow"}, nil).Once()
	h.clubs.EXPECT().GetByID(mock.Anything, oppClubID).
		Return(&club.Club{ID: oppClubID, City: "Kyoto"}, nil).Once()

	// Квалифицированы: «Первый камень», «Первая кровь» (Wins=1) и три special-бейджа.
	h.expectBadgeAward(badge.CodeFirstStone, 101)
	h.expectBadgeAward(badge.CodeFirstBlood, 102)
	h.expectBadgeAward(badge.CodeLucky, 110)
	h.expectBadgeAward(badge.CodeFriendlyMatch, 111)
	h.expectBadgeAward(badge.CodeGlobalist, 112)
	h.clock.EXPECT().Now().Return(now).Once()
	h.expectTx()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.AwardPersonalBadgesInput{
		PlayerID: apPlayerID, SeasonID: apSeasonID,
	})

	require.NoError(t, err)
	require.Len(t, got.Awarded, 5) // first_stone + lucky + friendly + globalist
}

func TestAwardPersonalBadges_Idempotent(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	h := newAPHarness(t, now)

	member := apMakeMember(now, 10, 10, 7, 2)
	player := apMakePlayer(now, nil)

	h.members.EXPECT().GetByPlayerAndSeason(mock.Anything, apPlayerID, apSeasonID).
		Return(member, nil).Once()
	h.players.EXPECT().GetByID(mock.Anything, apPlayerID).Return(player, nil).Once()
	h.matches.EXPECT().ListByPlayer(mock.Anything, apPlayerID, ports.MatchFilter{}).
		Return([]*dommatch.Match{}, nil).Once()

	// Все бейджи уже выданы.
	h.expectBadgeAlready(badge.CodeFirstStone, 101)
	h.expectBadgeAlready(badge.CodeFirstBlood, 102)
	h.expectBadgeAlready(badge.CodeLizard, 103)
	h.expectBadgeAlready(badge.CodeMarathoner, 104)
	h.expectBadgeAlready(badge.CodeStoneEater, 105)
	h.expectBadgeAlready(badge.CodeFire, 106)
	h.expectBadgeAlready(badge.CodeUnstoppable, 107)
	h.expectBadgeAlready(badge.CodeSixSeven, 108)
	h.expectBadgeAlready(badge.CodeHeadhunter, 109)

	h.clock.EXPECT().Now().Return(now).Once()
	h.expectTx()

	uc := h.build(t)
	got, err := uc.Execute(context.Background(), dto.AwardPersonalBadgesInput{
		PlayerID: apPlayerID, SeasonID: apSeasonID,
	})

	require.NoError(t, err)
	require.Empty(t, got.Awarded)
}
