package match

import (
	"context"
	"errors"
	"fmt"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	dommatch "github.com/J0hnLenin/67League/internal/domain/match"
	"github.com/J0hnLenin/67League/internal/domain/notification"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// SubmitResult вносит результат: playing → pending_confirmation.
// Результат вносит победитель.
type SubmitResult struct {
	matches ports.MatchRepository
	events  ports.MatchEventRepository
	notifs  ports.NotificationRepository
	clock   ports.Clock
	tx      ports.TransactionManager
}

type SubmitResultDeps struct {
	Matches ports.MatchRepository
	Events  ports.MatchEventRepository
	Notifs  ports.NotificationRepository
	Clock   ports.Clock
	Tx      ports.TransactionManager
}

func NewSubmitResult(d SubmitResultDeps) (*SubmitResult, error) {
	if d.Matches == nil || d.Events == nil || d.Notifs == nil ||
		d.Clock == nil || d.Tx == nil {
		return nil, errors.New("submit result: missing dependencies")
	}
	return &SubmitResult{
		matches: d.Matches,
		events:  d.Events,
		notifs:  d.Notifs,
		clock:   d.Clock,
		tx:      d.Tx,
	}, nil
}

func (uc *SubmitResult) Execute(ctx context.Context, in dto.SubmitResultInput) (*dto.SubmitResultOutput, error) {
	var out *dto.SubmitResultOutput
	now := uc.clock.Now()

	err := uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		m, err := uc.matches.GetByID(ctx, in.MatchID)
		if err != nil {
			return fmt.Errorf("submit result: get: %w", err)
		}
		if err := m.SubmitResult(in.ActorID, in.WinnerID, in.ScoreDiff, in.SGFURL, in.Link, now); err != nil {
			return fmt.Errorf("submit result: %w", err)
		}
		if err := uc.matches.Update(ctx, m); err != nil {
			return fmt.Errorf("submit result: save: %w", err)
		}
		if err := uc.events.Create(ctx, dommatch.NewEvent(
			m.ID, dommatch.EventResultSubmitted, in.ActorID, nil, now,
		)); err != nil {
			return fmt.Errorf("submit result: save event: %w", err)
		}
		// Уведомляем соперника — он должен подтвердить.
		opp, ok := m.OpponentOf(in.ActorID)
		if !ok {
			return fmt.Errorf("submit result: resolve opponent: %w", shared.ErrInvalidState)
		}
		n := notification.New(opp, notification.TypeResultSubmitted, map[string]any{
			"match_id": int64(m.ID),
		}, notification.ChannelEmail, now)
		if err := uc.notifs.Create(ctx, n); err != nil {
			return fmt.Errorf("submit result: notify: %w", err)
		}
		out = &dto.SubmitResultOutput{Match: m}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
