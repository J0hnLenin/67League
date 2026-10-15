package club_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/usecases/club"
	dombadge "github.com/J0hnLenin/67League/internal/domain/badge"
	domclub "github.com/J0hnLenin/67League/internal/domain/club"
	domplayer "github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/testsupport/mocks"
)

const (
	jcPlayerID = shared.PlayerID(1001)
	jcClubID   = shared.ClubID(500)
	jcBadgeID  = shared.BadgeID(900)
)

type joinClubHarness struct {
	clubs      *mocks.ClubRepository
	players    *mocks.PlayerRepository
	badges     *mocks.BadgeRepository
	clubBadges *mocks.ClubBadgeRepository
	clock      *mocks.Clock
	tx         *mocks.TransactionManager
	now        time.Time
}

func newJoinClubHarness(t *testing.T, now time.Time) *joinClubHarness {
	t.Helper()
	return &joinClubHarness{
		clubs:      mocks.NewClubRepository(t),
		players:    mocks.NewPlayerRepository(t),
		badges:     mocks.NewBadgeRepository(t),
		clubBadges: mocks.NewClubBadgeRepository(t),
		clock:      mocks.NewClock(t),
		tx:         mocks.NewTransactionManager(t),
		now:        now,
	}
}

func (h *joinClubHarness) build(t *testing.T) *club.JoinClub {
	t.Helper()
	uc, err := club.NewJoinClub(club.JoinClubDeps{
		Clubs:      h.clubs,
		Players:    h.players,
		Badges:     h.badges,
		ClubBadges: h.clubBadges,
		Clock:      h.clock,
		Tx:         h.tx,
	})
	require.NoError(t, err)
	return uc
}

func (h *joinClubHarness) expectTx() {
	h.tx.EXPECT().WithinTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Maybe()
}

func jcMakePlayer(t *testing.T, now time.Time, status domplayer.Status, inClub *shared.ClubID) *domplayer.Player {
	t.Helper()
	p, err := domplayer.New(domplayer.NewPlayerParams{
		FullName: "Alice", Email: "alice@example.com",
		PasswordHash: "hashed:x", Nickname: "alice", City: "Moscow",
	}, now)
	require.NoError(t, err)
	p.ID = jcPlayerID
	p.Status = status
	p.ClubID = inClub
	return p
}

func jcMakeClub(t *testing.T, now time.Time, status domclub.Status) *domclub.Club {
	t.Helper()
	c, err := domclub.New("Shibuya", "Tokyo", "", jcPlayerID, now)
	require.NoError(t, err)
	c.ID = jcClubID
	c.Status = status
	return c
}

