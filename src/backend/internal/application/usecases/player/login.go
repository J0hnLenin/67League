package player

import (
	"context"
	"errors"
	"fmt"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type LoginPlayer struct {
	players ports.PlayerRepository
	hasher  ports.PasswordHasher
	tokens  ports.TokenIssuer
}

type LoginPlayerDeps struct {
	Players ports.PlayerRepository
	Hasher  ports.PasswordHasher
	Tokens  ports.TokenIssuer
}

func NewLoginPlayer(d LoginPlayerDeps) (*LoginPlayer, error) {
	if d.Players == nil || d.Hasher == nil || d.Tokens == nil {
		return nil, errors.New("login player: missing dependencies")
	}
	return &LoginPlayer{players: d.Players, hasher: d.Hasher, tokens: d.Tokens}, nil
}

// Execute возвращает ErrUnauthorized при неверной паре email/пароль
// и ErrForbidden для забаненных игроков.
func (uc *LoginPlayer) Execute(ctx context.Context, in dto.LoginPlayerInput) (*dto.LoginPlayerOutput, error) {
	p, err := uc.players.GetByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, fmt.Errorf("login: %w", shared.ErrUnauthorized)
		}
		return nil, fmt.Errorf("login: get player: %w", err)
	}

	ok, err := uc.hasher.Verify(in.Password, p.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("login: verify password: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("login: %w", shared.ErrUnauthorized)
	}

	if p.Status == player.StatusBanned {
		return nil, fmt.Errorf("login: %w", shared.ErrForbidden)
	}

	token, err := uc.tokens.IssuePlayer(p.ID)
	if err != nil {
		return nil, fmt.Errorf("login: issue token: %w", err)
	}

	return &dto.LoginPlayerOutput{Player: p, Token: token}, nil
}
