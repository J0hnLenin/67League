package match

import (
	"context"
	"errors"
	"fmt"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	dommatch "github.com/J0hnLenin/67League/internal/domain/match"
)

// StartMatch переводит партию в playing. Действие доступно любому участнику.
type StartMatch struct {
	matches ports.MatchRepository
	events  ports.MatchEventRepository
	clock   ports.Clock
	tx      ports.TransactionManager
}

type StartMatchDeps struct {
	Matches ports.MatchRepository
	Events  ports.MatchEventRepository
	Clock   ports.Clock
	Tx      ports.TransactionManager
}

func NewStartMatch(d StartMatchDeps) (*StartMatch, error) {
	if d.Matches == nil || d.Events == nil || d.Clock == nil || d.Tx == nil {
		return nil, errors.New("start match: missing dependencies")
	}
	return &StartMatch{
		matches: d.Matches,
		events:  d.Events,
		clock:   d.Clock,
		tx:      d.Tx,
	}, nil
}

func (uc *StartMatch) Execute(ctx context.Context, in dto.StartMatchInput) (*dto.StartMatchOutput, error) {
	var out *dto.StartMatchOutput
	now := uc.clock.Now()

	err := uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		m, err := uc.matches.GetByID(ctx, in.MatchID)
		if err != nil {
			return fmt.Errorf("start match: get: %w", err)
		}
		if err := m.Start(in.ActorID, now); err != nil {
			return fmt.Errorf("start match: %w", err)
		}
		if err := uc.matches.Update(ctx, m); err != nil {
			return fmt.Errorf("start match: save: %w", err)
		}
		if err := uc.events.Create(ctx, dommatch.NewEvent(
			m.ID, dommatch.EventStarted, in.ActorID, nil, now,
		)); err != nil {
			return fmt.Errorf("start match: save event: %w", err)
		}
		out = &dto.StartMatchOutput{Match: m}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
