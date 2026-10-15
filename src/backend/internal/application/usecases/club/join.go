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

const (
	// BadgeCodeFullRoster — код бейджа «👥 Полный состав».
	BadgeCodeFullRoster = "club_full_roster"
	// FullRosterThreshold — сколько участников нужно для бейджа.
	FullRosterThreshold = 10
)

// JoinClub переводит игрока в клуб.
//
// Побочные эффекты (в одной транзакции):
//   - Player.ClubID устанавливается в целевой клуб;
//   - при достижении порога участников клубу выдаётся бейдж «Полный состав».
type JoinClub struct {
	clubs      ports.ClubRepository
	players    ports.PlayerRepository
	badges     ports.BadgeRepository
	clubBadges ports.ClubBadgeRepository
	clock      ports.Clock
	tx         ports.TransactionManager
}

type JoinClubDeps struct {
	Clubs      ports.ClubRepository
	Players    ports.PlayerRepository
	Badges     ports.BadgeRepository
	ClubBadges ports.ClubBadgeRepository
	Clock      ports.Clock
	Tx         ports.TransactionManager
}

func NewJoinClub(d JoinClubDeps) (*JoinClub, error) {
	if d.Clubs == nil || d.Players == nil || d.Badges == nil ||
		d.ClubBadges == nil || d.Clock == nil || d.Tx == nil {
		return nil, errors.New("join club: missing dependencies")
	}
	return &JoinClub{
		clubs:      d.Clubs,
		players:    d.Players,
		badges:     d.Badges,
		clubBadges: d.ClubBadges,
		clock:      d.Clock,
		tx:         d.Tx,
	}, nil
}

func (uc *JoinClub) Execute(ctx context.Context, in dto.JoinClubInput) (*dto.JoinClubOutput, error) {
	p, err := uc.players.GetByID(ctx, in.PlayerID)
	if err != nil {
		return nil, fmt.Errorf("join club: get player: %w", err)
	}
	if !p.Status.CanPlay() {
		return nil, fmt.Errorf("join club: player status %q: %w", p.Status, shared.ErrForbidden)
	}

	c, err := uc.clubs.GetByID(ctx, in.ClubID)
	if err != nil {
		return nil, fmt.Errorf("join club: get club: %w", err)
	}
	if c.Status != domclub.StatusActive {
		return nil, fmt.Errorf("join club: club status %q: %w", c.Status, shared.ErrForbidden)
	}

	if p.ClubID != nil && *p.ClubID == in.ClubID {
		return nil, fmt.Errorf("join club: already a member: %w", shared.ErrConflict)
	}

	now := uc.clock.Now()
	out := &dto.JoinClubOutput{}

	err = uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		before, err := uc.players.CountByClub(ctx, in.ClubID)
		if err != nil {
			return fmt.Errorf("join club: count members: %w", err)
		}

		if err := p.JoinClub(in.ClubID, now); err != nil {
			return fmt.Errorf("join club: %w", err)
		}
		if err := uc.players.Update(ctx, p); err != nil {
			return fmt.Errorf("join club: save player: %w", err)
		}

		out.Player = p
		out.MemberCount = before + 1

		if out.MemberCount < FullRosterThreshold {
			return nil
		}
		awarded, err := uc.awardFullRosterBadge(ctx, in.ClubID, now)
		if err != nil {
			return err
		}
		out.FullBadgeAwarded = awarded
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// awardFullRosterBadge выдаёт клубу бейдж «Полный состав», если его ещё нет.
// Отсутствие бейджа в справочнике — не ошибка, а конфигурация.
func (uc *JoinClub) awardFullRosterBadge(ctx context.Context, clubID shared.ClubID, now time.Time) (bool, error) {
	b, err := uc.badges.GetByCode(ctx, BadgeCodeFullRoster)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("join club: get badge: %w", err)
	}

	has, err := uc.clubBadges.Has(ctx, clubID, b.ID)
	if err != nil {
		return false, fmt.Errorf("join club: check badge: %w", err)
	}
	if has {
		return false, nil
	}

	cb := badge.NewClubBadge(clubID, b.ID, nil, now)
	if err := uc.clubBadges.Create(ctx, cb); err != nil {
		return false, fmt.Errorf("join club: award badge: %w", err)
	}
	return true, nil
}
