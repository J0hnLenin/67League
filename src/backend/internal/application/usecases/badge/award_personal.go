package badge

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	dombadge "github.com/J0hnLenin/67League/internal/domain/badge"
	"github.com/J0hnLenin/67League/internal/domain/match"
	"github.com/J0hnLenin/67League/internal/domain/player"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// AwardPersonalBadges выдаёт личные бейджи игроку по итогам сезона.
//
// Идемпотентен: повторный вызов не создаёт дубликатов.
// Отсутствие бейджа в справочнике — не ошибка, а конфигурация.
//
// Вызывается после подтверждения партии или по итогам сезона.
type AwardPersonalBadges struct {
	badges       ports.BadgeRepository
	playerBadges ports.PlayerBadgeRepository
	members      ports.DivisionMemberRepository
	matches      ports.MatchRepository
	players      ports.PlayerRepository
	clubs        ports.ClubRepository
	clock        ports.Clock
	tx           ports.TransactionManager
}

type AwardPersonalBadgesDeps struct {
	Badges       ports.BadgeRepository
	PlayerBadges ports.PlayerBadgeRepository
	Members      ports.DivisionMemberRepository
	Matches      ports.MatchRepository
	Players      ports.PlayerRepository
	Clubs        ports.ClubRepository
	Clock        ports.Clock
	Tx           ports.TransactionManager
}

func NewAwardPersonalBadges(d AwardPersonalBadgesDeps) (*AwardPersonalBadges, error) {
	if d.Badges == nil || d.PlayerBadges == nil || d.Members == nil ||
		d.Matches == nil || d.Players == nil || d.Clubs == nil ||
		d.Clock == nil || d.Tx == nil {
		return nil, errors.New("award personal badges: missing dependencies")
	}
	return &AwardPersonalBadges{
		badges: d.Badges, playerBadges: d.PlayerBadges, members: d.Members,
		matches: d.Matches, players: d.Players, clubs: d.Clubs,
		clock: d.Clock, tx: d.Tx,
	}, nil
}

func (uc *AwardPersonalBadges) Execute(
	ctx context.Context,
	in dto.AwardPersonalBadgesInput,
) (*dto.AwardPersonalBadgesOutput, error) {
	member, err := uc.members.GetByPlayerAndSeason(ctx, in.PlayerID, in.SeasonID)
	if err != nil {
		return nil, fmt.Errorf("award personal badges: get member: %w", err)
	}

	p, err := uc.players.GetByID(ctx, in.PlayerID)
	if err != nil {
		return nil, fmt.Errorf("award personal badges: get player: %w", err)
	}

	all, err := uc.matches.ListByPlayer(ctx, in.PlayerID, ports.MatchFilter{})
	if err != nil {
		return nil, fmt.Errorf("award personal badges: list matches: %w", err)
	}

	// Отбираем подтверждённые партии этого сезона.
	seasonMatches := make([]*match.Match, 0, len(all))
	for _, m := range all {
		if m.SeasonID == in.SeasonID && m.Status == match.StatusConfirmed {
			seasonMatches = append(seasonMatches, m)
		}
	}

	// Кому выдавать.
	// Особые случаи: счастливчик, товарищеский матч, глобалист.
	// Их надо вычислить по партиям.
	special, err := uc.evaluateMatchDependentBadges(ctx, p, seasonMatches)
	if err != nil {
		return nil, err
	}

	// Проверяем условия по статистике.
	qualified := []string{}
	if member.MatchesPlayed >= 1 {
		qualified = append(qualified, CodeFirstStone)
	}
	if member.Wins >= 1 {
		qualified = append(qualified, CodeFirstBlood)
	}
	if member.Wins >= 3 {
		qualified = append(qualified, CodeLizard)
	}
	if member.MatchesPlayed >= 10 {
		qualified = append(qualified, CodeMarathoner)
	}
	if member.Wins >= 10 {
		qualified = append(qualified, CodeStoneEater)
	}
	if member.StreakMax >= 3 {
		qualified = append(qualified, CodeFire)
	}
	if member.StreakMax >= 5 {
		qualified = append(qualified, CodeUnstoppable)
	}
	if member.StreakMax >= 6 {
		qualified = append(qualified, CodeSixSeven)
	}
	if member.StreaksBroken >= 1 {
		qualified = append(qualified, CodeHeadhunter)
	}
	qualified = append(qualified, special...)

	now := uc.clock.Now()
	out := &dto.AwardPersonalBadgesOutput{}

	err = uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		for _, code := range qualified {
			awarded, err := uc.award(ctx, in.PlayerID, code, now)
			if err != nil {
				return err
			}
			if awarded != nil {
				out.Awarded = append(out.Awarded, awarded)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// award выдаёт бейдж игроку, если ещё не выдан. Возвращает nil, если уже есть
// или бейджа нет в справочнике.
func (uc *AwardPersonalBadges) award(
	ctx context.Context,
	pid shared.PlayerID,
	code string,
	now time.Time,
) (*dombadge.PlayerBadge, error) {
	b, err := uc.badges.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, nil // нет в справочнике — не критично
		}
		return nil, fmt.Errorf("get badge %q: %w", code, err)
	}
	has, err := uc.playerBadges.Has(ctx, pid, b.ID)
	if err != nil {
		return nil, fmt.Errorf("check badge %q: %w", code, err)
	}
	if has {
		return nil, nil
	}
	pb := dombadge.NewPlayerBadge(pid, b.ID, nil, now) // личный — без сезонной привязки
	if err := uc.playerBadges.Create(ctx, pb); err != nil {
		return nil, fmt.Errorf("create badge %q: %w", code, err)
	}
	return pb, nil
}

// evaluateMatchDependentBadges вычисляет бейджи, зависящие от состава партий.
func (uc *AwardPersonalBadges) evaluateMatchDependentBadges(
	ctx context.Context,
	p *player.Player,
	matches []*match.Match,
) ([]string, error) {
	var out []string

	for _, m := range matches {
		if m.WinnerID == nil {
			continue
		}
		// Счастливчик: победил с разницей 0.5.
		if *m.WinnerID == p.ID && m.ScoreDiff != nil && *m.ScoreDiff == 0.5 {
			out = append(out, CodeLucky)
		}

		// Определяем соперника.
		oppID, ok := m.OpponentOf(p.ID)
		if !ok {
			continue
		}
		opp, err := uc.players.GetByID(ctx, oppID)
		if err != nil {
			return nil, fmt.Errorf("evaluate badges: get opponent: %w", err)
		}

		// Товарищеский матч: оба в клубах, но в разных.
		if p.ClubID != nil && opp.ClubID != nil && *p.ClubID != *opp.ClubID {
			out = append(out, CodeFriendlyMatch)
		}

		// Глобалист: оба клуба существуют и у них разные города.
		if p.ClubID != nil && opp.ClubID != nil {
			pc, err := uc.clubs.GetByID(ctx, *p.ClubID)
			if err != nil {
				return nil, fmt.Errorf("evaluate badges: get player club: %w", err)
			}
			oc, err := uc.clubs.GetByID(ctx, *opp.ClubID)
			if err != nil {
				return nil, fmt.Errorf("evaluate badges: get opponent club: %w", err)
			}
			if pc.City != "" && oc.City != "" && pc.City != oc.City {
				out = append(out, CodeGlobalist)
			}
		}
	}

	// Дедуплицируем, сохраняя порядок.
	seen := map[string]bool{}
	uniq := out[:0]
	for _, c := range out {
		if !seen[c] {
			seen[c] = true
			uniq = append(uniq, c)
		}
	}
	return uniq, nil
}
