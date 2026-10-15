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
	domplayer "github.com/J0hnLenin/67League/internal/domain/player"
	domseason "github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/testsupport/mocks"
)

const (
	jsPlayerID   = shared.PlayerID(1001)
	jsSeasonID   = shared.SeasonID(70)
	jsDivisionID = shared.DivisionID(310)
)

type joinSeasonHarness struct {
	players   *mocks.PlayerRepository
	seasons   *mocks.SeasonRepository
	divisions *mocks.DivisionRepository
	members   *mocks.DivisionMemberRepository
	clock     *mocks.Clock
	tx        *mocks.TransactionManager
	now       time.Time
}

func newJoinSeasonHarness(t *testing.T, now time.Time) *joinSeasonHarness {
	t.Helper()
	return &joinSeasonHarness{
		players:   mocks.NewPlayerRepository(t),
		seasons:   mocks.NewSeasonRepository(t),
		divisions: mocks.NewDivisionRepository(t),
		members:   mocks.NewDivisionMemberRepository(t),
		clock:     mocks.NewClock(t),
		tx:        mocks.NewTransactionManager(t),
		now:       now,
	}
}

func (h *joinSeasonHarness) build(t *testing.T) *season.JoinSeason {
	t.Helper()
	uc, err := season.NewJoinSeason(season.JoinSeasonDeps{
		Players:   h.players,
		Seasons:   h.seasons,
		Divisions: h.divisions,
		Members:   h.members,
		Clock:     h.clock,
		Tx:        h.tx,
	})
	require.NoError(t, err)
	return uc
}

func (h *joinSeasonHarness) expectTx() {
	h.tx.EXPECT().WithinTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Maybe()
}

func jsMakePlayer(t *testing.T, now time.Time, status domplayer.Status, rating int) *domplayer.Player {
	t.Helper()
	p, err := domplayer.New(domplayer.NewPlayerParams{
		FullName: "Alice", Email: "alice@example.com",
		PasswordHash: "hashed:x", Nickname: "alice", City: "Moscow",
	}, now)
	require.NoError(t, err)
	p.ID = jsPlayerID
	p.Status = status
	p.Rating = shared.Rating(rating)
	return p
}

func jsMakeSeason(t *testing.T, now time.Time, status domseason.Status) *domseason.Season {
	t.Helper()
	s, err := domseason.New(1, "S1",
		now, now.AddDate(0, 1, 0),
		domseason.DefaultSettings(), now)
	require.NoError(t, err)
	s.ID = jsSeasonID
	s.Status = status
	return s
}

func jsMakeDivision(now time.Time, code string, order int) *division.Division {
	return &division.Division{
		ID:        jsDivisionID,
		SeasonID:  jsSeasonID,
		Code:      code,
		Order:     order,
		Status:    division.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestJoinSeason(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		playerStatus domplayer.Status
		playerRating int
		seasonStatus domseason.Status
		membership   bool // уже член сезона
		ratingMatch  bool // FindForRating находит дивизион
		wantErr      error
		wantDivCode  string
	}{
		{
			name:         "успех: активный сезон, рейтинг попадает",
			playerStatus: domplayer.StatusActive,
			playerRating: 1500,
			seasonStatus: domseason.StatusActive,
			ratingMatch:  true,
			wantDivCode:  "A",
		},
		{
			name:         "успех: registration, рейтинг 0 → низший дивизион",
			playerStatus: domplayer.StatusActive,
			playerRating: 0,
			seasonStatus: domseason.StatusRegistration,
			ratingMatch:  false,
			wantDivCode:  "B",
		},
		{
			name:         "успех: announced, рейтинг 0 → низший дивизион",
			playerStatus: domplayer.StatusActive,
			playerRating: 0,
			seasonStatus: domseason.StatusAnnounced,
			ratingMatch:  false,
			wantDivCode:  "B",
		},
		{
			name:         "сезон закрыт — invalid state",
			playerStatus: domplayer.StatusActive,
			seasonStatus: domseason.StatusFinished,
			wantErr:      shared.ErrInvalidState,
		},
		{
			name:         "уже член сезона",
			playerStatus: domplayer.StatusActive,
			seasonStatus: domseason.StatusActive,
			membership:   true,
			wantErr:      shared.ErrAlreadyExists,
		},
		{
			name:         "забаненный — forbidden",
			playerStatus: domplayer.StatusBanned,
			seasonStatus: domseason.StatusActive,
			wantErr:      shared.ErrForbidden,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := newJoinSeasonHarness(t, now)
			h.clock.EXPECT().Now().Return(now).Maybe()

			player := jsMakePlayer(t, now, tt.playerStatus, tt.playerRating)
			seasonObj := jsMakeSeason(t, now, tt.seasonStatus)

			h.players.EXPECT().GetByID(mock.Anything, jsPlayerID).
				Return(player, nil).Once()

			var proceededToSeason bool
			if tt.playerStatus.CanJoinSeason() {
				proceededToSeason = true
				h.seasons.EXPECT().GetByID(mock.Anything, jsSeasonID).
					Return(seasonObj, nil).Once()
			}

			// Дальше только если сезон в статусе CanJoin и игрок может вступить.
			if proceededToSeason && tt.seasonStatus.CanJoin() {
				if tt.membership {
					h.members.EXPECT().
						GetByPlayerAndSeason(mock.Anything, jsPlayerID, jsSeasonID).
						Return(&division.DivisionMember{ID: 1, PlayerID: jsPlayerID}, nil).Once()
				} else {
					h.members.EXPECT().
						GetByPlayerAndSeason(mock.Anything, jsPlayerID, jsSeasonID).
						Return(nil, shared.ErrNotFound).Once()

					divCode := tt.wantDivCode
					if divCode == "" {
						divCode = "A"
					}
					div := jsMakeDivision(now, divCode, 1)

					if tt.ratingMatch {
						h.divisions.EXPECT().
							FindForRating(mock.Anything, jsSeasonID, shared.Rating(tt.playerRating)).
							Return(div, nil).Once()
					} else {
						h.divisions.EXPECT().
							FindForRating(mock.Anything, jsSeasonID, shared.Rating(tt.playerRating)).
							Return(nil, shared.ErrNotFound).Once()
						h.divisions.EXPECT().
							GetLowest(mock.Anything, jsSeasonID).
							Return(div, nil).Once()
					}

					h.members.EXPECT().
						Create(mock.Anything, mock.AnythingOfType("*division.DivisionMember")).
						Run(func(_ context.Context, m *division.DivisionMember) {
							require.Equal(t, jsPlayerID, m.PlayerID)
							require.Equal(t, jsDivisionID, m.DivisionID)
						}).
						Return(nil).Once()

					h.expectTx()
				}
			}

			uc := h.build(t)
			got, err := uc.Execute(context.Background(), dto.JoinSeasonInput{
				PlayerID: jsPlayerID,
				SeasonID: jsSeasonID,
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
			require.NotNil(t, got.Member)
			require.Equal(t, jsPlayerID, got.Member.PlayerID)
			require.NotNil(t, got.Division)
			require.Equal(t, tt.wantDivCode, got.Division.Code)
			require.NotNil(t, got.Season)
			require.Equal(t, jsSeasonID, got.Season.ID)
		})
	}
}
