package season

import (
	"context"
	"errors"
	"fmt"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/domain/division"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// JoinSeason добавляет игрока в сезон и распределяет его в дивизион.
//
// Побочные эффекты (в одной транзакции):
//   - создаётся DivisionMember (ровно один на игрока за сезон);
//   - повторный вызов для того же игрока/сезона → ErrAlreadyExists.
type JoinSeason struct {
	players   ports.PlayerRepository
	seasons   ports.SeasonRepository
	divisions ports.DivisionRepository
	members   ports.DivisionMemberRepository
	clock     ports.Clock
	tx        ports.TransactionManager
}

type JoinSeasonDeps struct {
	Players   ports.PlayerRepository
	Seasons   ports.SeasonRepository
	Divisions ports.DivisionRepository
	Members   ports.DivisionMemberRepository
	Clock     ports.Clock
	Tx        ports.TransactionManager
}

func NewJoinSeason(d JoinSeasonDeps) (*JoinSeason, error) {
	if d.Players == nil || d.Seasons == nil || d.Divisions == nil ||
		d.Members == nil || d.Clock == nil || d.Tx == nil {
		return nil, errors.New("join season: missing dependencies")
	}
	return &JoinSeason{
		players:   d.Players,
		seasons:   d.Seasons,
		divisions: d.Divisions,
		members:   d.Members,
		clock:     d.Clock,
		tx:        d.Tx,
	}, nil
}

func (uc *JoinSeason) Execute(ctx context.Context, in dto.JoinSeasonInput) (*dto.JoinSeasonOutput, error) {
	p, err := uc.players.GetByID(ctx, in.PlayerID)
	if err != nil {
		return nil, fmt.Errorf("join season: get player: %w", err)
	}
	if !p.Status.CanJoinSeason() {
		return nil, fmt.Errorf("join season: player status %q: %w", p.Status, shared.ErrForbidden)
	}

	s, err := uc.seasons.GetByID(ctx, in.SeasonID)
	if err != nil {
		return nil, fmt.Errorf("join season: get season: %w", err)
	}
	if !s.Status.CanJoin() {
		return nil, fmt.Errorf("join season: season status %q: %w", s.Status, shared.ErrInvalidState)
	}

	existing, err := uc.members.GetByPlayerAndSeason(ctx, in.PlayerID, in.SeasonID)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return nil, fmt.Errorf("join season: check membership: %w", err)
	}
	if err == nil && existing != nil {
		return nil, fmt.Errorf("join season: already a member: %w", shared.ErrAlreadyExists)
	}

	div, err := uc.resolveDivision(ctx, s.ID, p.Rating)
	if err != nil {
		return nil, err
	}

	now := uc.clock.Now()
	var m *division.DivisionMember

	err = uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		m = division.NewMember(div.ID, p.ID, now)
		if err := uc.members.Create(ctx, m); err != nil {
			return fmt.Errorf("join season: create member: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.JoinSeasonOutput{Member: m, Division: div, Season: s}, nil
}

// resolveDivision подбирает дивизион: по рейтингу, иначе — самый низкий.
func (uc *JoinSeason) resolveDivision(
	ctx context.Context,
	seasonID shared.SeasonID,
	rating shared.Rating,
) (*division.Division, error) {
	div, err := uc.divisions.FindForRating(ctx, seasonID, rating)
	if err == nil {
		return div, nil
	}
	if !errors.Is(err, shared.ErrNotFound) {
		return nil, fmt.Errorf("join season: find division: %w", err)
	}
	div, err = uc.divisions.GetLowest(ctx, seasonID)
	if err != nil {
		return nil, fmt.Errorf("join season: get lowest division: %w", err)
	}
	return div, nil
}
