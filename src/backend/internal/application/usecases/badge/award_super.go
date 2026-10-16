package badge

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/J0hnLenin/67League/internal/application/dto"
	"github.com/J0hnLenin/67League/internal/application/ports"
	dombadge "github.com/J0hnLenin/67League/internal/domain/badge"
	"github.com/J0hnLenin/67League/internal/domain/division"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// AwardSuperTitle присваивает супертитул «Чемпион 67» лучшему игроку
// лучшего клуба сезона. Идемпотентен: если супертитул за сезон уже выдан,
// повторно не выдаётся.
type AwardSuperTitle struct {
	badges       ports.BadgeRepository
	playerBadges ports.PlayerBadgeRepository
	superTitles  ports.SuperTitleRepository
	members      ports.DivisionMemberRepository
	divisions    ports.DivisionRepository
	players      ports.PlayerRepository
	clubs        ports.ClubRepository
	clock        ports.Clock
	tx           ports.TransactionManager
}

type AwardSuperTitleDeps struct {
	Badges       ports.BadgeRepository
	PlayerBadges ports.PlayerBadgeRepository
	SuperTitles  ports.SuperTitleRepository
	Members      ports.DivisionMemberRepository
	Divisions    ports.DivisionRepository
	Players      ports.PlayerRepository
	Clubs        ports.ClubRepository
	Clock        ports.Clock
	Tx           ports.TransactionManager
}

func NewAwardSuperTitle(d AwardSuperTitleDeps) (*AwardSuperTitle, error) {
	if d.Badges == nil || d.PlayerBadges == nil || d.SuperTitles == nil ||
		d.Members == nil || d.Divisions == nil || d.Players == nil ||
		d.Clubs == nil || d.Clock == nil || d.Tx == nil {
		return nil, errors.New("award super title: missing dependencies")
	}
	return &AwardSuperTitle{
		badges: d.Badges, playerBadges: d.PlayerBadges, superTitles: d.SuperTitles,
		members: d.Members, divisions: d.Divisions, players: d.Players,
		clubs: d.Clubs, clock: d.Clock, tx: d.Tx,
	}, nil
}

func (uc *AwardSuperTitle) Execute(
	ctx context.Context,
	in dto.AwardSuperTitleInput,
) (*dto.AwardSuperTitleOutput, error) {
	// Идемпотентность на уровне сезона.
	existing, err := uc.superTitles.GetBySeason(ctx, in.SeasonID)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return nil, fmt.Errorf("award super title: get existing: %w", err)
	}
	if existing != nil {
		return &dto.AwardSuperTitleOutput{
			SuperTitle: existing,
			Awarded:    false,
		}, nil
	}

	divs, err := uc.divisions.ListBySeason(ctx, in.SeasonID)
	if err != nil {
		return nil, fmt.Errorf("award super title: list divisions: %w", err)
	}
	if len(divs) == 0 {
		return nil, fmt.Errorf("award super title: no divisions: %w", shared.ErrInvalidState)
	}

	// Собираем всех активных участников с их клубами.
	type entry struct {
		member *division.DivisionMember
		clubID shared.ClubID
	}
	var entries []entry
	for _, d := range divs {
		ms, err := uc.members.ListByDivision(ctx, d.ID)
		if err != nil {
			return nil, fmt.Errorf("award super title: list members: %w", err)
		}
		for _, m := range ms {
			if m.Status != division.MemberActive {
				continue
			}
			p, err := uc.players.GetByID(ctx, m.PlayerID)
			if err != nil {
				return nil, fmt.Errorf("award super title: get player: %w", err)
			}
			if p.ClubID == nil {
				continue
			}
			entries = append(entries, entry{member: m, clubID: *p.ClubID})
		}
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("award super title: no eligible players: %w", shared.ErrInvalidState)
	}

	// Клубный зачёт: сумма LP активных.
	clubLP := map[shared.ClubID]shared.LP{}
	for _, e := range entries {
		clubLP[e.clubID] += e.member.LP
	}

	// Лучший клуб.
	var bestClub shared.ClubID
	var bestLP shared.LP
	first := true
	for clubID, lp := range clubLP {
		if first || lp > bestLP {
			bestClub = clubID
			bestLP = lp
			first = false
		}
	}

	// Игрок с максимальным LP в лучшем клубе.
	// Тайбрейк: больше партий.
	var winner *division.DivisionMember
	for _, e := range entries {
		if e.clubID != bestClub {
			continue
		}
		if winner == nil ||
			e.member.LP > winner.LP ||
			(e.member.LP == winner.LP && e.member.MatchesPlayed > winner.MatchesPlayed) {
			winner = e.member
		}
	}
	if winner == nil {
		return nil, fmt.Errorf("award super title: no winner: %w", shared.ErrInvalidState)
	}

	now := uc.clock.Now()
	out := &dto.AwardSuperTitleOutput{
		Player:  &winner.PlayerID,
		Club:    &bestClub,
		Awarded: true,
	}

	err = uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		st := dombadge.NewSuperTitle(in.SeasonID, winner.PlayerID, bestClub, now)
		if err := uc.superTitles.Create(ctx, st); err != nil {
			return fmt.Errorf("award super title: create: %w", err)
		}
		out.SuperTitle = st

		// Супербейдж.
		b, err := uc.badges.GetByCode(ctx, CodeSuperChampion67)
		if err != nil {
			if errors.Is(err, shared.ErrNotFound) {
				return nil // не критично
			}
			return fmt.Errorf("award super title: get badge: %w", err)
		}
		pb := dombadge.NewPlayerBadge(winner.PlayerID, b.ID, &in.SeasonID, now)
		if err := uc.playerBadges.Create(ctx, pb); err != nil {
			return fmt.Errorf("award super title: create badge: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// sortEntriesByLP — упорядочивает entries по LP (для потенциального использования).
func sortEntriesByLP(entries []*division.DivisionMember) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].LP != entries[j].LP {
			return entries[i].LP > entries[j].LP
		}
		return entries[i].MatchesPlayed > entries[j].MatchesPlayed
	})
}
