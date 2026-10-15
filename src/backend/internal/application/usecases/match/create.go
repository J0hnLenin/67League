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

// CreateMatch создаёт вызов на партию.
//
// Инварианты:
//   - создатель — участник (black или white);
//   - сезон в статусе CanCreateMatch (active);
//   - оба игрока в одном дивизионе этого сезона;
//   - лимит n партий с одним соперником не превышен.
type CreateMatch struct {
	matches ports.MatchRepository
	events  ports.MatchEventRepository
	members ports.DivisionMemberRepository
	seasons ports.SeasonRepository
	notifs  ports.NotificationRepository
	clock   ports.Clock
	tx      ports.TransactionManager
}

type CreateMatchDeps struct {
	Matches ports.MatchRepository
	Events  ports.MatchEventRepository
	Members ports.DivisionMemberRepository
	Seasons ports.SeasonRepository
	Notifs  ports.NotificationRepository
	Clock   ports.Clock
	Tx      ports.TransactionManager
}

func NewCreateMatch(d CreateMatchDeps) (*CreateMatch, error) {
	if d.Matches == nil || d.Events == nil || d.Members == nil ||
		d.Seasons == nil || d.Notifs == nil || d.Clock == nil || d.Tx == nil {
		return nil, errors.New("create match: missing dependencies")
	}
	return &CreateMatch{
		matches: d.Matches,
		events:  d.Events,
		members: d.Members,
		seasons: d.Seasons,
		notifs:  d.Notifs,
		clock:   d.Clock,
		tx:      d.Tx,
	}, nil
}

func (uc *CreateMatch) Execute(ctx context.Context, in dto.CreateMatchInput) (*dto.CreateMatchOutput, error) {
	if in.BlackPlayerID == in.WhitePlayerID {
		return nil, fmt.Errorf("create match: same players: %w", shared.ErrValidation)
	}
	if in.CreatedBy != in.BlackPlayerID && in.CreatedBy != in.WhitePlayerID {
		return nil, fmt.Errorf("create match: creator is not a participant: %w", shared.ErrForbidden)
	}

	s, err := uc.seasons.GetByID(ctx, in.SeasonID)
	if err != nil {
		return nil, fmt.Errorf("create match: get season: %w", err)
	}
	if !s.Status.CanCreateMatch() {
		return nil, fmt.Errorf("create match: season status %q: %w", s.Status, shared.ErrInvalidState)
	}

	blackMember, err := uc.members.GetByPlayerAndSeason(ctx, in.BlackPlayerID, in.SeasonID)
	if err != nil {
		return nil, fmt.Errorf("create match: black not in season: %w", err)
	}
	whiteMember, err := uc.members.GetByPlayerAndSeason(ctx, in.WhitePlayerID, in.SeasonID)
	if err != nil {
		return nil, fmt.Errorf("create match: white not in season: %w", err)
	}
	if blackMember.DivisionID != in.DivisionID || whiteMember.DivisionID != in.DivisionID {
		return nil, fmt.Errorf("create match: players must be in same division: %w", shared.ErrValidation)
	}

	count, err := uc.matches.CountBetweenPlayersInSeason(ctx, in.SeasonID, in.BlackPlayerID, in.WhitePlayerID)
	if err != nil {
		return nil, fmt.Errorf("create match: count matches: %w", err)
	}
	if count >= s.Settings.N {
		return nil, fmt.Errorf("create match: limit %d with opponent reached: %w", s.Settings.N, shared.ErrLimitExceeded)
	}

	now := uc.clock.Now()
	m, err := dommatch.New(dommatch.NewMatchParams{
		SeasonID:      in.SeasonID,
		DivisionID:    in.DivisionID,
		BlackPlayerID: in.BlackPlayerID,
		WhitePlayerID: in.WhitePlayerID,
		CreatedBy:     in.CreatedBy,
		Server:        in.Server,
		TimeControl:   in.TimeControl,
		Handicap:      in.Handicap,
		Komi:          in.Komi,
	}, now)
	if err != nil {
		return nil, fmt.Errorf("create match: build entity: %w", err)
	}

	opponentID, _ := m.OpponentOf(in.CreatedBy)

	err = uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := uc.matches.Create(ctx, m); err != nil {
			return fmt.Errorf("create match: save: %w", err)
		}
		if err := uc.events.Create(ctx, dommatch.NewEvent(
			m.ID, dommatch.EventCreated, in.CreatedBy, nil, now,
		)); err != nil {
			return fmt.Errorf("create match: save event: %w", err)
		}
		n := notification.New(opponentID, notification.TypeNewChallenge, map[string]any{
			"match_id": int64(m.ID),
		}, notification.ChannelEmail, now)
		if err := uc.notifs.Create(ctx, n); err != nil {
			return fmt.Errorf("create match: notify: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.CreateMatchOutput{Match: m}, nil
}
