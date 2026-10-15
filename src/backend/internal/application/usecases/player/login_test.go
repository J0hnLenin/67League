package player_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/usecases/player"
	domplayer "github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
	"github.com/J0hnLenin/67League/internal/testsupport/mocks"
)

func TestLoginPlayer(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	makePlayer := func(t *testing.T, status domplayer.Status) *domplayer.Player {
		t.Helper()
		p, err := domplayer.New(domplayer.NewPlayerParams{
			FullName:     "Alice",
			Email:        "alice@example.com",
			PasswordHash: "hashed:secret123",
			Nickname:     "alice",
			City:         "Moscow",
		}, now)
		require.NoError(t, err)
		p.Status = status
		return p
	}

	tests := []struct {
		name      string
		setup     func(t *testing.T, p *mocks.PlayerRepository, h *mocks.PasswordHasher, tok *mocks.TokenIssuer)
		input     dto.LoginPlayerInput
		wantErr   error
		wantToken bool
	}{
		{
			name: "успех",
			setup: func(t *testing.T, p *mocks.PlayerRepository, h *mocks.PasswordHasher, tok *mocks.TokenIssuer) {
				pl := makePlayer(t, domplayer.StatusActive)
				p.EXPECT().GetByEmail(mock.Anything, pl.Email).Return(pl, nil).Once()
				h.EXPECT().Verify("secret123", "hashed:secret123").Return(true, nil).Once()
				tok.EXPECT().IssuePlayer(pl.ID).Return("tok", nil).Once()
			},
			input:     dto.LoginPlayerInput{Email: "alice@example.com", Password: "secret123"},
			wantToken: true,
		},
		{
			name: "неверный пароль",
			setup: func(t *testing.T, p *mocks.PlayerRepository, h *mocks.PasswordHasher, tok *mocks.TokenIssuer) {
				pl := makePlayer(t, domplayer.StatusActive)
				p.EXPECT().GetByEmail(mock.Anything, pl.Email).Return(pl, nil).Once()
				h.EXPECT().Verify("wrong", "hashed:secret123").Return(false, nil).Once()
			},
			input:   dto.LoginPlayerInput{Email: "alice@example.com", Password: "wrong"},
			wantErr: shared.ErrUnauthorized,
		},
		{
			name: "неизвестный email",
			setup: func(t *testing.T, p *mocks.PlayerRepository, h *mocks.PasswordHasher, tok *mocks.TokenIssuer) {
				p.EXPECT().GetByEmail(mock.Anything, shared.Email("unknown@example.com")).
					Return(nil, shared.ErrNotFound).Once()
			},
			input:   dto.LoginPlayerInput{Email: "unknown@example.com", Password: "secret123"},
			wantErr: shared.ErrUnauthorized,
		},
		{
			name: "забаненный",
			setup: func(t *testing.T, p *mocks.PlayerRepository, h *mocks.PasswordHasher, tok *mocks.TokenIssuer) {
				pl := makePlayer(t, domplayer.StatusBanned)
				p.EXPECT().GetByEmail(mock.Anything, pl.Email).Return(pl, nil).Once()
				h.EXPECT().Verify("secret123", "hashed:secret123").Return(true, nil).Once()
			},
			input:   dto.LoginPlayerInput{Email: "alice@example.com", Password: "secret123"},
			wantErr: shared.ErrForbidden,
		},
		{
			name: "неактивный входит",
			setup: func(t *testing.T, p *mocks.PlayerRepository, h *mocks.PasswordHasher, tok *mocks.TokenIssuer) {
				pl := makePlayer(t, domplayer.StatusInactive)
				p.EXPECT().GetByEmail(mock.Anything, pl.Email).Return(pl, nil).Once()
				h.EXPECT().Verify("secret123", "hashed:secret123").Return(true, nil).Once()
				tok.EXPECT().IssuePlayer(pl.ID).Return("tok", nil).Once()
			},
			input:     dto.LoginPlayerInput{Email: "alice@example.com", Password: "secret123"},
			wantToken: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := mocks.NewPlayerRepository(t)
			hasher := mocks.NewPasswordHasher(t)
			tokens := mocks.NewTokenIssuer(t)

			tt.setup(t, repo, hasher, tokens)

			uc, err := player.NewLoginPlayer(player.LoginPlayerDeps{
				Players: repo,
				Hasher:  hasher,
				Tokens:  tokens,
			})
			require.NoError(t, err)

			got, err := uc.Execute(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.wantErr),
					"expected %v, got %v", tt.wantErr, err)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.NotNil(t, got.Player)
			if tt.wantToken {
				require.NotEmpty(t, got.Token)
			}
		})
	}
}