func TestJoinClub(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	fullBadge := dombadge.Badge{
		ID: jcBadgeID, Code: club.BadgeCodeFullRoster,
		Icon: "👥", Name: "Полный состав",
		Type: dombadge.TypeClub, Rarity: dombadge.RarityCommon,
		CreatedAt: now,
	}

	tests := []struct {
		name             string
		playerStatus     domplayer.Status
		playerClub       *shared.ClubID
		clubStatus       domclub.Status
		membersBefore    int
		badgeInCatalog   bool
		badgeAlreadyHas  bool
		wantErr          error
		wantMemberCount  int
		wantBadgeAwarded bool
	}{
		{
			name:            "успех: 0 → 1, бейдж не выдаётся",
			playerStatus:    domplayer.StatusActive,
			clubStatus:      domclub.StatusActive,
			membersBefore:   0,
			wantMemberCount: 1,
		},
		{
			name:             "успех: 9 → 10, бейдж выдан",
			playerStatus:     domplayer.StatusActive,
			clubStatus:       domclub.StatusActive,
			membersBefore:    9,
			badgeInCatalog:   true,
			wantMemberCount:  10,
			wantBadgeAwarded: true,
		},
		{
			name:             "успех: 10 → 11, бейдж уже есть",
			playerStatus:     domplayer.StatusActive,
			clubStatus:       domclub.StatusActive,
			membersBefore:    10,
			badgeInCatalog:   true,
			badgeAlreadyHas:  true,
			wantMemberCount:  11,
			wantBadgeAwarded: false,
		},
		{
			name:            "успех: 9 → 10, справочник без бейджа",
			playerStatus:    domplayer.StatusActive,
			clubStatus:      domclub.StatusActive,
			membersBefore:   9,
			badgeInCatalog:  false,
			wantMemberCount: 10,
		},
		{
			name:         "уже в этом клубе — конфликт",
			playerStatus: domplayer.StatusActive,
			playerClub:   ptrClubID(jcClubID),
			clubStatus:   domclub.StatusActive,
			wantErr:      shared.ErrConflict,
		},
		{
			name:         "забаненный — forbidden",
			playerStatus: domplayer.StatusBanned,
			clubStatus:   domclub.StatusActive,
			wantErr:      shared.ErrForbidden,
		},
		{
			name:         "клуб неактивен — forbidden",
			playerStatus: domplayer.StatusActive,
			clubStatus:   domclub.StatusArchived,
			wantErr:      shared.ErrForbidden,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := newJoinClubHarness(t, now)
			h.clock.EXPECT().Now().Return(now).Maybe()

			player := jcMakePlayer(t, now, tt.playerStatus, tt.playerClub)
			clubObj := jcMakeClub(t, now, tt.clubStatus)

			h.players.EXPECT().GetByID(mock.Anything, jcPlayerID).
				Return(player, nil).Once()

			// После проверки статуса — читаем клуб.
			if tt.playerStatus.CanPlay() {
				h.clubs.EXPECT().GetByID(mock.Anything, jcClubID).
					Return(clubObj, nil).Once()
			}

			// Дальше идём только если статус клуба active и игрок не в нём.
			proceed := tt.playerStatus.CanPlay() &&
				tt.clubStatus == domclub.StatusActive &&
				!(tt.playerClub != nil && *tt.playerClub == jcClubID)

			var capturedPlayer *domplayer.Player
			if proceed {
				h.players.EXPECT().CountByClub(mock.Anything, jcClubID).
					Return(tt.membersBefore, nil).Once()

				h.players.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*player.Player")).
					Run(func(_ context.Context, p *domplayer.Player) { capturedPlayer = p }).
					Return(nil).Once()

				if tt.membersBefore+1 >= club.FullRosterThreshold {
					if tt.badgeInCatalog {
						h.badges.EXPECT().
							GetByCode(mock.Anything, club.BadgeCodeFullRoster).
							Return(&fullBadge, nil).Once()
						h.clubBadges.EXPECT().
							Has(mock.Anything, jcClubID, jcBadgeID).
							Return(tt.badgeAlreadyHas, nil).Once()

						if !tt.badgeAlreadyHas {
							h.clubBadges.EXPECT().
								Create(mock.Anything, mock.AnythingOfType("*badge.ClubBadge")).
								Run(func(_ context.Context, cb *dombadge.ClubBadge) {
									require.Equal(t, jcClubID, cb.ClubID)
									require.Equal(t, jcBadgeID, cb.BadgeID)
									require.Nil(t, cb.SeasonID)
								}).
								Return(nil).Once()
						}
					} else {
						h.badges.EXPECT().
							GetByCode(mock.Anything, club.BadgeCodeFullRoster).
							Return(nil, shared.ErrNotFound).Once()
					}
				}

				h.expectTx()
			}

			uc := h.build(t)
			got, err := uc.Execute(context.Background(), dto.JoinClubInput{
				PlayerID: jcPlayerID,
				ClubID:   jcClubID,
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
			require.Equal(t, tt.wantMemberCount, got.MemberCount)
			require.Equal(t, tt.wantBadgeAwarded, got.FullBadgeAwarded)
			require.NotNil(t, got.Player)
			require.NotNil(t, got.Player.ClubID)
			require.Equal(t, jcClubID, *got.Player.ClubID)

			require.NotNil(t, capturedPlayer, "player was not updated")
			require.NotNil(t, capturedPlayer.ClubID)
			require.Equal(t, jcClubID, *capturedPlayer.ClubID)
		})
	}
}

func ptrClubID(id shared.ClubID) *shared.ClubID { return &id }
