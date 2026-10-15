package match

import (
	"context"
	"errors"
	"fmt"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	dommatch "github.com/J0hnLenin/67League/internal/domain/match"
	"github.com/J0hnLenin/67League/internal/domain/notification"
)

// AcceptMatch принимает вызов: proposed → accepted.
type AcceptMatch struct {
	matches ports.MatchRepository
	events  ports.MatchEventRepository
	notifs  ports.NotificationRepository
	clock   ports.Clock
	tx      ports.TransactionManager
}

type AcceptMatchDeps struct {
	Matches ports.MatchRepository
	Events  ports.MatchEventRepository
	Notifs  ports.NotificationRepository
	Clock   ports.Clock
	Tx      ports.TransactionManager
}

func NewAcceptMatch(d AcceptMatchDeps) (*AcceptMatch, error) {
	if d.Matches == nil || d.Events == nil || d.Notifs == nil ||
		d.Clock == nil || d.Tx == nil {
		return nil, errors.New("accept match: missing dependencies")
	}
	return &AcceptMatch{
		matches: d.Matches,
		events:  d.Events,
		notifs:  d.Notifs,
		clock:   d.Clock,
		tx:      d.Tx,
	}, nil
}

func (uc *AcceptMatch) Execute(ctx context.Context, in dto.AcceptMatchInput) (*dto.AcceptMatchOutput, error) {
	var out *dto.AcceptMatchOutput
	now := uc.clock.Now()

	err := uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		m, err := uc.matches.GetByID(ctx, in.MatchID)
		if err != nil {
			return fmt.Errorf("accept match: get: %w", err)
		}
		if err := m.Accept(in.ActorID, now); err != nil {
			return fmt.Errorf("accept match: %w", err)
		}
		if err := uc.matches.Update(ctx, m); err != nil {
			return fmt.Errorf("accept match: save: %w", err)
		}
		if err := uc.events.Create(ctx, dommatch.NewEvent(
			m.ID, dommatch.EventAccepted, in.ActorID, nil, now,
		)); err != nil {
			return fmt.Errorf("accept match: save event: %w", err)
		}
		// Уведомляем создателя — он получит «вызов принят».
		n := notification.New(m.CreatedBy, notification.TypeChallengeAccepted, map[string]any{
			"match_id": int64(m.ID),
		}, notification.ChannelEmail, now)
		if err := uc.notifs.Create(ctx, n); err != nil {
			return fmt.Errorf("accept match: notify: %w", err)
		}
		out = &dto.AcceptMatchOutput{Match: m}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
