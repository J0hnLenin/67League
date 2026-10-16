package badge

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	dombadge "github.com/J0hnLenin/67League/internal/domain/badge"
	"github.com/J0hnLenin/67League/internal/domain/division"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// AwardSeasonBadges выдаёт сезонные бейджи за места в дивизионе и клубе.
// Должен вызываться после CalculateSeasonResults (у мемберов проставлены
// FinalPlace и Status).
//
// Побочные эффекты (одна транзакция):
//   - сбрасывает is_active у прошлых сезонных бейджей игрока;
//   - выдаёт новые сезонные бейджи за 1/2/3 места в дивизионе;
//   - выдаёт новые сезонные бейджи за 1/2/3 места в клубе.
type AwardSeasonBadges struct {
	badges       ports.BadgeRepository
	playerBadges ports.PlayerBadgeRepository
	members      ports.DivisionMemberRepository
	divisions    ports.DivisionRepository
	players      ports.PlayerRepository
	clubs        ports.ClubRepository
	clock        ports.Clock
	tx           ports.TransactionManager
}

type AwardSeasonBadgesDeps struct {
	Badges       ports.BadgeRepository
	PlayerBadges ports.PlayerBadgeRepository
	Members      ports.DivisionMemberRepository
	Divisions    ports.DivisionRepository
	Players      ports.PlayerRepository
	Clubs        ports.ClubRepository
	Clock        ports.Clock
	Tx           ports.TransactionManager
}

func NewAwardSeasonBadges(d AwardSeasonBadgesDeps) (*AwardSeasonBadges, error) {
	if d.Badges == nil || d.PlayerBadges == nil || d.Members == nil ||
		d.Divisions == nil || d.Players == nil || d.Clubs == nil ||
		d.Clock == nil || d.Tx == nil {
		return nil, errors.New("award season badges: missing dependencies")
	}
	return &AwardSeasonBadges{
		badges: d.Badges, playerBadges: d.PlayerBadges, members: d.Members,
		divisions: d.Divisions, players: d.Players, clubs: d.Clubs,
		clock: d.Clock, tx: d.Tx,
	}, nil
}

func (uc *AwardSeasonBadges) Execute(
	ctx context.Context,
	in dto.AwardSeasonBadgesInput,
) (*dto.AwardSeasonBadgesOutput, error) {
	divs, err := uc.divisions.ListBySeason(ctx, in.SeasonID)
	if err != nil {
		return nil, fmt.Errorf("award season badges: list divisions: %w", err)
	}
	if len(divs) == 0 {
		return nil, fmt.Errorf("award season badges: no divisions: %w", shared.ErrInvalidState)
	}

	now := uc.clock.Now()
	out := &dto.AwardSeasonBadgesOutput{}

	err = uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		var allMembers []*division.DivisionMember

		// 1. Бейджи за место в дивизионе + сбор всех участников
		//    (чтобы не дёргать репозиторий дважды).
		for _, d := range divs {
			rows, err := uc.members.ListByDivision(ctx, d.ID)
			if err != nil {
				return fmt.Errorf("list members of %q: %w", d.Code, err)
			}
			allMembers = append(allMembers, rows...)
			if err := uc.awardDivisionPlaces(ctx, rows, in.SeasonID, now, out); err != nil {
				return err
			}
		}

		// 2. Бейджи за место в клубе.
		if err := uc.awardClubPlaces(ctx, allMembers, in.SeasonID, now, out); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// awardDivisionPlaces выдаёт бейджи топ-3 игрокам дивизиона.
// Исключённые (Status=excluded) в топ не попадают.
func (uc *AwardSeasonBadges) awardDivisionPlaces(
	ctx context.Context,
	rows []*division.DivisionMember,
	seasonID shared.SeasonID,
	now time.Time,
	out *dto.AwardSeasonBadgesOutput,
) error {
	active := filterActive(rows)
	sort.SliceStable(active, func(i, j int) bool {
		if active[i].FinalPlace == nil || active[j].FinalPlace == nil {
			return false
		}
		return *active[i].FinalPlace < *active[j].FinalPlace
	})

	pairs := []struct {
		place int
		code  string
	}{
		{1, CodeSeasonDivGold},
		{2, CodeSeasonDivSilver},
		{3, CodeSeasonDivBronze},
	}
	for _, p := range pairs {
		if len(active) < p.place {
			continue
		}
		m := active[p.place-1]
		awarded, err := uc.awardSeasonal(ctx, m.PlayerID, p.code, seasonID, now)
		if err != nil {
			return err
		}
		if awarded != nil {
			out.Awarded = append(out.Awarded, awarded)
		}
	}
	return nil
}

// awardClubPlaces для каждого клуба выдаёт бейдж за 1/2/3 место внутри клуба.
func (uc *AwardSeasonBadges) awardClubPlaces(
	ctx context.Context,
	members []*division.DivisionMember,
	seasonID shared.SeasonID,
	now time.Time,
	out *dto.AwardSeasonBadgesOutput,
) error {
	// Группируем активных участников по клубу.
	byClub := map[shared.ClubID][]*division.DivisionMember{}
	for _, m := range members {
		p, err := uc.players.GetByID(ctx, m.PlayerID)
		if err != nil {
			return fmt.Errorf("get player %d: %w", m.PlayerID, err)
		}
		if p.ClubID == nil {
			continue
		}
		byClub[*p.ClubID] = append(byClub[*p.ClubID], m)
	}

	for _, list := range byClub {
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].LP != list[j].LP {
				return list[i].LP > list[j].LP
			}
			return list[i].MatchesPlayed > list[j].MatchesPlayed
		})
		pairs := []struct {
			place int
			code  string
		}{
			{1, CodeSeasonClubGold},
			{2, CodeSeasonClubSilver},
			{3, CodeSeasonClubBronze},
		}
		for _, p := range pairs {
			if len(list) < p.place {
				continue
			}
			m := list[p.place-1]
			awarded, err := uc.awardSeasonal(ctx, m.PlayerID, p.code, seasonID, now)
			if err != nil {
				return err
			}
			if awarded != nil {
				out.Awarded = append(out.Awarded, awarded)
			}
		}
	}
	return nil
}

// awardSeasonal выдаёт сезонный бейдж, если его нет за этот сезон.
func (uc *AwardSeasonBadges) awardSeasonal(
	ctx context.Context,
	pid shared.PlayerID,
	code string,
	seasonID shared.SeasonID,
	now time.Time,
) (*dombadge.PlayerBadge, error) {
	b, err := uc.badges.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get badge %q: %w", code, err)
	}
	// Идемпотентность: если бейдж уже выдан этому игроку за этот сезон.
	existing, err := uc.playerBadges.ListByPlayerAndSeason(ctx, pid, seasonID)
	if err != nil {
		return nil, fmt.Errorf("list seasonal badges: %w", err)
	}
	for _, e := range existing {
		if e.BadgeID == b.ID {
			return nil, nil
		}
	}

	pb := dombadge.NewPlayerBadge(pid, b.ID, &seasonID, now)
	if err := uc.playerBadges.Create(ctx, pb); err != nil {
		return nil, fmt.Errorf("create seasonal badge: %w", err)
	}
	return pb, nil
}

func filterActive(rows []*division.DivisionMember) []*division.DivisionMember {
	out := make([]*division.DivisionMember, 0, len(rows))
	for _, m := range rows {
		if m.Status == division.MemberActive {
			out = append(out, m)
		}
	}
	return out
}
