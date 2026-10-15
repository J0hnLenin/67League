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

// CancelMatch отменяет партию. Игрок — только если прошло меньше
// match_cancel_days с момента завершения. Админ — в любой момент.
type CancelMatch struct {
	matches ports.MatchRepository
	events  ports.MatchEventRepository
	seasons ports.SeasonRepository
	notifs  ports.NotificationRepository
	clock   ports.Clock
	tx      ports.TransactionManager
}

type CancelMatchDeps struct {
	Matches ports.MatchRepository
	Events  ports.MatchEventRepository
	Seasons ports.SeasonRepository
	Notifs  ports.NotificationRepository
	Clock   ports.Clock
	Tx      ports.TransactionManager
}

func NewCancelMatch(d CancelMatchDeps) (*CancelMatch, error) {
	if d.Matches == nil || d.Events == nil || d.Seasons == nil ||
		d.Notifs == nil || d.Clock == nil || d.Tx == nil {
		return nil, errors.New("cancel match: missing dependencies")
	}
	return &CancelMatch{
		matches: d.Matches,
		events:  d.Events,
		seasons: d.Seasons,
		notifs:  d.Notifs,
		clock:   d.Clock,
		tx:      d.Tx,
	}, nil
}

func (uc *CancelMatch) Execute(ctx context.Context, in dto.CancelMatchInput) (*dto.CancelMatchOutput, error) {
	var out *dto.CancelMatchOutput
	now := uc.clock.Now()

	err := uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		m, err := uc.matches.GetByID(ctx, in.MatchID)
		if err != nil {
			return fmt.Errorf("cancel match: get: %w", err)
		}

		s, err := uc.seasons.GetByID(ctx, m.SeasonID)
		if err != nil {
			return fmt.Errorf("cancel match: get season: %w", err)
		}

		if err := m.Cancel(in.ActorID, in.IsAdmin, s.Settings.MatchCancelDays, now); err != nil {
			return fmt.Errorf("cancel match: %w", err)
		}
		if in.Reason != "" {
			m.CancelReason = in.Reason
		}

		if err := uc.matches.Update(ctx, m); err != nil {
			return fmt.Errorf("cancel match: save: %w", err)
		}
		if err := uc.events.Create(ctx, dommatch.NewEvent(
			m.ID, dommatch.EventCancelled, in.ActorID,
			map[string]any{"reason": in.Reason}, now,
		)); err != nil {
			return fmt.Errorf("cancel match: save event: %w", err)
		}

		// Уведомляем соперника — если отменял игрок.
		if !in.IsAdmin {
			if opp, ok := m.OpponentOf(in.ActorID); ok {
				n := notification.New(opp, notification.TypeResultConfirmed, map[string]any{
					"match_id": int64(m.ID),
					"status":   "cancelled",
					"reason":   in.Reason,
				}, notification.ChannelEmail, now)
				if err := uc.notifs.Create(ctx, n); err != nil {
					return fmt.Errorf("cancel match: notify: %w", err)
				}
			}
		}

		out = &dto.CancelMatchOutput{Match: m}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
