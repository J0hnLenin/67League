package player_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/application/usecases/player"
	"github.com/J0hnLenin/67League/internal/domain/club"
	"github.com/J0hnLenin/67League/internal/domain/division"
	domplayer "github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/season"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/testsupport/mocks"
)

// ---- harness ---------------------------------------------------------------

type registerHarness struct {
	players   *mocks.PlayerRepository
	clubs     *mocks.ClubRepository
	seasons   *mocks.SeasonRepository
	divisions *mocks.DivisionRepository
	members   *mocks.DivisionMemberRepository
	rating    *mocks.RatingProvider
	hasher    *mocks.PasswordHasher
	clock     *mocks.Clock
	tx        *mocks.TransactionManager
	now       time.Time
}

func newRegisterHarness(t *testing.T, now time.Time) *registerHarness {
	return &registerHarness{
		players:   mocks.NewPlayerRepository(t),
		clubs:     mocks.NewClubRepository(t),
		seasons:   mocks.NewSeasonRepository(t),
		divisions: mocks.NewDivisionRepository(t),
		members:   mocks.NewDivisionMemberRepository(t),
		rating:    mocks.NewRatingProvider(t),
		hasher:    mocks.NewPasswordHasher(t),
		clock:     mocks.NewClock(t),
		tx:        mocks.NewTransactionManager(t),
		now:       now,
	}
}

func (h *registerHarness) build(t *testing.T) *player.RegisterPlayer {
	t.Helper()
	uc, err := player.NewRegisterPlayer(player.RegisterPlayerDeps{
		Players:   h.players,
		Clubs:     h.clubs,
		Seasons:   h.seasons,
		Divisions: h.divisions,
		Members:   h.members,
		Rating:    h.rating,
		Hasher:    h.hasher,
		Clock:     h.clock,
		Tx:        h.tx,
	})
	require.NoError(t, err)
	return uc
}

func (h *registerHarness) expectTxPassthrough() {
	h.tx.EXPECT().WithinTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
}

func (h *registerHarness) expectClockMaybe() {
	h.clock.EXPECT().Now().Return(h.now).Maybe()
}

// ---- тест ------------------------------------------------------------------

