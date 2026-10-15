package club

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/domain/badge"
	domclub "github.com/J0hnLenin/67League/internal/domain/club"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// BadgeCodeFounding — код бейджа «Основание» в справочнике.
const BadgeCodeFounding = "club_founding"

type CreateClub struct {
	clubs     ports.ClubRepository
	players   ports.PlayerRepository
	badges    ports.BadgeRepository
	clubBadge ports.ClubBadgeRepository
	clock     ports.Clock
	tx        ports.TransactionManager
}

type CreateClubDeps struct {
	Clubs     ports.ClubRepository
	Players   ports.PlayerRepository
	Badges    ports.BadgeRepository
	ClubBadge ports.ClubBadgeRepository
	Clock     ports.Clock
	Tx        ports.TransactionManager
}

func NewCreateClub(d CreateClubDeps) (*CreateClub, error) {
	if d.Clubs == nil || d.Players == nil || d.Badges == nil ||
		d.ClubBadge == nil || d.Clock == nil || d.Tx == nil {
		return nil, errors.New("create club: missing dependencies")
	}
	return &CreateClub{
		clubs: d.Clubs, players: d.Players, badges: d.Badges,
		clubBadge: d.ClubBadge, clock: d.Clock, tx: d.Tx,
	}, nil
}

func (uc *CreateClub) Execute(ctx context.Context, in dto.CreateClubInput) (*dto.CreateClubOutput, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("create club: name: %w", shared.ErrValidation)
	}

	exists, err := uc.clubs.ExistsByName(ctx, in.Name)
	if err != nil {
		return nil, fmt.Errorf("create club: check name: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("create club: name %q: %w", in.Name, shared.ErrAlreadyExists)
	}

	creator, err := uc.players.GetByID(ctx, in.CreatorID)
	if err != nil {
		return nil, fmt.Errorf("create club: get creator: %w", err)
	}
	if !creator.Status.CanCreateClub() {
		return nil, fmt.Errorf("create club: creator status: %w", shared.ErrForbidden)
	}

	now := uc.clock.Now()
	c, err := domclub.New(in.Name, in.City, in.University, in.CreatorID, now)
	if err != nil {
		return nil, fmt.Errorf("create club: build entity: %w", err)
	}

	err = uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := uc.clubs.Create(ctx, c); err != nil {
			return fmt.Errorf("create club: %w", err)
		}
		// Создатель автоматически становится участником клуба.
		if err := creator.JoinClub(c.ID, now); err != nil {
			return fmt.Errorf("create club: join creator: %w", err)
		}
		if err := uc.players.Update(ctx, creator); err != nil {
			return fmt.Errorf("create club: update creator: %w", err)
		}
		// Автовыдача бейджа «Основание». Не критично, если справочника нет.
		if b, err := uc.badges.GetByCode(ctx, BadgeCodeFounding); err == nil {
			cb := newClubBadge(c.ID, b.ID, now)
			if err := uc.clubBadge.Create(ctx, cb); err != nil {
				return fmt.Errorf("create club: award founding badge: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.CreateClubOutput{Club: c}, nil
}

func newClubBadge(clubID shared.ClubID, badgeID shared.BadgeID, now time.Time) *badge.ClubBadge {
	return badge.NewClubBadge(clubID, badgeID, nil, now)
}
