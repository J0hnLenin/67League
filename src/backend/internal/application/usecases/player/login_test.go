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

const (
	loginTestPlayerID = shared.PlayerID(2001)
	loginTestPassword = "secret123"
	loginTestHash     = "hashed:secret123"
	loginTestEmail    = shared.Email("alice@example.com")
)

func TestLoginPlayer(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// makePlayer создаёт игрока с уже назначенным ID и хешем пароля.
	// В проде ID назначает репозиторий при Create, здесь — вручную.
	makePlayer := func(t *testing.T, status domplayer.Status) *domplayer.Player {
		t.Helper()
		p, err := domplayer.New(domplayer.NewPlayerParams{
			FullName:     "Alice",
			Email:        loginTestEmail,
			PasswordHash: loginTestHash,
			Nickname:     "alice",
			City:         "Moscow",
		}, now)
		require.NoError(t, err)
		p.ID = loginTestPlayerID
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
			name: "успех: активный игрок",
			setup: func(t *testing.T, p *mocks.PlayerRepository, h *mocks.PasswordHasher, tok *mocks.TokenIssuer) {
				pl := makePlayer(t, domplayer.StatusActive)
				p.EXPECT().GetByEmail(mock.Anything, loginTestEmail).Return(pl, nil).Once()
				h.EXPECT().Verify(loginTestPassword, loginTestHash).Return(true, nil).Once()
				tok.EXPECT().IssuePlayer(loginTestPlayerID).Return("tok", nil).Once()
			},
			input:     dto.LoginPlayerInput{Email: loginTestEmail, Password: loginTestPassword},
			wantToken: true,
		},
		{
			name: "неактивный игрок тоже может войти",
			setup: func(t *testing.T, p *mocks.PlayerRepository, h *mocks.PasswordHasher, tok *mocks.TokenIssuer) {
				pl := makePlayer(t, domplayer.StatusInactive)
				p.EXPECT().GetByEmail(mock.Anything, loginTestEmail).Return(pl, nil).Once()
				h.EXPECT().Verify(loginTestPassword, loginTestHash).Return(true, nil).Once()
				tok.EXPECT().IssuePlayer(loginTestPlayerID).Return("tok", nil).Once()
			},
			input:     dto.LoginPlayerInput{Email: loginTestEmail, Password: loginTestPassword},
			wantToken: true,
		},
		{
			name: "неверный пароль",
			setup: func(t *testing.T, p *mocks.PlayerRepository, h *mocks.PasswordHasher, tok *mocks.TokenIssuer) {
				pl := makePlayer(t, domplayer.StatusActive)
				p.EXPECT().GetByEmail(mock.Anything, loginTestEmail).Return(pl, nil).Once()
				h.EXPECT().Verify("wrong", loginTestHash).Return(false, nil).Once()
			},
			input:   dto.LoginPlayerInput{Email: loginTestEmail, Password: "wrong"},
			wantErr: shared.ErrUnauthorized,
		},
		{
			name: "неизвестный email",
			setup: func(t *testing.T, p *mocks.PlayerRepository, h *mocks.PasswordHasher, tok *mocks.TokenIssuer) {
				p.EXPECT().GetByEmail(mock.Anything, shared.Email("unknown@example.com")).
					Return(nil, shared.ErrNotFound).Once()
			},
			input:   dto.LoginPlayerInput{Email: "unknown@example.com", Password: loginTestPassword},
			wantErr: shared.ErrUnauthorized,
		},
		{
			name: "забаненный игрок",
			setup: func(t *testing.T, p *mocks.PlayerRepository, h *mocks.PasswordHasher, tok *mocks.TokenIssuer) {
				pl := makePlayer(t, domplayer.StatusBanned)
				p.EXPECT().GetByEmail(mock.Anything, loginTestEmail).Return(pl, nil).Once()
				h.EXPECT().Verify(loginTestPassword, loginTestHash).Return(true, nil).Once()
				// IssuePlayer не должен вызываться — мок это проверит сам.
			},
			input:   dto.LoginPlayerInput{Email: loginTestEmail, Password: loginTestPassword},
			wantErr: shared.ErrForbidden,
		},
		{
			name: "ошибка PasswordHasher при проверке",
			setup: func(t *testing.T, p *mocks.PlayerRepository, h *mocks.PasswordHasher, tok *mocks.TokenIssuer) {
				pl := makePlayer(t, domplayer.StatusActive)
				p.EXPECT().GetByEmail(mock.Anything, loginTestEmail).Return(pl, nil).Once()
				h.EXPECT().Verify(loginTestPassword, loginTestHash).
					Return(false, errors.New("bcrypt: invalid hash")).Once()
			},
			input:   dto.LoginPlayerInput{Email: loginTestEmail, Password: loginTestPassword},
			wantErr: nil, // проверяем отдельно ниже
		},
		{
			name: "ошибка TokenIssuer",
			setup: func(t *testing.T, p *mocks.PlayerRepository, h *mocks.PasswordHasher, tok *mocks.TokenIssuer) {
				pl := makePlayer(t, domplayer.StatusActive)
				p.EXPECT().GetByEmail(mock.Anything, loginTestEmail).Return(pl, nil).Once()
				h.EXPECT().Verify(loginTestPassword, loginTestHash).Return(true, nil).Once()
				tok.EXPECT().IssuePlayer(loginTestPlayerID).
					Return("", errors.New("token: signer unavailable")).Once()
			},
			input:   dto.LoginPlayerInput{Email: loginTestEmail, Password: loginTestPassword},
			wantErr: nil, // проверяем отдельно ниже
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

			// Кейсы с внутренними ошибками: тип ошибки не важен,
			// важно, что use case не проглотил её и не вернул игрока.
			switch tt.name {
			case "ошибка PasswordHasher при проверке",
				"ошибка TokenIssuer":
				require.Error(t, err)
				require.Nil(t, got)
				return
			}

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
			require.Equal(t, loginTestPlayerID, got.Player.ID)
			if tt.wantToken {
				require.NotEmpty(t, got.Token)
			}
		})
	}
}