func TestRegisterPlayer(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	baseInput := dto.RegisterPlayerInput{
		FullName: "Alice",
		Email:    shared.Email("alice@example.com"),
		Password: "secret123",
		Nickname: shared.Nickname("alice"),
		City:     "Moscow",
	}

	type want struct {
		rating       shared.Rating
		inClub       bool
		divisionCode string
	}

	tests := []struct {
		name    string
		setup   func(t *testing.T, h *registerHarness, in dto.RegisterPlayerInput)
		input   func(h *registerHarness) dto.RegisterPlayerInput
		wantErr error
		want    want
	}{
		{
			name: "успех: без клуба, без сезона, без рейтинга",
			setup: func(t *testing.T, h *registerHarness, in dto.RegisterPlayerInput) {
				h.expectClockMaybe()
				h.players.EXPECT().ExistsByEmail(mock.Anything, in.Email).Return(false, nil).Once()
				h.players.EXPECT().ExistsByNickname(mock.Anything, in.Nickname).Return(false, nil).Once()
				h.rating.EXPECT().FetchByRFGNick(mock.Anything, string(in.Nickname)).
					Return(nil, shared.ErrNotFound).Once()
				h.hasher.EXPECT().Hash(in.Password).Return("hashed:"+in.Password, nil).Once()
				h.seasons.EXPECT().GetCurrent(mock.Anything).Return(nil, shared.ErrNotFound).Once()
				h.players.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*player.Player")).
					Return(nil).Once()
				h.expectTxPassthrough()
			},
			input: func(_ *registerHarness) dto.RegisterPlayerInput { return baseInput },
			want:  want{rating: 0},
		},
		{
			name: "успех: рейтинг из РФГ по ID",
			setup: func(t *testing.T, h *registerHarness, in dto.RegisterPlayerInput) {
				h.expectClockMaybe()
				h.players.EXPECT().ExistsByEmail(mock.Anything, in.Email).Return(false, nil).Once()
				h.players.EXPECT().ExistsByNickname(mock.Anything, in.Nickname).Return(false, nil).Once()
				h.rating.EXPECT().FetchByRFGID(mock.Anything, "42").
					Return(&ports.ExternalRating{Rating: 1500, Source: "rfg"}, nil).Once()
				h.hasher.EXPECT().Hash(in.Password).Return("hashed:"+in.Password, nil).Once()
				h.seasons.EXPECT().GetCurrent(mock.Anything).Return(nil, shared.ErrNotFound).Once()
				h.players.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*player.Player")).
					Return(nil).Once()
				h.expectTxPassthrough()
			},
			input: func(_ *registerHarness) dto.RegisterPlayerInput {
				in := baseInput
				in.RFGID = "42"
				return in
			},
			want: want{rating: 1500},
		},
		{
			name: "email занят",
			setup: func(t *testing.T, h *registerHarness, in dto.RegisterPlayerInput) {
				h.players.EXPECT().ExistsByEmail(mock.Anything, in.Email).Return(true, nil).Once()
			},
			input:   func(_ *registerHarness) dto.RegisterPlayerInput { return baseInput },
			wantErr: shared.ErrAlreadyExists,
		},
		{
			name: "ник занят",
			setup: func(t *testing.T, h *registerHarness, in dto.RegisterPlayerInput) {
				h.players.EXPECT().ExistsByEmail(mock.Anything, in.Email).Return(false, nil).Once()
				h.players.EXPECT().ExistsByNickname(mock.Anything, in.Nickname).Return(true, nil).Once()
			},
			input:   func(_ *registerHarness) dto.RegisterPlayerInput { return baseInput },
			wantErr: shared.ErrAlreadyExists,
		},
		{
			name:  "пустой email",
			setup: func(_ *testing.T, _ *registerHarness, _ dto.RegisterPlayerInput) {},
			input: func(_ *registerHarness) dto.RegisterPlayerInput {
				in := baseInput
				in.Email = ""
				return in
			},
			wantErr: shared.ErrValidation,
		},
		{
			name: "короткий пароль",
			setup: func(t *testing.T, h *registerHarness, in dto.RegisterPlayerInput) {
				h.players.EXPECT().ExistsByEmail(mock.Anything, in.Email).Return(false, nil).Once()
				h.players.EXPECT().ExistsByNickname(mock.Anything, in.Nickname).Return(false, nil).Once()
			},
			input: func(_ *registerHarness) dto.RegisterPlayerInput {
				in := baseInput
				in.Password = "123"
				return in
			},
			wantErr: shared.ErrValidation,
		},
		{
			name: "вступление в клуб",
			setup: func(t *testing.T, h *registerHarness, in dto.RegisterPlayerInput) {
				h.expectClockMaybe()
				h.players.EXPECT().ExistsByEmail(mock.Anything, in.Email).Return(false, nil).Once()
				h.players.EXPECT().ExistsByNickname(mock.Anything, in.Nickname).Return(false, nil).Once()
				h.rating.EXPECT().FetchByRFGNick(mock.Anything, string(in.Nickname)).
					Return(nil, shared.ErrNotFound).Once()
				h.hasher.EXPECT().Hash(in.Password).Return("hashed:"+in.Password, nil).Once()
				h.seasons.EXPECT().GetCurrent(mock.Anything).Return(nil, shared.ErrNotFound).Once()
				h.players.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*player.Player")).
					Return(nil).Once()
				h.clubs.EXPECT().GetByID(mock.Anything, *in.ClubID).
					Return(&club.Club{ID: *in.ClubID, Name: "Shibuya"}, nil).Once()
				h.players.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*player.Player")).
					Return(nil).Once()
				h.expectTxPassthrough()
			},
			input: func(_ *registerHarness) dto.RegisterPlayerInput {
				in := baseInput
				id := shared.NewClubID()
				in.ClubID = &id
				return in
			},
			want: want{inClub: true},
		},
		{
			name: "несуществующий клуб",
			setup: func(t *testing.T, h *registerHarness, in dto.RegisterPlayerInput) {
				h.expectClockMaybe()
				h.players.EXPECT().ExistsByEmail(mock.Anything, in.Email).Return(false, nil).Once()
				h.players.EXPECT().ExistsByNickname(mock.Anything, in.Nickname).Return(false, nil).Once()
				h.rating.EXPECT().FetchByRFGNick(mock.Anything, string(in.Nickname)).
					Return(nil, shared.ErrNotFound).Once()
				h.hasher.EXPECT().Hash(in.Password).Return("hashed:"+in.Password, nil).Once()
				h.seasons.EXPECT().GetCurrent(mock.Anything).Return(nil, shared.ErrNotFound).Once()
				h.players.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*player.Player")).
					Return(nil).Once()
				h.clubs.EXPECT().GetByID(mock.Anything, *in.ClubID).
					Return(nil, shared.ErrNotFound).Once()
				h.expectTxPassthrough()
			},
			input: func(_ *registerHarness) dto.RegisterPlayerInput {
				in := baseInput
				id := shared.NewClubID()
				in.ClubID = &id
				return in
			},
			wantErr: shared.ErrNotFound,
		},
		{
			name: "сезон active: назначен дивизион по рейтингу",
			setup: func(t *testing.T, h *registerHarness, in dto.RegisterPlayerInput) {
				h.expectClockMaybe()

				seasonID := shared.NewSeasonID()
				div := &division.Division{
					ID:       shared.NewDivisionID(),
					SeasonID: seasonID,
					Code:     "A",
					Order:    1,
				}

				h.players.EXPECT().ExistsByEmail(mock.Anything, in.Email).Return(false, nil).Once()
				h.players.EXPECT().ExistsByNickname(mock.Anything, in.Nickname).Return(false, nil).Once()
				h.rating.EXPECT().FetchByRFGID(mock.Anything, "42").
					Return(&ports.ExternalRating{Rating: 1500, Source: "rfg"}, nil).Once()
				h.hasher.EXPECT().Hash(in.Password).Return("hashed:"+in.Password, nil).Once()
				h.seasons.EXPECT().GetCurrent(mock.Anything).
					Return(&season.Season{ID: seasonID, Status: season.StatusActive}, nil).Once()
				h.divisions.EXPECT().FindForRating(mock.Anything, seasonID, shared.Rating(1500)).
					Return(div, nil).Once()
				h.players.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*player.Player")).
					Return(nil).Once()
				h.members.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*division.DivisionMember")).
					Return(nil).Once()
				h.expectTxPassthrough()
			},
			input: func(_ *registerHarness) dto.RegisterPlayerInput {
				in := baseInput
				in.RFGID = "42"
				return in
			},
			want: want{rating: 1500, divisionCode: "A"},
		},
		{
			name: "сезон active: рейтинг не найден → низший дивизион",
			setup: func(t *testing.T, h *registerHarness, in dto.RegisterPlayerInput) {
				h.expectClockMaybe()

				seasonID := shared.NewSeasonID()
				lowest := &division.Division{
					ID:       shared.NewDivisionID(),
					SeasonID: seasonID,
					Code:     "B",
					Order:    2,
				}

				h.players.EXPECT().ExistsByEmail(mock.Anything, in.Email).Return(false, nil).Once()
				h.players.EXPECT().ExistsByNickname(mock.Anything, in.Nickname).Return(false, nil).Once()
				h.rating.EXPECT().FetchByRFGNick(mock.Anything, string(in.Nickname)).
					Return(nil, shared.ErrNotFound).Once()
				h.hasher.EXPECT().Hash(in.Password).Return("hashed:"+in.Password, nil).Once()
				h.seasons.EXPECT().GetCurrent(mock.Anything).
					Return(&season.Season{ID: seasonID, Status: season.StatusActive}, nil).Once()
				h.divisions.EXPECT().FindForRating(mock.Anything, seasonID, shared.Rating(0)).
					Return(nil, shared.ErrNotFound).Once()
				h.divisions.EXPECT().GetLowest(mock.Anything, seasonID).Return(lowest, nil).Once()
				h.players.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*player.Player")).
					Return(nil).Once()
				h.members.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*division.DivisionMember")).
					Return(nil).Once()
				h.expectTxPassthrough()
			},
			input: func(_ *registerHarness) dto.RegisterPlayerInput { return baseInput },
			want:  want{rating: 0, divisionCode: "B"},
		},
		{
			name: "сезон finished: дивизион не назначается",
			setup: func(t *testing.T, h *registerHarness, in dto.RegisterPlayerInput) {
				h.expectClockMaybe()

				seasonID := shared.NewSeasonID()

				h.players.EXPECT().ExistsByEmail(mock.Anything, in.Email).Return(false, nil).Once()
				h.players.EXPECT().ExistsByNickname(mock.Anything, in.Nickname).Return(false, nil).Once()
				h.rating.EXPECT().FetchByRFGNick(mock.Anything, string(in.Nickname)).
					Return(nil, shared.ErrNotFound).Once()
				h.hasher.EXPECT().Hash(in.Password).Return("hashed:"+in.Password, nil).Once()
				h.seasons.EXPECT().GetCurrent(mock.Anything).
					Return(&season.Season{ID: seasonID, Status: season.StatusFinished}, nil).Once()
				h.players.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*player.Player")).
					Return(nil).Once()
				h.expectTxPassthrough()
			},
			input: func(_ *registerHarness) dto.RegisterPlayerInput { return baseInput },
			want:  want{rating: 0},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := newRegisterHarness(t, baseTime)
			in := tt.input(h)
			if tt.setup != nil {
				tt.setup(t, h, in)
			}

			uc := h.build(t)
			got, err := uc.Execute(context.Background(), in)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.wantErr),
					"expected %v, got %v", tt.wantErr, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.NotNil(t, got.Player)
			require.Equal(t, tt.want.rating, got.Player.Rating)

			if tt.want.inClub {
				require.NotNil(t, got.Player.ClubID)
			}
			if tt.want.divisionCode == "" {
				require.Nil(t, got.Division)
			} else {
				require.NotNil(t, got.Division)
			}
		})
	}
}

// неиспользуемые сущности — используются в других тестах пакета.
var _ = domplayer.StatusActive
